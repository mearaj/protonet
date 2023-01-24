package database

import (
	"encoding/hex"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
)

type Contact struct {
	IDStr     string
	PublicKey string
	Name      string
	Image     []byte
}
type Contacts map[string]*Contact

func AddContact(host host.Host, userID string, clientID string, name string) (contact *Contact, err error) {
	peerID, err := peer.Decode(clientID)
	if err != nil {
		log.Println("error in SaveContactToDisk, invalid clientID", err)
		return nil, err
	}

	publicKey := host.Peerstore().PubKey(peerID)
	publicKeyBytes, err := publicKey.Raw()
	if err != nil {
		log.Println("error in SaveContactToDisk, in publicKey.Raw", err)
		return nil, err
	}
	publicKeyStr := hex.EncodeToString(publicKeyBytes)
	contact = &Contact{
		IDStr:     clientID,
		PublicKey: publicKeyStr,
		Name:      name,
		Image:     nil,
	}
	return contact, err
}
