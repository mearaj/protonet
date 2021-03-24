package service

import (
	"context"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/libp2p/go-libp2p-kad-dht"
	mplex "github.com/libp2p/go-libp2p-mplex"
	"github.com/libp2p/go-libp2p-secio"
	"github.com/libp2p/go-libp2p-webrtc-direct"
	"github.com/pion/webrtc/v3"
	log "github.com/sirupsen/logrus"
	"protonet.live/database"
	"runtime"
	"time"
)

type ChatService struct {
	db                *database.Database
	isServiceReady    bool
	Host              host.Host
	user              *database.Account
	accounts          database.Accounts
	contacts          database.Contacts
	txtChatServiceMap TxtChatServiceMap
	changesNotifer    chan struct{}
	showNotification  chan *database.TxtMsg
}

func NewChatService() *ChatService {
	cs := &ChatService{
		accounts:         database.Accounts{},
		changesNotifer:   make(chan struct{}, 10),
		showNotification: make(chan *database.TxtMsg, 10),
	}
	db, initCh := database.NewDatabase()
	cs.db = db

	go func() {
		<-initCh
		cs.initChatService()
	}()
	return cs
}

func (cs *ChatService) IsServiceReady() bool {
	return cs.isServiceReady && cs.db.IsDatabaseReady()
}
func (cs *ChatService) setIsServiceReady(isReady bool) {
	cs.isServiceReady = isReady
}

// go routine
func (cs *ChatService) initChatService() {
	cs.setIsServiceReady(false)
	for !cs.db.IsDatabaseReady() {
		time.Sleep(time.Second / 10)
	}
	cs.loadUserAccounts()
	cs.removeStreamHandlers()
	cs.createHost()
	cs.initTextChatServicesMap()
	cs.setStreamHandlers()
	// FixME: this function takes too long
	go cs.addContactsToLibAdds()
	select {
	case cs.changesNotifer <- struct{}{}:
	default:
	}
	cs.setIsServiceReady(true)
}

//func (cs *ChatService) GetSavePayloadChan() chan<- ChatServicePayload {
//	return cs.chatPayloadChan
//}

func (cs *ChatService) loadUserAccounts() () {
	cs.accounts = <-cs.db.LoadAccountsFromDisk()
	if len(cs.accounts) == 0 {
		cs.user = <-cs.db.LoadUserFromDisk()
		if cs.user == nil || cs.user.PvtKeyHex == "" {
			cs.user = cs.db.CreateNewAccount()
		}
		cs.accounts[cs.user.ID] = cs.user
		cs.db.SaveUserToDisk(cs.user)
		cs.db.SaveAccountsToDisk(cs.accounts)
	}
	cs.user = <-cs.db.LoadUserFromDisk()

	if cs.user == nil || cs.user.PvtKeyHex == "" {
		for _, user := range cs.accounts {
			cs.user = user
			cs.db.SaveUserToDisk(user)
			return
		}
	}
}

func (cs *ChatService) SaveAccountToDisk(acc *database.Account) {
	cs.db.SaveAccountToDisk(acc)
	if acc.ID == cs.user.ID {
		cs.db.SaveUserToDisk(cs.user)
	}
}

func (cs *ChatService) CreateNewAccount() (acc *database.Account) {
	acc = cs.db.CreateNewAccount()
	if acc != nil {
		cs.SaveAccountToDisk(acc)
	}
	cs.loadUserAccounts()
	return acc
}

func (cs *ChatService) createHost() () {
	pvtKey, err := database.GetPrivateKeyFromHex(cs.GetCurrentUser().PvtKeyHex)
	if err != nil {
		log.Println("error in createHost, in GetPrivateKeyFromHex err:", err)
		return
	}
	webTrs := libp2pwebrtcdirect.NewTransport(
		webrtc.Configuration{},
		new(mplex.Transport),
	)
	retry := 0
	for retry < 5 {
		// Attempt to open ports using uPNP for NATed hosts.
		var natPortMap libp2p.Option

		if runtime.GOOS != "js" {
			natPortMap = libp2p.NATPortMap()
		}

		// create a new libp2p Host that listens on a random TCP port
		cs.Host, err = libp2p.New(context.Background(),
			// Use the keypair we generated
			// libp2p.Identity(ua.PvtKey),
			libp2p.Identity(pvtKey),
			// Multiple listen addresses
			libp2p.ListenAddrStrings(
				"/ip4/0.0.0.0/tcp/0", // regular tcp connections
				"/ip4/0.0.0.0/udp/0", // regular tcp connections
				"/ip4/0.0.0.0/tcp/0/ws",
				"/ip4/0.0.0.0/udp/0/quic",
				"/ip4/0.0.0.0/tcp/0/http/p2p-webrtc-direct",
			),
			// support secio connections
			libp2p.Security(secio.ID, secio.New),
			// support any other default transports (TCP)
			libp2p.DefaultTransports,
			// Let's prevent our peer from having too many
			// connections by attaching a connection manager.
			libp2p.ConnectionManager(connmgr.NewConnManager(
				500,            // Lowwater
				1000,           // HighWater,
				time.Minute*10, // GracePeriod previous value time.Minute * 10
			)),
			natPortMap,
			libp2p.Transport(webTrs),
			// Let this Host use the DHT to find other hosts
			libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
				var idht *dht.IpfsDHT
				idht, err := dht.New(context.Background(), h)
				return idht, err
			}),
			// Let this Host use relays and advertise itself on relays if
			// it finds it is behind NAT. Use libp2p.Relay(options...) to
			// enable active relays and more.
			libp2p.DefaultEnableRelay,
			libp2p.DefaultMuxers,
			libp2p.EnableNATService(),
			libp2p.DefaultSecurity,
			libp2p.DefaultPeerstore,
			libp2p.DefaultListenAddrs,
			//libp2p.NoSecurity,
		)
		if err != nil {
			log.Println("err creating libp2p Host, err is", err)
			log.Println("Retrying creating Host, max tries is", 5, "this is", retry+1, "attempt")
		} else {
			break
		}
		retry++
	}
	if err != nil {
		log.Println(err)
	}
	log.Println(cs.Host.Network().ListenAddresses())
	return
}
func (cs *ChatService) initTextChatServicesMap() {
	user := cs.GetCurrentUser()
	cs.contacts = <-cs.db.LoadContactsFromDisk(user.ID)
	cs.txtChatServiceMap = NewTxtChatServiceMap(user, cs.contacts, cs.Host, cs.db, cs.changesNotifer, cs.showNotification)
}
func (cs *ChatService) setStreamHandlers() {
	for _, eachProtocol := range MessagesProtocol {
		cs.Host.SetStreamHandler(eachProtocol, cs.handleHostStream)
	}
}
func (cs *ChatService) removeStreamHandlers() {
	if cs.Host != nil {
		for _, eachProtocol := range MessagesProtocol {
			cs.Host.RemoveStreamHandler(eachProtocol)
		}
		err := cs.Host.Close()
		if err != nil {
			log.Println(err)
		}
	}
}

// go routine
func (cs *ChatService) addContactsToLibAdds() {
	contacts := <-cs.db.LoadContactsFromDisk(cs.GetCurrentUser().ID)
	for _, contact := range contacts {
		cs.addContactToLibAdds(contact.IDStr)
	}
}

func (cs *ChatService) addContactToLibAdds(peerIDStr string) {
	peerID, err := peer.Decode(peerIDStr)
	if err != nil {
		log.Println("Error decoding clientID in cs.addContactToLibAdds, err:",
			err)
	}

	addrInfo := peer.AddrInfo{ID: peerID}.Addrs
	cs.Host.Peerstore().AddAddrs(peerID, addrInfo, peerstore.PermanentAddrTTL)

	for _, addr := range dht.DefaultBootstrapPeers {
		pi, _ := peer.AddrInfoFromP2pAddr(addr)
		_ = cs.Host.Connect(context.Background(), *pi)
	}
}

func (cs *ChatService) GetCurrentUser() *database.Account {
	return cs.user
}

func (cs *ChatService) GetAccounts() database.Accounts {
	return cs.accounts
}

func (cs *ChatService) GetAccountsArray() []*database.Account {
	accountsArr := make([]*database.Account, 0, len(cs.accounts))
	for _, acc := range cs.accounts {
		accountsArr = append(accountsArr, acc)
	}
	return accountsArr
}

func (cs *ChatService) AddContact(clientID string, name string) (err error) {
	if _, ok := cs.txtChatServiceMap[clientID]; !ok {
		contact, _ := database.AddContact(cs.Host, cs.GetCurrentUser().ID, clientID, name)
		cs.db.SaveContactToDisk(cs.GetCurrentUser().ID, contact)
		cs.txtChatServiceMap[clientID] =
			NewTxtChatService(cs.GetCurrentUser(), contact, cs.Host, cs.db, cs.changesNotifer, cs.showNotification)
		go cs.addContactToLibAdds(contact.IDStr)
	} else {
		contact := cs.txtChatServiceMap[clientID].GetClient()
		contact.Name = name
		contact.IDStr = clientID
		go cs.addContactToLibAdds(contact.IDStr)
		cs.db.SaveContactToDisk(cs.GetCurrentUser().ID, contact)
	}
	return err
}

func (cs *ChatService) GetTxtChatServicesMap() TxtChatServiceMap {
	return cs.txtChatServiceMap
}

func (cs *ChatService) DeleteTxtChatServicesMapItems(clientIDs []string) TxtChatServiceMap {
	for _, clientID := range clientIDs {
		if txtChatService, ok := cs.txtChatServiceMap[clientID]; ok {
			cs.db.DeleteContact(cs.GetCurrentUser().ID, txtChatService.GetClient())
			delete(cs.txtChatServiceMap, clientID)
		}
	}
	return cs.txtChatServiceMap
}

// go routine
func (cs *ChatService) SetCurrentUser(user *database.Account) {
	cs.user = user
	cs.db.SaveUserToDisk(user)
	go cs.initChatService()
}

// This method assumes protocolID as
func (cs *ChatService) handleHostStream(stream network.Stream) {
	clientId := stream.Conn().RemotePeer().String()
	textContactChatService, ok := cs.txtChatServiceMap[clientId]
	if !ok {
		log.Println("cs.handleHostStream, client not found")
		err := stream.Reset()
		if err != nil {
			log.Println("cs.handleHostStream, stream.Reset, err:")
		}
		return
	}
	if stream.Protocol() == TxtMsgServerProtocol ||
		stream.Protocol() == TxtMsgClientProtocol {
		// stream.SetProtocol(textContactChatService.GetUserMsgProtocol())
		textContactChatService.StreamInChan <- stream
	} else if stream.Protocol() == TxtMsgLiveServerProtocol ||
		stream.Protocol() == TxtMsgLiveClientProtocol {
		// stream.SetProtocol(textContactChatService.GetUserMsgLiveProtocol())
		textContactChatService.StreamLiveInChan <- stream
	}
	//else if stream.Protocol() == TxtMsgActionServerProtocol ||
	//	stream.Protocol() == TxtMsgActionClientProtocol {
	//	//stream.SetProtocol(textContactChatService.GetUserActionProtocol())
	//	textContactChatService.StreamActionInChan <- stream
	//}
}

func (cs *ChatService) GetChangesNotifier() <-chan struct{} {
	return cs.changesNotifer
}

func (cs ChatService) GetShowNotification() <-chan *database.TxtMsg {
	return cs.showNotification
}
