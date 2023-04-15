package db

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/mearaj/protonet/alog"
	"github.com/mearaj/protonet/internal/model"
	"github.com/mearaj/protonet/internal/pubsub"
	"strings"
	"time"
)

type Account = model.Account

// AddOrUpdateAccount saves as primary account
func (d *ProtoDB) AddUpdateAccount(acc *Account) (err error) {
	if acc == nil || len(acc.PublicKey) == 0 {
		return ErrInvalidAccount
	}
	prevAccount, _ := d.Account()
	var currentAccountChanged bool
	err = d.getErrorState()
	if err != nil {
		return err
	}
	dB := d.getState().dB
	defer func() {
		if r := recover(); r != nil {
			alog.Logger().Errorln(r)
		}
		if currentAccountChanged {
			event := pubsub.Event{Data: pubsub.CurrentAccountChangedEventData{
				PrevAccountPublicKey:    prevAccount.PublicKey,
				CurrentAccountPublicKey: acc.PublicKey,
			}, Topic: pubsub.CurrentAccountChangedEventTopic}
			d.EventBroker.Fire(event)
		}
	}()
	acc.UpdatedAt = time.Now()
	if acc.CreatedAt.IsZero() {
		acc.CreatedAt = time.Now()
	}
	fullKey, err := acc.GetDBFullKey()
	if err != nil {
		return
	}
	var accountKeys []string
	txn := dB.NewTransaction(true)
	defer txn.Discard()
	opts := badger.DefaultIteratorOptions
	it := txn.NewIterator(opts)
	for it.Seek([]byte(KeyPrefixAccounts)); it.ValidForPrefix([]byte(KeyPrefixAccounts)); it.Next() {
		item := it.Item()
		k := string(item.KeyCopy(nil))
		accountKeys = append(accountKeys, k)
	}
	it.Close()
	for _, key := range accountKeys {
		if strings.Contains(key, acc.PublicKey) {
			err = txn.Delete([]byte(key))
		}
	}
	val := EncodeToBytes(acc)
	err = txn.Set([]byte(fullKey), val)
	if err != nil {
		return err
	}
	err = txn.Commit()
	if err == nil && prevAccount.PublicKey != acc.PublicKey {
		currentAccountChanged = true
	}
	return err
}

func (d *ProtoDB) Account() (acc Account, err error) {
	err = d.getErrorState()
	if err != nil {
		return acc, err
	}
	accs, err := d.Accounts()
	if err != nil {
		return acc, err
	}
	if len(accs) == 0 {
		return acc, ErrAccountDoesNotExist
	}
	return accs[0], err
}

func (d *ProtoDB) Accounts() (accounts []Account, err error) {
	err = d.getErrorState()
	if err != nil {
		return accounts, err
	}
	prefix := KeyPrefixAccounts
	allKeys, err := d.prefixScanSorted(prefix, KeySeparator, 0, 1, true)
	if err != nil {
		return accounts, err
	}
	for _, k := range allKeys {
		var account Account
		err = d.ViewRecord([]byte(k), &account)
		if err != nil {
			return accounts, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// DeleteAccounts cascade deletes all Contacts and Messages belong to those Accounts
func (d *ProtoDB) DeleteAccounts(accounts []Account) (err error) {
	if len(accounts) == 0 {
		return nil
	}
	defer func() {
		d.EventBroker.Fire(pubsub.Event{
			Data:  pubsub.AccountsChangedEventData{},
			Topic: pubsub.AccountsChangedEventTopic,
		})
	}()

	err = d.getErrorState()
	if err != nil {
		return err
	}
	dB := d.getState().dB
	for _, eachAccount := range accounts {
		err = dB.Update(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				k := item.Key()
				if strings.Contains(string(k), eachAccount.PublicKey) {
					err = txn.Delete(k)
				}
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return err
}

func (d *ProtoDB) AccountExists(publicKey string) (exists bool, err error) {
	err = d.getErrorState()
	if err != nil {
		return exists, err
	}
	dB := d.getState().dB
	if len(publicKey) == 0 {
		return exists, ErrInvalidKey
	}
	acc := Account{PublicKey: publicKey}
	accountKey := acc.GetAccountDBPrefixKey()
	err = dB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		it.Seek([]byte(accountKey))
		exists = it.ValidForPrefix([]byte(accountKey))
		it.Close()
		return nil
	})
	return exists, err
}
