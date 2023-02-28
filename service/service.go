package service

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/google/uuid"
	libcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/mearaj/protonet/alog"
	"github.com/mearaj/protonet/utils"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
	"sync"
)

type Service interface {
	Initialized() bool
	Account() Account
	Accounts() <-chan []Account
	Contacts(accountPublicKey string, offset, limit int) <-chan []Contact
	//SetPrivateKey(privateKeyHex string) error
	SetUserPassword(passwd string) <-chan error
	Messages(contactPubKey string, offset, limit int) <-chan []Message
	CreateAccount(privateKeyHex string) <-chan error
	SendMessage(contactPublicKey string, msg string, createdTimestamp string) <-chan error
	SaveContact(contactPublicKey string, identified bool) <-chan error
	AutoCreateAccount() <-chan error
	AccountKeyExists(publicKey string) <-chan bool
	SetAsCurrentAccount(account Account) <-chan error
	Subscribe(topics ...EventTopic) Subscriber
	LastMessage(contactPublicKey string) <-chan Message
	DeleteAccounts([]Account) <-chan error
	DeleteContacts([]Contact) <-chan error
	UnreadMessagesCount(contactPublicKey string) <-chan int64
	MessagesCount(contactPublicKey string) <-chan int64
	ContactsCount(addrPublicKey string) <-chan int64
	AccountsCount() <-chan int64
	MarkPrevMessagesAsRead(contactAddr string) <-chan error
	UserPasswordSet() bool
}

// Service Always call GetServiceInstance function to create Service
type service struct {
	account       Account
	accountMutex  sync.RWMutex
	database      interface{}
	databaseMutex sync.RWMutex
	host          host.Host
	hostMutex     sync.RWMutex
	eventBroker   *eventBroker
	// syncStreams key is publicKey of peer
	syncStreams      utils.Map[string, network.Stream]
	syncStreamsOutCh utils.Map[string, chan Sync]
	// chatStreams key is publicKey of peer
	chatStreams       utils.Map[string, network.Stream]
	chatStreamsOutCh  utils.Map[string, chan Message]
	userPassword      string
	userPasswordMutex sync.RWMutex
}

var serviceInstance = service{
	eventBroker:      newEventBroker(),
	syncStreams:      utils.NewMap[string, network.Stream](),
	syncStreamsOutCh: utils.NewMap[string, chan Sync](),
	chatStreams:      utils.NewMap[string, network.Stream](),
	chatStreamsOutCh: utils.NewMap[string, chan Message](),
}

func init() {
	go serviceInstance.init()
}

func GetServiceInstance() Service {
	return &serviceInstance
}

func (s *service) Account() Account {
	s.accountMutex.RLock()
	defer s.accountMutex.RUnlock()
	return s.account
}

func (s *service) setAccount(account Account) {
	currAccount := s.Account()
	s.accountMutex.Lock()
	s.account = account
	s.accountMutex.Unlock()
	if currAccount.PublicKey != account.PublicKey &&
		strings.TrimSpace(account.PublicKey) != "" {
		<-s.saveAccountToDB(account)
		event := Event{Data: AccountChangedEventData{}, Topic: AccountChangedEventTopic}
		s.eventBroker.Fire(event)
	}
}

func (s *service) Host() host.Host {
	s.hostMutex.RLock()
	defer s.hostMutex.RUnlock()
	if s.host == nil {
		return nil
	}
	return s.host
}

func (s *service) setHost(host host.Host) {
	s.hostMutex.Lock()
	defer s.hostMutex.Unlock()
	s.host = host
}

func (s *service) Initialized() bool {
	s.databaseMutex.RLock()
	defer s.databaseMutex.RUnlock()
	return s.database != nil
}

func (s *service) CreateAccount(pvtKeyHex string) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		var err error
		defer func() {
			recoverPanicCloseCh(errCh, err, alog.Logger())
		}()
		if strings.TrimSpace(pvtKeyHex) == "" {
			err = errors.New("private key is empty")
			return
		}
		if !s.Initialized() {
			err = errors.New("database engine not running")
			return
		}
		if s.getUserPassword() == "" {
			err = errors.New("password is not set")
			return
		}
		algo := libcrypto.Secp256k1
		pvtKey, err := GetPrivateKeyFromStr(pvtKeyHex, algo)
		if err != nil {
			alog.Logger().Errorln(err)
			return
		}
		pubKeyBytes, err := pvtKey.GetPublic().Raw()
		if err != nil {
			alog.Logger().Errorln(err)
			return
		}
		publicKeyStr := hex.EncodeToString(pubKeyBytes)
		pvtKeyBytes, err := hex.DecodeString(pvtKeyHex)
		if err != nil {
			return
		}
		pvtKeyEncBs, err := Encrypt([]byte(s.getUserPassword()), pvtKeyBytes)
		if err != nil {
			return
		}
		pvtKeyEnc := hex.EncodeToString(pvtKeyEncBs)
		account := Account{
			PrivateKeyEnc: pvtKeyEnc,
			PublicKey:     publicKeyStr,
		}
		s.setAccount(account)
	}()
	return errCh
}

func (s *service) AutoCreateAccount() <-chan error {
	errCh := make(chan error, 1)
	go func() {
		var err error
		defer func() {
			recoverPanicCloseCh(errCh, err, alog.Logger())
		}()
		if !s.Initialized() {
			err = errors.New("database engine not running")
			return
		}
		if s.getUserPassword() == "" {
			err = errors.New("password is not set")
			return
		}
		_, pvtKeyStr, _, publicKeyStr, _, err := CreatePrivateKey(libcrypto.Secp256k1)
		if err != nil {
			alog.Logger().Errorln(err)
			return
		}
		pvtKeyBytes, err := hex.DecodeString(pvtKeyStr)
		if err != nil {
			return
		}
		pvtKeyEncBs, err := Encrypt([]byte(s.getUserPassword()), pvtKeyBytes)
		if err != nil {
			return
		}
		pvtKeyEnc := hex.EncodeToString(pvtKeyEncBs)

		account := Account{
			PrivateKeyEnc: pvtKeyEnc,
			PublicKey:     publicKeyStr,
		}
		s.setAccount(account)
		s.eventBroker.Fire(Event{
			Data:  AccountsChangedEventData{},
			Topic: AccountsChangedEventTopic,
		})
	}()
	return errCh
}

//func (s *service) SetPrivateKey(privateKeyHex string) error {
//	acc := s.Account()
//	if strings.TrimSpace(acc.PublicKey) == "" {
//		return errors.New("accounts is empty")
//	}
//	privateKey, err := GetPrivateKeyFromStr(privateKeyHex, libcrypto.Secp256k1)
//	if err != nil {
//		return err
//	}
//	pubKeyBytes, err := privateKey.GetPublic().Raw()
//	if err != nil {
//		return err
//	}
//	pubKeyHex := hex.EncodeToString(pubKeyBytes)
//	if acc.PublicKey != pubKeyHex {
//		return errors.New("invalid private key")
//	}
//	acc.PrivateKey = privateKeyHex
//	s.setAccount(acc)
//	return nil
//}

func (s *service) SendMessage(contactPublicKey string, msg string, created string) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		var err error
		defer func() {
			recoverPanicCloseCh(errCh, err, alog.Logger())
		}()
		identity := s.Account()
		dbMsg := Message{
			AccountPublicKey: identity.PublicKey,
			ContactPublicKey: contactPublicKey,
			From:             identity.PublicKey,
			To:               contactPublicKey,
			Text:             msg,
			Created:          created,
			ID:               uuid.New(),
		}
		err = <-s.saveMessage(dbMsg)
		if err != nil {
			alog.Logger().Errorln(err)
		}
		eventData := MessagesCountChangedEventData{
			AccountPublicKey: dbMsg.AccountPublicKey,
			ContactPublicKey: dbMsg.ContactPublicKey,
		}
		event := Event{Data: eventData, Topic: MessagesCountChangedEventTopic}
		s.eventBroker.Fire(event)
		hst := s.Host()
		if hst != nil {
			if _, ok := s.chatStreams.Value(contactPublicKey); !ok {
				ch := make(chan Sync, 10)
				s.syncStreamsOutCh.Add(contactPublicKey, ch)
			}

			publicKey, err := GetPublicKeyFromStr(contactPublicKey, libcrypto.Secp256k1)
			if err != nil {
				alog.Logger().Errorln(err)
				return
			}
			peerID, err := peer.IDFromPublicKey(publicKey)
			if err != nil {
				alog.Logger().Errorln(err)
				return
			}
			if _, ok := s.chatStreams.Value(contactPublicKey); !ok {
				stream, err := hst.NewStream(context.Background(), peerID, ProtocolChat)
				if err != nil {
					alog.Logger().Errorln(err)
					return
				}
				s.handleHostChatStream(stream)
			}
			var msgCh chan Message
			var ok bool
			if msgCh, ok = s.chatStreamsOutCh.Value(contactPublicKey); !ok {
				msgCh = make(chan Message, 10)
			}
			select {
			case msgCh <- dbMsg:
			default:
			}

		}
	}()
	return errCh
}

func (s *service) handleHostChatStream(stream network.Stream) {
	pubKey := stream.Conn().RemotePublicKey()
	pubKeyBytes, err := pubKey.Raw()
	if err != nil {
		alog.Logger().Errorln(err)
		return
	}
	pubKeyStr := hex.EncodeToString(pubKeyBytes)
	s.chatStreams.Add(pubKeyStr, stream)
	go s.writeChatStream(stream, pubKeyStr)
	go s.readChatStream(stream, pubKeyStr)
}

func (s *service) readChatStream(stream network.Stream, pubKeyHex string) {
	var err error
	defer func() {
		recoverPanic(alog.Logger())
		if err != nil && err.Error() == "stream reset" {
			s.chatStreams.Delete(pubKeyHex)
		}
	}()
	for err == nil || (err.Error() != "stream reset") {
		b := make([]byte, 8)
		_, err = io.ReadFull(stream, b)
		if err != nil {
			alog.Logger().Errorln(err)
			continue
		}
		sizeOfMsg := binary.LittleEndian.Uint32(b)
		if sizeOfMsg > 0 {
			pb := make([]byte, sizeOfMsg)
			_, err = io.ReadFull(stream, pb)
			if err != nil {
				alog.Logger().Errorln(err)
				continue
			}
			msg := Message{}
			acc := s.Account()
			var pvtKeyStr string
			pvtKeyStr, err = acc.PrivateKey(s.getUserPassword())
			if err != nil {
				alog.Logger().Errorln(err)
				continue
			}
			err = GetDecryptedStruct(pvtKeyStr, pb, &msg, libcrypto.Secp256k1)
			if err != nil {
				alog.Logger().Errorln(err)
				continue
			}
			err = VerifyMessage(&msg)
			if err != nil {
				alog.Logger().Errorln(err)
				continue
			}
			remotePublicKey := stream.Conn().RemotePublicKey()
			remotePublicKeyBytes, err := remotePublicKey.Raw()
			if err != nil {
				alog.Logger().Errorln(err)
				continue
			}
			remotePublicKeyHex := hex.EncodeToString(remotePublicKeyBytes)
			msgIsValid := s.Account().PublicKey == msg.ContactPublicKey &&
				msg.AccountPublicKey == remotePublicKeyHex
			if !msgIsValid {
				alog.Logger().Errorln("public key mismatch")
				continue
			}
			msg.AccountPublicKey = acc.PublicKey
			msg.ContactPublicKey = remotePublicKeyHex
			msg.From = remotePublicKeyHex
			msg.To = acc.PublicKey
			<-s.saveMessage(msg)
			syncMsgs := SyncMessages{
				MessagesReceived: []string{msg.ID.String()},
				MessagesRead:     []string{},
				MessagesNotFound: []string{},
			}
			syncResp := Sync{
				Type: SyncResponse,
				Data: syncMsgs,
			}
			if ch, ok := s.syncStreamsOutCh.Value(remotePublicKeyHex); ok {
				select {
				case ch <- syncResp:
				default:
				}
			}
		}
	}
}

func (s *service) writeChatStream(stream network.Stream, pubKeyHex string) {
	var err error
	account := s.Account()
	defer func() {
		if err != nil {
			alog.Logger().Errorln(err)
		}
		if err != nil && err.Error() == "stream reset" {
			s.chatStreams.Delete(pubKeyHex)
		}
	}()
	rw := bufio.NewWriter(stream)
	// if current account is changed, then return
	if account.PublicKey != s.Account().PublicKey {
		return
	}
	var ch chan Message
	var ok bool
	if ch, ok = s.chatStreamsOutCh.Value(pubKeyHex); !ok {
		ch = make(chan Message, 10)
		s.chatStreamsOutCh.Add(pubKeyHex, ch)
	}
	for err == nil || err.Error() != "stream reset" {
		// if current account is changed, then return
		if account.PublicKey != s.Account().PublicKey {
			return
		}
		if dbMsg, ok := <-ch; ok {
			acc := s.Account()
			var pvtKeyStr string
			pvtKeyStr, err = acc.PrivateKey(s.getUserPassword())
			if err != nil {
				alog.Logger().Errorln(err)
				continue
			}
			err = SignMessage(pvtKeyStr, &dbMsg)
			if err != nil {
				alog.Logger().Errorln(err)
				continue
			}
			var bytes []byte
			bytes, err = GetEncryptedStruct(dbMsg.ContactPublicKey, dbMsg, libcrypto.Secp256k1)
			if err != nil {
				alog.Logger().Errorln(err)
				continue
			}
			messageSize := uint32(len(bytes))
			b := make([]byte, 8)
			binary.LittleEndian.PutUint32(b, messageSize)
			b = append(b, bytes...)
			_, err = rw.Write(b)
			if err != nil {
				alog.Logger().Errorln(err)
				if err.Error() == "stream reset" {
					return
				}
				continue
			}
			err = rw.Flush()
			if err != nil {
				alog.Logger().Errorln(err)
				if err.Error() == "stream reset" {
					return
				}
				continue
			}
		}
	}
}

func (s *service) Subscribe(topics ...EventTopic) Subscriber {
	subscr := newSubscriber()
	_ = subscr.Subscribe(topics...)
	s.eventBroker.addSubscriber(subscr)
	return subscr
}

func recoverPanic(entry *logrus.Entry) {
	if r := recover(); r != nil {
		entry.Errorln("recovered from panic", r)
	}
}
func recoverPanicCloseCh[S any](stateChan chan<- S, state S, entry *logrus.Entry) {
	recoverPanic(entry)
	stateChan <- state
	close(stateChan)
}
