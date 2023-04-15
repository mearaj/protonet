package chat

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	ipfsdatastore "github.com/ipfs/go-datastore"
	ipfssync "github.com/ipfs/go-datastore/sync"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	libcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/mearaj/protonet/alog"
	"github.com/mearaj/protonet/internal/common"
	"github.com/mearaj/protonet/internal/pubsub"
	"github.com/mearaj/protonet/internal/wallet"
	"github.com/mearaj/protonet/utils"
	"io"
	"sync"
	"time"
)

var ErrStreamReset = network.ErrReset

type Chat interface {
	SendNewMessage(account *Account, message *Message)
}
type chat struct {
	host      host.Host
	hostError error
	hostMutex sync.RWMutex
	// chatStreams key is publicKey of peer
	chatStreams      utils.Map[string, network.Stream]
	chatStreamsOutCh utils.Map[string, chan Message]
}

var GlobalChat = chat{
	chatStreams:      utils.NewMap[string, network.Stream](),
	chatStreamsOutCh: utils.NewMap[string, chan Message](),
}

func init() {
	go GlobalChat.runChat()
}

var _ Chat = &chat{}

func (c *chat) Host() (host.Host, error) {
	c.hostMutex.RLock()
	defer c.hostMutex.RUnlock()
	return c.host, c.hostError
}

func (c *chat) setHost(host host.Host, err error) {
	c.hostMutex.Lock()
	defer c.hostMutex.Unlock()
	c.host = host
	c.hostError = err
}

func (c *chat) handleHostChatStream(stream network.Stream) {
	pubKey := stream.Conn().RemotePublicKey()
	pubKeyBytes, err := pubKey.Raw()
	if err != nil {
		alog.Logger().Errorln(err)
		return
	}
	pubKeyStr := hex.EncodeToString(pubKeyBytes)
	c.chatStreams.Set(pubKeyStr, stream)
	go c.writeChatStream(stream, pubKeyStr)
	go c.readChatStream(stream, pubKeyStr)
}

func (c *chat) readChatStream(stream network.Stream, contactPubKeyHex string) {
	var err error
	defer func() {
		if err != nil {
			alog.Logger().Errorln(err)
		}
		if err != nil && errors.Is(err, ErrStreamReset) {
			c.chatStreams.Delete(contactPubKeyHex)
		}
	}()
	for err == nil || !errors.Is(err, ErrStreamReset) {
		b := make([]byte, 8)
		_, err = io.ReadFull(stream, b)
		if err != nil {
			continue
		}
		sizeOfMsg := binary.LittleEndian.Uint32(b)
		if sizeOfMsg > 0 {
			pb := make([]byte, sizeOfMsg)
			_, err = io.ReadFull(stream, pb)
			if err != nil {
				continue
			}
			networkMsg := Message{}
			var acc Account
			acc, err = wallet.GlobalWallet.Account()
			if err != nil {
				continue
			}
			//var pvtKeyStr string
			//pvtKeyStr, err = wallet.GlobalWallet.GetPrivateKey(acc)
			//if err != nil {
			//	continue
			//}
			err = common.GetDecryptedStruct(acc.PrivateKey, pb, &networkMsg, libcrypto.ECDSA)
			if err != nil {
				continue
			}
			verPublicKey := acc.PublicKey
			isMsgCreatedByMe := verPublicKey == networkMsg.Sender
			if isMsgCreatedByMe {
				verPublicKey = networkMsg.Recipient
			}
			err = common.VerifyMessage(&networkMsg, verPublicKey, libcrypto.ECDSA)
			if err != nil {
				continue
			}
			remotePublicKey := stream.Conn().RemotePublicKey()
			var remotePublicKeyBytes []byte
			remotePublicKeyBytes, err = remotePublicKey.Raw()
			if err != nil {
				continue
			}
			remotePublicKeyHex := hex.EncodeToString(remotePublicKeyBytes)
			// Message is either created by user or his peer
			if acc2, err := wallet.GlobalWallet.Account(); acc2.PublicKey != acc.PublicKey || err != nil {
				continue
			}
			msgIsValid := (acc.PublicKey == networkMsg.Recipient && networkMsg.Sender == remotePublicKeyHex) ||
				(acc.PublicKey == networkMsg.Sender && networkMsg.Recipient == remotePublicKeyHex)
			if !msgIsValid {
				err = errors.New("invalid message")
				continue
			}
			dbMessage := Message{
				ID:        networkMsg.ID,
				Sender:    networkMsg.Sender,
				Recipient: networkMsg.Recipient,
				CreatedAt: networkMsg.CreatedAt}
			var key string
			key, err = dbMessage.GetDBFullKey(acc.PublicKey)
			if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
				continue
			}
			err = wallet.GlobalWallet.ViewRecord([]byte(key), &dbMessage)
			// this shouldn't happen,
			msgExist := err == nil
			// if networkMsg is not created by me
			if !isMsgCreatedByMe {
				// if networkMsg already exist, then it implies user's peer is requesting for the updated state
				if msgExist {
					if networkMsg.State < MessageStateReceived {
						networkMsg.State = MessageStateReceived
					}
					if networkMsg.State < dbMessage.State {
						networkMsg.State = dbMessage.State
					}
				} else {
					// if networkMsg is new, then update the state to MessageStateReceived
					networkMsg.State = MessageStateReceived
				}
			} else {
				// if the message is created by me, then it implies user's peer is providing updated State
				if networkMsg.State < dbMessage.State {
					networkMsg.State = dbMessage.State
				}
			}
			var msgCh chan Message
			var ok bool
			if msgCh, ok = c.chatStreamsOutCh.Get(networkMsg.Sender); !ok {
				msgCh = make(chan Message, 10)
				c.chatStreamsOutCh.Set(networkMsg.Sender, msgCh)
			}
			select {
			case msgCh <- networkMsg:
			default:
			}
			err = wallet.GlobalWallet.SaveOrUpdateMessage(acc.PublicKey, &networkMsg)
		}
	}
}

func (c *chat) writeChatStream(stream network.Stream, contactPubKeyHex string) {
	var err error
	defer func() {
		if err != nil {
			alog.Logger().Errorln(err)
		}
		if err != nil && errors.Is(err, ErrStreamReset) {
			c.chatStreams.Delete(contactPubKeyHex)
		}
	}()
	account, err := wallet.GlobalWallet.Account()
	if err != nil {
		return
	}
	rw := bufio.NewWriter(stream)
	// if current account is changed, then return
	if acc, err := wallet.GlobalWallet.Account(); acc.PublicKey != account.PublicKey || err != nil {
		return
	}
	var ch chan Message
	var ok bool
	if ch, ok = c.chatStreamsOutCh.Get(contactPubKeyHex); !ok {
		ch = make(chan Message, 10)
		c.chatStreamsOutCh.Set(contactPubKeyHex, ch)
	}
	for err == nil || !errors.Is(err, ErrStreamReset) {
		// if current account is changed, then return
		if acc, err := wallet.GlobalWallet.Account(); acc.PublicKey != account.PublicKey || err != nil {
			return
		}
		if dbMsg, ok := <-ch; ok {
			// reports the error encountered in prev iteration
			if err != nil {
				alog.Logger().Errorln(err)
			}
			//var pvtKeyStr string
			//pvtKeyStr, err = wallet.GlobalWallet.GetPrivateKey(account)
			//if err != nil {
			//	continue
			//}
			err = common.SignMessage(account.PrivateKey, &dbMsg, libcrypto.ECDSA)
			if err != nil {
				continue
			}
			var bytes []byte
			bytes, err = common.GetEncryptedStruct(contactPubKeyHex, dbMsg, libcrypto.ECDSA)
			if err != nil {
				continue
			}
			messageSize := uint32(len(bytes))
			b := make([]byte, 8)
			binary.LittleEndian.PutUint32(b, messageSize)
			b = append(b, bytes...)
			_, err = rw.Write(b)
			if err != nil {
				if errors.Is(err, ErrStreamReset) {
					return
				}
				continue
			}
			err = rw.Flush()
			if err != nil {
				if errors.Is(err, ErrStreamReset) {
					return
				}
				continue
			}
		}
	}
}

//func (c *chat) getPrivateKeyFromPasswd(a Account, passwd string) (string, error) {
//	if a.PrivateKey == "" {
//		return "", errors.New("encrypted private key doesn't exist")
//	}
//	pvtKeyBs, err := hex.DecodeString(a.PrivateKey)
//	if err != nil {
//		return "", err
//	}
//	pvtKeyBytes, err := common.Decrypt([]byte(passwd), pvtKeyBs)
//	if err != nil {
//		return "", err
//	}
//	pvtKeyHex := hex.EncodeToString(pvtKeyBytes)
//	pvtKey, err := common.GetPrivateKeyFromStr(pvtKeyHex, libcrypto.ECDSA)
//	if err != nil {
//		return "", err
//	}
//	pubKey := pvtKey.GetPublic()
//	pubKeyBs, err := pubKey.Raw()
//	if err != nil {
//		return "", err
//	}
//	pubKeyStr := hex.EncodeToString(pubKeyBs)
//	if a.PublicKey != pubKeyStr {
//		return "", errors.New("invalid password")
//	}
//	return pvtKeyHex, err
//}

func (c *chat) SendNewMessage(identity *Account, message *Message) {
	go func() {
		var err error
		defer func() {
			if r := recover(); r != nil {
				alog.Logger().Errorln(r)
			}
		}()
		message.Sender = identity.PublicKey
		message.ID = uuid.New().String()
		err = wallet.GlobalWallet.SaveOrUpdateMessage(identity.PublicKey, message)
		if err != nil {
			alog.Logger().Errorln(err)
		}
		hst, err := c.Host()
		if err == nil {
			publicKey, err := common.GetPublicKeyFromStr(message.Recipient, libcrypto.ECDSA)
			if err != nil {
				alog.Logger().Errorln(err)
				return
			}
			peerID, err := peer.IDFromPublicKey(publicKey)
			if err != nil {
				alog.Logger().Errorln(err)
				return
			}
			if _, ok := c.chatStreams.Get(message.Recipient); !ok {
				stream, err := hst.NewStream(context.Background(), peerID, ProtocolChat)
				if err != nil {
					alog.Logger().Errorln(err)
					return
				}
				c.handleHostChatStream(stream)
			}
			var msgCh chan Message
			var ok bool
			if msgCh, ok = c.chatStreamsOutCh.Get(message.Recipient); !ok {
				msgCh = make(chan Message, 10)
				c.chatStreamsOutCh.Set(message.Recipient, msgCh)
			}
			select {
			case msgCh <- *message:
			default:
			}
		}
	}()
}

var ErrHostNotInitialized = errors.New("host not initialized")

func (c *chat) makeHost() (host.Host, error) {
	if !wallet.GlobalWallet.IsOpen() {
		return nil, errors.New("password is not set")
	}
	account, err := wallet.GlobalWallet.Account()
	if err != nil {
		return nil, err
	}
	//pvtKeyStr, err := wallet.GlobalWallet.GetPrivateKey(account)
	//if err != nil {
	//	return nil, err
	//}
	pvtKey, err := common.GetPrivateKeyFromStr(account.PrivateKey, libcrypto.ECDSA)
	if err != nil {
		alog.Logger().Errorln(err)
		return nil, err
	}
	hst, err := libp2p.New(
		libp2p.Identity(pvtKey),
		libp2p.NATPortMap(),
		// Let this Host use the DHT to find other hosts
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			var idht *dht.IpfsDHT
			idht, err := dht.New(context.Background(), h)
			return idht, err
		}),
	)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	dstore := ipfssync.MutexWrap(ipfsdatastore.NewMapDatastore())
	dHT := dht.NewDHT(ctx, hst, dstore)
	routedHost := routedhost.Wrap(hst, dHT)
	err = dHT.Bootstrap(ctx)
	if err != nil {
		alog.Logger().Errorln(err)
	}
	return routedHost, nil
}

// runChat keeps the chat running as long as app is running,
// should be called only once for the entire lifecycle of app
func (c *chat) runChat() {
reloadClientService:
	var sub *pubsub.Subscriber
	var tckr *time.Ticker
	account, _ := wallet.GlobalWallet.Account()
	for account.PublicKey == "" {
		time.Sleep(time.Millisecond * 100)
		account, _ = wallet.GlobalWallet.Account()
	}
	if tckr != nil {
		tckr.Stop()
	}
	hst, err := c.Host()
	if hst != nil {
		c.setHost(nil, ErrHostNotInitialized)
		hst.RemoveStreamHandler(ProtocolChat)
		err := hst.Close()
		if err != nil {
			alog.Logger().Errorln(err)
		}
	}
	c.chatStreams.Clear()
	chatChannels := c.chatStreamsOutCh.Values()
	c.chatStreamsOutCh.Clear() // clear the map before closing channel
	for _, ch := range chatChannels {
		go close(ch)
	}
	hst, err = c.makeHost()
	for err != nil {
		time.Sleep(time.Millisecond * 100)
		hst, err = c.makeHost()
	}
	c.setHost(hst, nil)
	for _, addr := range dht.DefaultBootstrapPeers {
		pi, _ := peer.AddrInfoFromP2pAddr(addr)
		_ = hst.Connect(context.Background(), *pi)
	}
	if sub == nil || sub.IsClosed() {
		sub = pubsub.NewSubscriber()
		_ = sub.Subscribe()
		wallet.GlobalWallet.EventBroker.AddSubscriber(sub)
	}
	fmt.Println("Listening at :")
	for _, addr := range hst.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, hst.ID().String())
	}
	hst.SetStreamHandler(ProtocolChat, c.handleHostChatStream)
	limit := int64(50)
	tckr = time.NewTicker(time.Second * 1)
	if acc, err := wallet.GlobalWallet.Account(); acc.PublicKey != account.PublicKey || err != nil {
		goto reloadClientService
	}
	for {
		select {
		case <-sub.Events():
			if acc, err := wallet.GlobalWallet.Account(); acc.PublicKey != account.PublicKey || err != nil {
				goto reloadClientService
			}
		case <-tckr.C:
			contactsCount, _ := wallet.GlobalWallet.ContactsCount(account.PublicKey)
			for offset := int64(0); offset < contactsCount; offset += limit {
				contacts, _ := wallet.GlobalWallet.Contacts(account.PublicKey, int(offset), int(limit))
				for _, eachContact := range contacts {
					if acc, err := wallet.GlobalWallet.Account(); acc.PublicKey != account.PublicKey || err != nil {
						goto reloadClientService
					}
					var publicKey libcrypto.PubKey
					publicKey, err = common.GetPublicKeyFromStr(eachContact.PublicKey, libcrypto.ECDSA)
					if err != nil {
						alog.Logger().Errorln(err)
						continue
					}
					peerID, err := peer.IDFromPublicKey(publicKey)
					if err != nil {
						alog.Logger().Errorln(err)
						continue
					}
					if _, ok := c.chatStreams.Get(eachContact.PublicKey); !ok {
						stream, err := hst.NewStream(context.Background(), peerID, ProtocolChat)
						if err != nil {
							alog.Logger().Errorln(err)
							continue
						}
						c.handleHostChatStream(stream)
					}
					if _, ok := c.chatStreamsOutCh.Get(eachContact.PublicKey); !ok {
						ch := make(chan Message, 10)
						c.chatStreamsOutCh.Set(eachContact.PublicKey, ch)
					}
					msgLimit := int64(100)
					count, _ := wallet.GlobalWallet.MessagesCount(account.PublicKey, eachContact.PublicKey)
					// Resend the message for which we haven't received the read ack
					for msgOffset := int64(0); err == nil && msgOffset < count; msgOffset += msgLimit {
						if acc, err := wallet.GlobalWallet.Account(); acc.PublicKey != account.PublicKey || err != nil {
							goto reloadClientService
						}
						msgs, _ := wallet.GlobalWallet.Messages(account.PublicKey, eachContact.PublicKey, int(msgOffset), int(msgLimit))
						for _, msg := range msgs {
							if acc, err := wallet.GlobalWallet.Account(); acc.PublicKey != account.PublicKey || err != nil {
								goto reloadClientService
							}
							// should send if I haven't received read acknowledgement for my messages
							shouldSend := msg.State < MessageStateRead && msg.Sender == account.PublicKey
							if shouldSend {
								msgCh, _ := c.chatStreamsOutCh.Get(eachContact.PublicKey)
								select {
								case msgCh <- msg:
								default:
								}
							}
						}
					}
				}
			}
		default:
			if acc, err := wallet.GlobalWallet.Account(); acc.PublicKey != account.PublicKey || err != nil {
				goto reloadClientService
			}
		}
	}
}
