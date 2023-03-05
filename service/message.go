package service

import (
	"github.com/google/uuid"
)

// Message properties AccountPublicKey and ContactPublicKey keyConfigMessages
type Message struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	AccountPublicKey string    `gorm:"foreignKey;primaryKey"`
	ContactPublicKey string    `gorm:"foreignKey;primaryKey"`
	From             string
	To               string
	Created          string
	Text             string
	Read             bool `gorm:"type:boolean;default:false;not null;"`
	Sign             []byte
	State            MessageState `gorm:"type:int;default:0;"`
	Audio            []byte
}

type MessageState int

const (
	// MessageStateless indicates a new message is created or received
	MessageStateless MessageState = iota
	// MessageReceivedSent Depends upon the context
	//  If the message owner is msg creator(outgoing message), then it indicates whether creator has received
	//  an acknowledgement from msg receiver.
	//  If the message owner is msg receiver(incoming message), then it indicates whether creator has sent
	//  an acknowledgement to msg receiver.
	MessageReceivedSent
	// MessageRead Depends upon the context.
	// The logic is similar to MessageReceivedSent except that it deals with msg's read state
	MessageRead
)
