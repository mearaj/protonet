package db

import (
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/mearaj/protonet/alog"
	"github.com/mearaj/protonet/internal/model"
	"github.com/mearaj/protonet/internal/pubsub"
	"strings"
	"time"
)

type Contact = model.Contact

func (d *ProtoDB) Contacts(accountPublicKey string, offset, limit int) (contacts []Contact, err error) {
	contact := Contact{AccountPublicKey: accountPublicKey}
	keyPrefix := contact.GetDBPrefixKey()
	allKeys, err := d.prefixScanSorted(keyPrefix, KeySeparator, 1, 3, true)
	if err != nil {
		return contacts, err
	}
	if len(allKeys) <= offset {
		return contacts, errors.New("invalid offset")
	}
	availableLimit := len(allKeys[offset:])
	if limit > availableLimit {
		limit = availableLimit
	}
	allKeys = allKeys[offset : offset+limit]
	for _, k := range allKeys {
		var contact Contact
		err = d.ViewRecord([]byte(k), &contact)
		if err == nil {
			contacts = append(contacts, contact)
		}
	}
	return contacts, err
}

func (d *ProtoDB) AddUpdateContact(c *Contact) (err error) {
	err = d.getErrorState()
	if err != nil {
		return err
	}
	dB := d.getState().dB
	fullKey, err := c.GetDBFullKey()
	if err != nil {
		return err
	}
	duplicateKeys, err := d.prefixScan(fullKey, KeySeparator, 2)
	if err != nil {
		return err
	}
	txn := dB.NewTransaction(true)
	defer txn.Discard()
	if len(duplicateKeys) > 0 {
		for _, key := range duplicateKeys {
			err = txn.Delete([]byte(key))
			if err != nil {
				return err
			}
		}
	}
	c.UpdatedAt = time.Now()
	fullKey, err = c.GetDBFullKey()
	if err != nil {
		return err
	}
	val := EncodeToBytes(c)
	err = txn.Set([]byte(fullKey), val)
	if err != nil {
		return err
	}
	err = txn.Commit()
	if err != nil {
		return err
	}
	eventData := pubsub.SaveContactEventData{Contact: *c}
	event := pubsub.Event{
		Data:  eventData,
		Topic: pubsub.SaveContactTopic,
	}
	d.EventBroker.Fire(event)
	return err
}

func (d *ProtoDB) ContactsCount(accountPublicKey string) (count int64, err error) {
	err = d.getErrorState()
	if err != nil {
		return count, err
	}
	dB := d.getState().dB
	defer func() {
		if r := recover(); r != nil {
			alog.Logger().Errorln(r)
		}
	}()
	c := Contact{AccountPublicKey: accountPublicKey}
	prefixKey := c.GetDBPrefixKey()
	// This should never happen
	if err != nil {
		return
	}
	err = dB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := string(item.Key())
			if strings.HasPrefix(k, prefixKey) {
				count++
			}
		}
		return nil
	})
	return
}

func (d *ProtoDB) DeleteContacts(accountPublicKey string, contacts []Contact) (count int64, err error) {
	err = d.getErrorState()
	if err != nil {
		return count, err
	}
	dB := d.getState().dB
	if len(contacts) == 0 {
		return count, errors.New("contacts is empty")
	}
	if len(accountPublicKey) == 0 {
		return count, errors.New("account public key is empty")
	}
	defer func() {
		if err != nil {
			alog.Logger().Errorln(err)
		}
		if count > 0 {
			d.EventBroker.Fire(pubsub.Event{
				Data: pubsub.ContactsChangeEventData{
					AccountPublicKey: accountPublicKey,
				},
				Topic: pubsub.ContactsChangedEventTopic,
			})
		}
	}()
	// This will delete all the contacts and all messages that belongs to deleted contact
	for _, eachContact := range contacts {
		keyComponent := fmt.Sprintf("%s%s%s", accountPublicKey, KeySeparator, eachContact.PublicKey)
		err = dB.Update(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				k := item.Key()
				if strings.Contains(string(k), keyComponent) {
					err = txn.Delete(k)
					if err == nil {
						count++
					}
				}
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return count, err
		}
	}
	return count, nil
}
