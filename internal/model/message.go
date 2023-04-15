package model

import (
	"fmt"
	"time"
)

const (
	MessageStateStateless = iota
	MessageStateReceived
	MessageStateRead
)

type Message struct {
	ID        string
	Sender    string
	Recipient string
	CreatedAt time.Time
	Text      string
	Sign      []byte
	Audio     []byte
	// Holds the read state of the Recipient, this is the only field that can be different
	// between sender-and-receiver, and it's the only field that can be changed and is always
	// decided by the recipient
	State int64
}

func (msg *Message) GetDBFullKey(accPublicKey string) (key string, err error) {
	isError := len(accPublicKey) == 0 || len(msg.Sender) == 0 || len(msg.Recipient) == 0 ||
		len(msg.ID) == 0 || msg.CreatedAt.IsZero() || (msg.Recipient != accPublicKey && msg.Sender != accPublicKey)
	if isError {
		return key, ErrInvalidMessage
	}
	contactPublicKey := msg.Sender
	isMeMsgCreator := msg.Sender == accPublicKey
	if isMeMsgCreator {
		contactPublicKey = msg.Recipient
	}
	createdAt := msg.CreatedAt.UnixNano()
	key = fmt.Sprintf("%s%s%s%s%s%s%d%s%s%s%s%s%s",
		KeyPrefixMessages,
		KeySeparator, accPublicKey,
		KeySeparator, contactPublicKey,
		KeySeparator, createdAt,
		KeySeparator, msg.Sender,
		KeySeparator, msg.Recipient,
		KeySeparator, msg.ID,
	)
	return key, nil
}

func (msg *Message) GetDBPrefixKey(accPublicKey string, contactPublicKey string) (key string, separatorCount int) {
	key = KeyPrefixMessages
	if len(accPublicKey) == 0 {
		return key, separatorCount
	}
	separatorCount++
	key = fmt.Sprintf("%s%s%s", key, KeySeparator, accPublicKey)
	if len(contactPublicKey) == 0 {
		contactPublicKey = msg.Sender
		isMeMsgCreator := msg.Sender == accPublicKey
		if isMeMsgCreator {
			contactPublicKey = msg.Recipient
		}
	}
	if len(contactPublicKey) == 0 {
		return key, separatorCount
	}
	separatorCount++
	key = fmt.Sprintf("%s%s%s", key, KeySeparator, contactPublicKey)
	if msg.CreatedAt.IsZero() {
		return key, separatorCount
	}
	createdAt := msg.CreatedAt.UnixMilli()
	separatorCount++
	key = fmt.Sprintf("%s%s%d", key, KeySeparator, createdAt)
	if len(msg.Sender) == 0 {
		return key, separatorCount
	}
	separatorCount++
	key = fmt.Sprintf("%s%s%s", key, KeySeparator, msg.Sender)
	if len(msg.Recipient) == 0 {
		return key, separatorCount
	}
	separatorCount++
	key = fmt.Sprintf("%s%s%s", key, KeySeparator, msg.Recipient)
	if len(msg.ID) == 0 {
		return key, separatorCount
	}
	key = fmt.Sprintf("%s%s%s", key, KeySeparator, msg.ID)
	return key, separatorCount
}
