package service

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"gioui.org/app"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ethereum/go-ethereum/crypto"
	libcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/mearaj/protonet/alog"
	"github.com/mearaj/protonet/assets"
	"golang.org/x/crypto/scrypt"
	"image"
	"image/gif"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"net/http"
	"os"
	"path/filepath"
)

// returns the image in encoded format
// if there's an error then returns assets.AppIcon image as fallback
func fetchAvatarWithFallBack(url string) []byte {
	res, err := http.Get(url)
	if err != nil {
		alog.Logger().Errorln(err)
		return assets.AppIcon
	}
	img, ext, err := image.Decode(res.Body)
	if err != nil {
		alog.Logger().Errorln(err)
		return assets.AppIcon
	}
	var buff bytes.Buffer
	switch ext {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buff, img, nil)
		if err != nil {
			alog.Logger().Errorln(err)
			return assets.AppIcon
		}
	case "png":
		err = png.Encode(&buff, img)
		if err != nil {
			alog.Logger().Errorln(err)
			return assets.AppIcon
		}
	case "gif":
		err = gif.Encode(&buff, img, nil)
		if err != nil {
			alog.Logger().Errorln(err)
			return assets.AppIcon
		}
	}
	return buff.Bytes()
}

func VerifyMessage(message *Message) (err error) {
	publicKey, err := GetPublicKeyFromStr(message.AccountPublicKey, libcrypto.Secp256k1)
	if err != nil {
		return
	}
	sign := message.Sign
	message.Sign = nil
	data := EncodeToBytes(message)
	if len(data) == 0 {
		return fmt.Errorf("err in VerifyMessage, invalid message:%v", message)
	}
	_, err = publicKey.Verify(data, sign)
	if err != nil {
		return fmt.Errorf("err in VerifyMessage, Error authenticating data")
	}
	message.Sign = sign
	return err
}

// strc is a pointer to a struct
func DecodeToStruct(strc interface{}, s []byte) (err error) {
	dec := gob.NewDecoder(bytes.NewReader(s))
	err = dec.Decode(strc)
	if err != nil {
		alog.Logger().Errorln(err)
		return
	}
	return nil
}
func NewAEAD(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		alog.Logger().Errorln(err)
		return nil, err
	}
	return cipher.NewGCM(block)
}

func SignMessage(pvtKeyHex string, message *Message) (err error) {
	pvtKey, err := GetPrivateKeyFromStr(pvtKeyHex, libcrypto.Secp256k1)
	if err != nil {
		return err
	}

	message.Sign = nil
	data := EncodeToBytes(message)
	if len(data) == 0 {
		return fmt.Errorf("invalid message:%v", message)
	}
	sign, err := pvtKey.Sign(data)
	if err != nil {
		return err
	}
	message.Sign = sign
	return err
}

// Ref https://pkg.go.dev/github.com/decred/dcrd/dcrec/secp256k1/v3#example-package-EncryptDecryptMessage
func GetEncryptedStruct(pubKeyHex string, message interface{}, algo int) (data []byte, err error) {
	libPubKey, err := GetPublicKeyFromStr(pubKeyHex, algo)
	if err != nil {
		return nil, err
	}
	pubKeyBytes, err := libPubKey.Raw()
	if err != nil {
		return nil, err
	}
	switch algo {
	// Ref https://stackoverflow.com/questions/39410808/how-to-convert-a-interface-into-type-rsa-publickey-golang
	// https://stackoverflow.com/questions/40243857/how-to-encrypt-large-file-with-rsa
	case libcrypto.RSA:
		return nil, errors.New("rsa algorithm not supported")
	case libcrypto.Secp256k1:
		pubKey, err := secp256k1.ParsePubKey(pubKeyBytes)
		if err != nil {
			alog.Logger().Errorln(err)
			return nil, err
		}
		ephemeralPrivKey, err := secp256k1.GeneratePrivateKey()
		if err != nil {
			alog.Logger().Errorln(err)
			return nil, err
		}
		ephemeralPubKey := ephemeralPrivKey.PubKey().SerializeCompressed()
		cipherKey := sha256.Sum256(secp256k1.GenerateSharedSecret(ephemeralPrivKey, pubKey))
		aead, err := NewAEAD(cipherKey[:])
		if err != nil {
			alog.Logger().Errorln(err)
			return nil, err
		}
		nonce := make([]byte, aead.NonceSize())
		ciphertext := make([]byte, 4+len(ephemeralPubKey))
		binary.LittleEndian.PutUint32(ciphertext, uint32(len(ephemeralPubKey)))
		copy(ciphertext[4:], ephemeralPubKey)
		data = EncodeToBytes(message)
		ciphertext = aead.Seal(ciphertext, nonce, data, ephemeralPubKey)
		return ciphertext, err
	default:
		return nil, errors.New("algorithm not supported")
	}
}

func GetDecryptedStruct(pvtKeyHex string, msgEncrypted []byte, message interface{}, algo int) (err error) {
	libPrivateKey, err := GetPrivateKeyFromStr(pvtKeyHex, algo)
	if err != nil {
		return err
	}
	pkBytes, err := libPrivateKey.Raw()
	if err != nil {
		return
	}
	switch algo {
	case libcrypto.RSA:
		// Ref https://stackoverflow.com/questions/39410808/how-to-convert-a-interface-into-type-rsa-publickey-golang
		// https://stackoverflow.com/questions/40243857/how-to-encrypt-large-file-with-rsa
		return errors.New("rsa algorithm not supported")
	case libcrypto.Secp256k1:
		privKey := secp256k1.PrivKeyFromBytes(pkBytes)
		pubKeyLen := binary.LittleEndian.Uint32(msgEncrypted[:4])
		pubKeyBytes := msgEncrypted[4 : 4+pubKeyLen]
		pubKey, err := secp256k1.ParsePubKey(pubKeyBytes)
		if err != nil {
			return err
		}
		recoveredCipherKey := sha256.Sum256(secp256k1.GenerateSharedSecret(privKey, pubKey))
		// Open the sealed message.
		aead, err := NewAEAD(recoveredCipherKey[:])
		if err != nil {
			return err
		}
		nonce := make([]byte, aead.NonceSize())
		recoveredData, err := aead.Open(nil, nonce, msgEncrypted[4+pubKeyLen:], pubKeyBytes)
		if err != nil {
			return err
		}
		err = DecodeToStruct(message, recoveredData)
		if err != nil {
			return err
		}
		return err
	default:
		return errors.New("algorithm not supported")
	}
}

// Ref https://gist.github.com/SteveBate/042960baa7a4795c3565
func EncodeToBytes(str interface{}) []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(str)
	if err != nil {
		alog.Logger().Errorln(err)
		return nil
	}
	return buf.Bytes()
}

func CreatePrivateKey(typ int) (privateKey libcrypto.PrivKey, privateKeyStr string, publicKey libcrypto.PubKey, publicKeyStr string, id peer.ID, err error) {
	privateKey, publicKey, err = libcrypto.GenerateKeyPair(typ, 2048)
	if err != nil {
		return privateKey, privateKeyStr, publicKey, publicKeyStr, id, err
	}
	publicKeyBytes, err := publicKey.Raw()
	if err != nil {
		return privateKey, privateKeyStr, publicKey, publicKeyStr, id, err
	}
	privateKeyBytes, err := privateKey.Raw()
	if err != nil {
		return privateKey, privateKeyStr, publicKey, publicKeyStr, id, err
	}
	publicKeyStr = hex.EncodeToString(publicKeyBytes)
	privateKeyStr = hex.EncodeToString(privateKeyBytes)
	id, err = peer.IDFromPublicKey(publicKey)
	if err != nil {
		return privateKey, privateKeyStr, publicKey, publicKeyStr, id, err
	}

	return privateKey, privateKeyStr, publicKey, publicKeyStr, id, nil
}

func GetPublicKeyFromStr(publicKeyStr string, algo int) (libcrypto.PubKey, error) {
	publicKeyBytes, err := hex.DecodeString(publicKeyStr)
	if err != nil {
		return nil, err
	}
	switch algo {
	case libcrypto.RSA:
		publicKey, err := libcrypto.UnmarshalRsaPublicKey(publicKeyBytes)
		return publicKey, err
	case libcrypto.Secp256k1:
		publicKey, err := libcrypto.UnmarshalSecp256k1PublicKey(publicKeyBytes)
		return publicKey, err
	case libcrypto.Ed25519:
		publicKey, err := libcrypto.UnmarshalEd25519PublicKey(publicKeyBytes)
		return publicKey, err
	case libcrypto.ECDSA:
		publicKey, err := libcrypto.UnmarshalECDSAPublicKey(publicKeyBytes)
		return publicKey, err
	}
	return nil, errors.New("invalid public key")
}

func GetPrivateKeyFromStr(privateKeyStr string, algo int) (privateKey libcrypto.PrivKey, err error) {
	privateKeyBytes, err := hex.DecodeString(privateKeyStr)
	if err != nil {
		return nil, err
	}
	defer func() {
		if r := recover(); r != nil {
			alog.Logger().Errorln("recovered from error", r)
		}
	}()
	switch algo {
	case libcrypto.RSA:
		privateKey, err = libcrypto.UnmarshalRsaPrivateKey(privateKeyBytes)
		return privateKey, err
	case libcrypto.Secp256k1:
		privateKey, err = libcrypto.UnmarshalSecp256k1PrivateKey(privateKeyBytes)
		return privateKey, err
	case libcrypto.Ed25519:
		//priv := ed25519.NewKeyFromSeed(privateKeyBytes)
		//pubKeyBytes := ([]byte)(priv.Public().(ed25519.PublicKey))
		//ed25519PrivateKeyBytes := append(privateKeyBytes, pubKeyBytes...)
		privateKey, err = libcrypto.UnmarshalEd25519PrivateKey(privateKeyBytes)
		return privateKey, err
	case libcrypto.ECDSA:
		pvtKey, err := crypto.HexToECDSA(privateKeyStr)
		if err != nil {
			return nil, err
		}
		pvtKey.PublicKey.Curve = elliptic.P256()
		pvtKey.PublicKey.X, pvtKey.PublicKey.Y = pvtKey.PublicKey.Curve.ScalarBaseMult(pvtKey.D.Bytes())
		privateKey, _, err = libcrypto.ECDSAKeyPairFromKey(pvtKey)
		return privateKey, err
	}
	return nil, errors.New("algorithm not supported")
}
func openAppFile(filename string) (*os.File, error) {
	var err error
	defer func() {
		if err != nil {
			alog.Logger().Errorln(err)
		}
	}()
	dirPath, err := app.DataDir()
	if err != nil {
		return nil, err
	}
	dirPath = filepath.Join(dirPath, DBPathCfgDir)
	if _, err = os.Stat(dirPath); os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, 0700)
		if err != nil {
			return nil, err
		}
	}
	file, err := os.OpenFile(
		filepath.Join(dirPath, filename),
		os.O_CREATE|os.O_APPEND|os.O_RDWR,
		0700,
	)
	if err != nil {
		return nil, err
	}
	return file, err
}

// Encrypt https://bruinsslot.jp/post/golang-crypto/
func Encrypt(key, data []byte) ([]byte, error) {
	key, salt, err := DeriveKey(key, nil)
	if err != nil {
		return nil, err
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	ciphertext = append(ciphertext, salt...)

	return ciphertext, nil
}

// Decrypt https://bruinsslot.jp/post/golang-crypto/
func Decrypt(key, data []byte) ([]byte, error) {
	salt, data := data[len(data)-32:], data[:len(data)-32]

	key, _, err := DeriveKey(key, salt)
	if err != nil {
		return nil, err
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// DeriveKey https://bruinsslot.jp/post/golang-crypto/
func DeriveKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
	}

	key, err := scrypt.Key(password, salt, 16384, 8, 1, 32)
	if err != nil {
		return nil, nil, err
	}

	return key, salt, nil
}
