package service

import (
	"context"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-kad-dht"
	log "github.com/sirupsen/logrus"
	"protonet.live/database"
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

func (cs *ChatService) loadUserAccounts() {
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
