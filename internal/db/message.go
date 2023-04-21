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

const (
	MessageStateStateless = model.MessageStateStateless
	MessageStateReceived  = model.MessageStateReceived
	MessageStateRead      = model.MessageStateRead
)

type Message = model.Message

func (d *ProtoDB) Messages(accountPublicKey, contactPublicKey string, offset int, limit int) (messages []Message, err error) {
	err = d.getErrorState()
	if err != nil {
		return messages, err
	}
	dB := d.getState().dB
	message := Message{}
	keyPrefix, count := message.GetDBPrefixKey(accountPublicKey, contactPublicKey)
	allKeys, err := d.prefixScanSorted(keyPrefix, KeySeparator, count, 3, true)
	if err != nil {
		return nil, err
	}
	if len(allKeys) <= offset {
		return nil, errors.New("invalid offset")
	}
	availableLimit := len(allKeys[offset:])
	if limit > availableLimit {
		limit = availableLimit
	}
	allKeys = allKeys[offset : offset+limit]
	err = dB.View(func(txn *badger.Txn) (err error) {
		for _, k := range allKeys {
			var msg Message
			var item *badger.Item
			item, err = txn.Get([]byte(k))
			if err != nil {
				return err
			}
			err = item.Value(func(val []byte) (err error) {
				err = DecodeToStruct(&msg, val)
				return err
			})
			if err != nil {
				return err
			}
			messages = append(messages, msg)
		}
		return nil
	})
	return messages, err
}

// SaveOrUpdateMessage saves message to the database
func (d *ProtoDB) SaveOrUpdateMessage(accountPublicKey string, msg *Message) (err error) {
	if len(accountPublicKey) == 0 {
		return ErrInvalidAccount
	}
	if msg == nil {
		return ErrInvalidMessage
	}
	err = d.getErrorState()
	if err != nil {
		return err
	}
	var isMessageNew, msgStateChanged bool
	var contact *Contact
	dB := d.getState().dB
	defer func() {
		if err != nil {
			alog.Logger().Errorln(err)
		}
		// If new message created
		if isMessageNew {
			if msg.Sender == accountPublicKey {
				d.EventBroker.Fire(pubsub.Event{
					Data:  pubsub.SendNewMessageEventData{Message: *msg},
					Topic: pubsub.SendNewMessageEventTopic,
					Err:   err,
				})
			} else {
				d.EventBroker.Fire(pubsub.Event{
					Data:  pubsub.NewMessageReceivedEventData{Message: *msg},
					Topic: pubsub.NewMessageReceivedTopic,
					Err:   err,
				})
			}
		}
		if msgStateChanged { // If the message state changed
			d.EventBroker.Fire(pubsub.Event{
				Data:  pubsub.MessageStateChangedEventData{Message: *msg},
				Topic: pubsub.MessageStateChangedEventTopic,
				Err:   err,
			})
		}
		if contact != nil {
			// If new contact is added
			d.EventBroker.Fire(pubsub.Event{
				Data:  pubsub.SaveContactEventData{Contact: *contact},
				Topic: pubsub.SaveContactTopic,
				Err:   err,
			})
		}
	}()
	fullKey, err := msg.GetDBFullKey(accountPublicKey)
	if err != nil {
		return err
	}
	isMsgCreatedByMe := accountPublicKey == msg.Sender
	allKeys, err := d.prefixScanSorted(fullKey, KeySeparator, 2, 3, true)
	var duplicateKeys []string
	var isDuplicate bool
	for _, key := range allKeys {
		if strings.HasSuffix(key, msg.ID) {
			duplicateKeys = append(duplicateKeys, key)
		}
	}
	isMessageNew = true
	if len(duplicateKeys) > 0 {
		isDuplicate = true
		var dbMsg Message
		txn := dB.NewTransaction(true)
		defer txn.Discard()
		for _, key := range duplicateKeys[1:] {
			err = txn.Delete([]byte(key))
			if err != nil {
				return
			}
		}
		var it *badger.Item
		it, err = txn.Get([]byte(duplicateKeys[0]))
		if err != nil {
			return
		}
		var val []byte
		val, err = it.ValueCopy(nil)
		if err != nil {
			return
		}
		err = DecodeToStruct(&dbMsg, val)
		if err != nil {
			return
		}
		// Todo: Rethink implementation
		if dbMsg.State < msg.State {
			err = txn.Delete([]byte(duplicateKeys[0]))
			if err != nil {
				return
			}
			dbMsg.State = msg.State
			bs := EncodeToBytes(&dbMsg)
			err = txn.Set([]byte(duplicateKeys[0]), bs)
			if err != nil {
				return
			}
			msgStateChanged = true
		}
		err = txn.Commit()
		if err != nil {
			return
		}
		*msg = dbMsg
		isMessageNew = false
	}
	if !isDuplicate {
		if !isMsgCreatedByMe {
			msg.State = model.MessageStateReceived
			msgStateChanged = true
		}
		bs := EncodeToBytes(&msg)
		err = dB.Update(func(txn *badger.Txn) error {
			err = txn.Set([]byte(fullKey), bs)
			if err != nil {
				return err
			}
			return err
		})
	}

	// After saving/updating new message, we update/create contact to update contact's UpdatedAt
	if !isDuplicate {
		contact = &Contact{PublicKey: msg.Sender, AccountPublicKey: msg.Recipient}
		if isMsgCreatedByMe {
			contact.PublicKey = msg.Recipient
			contact.AccountPublicKey = msg.Sender
		}
		contact.CreatedAt = time.Now()
		contact.UpdatedAt = time.Now()
		err = d.AddUpdateContact(contact)
		if err != nil {
			alog.Logger().Errorln(err)
			return
		}
	}
	return
}

func (d *ProtoDB) LastMessage(accountPublicKey, contactPublicKey string) (msg Message, err error) {
	err = d.getErrorState()
	if err != nil {
		return msg, err
	}
	if accountPublicKey == "" {
		return msg, ErrInvalidAccount
	}
	dB := d.getState().dB
	defer func() {
		if r := recover(); r != nil {
			alog.Logger().Errorln(r)
		}
	}()
	prefixKey, _ := msg.GetDBPrefixKey(accountPublicKey, contactPublicKey)
	err = dB.View(func(txn *badger.Txn) error {
		options := badger.DefaultIteratorOptions
		it := txn.NewIterator(options)
		defer it.Close()
		for it.Seek([]byte(prefixKey)); it.ValidForPrefix([]byte(prefixKey)); {
			item := it.Item()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			it.Next()
			if !it.ValidForPrefix([]byte(prefixKey)) {
				err = DecodeToStruct(&msg, val)
				if err != nil {
					return err
				}
				return err
			}
		}
		return err
	})
	return msg, err
}

// UnreadMessagesCount unread messages counts (incoming messages from contact)
func (d *ProtoDB) UnreadMessagesCount(accountPublicKey, contactPublicKey string) (count int64, err error) {
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
	if accountPublicKey == "" {
		return
	}
	msg := Message{Sender: contactPublicKey, Recipient: accountPublicKey}
	prefixKey, _ := msg.GetDBPrefixKey(accountPublicKey, contactPublicKey)
	allKeys, err := d.prefixScanSorted(prefixKey, KeySeparator, 2, 3, true)
	var inboxKeys []string
	for _, key := range allKeys {
		isInbox := strings.Contains(key,
			fmt.Sprintf("%s%s%s", contactPublicKey, KeySeparator, accountPublicKey),
		)
		if isInbox {
			inboxKeys = append(inboxKeys, key)
		}
	}
	for _, key := range inboxKeys {
		err = dB.View(func(txn *badger.Txn) error {
			var msg2 Message
			item, err := txn.Get([]byte(key))
			if err != nil {
				return err
			}
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			err = DecodeToStruct(&msg2, val)
			if err != nil {
				return err
			}
			if msg2.State < model.MessageStateRead && msg2.Sender != accountPublicKey &&
				msg2.Sender == contactPublicKey {
				count++
			}
			return err
		})
	}
	return count, err
}

func (d *ProtoDB) MessagesCount(accountPublicKey, contactPublicKey string) (count int64, err error) {
	err = d.getErrorState()
	if err != nil {
		return count, err
	}
	dB := d.getState().dB
	if accountPublicKey == "" || contactPublicKey == "" {
		return
	}
	defer func() {
		if err != nil {
			alog.Logger().Errorln(err)
		}
	}()
	msg := Message{}
	if accountPublicKey == "" || contactPublicKey == "" {
		return
	}
	prefixKey, _ := msg.GetDBPrefixKey(accountPublicKey, contactPublicKey)
	err = dB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek([]byte(prefixKey)); it.ValidForPrefix([]byte(prefixKey)); it.Next() {
			count++
		}
		return nil
	})
	return
}

func (d *ProtoDB) MarkPrevMessagesAsRead(accountPublicKey, contactPublicKey string) (count int64, err error) {
	err = d.getErrorState()
	if err != nil {
		return count, err
	}
	dB := d.getState().dB
	defer func() {
		if err != nil {
			alog.Logger().Errorln(err)
		}
		if count > 0 {
			d.EventBroker.Fire(pubsub.Event{
				Data: pubsub.MessagesStateChangedEventData{
					AccountPublicKey: accountPublicKey,
					ContactPublicKey: contactPublicKey,
				},
				Topic: pubsub.MessagesStateChangedEventTopic,
			})
		}
	}()
	if accountPublicKey == "" {
		return count, errors.New("account public key is empty")
	}
	if contactPublicKey == "" {
		return count, errors.New("contact public key is empty")
	}
	msg := Message{Sender: contactPublicKey, Recipient: accountPublicKey}
	prefixKey, _ := msg.GetDBPrefixKey(accountPublicKey, contactPublicKey)
	allKeys, err := d.prefixScanSorted(prefixKey, KeySeparator, 2, 3, true)
	var inboxKeys []string
	for _, key := range allKeys {
		isInbox := strings.Contains(key,
			fmt.Sprintf("%s%s%s", accountPublicKey, KeySeparator, contactPublicKey),
		)
		if isInbox {
			inboxKeys = append(inboxKeys, key)
		}
	}
	for _, key := range inboxKeys {
		err = dB.Update(func(txn *badger.Txn) (err error) {
			var msg2 Message
			var item *badger.Item
			item, err = txn.Get([]byte(key))
			if err != nil {
				return err
			}
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			err = DecodeToStruct(&msg2, val)
			if err != nil {
				return err
			}
			if msg2.State < model.MessageStateRead && msg2.Sender != accountPublicKey &&
				msg2.Sender == contactPublicKey {
				msg2.State = model.MessageStateRead
				val := EncodeToBytes(&msg2)
				err = txn.Set(item.Key(), val)
				if err != nil {
					return err
				}
				count++
			}
			return err
		})
	}
	return count, err
}
