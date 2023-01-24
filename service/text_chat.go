package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/mearaj/protonet/database"
	log "github.com/sirupsen/logrus"
	"runtime"
	"sync"
	"time"
)

const (
	TxtMsgServerProtocol protocol.ID = "/github.com/mearaj/protonet/msg-server/0.0.1"
	TxtMsgClientProtocol             = "/github.com/mearaj/protonet/msg-client/0.0.1"

	TxtMsgLiveServerProtocol = "/github.com/mearaj/protonet/msg-live-server/0.0.1"
	TxtMsgLiveClientProtocol = "/github.com/mearaj/protonet/msg-live-client/0.0.1"

	//TxtMsgActionServerProtocol = "/github.com/mearaj/protonet/msg-action-server/0.0.1"
	//TxtMsgActionClientProtocol = "/github.com/mearaj/protonet/msg-action-client/0.0.1"
)

var MessagesProtocol = map[protocol.ID]protocol.ID{
	TxtMsgServerProtocol:     TxtMsgServerProtocol,
	TxtMsgClientProtocol:     TxtMsgClientProtocol,
	TxtMsgLiveServerProtocol: TxtMsgLiveServerProtocol,
	TxtMsgLiveClientProtocol: TxtMsgLiveClientProtocol,
	//TxtMsgActionServerProtocol: TxtMsgActionServerProtocol,
	//TxtMsgActionClientProtocol: TxtMsgActionClientProtocol,
}

func (tcs *TxtChatService) GetContactServicePeerID() (peerID peer.ID, err error) {
	peerID, err = peer.Decode(tcs.client.IDStr)
	if err != nil {
		log.Println("Error decoding clientID in GetContactServicePeerID, err:", err)
		return peerID, err
	}
	return peerID, err
}
func (tcs *TxtChatService) NewClientStream(protocolID protocol.ID) network.Stream {
	peerID, err := peer.Decode(tcs.client.IDStr)
	if err != nil {
		log.Println(err)
		return nil
	}
	addrInfo := peer.AddrInfo{ID: peerID}.Addrs
	tcs.Host.Peerstore().AddAddrs(peerID, addrInfo, peerstore.PermanentAddrTTL)
	stream, err := tcs.Host.NewStream(context.Background(), peerID, protocolID)
	//log.Println(targetPeerAddr)
	if err != nil {
		// log.Println("Error creating new stream in NewStream, err:", err)
		return nil
	}
	return stream
}

func (tcs *TxtChatService) initiateStreams() {
	readMsgChan := make(chan bool)
	writeMsgChan := make(chan bool)
	readMsgLiveChan := make(chan bool)
	writeMsgLiveChan := make(chan bool)
	msgsSyncHelperChan := make(chan bool)
	go tcs.readTxtMsgsStream(readMsgChan, tcs.GetUserMsgProtocol())
	go tcs.writeTxtMsgsStream(writeMsgChan, tcs.GetClientMsgProtocol())

	go tcs.readTxtMsgsStream(readMsgLiveChan, tcs.GetUserMsgLiveProtocol())
	go tcs.writeTxtMsgsStream(writeMsgLiveChan, tcs.GetClientMsgLiveProtocol())

	go tcs.msgsSyncHelper(msgsSyncHelperChan)

	for {
		if runtime.GOOS == "js" {
			time.Sleep(time.Millisecond)
		}
		select {
		case <-readMsgChan:
			readMsgChan = make(chan bool)
			go tcs.readTxtMsgsStream(readMsgChan, tcs.GetUserMsgProtocol())
		case <-writeMsgChan:
			writeMsgChan = make(chan bool)
			go tcs.writeTxtMsgsStream(writeMsgChan, tcs.GetClientMsgProtocol())

		case <-readMsgLiveChan:
			readMsgLiveChan = make(chan bool)
			go tcs.readTxtMsgsStream(readMsgLiveChan, tcs.GetUserMsgLiveProtocol())
		case <-writeMsgLiveChan:
			writeMsgLiveChan = make(chan bool)
			go tcs.writeTxtMsgsStream(writeMsgLiveChan, tcs.GetClientMsgLiveProtocol())

		case <-msgsSyncHelperChan:
			msgsSyncHelperChan = make(chan bool)
			go tcs.msgsSyncHelper(msgsSyncHelperChan)
		default:

		}
	}
}

func (tcs *TxtChatService) NewTextMessage(msg string) (message *database.TxtMsg, err error) {
	message = &database.TxtMsg{
		CreatorID:             peer.Encode(tcs.Host.ID()),
		Timestamp:             time.Now().Unix(),
		ID:                    uuid.New().String(),
		CreatorPublicKey:      tcs.User.PubKeyHex,
		Sign:                  nil,
		Message:               msg,
		TimestampNano:         time.Now().UnixNano(),
		AckReceivedOrSent:     false,
		ReadAckReceivedOrSent: false,
		MsgRead:               false,
		//Action: tcs.TxtMsgActionInChan,
	}
	err = tcs.AddNewTxtMsg(message)
	if err != nil {
		log.Println("error in NewTextMessage, in AddNewTxtMsg, err:", err)
	}
	return message, err
}

// it compares User's id with client id
func (tcs *TxtChatService) IsUserLower() bool {
	return tcs.User.ID <= tcs.client.IDStr
}

func (tcs *TxtChatService) GetUserMsgProtocol() protocol.ID {
	if tcs.IsUserLower() {
		return TxtMsgServerProtocol
	} else {
		return TxtMsgClientProtocol
	}
}
func (tcs *TxtChatService) GetClientMsgProtocol() protocol.ID {
	if tcs.IsUserLower() {
		return TxtMsgClientProtocol
	} else {
		return TxtMsgServerProtocol
	}
}

func (tcs *TxtChatService) GetUserMsgLiveProtocol() protocol.ID {
	if tcs.IsUserLower() {
		return TxtMsgLiveServerProtocol
	} else {
		return TxtMsgLiveClientProtocol
	}
}
func (tcs *TxtChatService) GetClientMsgLiveProtocol() protocol.ID {
	if tcs.IsUserLower() {
		return TxtMsgLiveClientProtocol
	} else {
		return TxtMsgLiveServerProtocol
	}
}

type TxtChatService struct {
	Host               host.Host
	txtMsgs            database.TxtMsgs
	txtMsgsMutex       sync.Mutex
	db                 *database.Database
	User               *database.Account
	client             *database.Contact
	clientMutex        sync.Mutex
	StreamInChan       chan network.Stream
	StreamLiveInChan   chan network.Stream
	StreamActionInChan chan network.Stream
	TxtMsgOutChan      chan *database.TxtMsg
	TxtMsgLiveOutChan  chan *database.TxtMsg
	changesNotifier    chan<- struct{}
	showNotification   chan<- *database.TxtMsg
}

// string is client id which is unique
type TxtChatServiceMap map[string]*TxtChatService

func NewTxtChatServiceMap(user *database.Account, contacts database.Contacts, host host.Host,
	db *database.Database, notifierChan chan<- struct{},
	showNotChan chan<- *database.TxtMsg) TxtChatServiceMap {
	txtChatServiceMap := make(TxtChatServiceMap)
	for _, contact := range contacts {
		txtChatServiceMap[contact.IDStr] = NewTxtChatService(user, contact, host, db, notifierChan, showNotChan)
	}
	return txtChatServiceMap
}

func NewTxtChatService(user *database.Account, client *database.Contact,
	host host.Host, db *database.Database, notifierChan chan<- struct{}, showNotChan chan<- *database.TxtMsg) (tccs *TxtChatService) {
	txtMsgs := <-db.LoadTxtMsgsFromDisk(user.ID, client.IDStr)
	tccs = &TxtChatService{
		Host:              host,
		client:            client,
		User:              user,
		txtMsgs:           txtMsgs,
		db:                db,
		StreamInChan:      make(chan network.Stream, 10),
		StreamLiveInChan:  make(chan network.Stream, 10),
		TxtMsgOutChan:     make(chan *database.TxtMsg, 10),
		TxtMsgLiveOutChan: make(chan *database.TxtMsg, 10),
		changesNotifier:   notifierChan,
		showNotification:  showNotChan,
	}
	go tccs.initiateStreams()
	return tccs
}

func (tcs *TxtChatService) GetClient() *database.Contact {
	tcs.clientMutex.Lock()
	defer tcs.clientMutex.Unlock()
	return tcs.client
}

func (tcs *TxtChatService) SetClient(client *database.Contact) {
	tcs.clientMutex.Lock()
	defer tcs.clientMutex.Unlock()
	tcs.client = client
	tcs.Notify()
}

func (tcs *TxtChatService) GetTxtMsg(msgID string) (msg *database.TxtMsg) {
	tcs.txtMsgsMutex.Lock()
	defer tcs.txtMsgsMutex.Unlock()
	if msg, ok := tcs.txtMsgs[msgID]; ok {
		return database.GetMsgCopy(msg)
	}
	return nil
}

func (tcs *TxtChatService) AddNewTxtMsg(msg *database.TxtMsg) (err error) {
	tcs.txtMsgsMutex.Lock()
	defer tcs.txtMsgsMutex.Unlock()
	if _, ok := tcs.txtMsgs[msg.ID]; !ok {
		tcs.txtMsgs[msg.ID] = msg
		tcs.db.SaveTxtMsgToDisk(tcs.User.ID, tcs.GetClient().IDStr, msg)
		tcs.Notify()
		return nil
	}
	return errors.New("error in AddNewTxtMsg, msg already exist")
}

func (tcs *TxtChatService) MarkAllClientMsgsAsRead() {
	tcs.txtMsgsMutex.Lock()
	defer tcs.txtMsgsMutex.Unlock()
	txtMsgs := database.MarkAllClientMsgsAsRead(tcs.GetClient().IDStr, tcs.txtMsgs)
	tcs.db.SaveAllTxtMsgsToDisk(tcs.User.ID, tcs.client.IDStr, txtMsgs)
	tcs.Notify()
}

// call this method if you want to mark all the previous acknowledged msgs created by us
// as read acknowledged
func (tcs *TxtChatService) MarkAllUserAckMsgsAsReadAck(timestampLastReadAck int64) {
	tcs.txtMsgsMutex.Lock()
	defer tcs.txtMsgsMutex.Unlock()
	txtMsgs := database.MarkAllUserAckMsgsAsReadAck(tcs.User.ID, tcs.txtMsgs,
		timestampLastReadAck)
	tcs.db.SaveAllTxtMsgsToDisk(tcs.User.ID, tcs.client.IDStr, txtMsgs)
	tcs.Notify()
}

func (tcs *TxtChatService) GetPendingMessages() (filteredMessages []*database.TxtMsg) {
	return database.GetPendingMessages(tcs.GetMsgsCopy())
}

func (tcs *TxtChatService) TextMessagesToArray() []*database.TxtMsg {
	return database.TextMessagesToArray(tcs.GetMsgsCopy())
}

func (tcs *TxtChatService) GetMsgsCopy() database.TxtMsgs {
	tcs.txtMsgsMutex.Lock()
	defer tcs.txtMsgsMutex.Unlock()
	txtMsgs := make(database.TxtMsgs, len(tcs.txtMsgs))
	for key, msg := range tcs.txtMsgs {
		txtMsgs[key] = &database.TxtMsg{
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
	return txtMsgs
}

func (tcs *TxtChatService) GetMsgCopy(msg *database.TxtMsg) *database.TxtMsg {
	tcs.txtMsgsMutex.Lock()
	defer tcs.txtMsgsMutex.Unlock()
	txtMsg := &database.TxtMsg{
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
	return txtMsg
}

func (tcs *TxtChatService) SaveUpdateTxtMsgToDisk(userID string, contactID string, message *database.TxtMsg) {
	tcs.txtMsgsMutex.Lock()
	defer tcs.txtMsgsMutex.Unlock()
	tcs.txtMsgs[message.ID] = message
	tcs.db.SaveTxtMsgToDisk(userID, contactID, message)
	tcs.Notify()
}

func (tcs *TxtChatService) Notify() {
	select {
	case tcs.changesNotifier <- struct{}{}:
	default:
	}
}
