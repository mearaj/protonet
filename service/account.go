package service

import (
	"encoding/hex"
	"errors"
	libcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"time"
)

type Account struct {
	UpdatedAt time.Time
	// PrivateKeyEnc is encrypted private key(private key itself is represented in hex string)
	PrivateKeyEnc string
	// PublicKey is hex representation of public key as string
	PublicKey    string `gorm:"primaryKey"`
	PublicImage  []byte
	PrivateImage []byte
}

func (a *Account) PrivateKey(passwd string) (string, error) {
	if a.PrivateKeyEnc == "" {
		return "", errors.New("encrypted private key doesn't exist")
	}
	pvtKeyBs, err := hex.DecodeString(a.PrivateKeyEnc)
	if err != nil {
		return "", err
	}
	pvtKeyBytes, err := Decrypt([]byte(passwd), pvtKeyBs)
	if err != nil {
		return "", err
	}
	pvtKeyHex := hex.EncodeToString(pvtKeyBytes)
	pvtKey, err := libcrypto.UnmarshalSecp256k1PrivateKey(pvtKeyBytes)
	if err != nil {
		return "", err
	}
	pubKey := pvtKey.GetPublic()
	pubKeyBs, err := pubKey.Raw()
	if err != nil {
		return "", err
	}
	pubKeyStr := hex.EncodeToString(pubKeyBs)
	if a.PublicKey != pubKeyStr {
		return "", errors.New("invalid password")
	}
	return pvtKeyHex, err
}
