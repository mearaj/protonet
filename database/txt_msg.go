package database

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"path/filepath"
	"sort"
)

type TxtMsg struct {
	CreatorID        string
	Timestamp        int64
	ID               string
	CreatorPublicKey string
	Sign             []byte
	Message          string
	// should be nil during communication
	// Received is for Msg Creator and Sent is for Msg receiver
	AckReceivedOrSent     bool
	ReadAckReceivedOrSent bool
	TimestampNano         int64
	// The receiver/client of the message hold these values to maintain the state of the message
	MsgRead bool
	Action  TxtMsgAction
}
type TxtMsgs map[string]*TxtMsg
type TxtMsgAction struct {
	Message string
	Type    int32
}

const (
	MessageAck      = "MessageAck"
	MessageReadAck  = "MessageReadAck"
	MessageNotFound = "MessageNotFound"
)
const (
	Request int32 = iota
	Response
)

// returns msg and ok as true if message exists, else returns nil with ok as false
func GetTxtMsgByKey(messages TxtMsgs, msgID string) (msg *TxtMsg) {
	if msg, ok := messages[msgID]; ok {
		return msg
	}
	return nil
}

func DeleteAllTxtMsgsFromDisk(userID string, contactID string) (err error) {
	DeleteDirIfExist(filepath.Join(TextMessagesDir, userID, contactID))
	return err
}

func DeleteTextMessageFromDisk(userID string, contactID string, messages TxtMsgs, message *TxtMsg) bool {
	delete(messages, message.ID)
	filePath := filepath.Join(TextMessagesDir, userID, contactID)
	ok := DeleteFileIfExist(filePath)
	return ok
}

func TextMessagesToArray(messages TxtMsgs) []*TxtMsg {
	if len(messages) < 1 {
		return make([]*TxtMsg, 0, len(messages))
	}
	textMessagesArr := make([]*TxtMsg, 0, len(messages))
	for _, textMessage := range messages {
		textMessagesArr = append(textMessagesArr, textMessage)
	}
	sort.Slice(textMessagesArr, func(i, j int) bool {
		return textMessagesArr[i].TimestampNano < textMessagesArr[j].TimestampNano
	})
	return textMessagesArr
}
func GetPendingMessages(messages TxtMsgs) (filteredMessages []*TxtMsg) {
	filteredMessages = make([]*TxtMsg, 0, len(messages))
	for _, msg := range messages {
		if !msg.ReadAckReceivedOrSent || !msg.AckReceivedOrSent {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	return filteredMessages
}

func MarkAllClientMsgsAsRead(contactID string, messages TxtMsgs) TxtMsgs {
	clientMessages := make(TxtMsgs)
	if len(messages) > 0 {
		for _, msg := range messages {
			if msg.CreatorID == contactID && !msg.MsgRead {
				msg.MsgRead = true
			}
		}
	}
	return clientMessages
}

// call this method if you want to mark all the previous acknowledged msgs created by us
// as read acknowledged
func MarkAllUserAckMsgsAsReadAck(userID string, messages TxtMsgs, timestampLastReadAck int64) (markedMessages TxtMsgs) {
	markedMessages = make(TxtMsgs)
	if len(messages) > 0 {
		for _, msg := range messages {
			if msg.CreatorID == userID && msg.AckReceivedOrSent &&
				msg.Timestamp <= timestampLastReadAck {
				msg.ReadAckReceivedOrSent = true
				markedMessages[msg.ID] = msg
			}
		}
	}
	return markedMessages
}

func GetMsgCopy(msg *TxtMsg) *TxtMsg {
	return &TxtMsg{
		CreatorID:             msg.CreatorID,
		Message:               msg.Message,
		Timestamp:             msg.Timestamp,
		TimestampNano:         msg.TimestampNano,
		ID:                    msg.ID,
		CreatorPublicKey:      msg.CreatorPublicKey,
		Sign:                  msg.Sign,
		AckReceivedOrSent:     msg.AckReceivedOrSent,
		ReadAckReceivedOrSent: msg.ReadAckReceivedOrSent,
		MsgRead:               msg.MsgRead,
	}
}

func SignMessage(pvtKeyHex string, message *TxtMsg) (err error) {
	pvtKey, err := GetPrivateKeyFromHex(pvtKeyHex)
	if err != nil {
		fmt.Println("err in GetDecryptedProtoMessage, in hex.DecodeString, err:", err)
		return err
	}

	message.Sign = nil
	data := EncodeToBytes(message)
	if len(data) == 0 {
		return errors.New(fmt.Sprintf("invalid message:%v", message))
	}
	sign, err := pvtKey.Sign(data)
	if err != nil {
		fmt.Println(err)
		return err
	}
	message.Sign = sign
	return err
}

func VerifyMessage(message *TxtMsg) error {
	publicKeyBytes, err := hex.DecodeString(message.CreatorPublicKey)
	if err != nil {
		return errors.New(fmt.Sprintf("err in VerifyMessage, err:%v", err))
	}
	key, err := crypto.UnmarshalSecp256k1PublicKey(publicKeyBytes)
	if err != nil {
		return errors.New(fmt.Sprintf("err in VerifyMessage, UnmarshalSecp256k1PublicKey err:%v",
			err))
	}

	// extract node id from the provided public key
	idFromKey, err := peer.IDFromPublicKey(key)
	if err != nil {
		return errors.New(fmt.Sprintf("err in VerifyMessage, IDFromPublicKey err:%v", err))
	}

	peerIDFromID, err := peer.Decode(message.CreatorID)
	if err != nil {
		return errors.New(fmt.Sprintf("err in VerifyMessage, Decode err:%v", err))
	}

	// verify that message author node id matches the provided node public key
	if idFromKey != peerIDFromID {
		fmt.Println(err, "Node id and provided public key mismatch")
		return errors.New(fmt.Sprintf("err in VerifyMessage, Node id and provided public key mismatch"))
	}

	sign := message.Sign
	message.Sign = nil
	data := EncodeToBytes(message)
	if len(data) == 0 {
		return errors.New(fmt.Sprintf("err in VerifyMessage, invalid message:%v", message))
	}

	_, err = key.Verify(data, sign)
	if err != nil {
		return errors.New(fmt.Sprintf("err in VerifyMessage, Error authenticating data"))
	}
	message.Sign = sign

	return err
}
