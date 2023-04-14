package db

import (
	"encoding/gob"
	"errors"
	"fmt"
	"gioui.org/app"
	"github.com/dgraph-io/badger/v4"
	"github.com/mearaj/protonet/internal/pubsub"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type Service interface {
	Account() (Account, error)
	Accounts() ([]Account, error)
	Contacts(accountPublicKey string, offset, limit int) ([]Contact, error)
	Messages(accountPublicKey, contactPubKey string, offset, limit int) ([]Message, error)
	AddUpdateAccount(account *Account) error
	AddUpdateContact(contact *Contact) (err error)
	AccountExists(publicKey string) (bool, error)
	LastMessage(accountPublicKey, contactPublicKey string) (Message, error)
	DeleteAccounts([]Account) error
	DeleteContacts(accountPublicKey string, contacts []Contact) (int64, error)
	SaveOrUpdateMessage(accountPublicKey string, msg *Message) (err error)
	UnreadMessagesCount(accountPublicKey, contactPublicKey string) (count int64, err error)
	MessagesCount(accountPublicKey, contactPublicKey string) (count int64, err error)
	ContactsCount(addrPublicKey string) (int64, error)
	MarkPrevMessagesAsRead(accountPublicKey, contactAddr string) (count int64, err error)
	ViewRecord(key []byte, ptrStruct interface{}) (err error)
	IsOpen() bool
	VerifyPassword(passwd string) error
}

type State int

var (
	ErrDBNotOpened           = errors.New("database not yet opened")
	ErrDBAlreadyOpened       = errors.New("database is already open")
	ErrEncryptionKeyMismatch = badger.ErrEncryptionKeyMismatch
	ErrPasswordMismatch      = errors.New("password mismatch")
	ErrPasswordInvalid       = errors.New("password invalid")
	ErrPasswordNotSet        = errors.New("password not set")
)

type protoDBState struct {
	dB  *badger.DB
	err error
}

type ProtoDB struct {
	password    string
	EventBroker *pubsub.EventBroker
	state       protoDBState
	stateMutex  sync.RWMutex
}

var _ Service = &ProtoDB{}

func init() {
	gob.Register(Account{})
	gob.Register(Contact{})
	gob.Register(Message{})
}

//var GlobalProtoDB = &ProtoDB{}

func New() *ProtoDB {
	return &ProtoDB{
		EventBroker: pubsub.NewEventBroker(),
	}
}

func (d *ProtoDB) Open(options badger.Options) error {
	_ = d.Close()
	state := d.getState()
	if state.err != nil {
		if errors.Is(state.err, ErrDBAlreadyOpened) {
			return state.err
		}
	}
	dB, err := badger.Open(options)
	if err != nil {
		d.setState(protoDBState{dB: nil, err: err})
		return err
	}
	d.setState(protoDBState{dB: dB, err: nil})
	d.EventBroker.Fire(pubsub.Event{
		Data:   pubsub.DatabaseOpenedEventData{},
		Topic:  pubsub.DatabaseOpened,
		Cached: false,
	})
	return nil
}

func (d *ProtoDB) Close() error {
	state := d.getState()
	if state.dB != nil {
		_ = d.Close()
	}
	state.err = nil
	state.dB = nil
	d.setState(state)
	return nil
}

func (d *ProtoDB) getState() protoDBState {
	d.stateMutex.RLock()
	defer d.stateMutex.RUnlock()
	return d.state
}
func (d *ProtoDB) setState(state protoDBState) {
	d.stateMutex.Lock()
	d.state = state
	d.stateMutex.Unlock()
}

func (d *ProtoDB) getErrorState() (err error) {
	state := d.getState()
	if state.dB == nil {
		return ErrDBNotOpened
	}
	if state.err != nil {
		return state.err
	}
	return nil
}

// prefixScan prefixPos is the position of separator to derive prefix key for scanning
func (d *ProtoDB) prefixScan(prefixOrFullKey string, keySeparator string, prefixPos int) (fullKeys []string, err error) {
	err = d.getErrorState()
	if err != nil {
		return fullKeys, err
	}
	db := d.getState().dB
	prefixKeyArr := strings.Split(prefixOrFullKey, keySeparator)
	if len(prefixKeyArr) < prefixPos+1 {
		return fullKeys, ErrInvalidKey
	}
	prefixKey := strings.Join(prefixKeyArr[0:prefixPos+1], keySeparator)
	err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek([]byte(prefixKey)); it.ValidForPrefix([]byte(prefixKey)); it.Next() {
			item := it.Item()
			k := string(item.KeyCopy(nil))
			fullKeys = append(fullKeys, k)
		}
		return nil
	})
	return fullKeys, err
}

func (d *ProtoDB) prefixScanSorted(prefixOrFullKey, sep string, prefixSepPos, sortSepPos int, shouldReverse bool) (derivedKeys []string, err error) {
	derivedKeys, err = d.prefixScan(prefixOrFullKey, sep, prefixSepPos)
	if err != nil {
		return
	}
	if len(derivedKeys) > 0 {
		arr := strings.Split(derivedKeys[0], KeySeparator)
		if sortSepPos >= len(arr) {
			return derivedKeys, errors.New("invalid separator pos")
		}
	}
	sort.Slice(derivedKeys, func(i, j int) bool {
		keyOneArr := strings.Split(derivedKeys[i], KeySeparator)
		keyTwoArr := strings.Split(derivedKeys[j], KeySeparator)
		keyOne := keyOneArr[sortSepPos]
		keyTwo := keyTwoArr[sortSepPos]
		if shouldReverse {
			return keyOne > keyTwo
		}
		return keyOne < keyTwo
	})
	return derivedKeys, err
}

// ViewRecord
//
//	ptrStruct should be a pointer to a struct registered with gob
func (d *ProtoDB) ViewRecord(key []byte, ptrStruct interface{}) (err error) {
	err = d.getErrorState()
	if err != nil {
		return err
	}
	dB := d.getState().dB
	err = dB.View(func(txn *badger.Txn) (err error) {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) (err error) {
			err = DecodeToStruct(ptrStruct, val)
			return err
		})
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (d *ProtoDB) IsOpen() bool {
	state := d.getState()
	if state.dB == nil {
		return false
	}
	return !state.dB.IsClosed()
}

func (d *ProtoDB) OpenFromPassword(passwd string) error {
	origPasswd := passwd
	if d.IsOpen() {
		return ErrDBAlreadyOpened
	}
	if len(passwd) == 0 {
		return ErrPasswdCannotBeEmpty
	}
	if len([]byte(passwd)) > MaxNumOfPasswdChars {
		return fmt.Errorf("password should be less than %d characters", MaxNumOfPasswdChars)
	}
	padDiff := MaxNumOfPasswdChars - len([]byte(passwd))
	var leftPad, rightPad string
	for i := 0; i < padDiff; i++ {
		if i%2 == 0 {
			leftPad += passwdPadCharacter
		} else {
			rightPad += passwdPadCharacter
		}
	}
	passwd = leftPad + passwd + rightPad
	dirPath, err := app.DataDir()
	if err != nil {
		return err
	}
	dbPath := filepath.Join(dirPath, PathAppDirName, PathDBDirName)
	options := badger.DefaultOptions(dbPath)
	options.EncryptionKey = []byte(passwd)
	options.IndexCacheSize = 100
	err = d.Open(options)
	if err != nil {
		return err
	}
	d.password = origPasswd
	d.EventBroker.Fire(pubsub.Event{
		Data:   pubsub.DatabaseOpenedEventData{},
		Topic:  pubsub.DatabaseOpened,
		Cached: false,
	})
	return nil
}
func (d *ProtoDB) DatabaseExists() bool {
	dirPath, err := app.DataDir()
	if err != nil {
		return false
	}
	dbPath := filepath.Join(dirPath, PathAppDirName, PathDBDirName)
	_, err = os.Stat(dbPath)
	return err == nil
}

// VerifyPassword
// returns nil if password is correct else may
// return ErrPasswordMismatch or ErrPasswordInvalid or ErrPasswordNotSet
func (d *ProtoDB) VerifyPassword(passwd string) error {
	if strings.TrimSpace(passwd) == "" {
		return ErrPasswordInvalid
	}
	if d.password == "" {
		return ErrPasswordNotSet
	}
	if d.password != passwd {
		return ErrPasswordMismatch
	}
	return nil
}
