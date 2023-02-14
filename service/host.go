package service

import (
	"context"
	"errors"
	"fmt"
	ipfsdatastore "github.com/ipfs/go-datastore"
	ipfssync "github.com/ipfs/go-datastore/sync"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	libcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/mearaj/protonet/alog"
	"time"
)

func (s *service) makeHost() (host.Host, error) {
	if s.getUserPassword() == "" {
		return nil, errors.New("password is not set")
	}
	account := s.Account()
	pvtKeyStr, err := account.PrivateKey(s.getUserPassword())
	if err != nil {
		return nil, err
	}
	pvtKey, err := GetPrivateKeyFromStr(pvtKeyStr, libcrypto.Secp256k1)
	if err != nil {
		alog.Logger().Errorln(err)
		return nil, err
	}
	hst, err := libp2p.New(
		libp2p.Identity(pvtKey),
		libp2p.NATPortMap(),
		libp2p.DefaultListenAddrs,
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.DefaultPeerstore,
		libp2p.DefaultEnableRelay,
		libp2p.DefaultResourceManager,
		libp2p.DefaultConnectionManager,
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

// runService keeps the service running as long as app is running,
// should be called only once for the entire lifecycle of app
func (s *service) runService() {
reloadClientService:
	var sub Subscriber
	var tckr *time.Ticker
	account := s.Account()
	for account.PublicKey == "" {
		time.Sleep(time.Millisecond * 100)
		account = s.Account()
	}
	if tckr != nil {
		tckr.Stop()
	}
	for s.GormDB() == nil {
		time.Sleep(time.Millisecond * 100)
	}
	hst := s.Host()
	var err error
	if hst != nil {
		s.setHost(nil)
		hst.RemoveStreamHandler(ProtocolChat)
		hst.RemoveStreamHandler(ProtocolSync)
		err := hst.Close()
		if err != nil {
			alog.Logger().Errorln(err)
		}
	}
	s.syncStreams.Clear()
	s.chatStreams.Clear()
	syncChannels := s.syncStreamsOutCh.Values()
	s.syncStreamsOutCh.Clear() // clear the map before closing channel
	for _, ch := range syncChannels {
		go close(ch)
	}
	chatChannels := s.chatStreamsOutCh.Values()
	s.chatStreamsOutCh.Clear() // clear the map before closing channel
	for _, ch := range chatChannels {
		go close(ch)
	}
	hst, err = s.makeHost()
	for err != nil {
		time.Sleep(time.Millisecond * 100)
		hst, err = s.makeHost()
	}
	s.setHost(hst)
	for _, addr := range dht.DefaultBootstrapPeers {
		pi, _ := peer.AddrInfoFromP2pAddr(addr)
		_ = hst.Connect(context.Background(), *pi)
	}
	if sub == nil || sub.IsClosed() {
		sub = s.Subscribe()
	}
	fmt.Println("Listening at :")
	for _, addr := range hst.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, hst.ID().Pretty())
	}
	hst.SetStreamHandler(ProtocolChat, s.handleHostChatStream)
	hst.SetStreamHandler(ProtocolSync, s.handleSyncStream)
	limit := 50
	tckr = time.NewTicker(time.Second * 1)
	for {
		select {
		case <-sub.Events():
			if s.Account().PublicKey != account.PublicKey {
				goto reloadClientService
			}
		case <-tckr.C:
			contactsCount := int(<-s.ContactsCount(account.PublicKey))
			for offset := 0; offset < contactsCount; offset += limit {
				contacts := <-s.Contacts(account.PublicKey, offset, limit)
				for _, eachContact := range contacts {
					if s.Account().PublicKey != account.PublicKey {
						goto reloadClientService
					}
					if _, ok := s.syncStreamsOutCh.Value(eachContact.PublicKey); !ok {
						ch := make(chan Sync, 10)
						s.syncStreamsOutCh.Add(eachContact.PublicKey, ch)
					}
					if _, ok := s.syncStreams.Value(eachContact.PublicKey); !ok {
						publicKey, err := GetPublicKeyFromStr(eachContact.PublicKey, libcrypto.Secp256k1)
						if err != nil {
							alog.Logger().Errorln(err)
							continue
						}
						peerID, err := peer.IDFromPublicKey(publicKey)
						if err != nil {
							alog.Logger().Errorln(err)
							continue
						}
						stream, err := hst.NewStream(context.Background(), peerID, ProtocolSync)
						if err != nil {
							alog.Logger().Errorln(err)
							continue
						}
						if _, ok := s.syncStreams.Value(eachContact.PublicKey); !ok {
							s.handleSyncStream(stream)
						}
					}
				}
			}
		default:
			if account.PublicKey != s.Account().PublicKey {
				goto reloadClientService
			}
		}
	}
}
