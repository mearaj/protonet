package database

import (
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	RSA       = crypto.RSA
	Ed25519   = crypto.Ed25519
	Secp256k1 = crypto.Secp256k1
	ECDSA     = crypto.ECDSA
)

type Account struct {
	PvtKeyHex string
	PubKeyHex string
	ID        string
	Name      string
	accType   int
	PvtImg    []byte
	PubImg    []byte
}
type Accounts map[string]*Account

func (db *Database) CreateNewAccount() (user *Account) {
	user = &Account{}
	if name, ok := os.LookupEnv("USER"); ok {
		user.Name = name
	} else {
		user.Name = "ProtoUser"
	}
	pvtKey, pvtKeyStr, publicKey, err := GeneratePrivateKey(Secp256k1)
	peerID, err := peer.IDFromPrivateKey(pvtKey)
	if err != nil {
		log.Println("error in db.CreateNewAccount, err", err)
		return nil
	}
	peerIDStr := peer.Encode(peerID)
	user.PvtKeyHex, user.ID, user.PubKeyHex = pvtKeyStr, peerIDStr, publicKey
	user.PubImg = make([]byte, 0, 0)
	user.PvtImg = make([]byte, 0, 0)
	return
}

func AccountsToArray(accounts Accounts) []*Account {
	accountsArr := make([]*Account, 0, len(accounts))
	for _, acc := range accounts {
		accountsArr = append(accountsArr, acc)
	}
	return accountsArr
}
