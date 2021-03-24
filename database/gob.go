package database

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"github.com/decred/dcrd/dcrec/secp256k1/v3"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

// Ref https://gist.github.com/SteveBate/042960baa7a4795c3565
func EncodeToBytes(str interface{}) ([]byte) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(str)
	if err != nil {
		log.Println(err)
		return nil
	}
	//log.Println("EncodeToBytes, uncompressed size (bytes): ", len(buf.Bytes()))
	return buf.Bytes()
}

func Compress(s []byte) []byte {
	zipbuf := bytes.Buffer{}
	zipped := gzip.NewWriter(&zipbuf)
	zipped.Write(s)
	zipped.Close()
	//log.Println("Compress, compressed size (bytes): ", len(zipbuf.Bytes()))
	return zipbuf.Bytes()
}

func Decompress(s []byte) []byte {
	rdr, _ := gzip.NewReader(bytes.NewReader(s))
	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		log.Println(err)
		return nil
	}
	rdr.Close()
	return data
}

// strc is a pointer to a struct
func DecodeToStruct(strc interface{}, s []byte) (err error) {
	dec := gob.NewDecoder(bytes.NewReader(s))
	err = dec.Decode(strc)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func WriteBytesToFile(dirPath string , filename string, bs []byte) (err error){
	file, err := CreateFileIfNotExist(dirPath, filename)
	if err != nil {
		log.Println("error in WriteBytesToFile, CreateFileIfNotExist", dirPath, filename, err)
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Println("error in WriteBytesToFile, file.Close", dirPath, filename, err)
		}
	}()

	err = ioutil.WriteFile(file.Name(), bs, os.ModePerm)
	if err != nil {
		log.Println("error in WriteBytesToFile, ioutil.WriteFile", dirPath, filename, err)
		return err
	}
	return err
}

func ReadFromFile(dirPath string, filename string) []byte {
	file, err := CreateFileIfNotExist(dirPath, filename)
	if err != nil {
		log.Println("error in ReadFromFile, CreateFileIfNotExist", dirPath, filename,  err)
		return nil
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Println("error in ReadFromFile, file.Close", dirPath, filename, err)
		}
	}()

	bs, err := ioutil.ReadFile(file.Name())
	if err != nil {
		log.Println("error in ReadFromFile, ioutil.ReadFile", dirPath, filename, err)
		return nil
	}
	return bs
}
// strc is a struct
func SaveStructToFile(dirPath string, fileName string, strc interface{})  (err error) {
	encData := EncodeToBytes(strc)
	encData = Compress(encData)
	return WriteBytesToFile(dirPath, fileName, encData)
}

// strc is a pointer to struct
func LoadStructFromFile(dirPath string, filename string, strc interface{}) error {
	bs := ReadFromFile(dirPath, filename)
	if len(bs) == 0 {
		return nil
	}
	bs = Decompress(bs)
	return DecodeToStruct(strc, bs)
}

// Ref https://pkg.go.dev/github.com/decred/dcrd/dcrec/secp256k1/v3#example-package-EncryptDecryptMessage
func GetEncryptedStruct(pubKeyHex string, message interface{}) (data []byte, err error) {
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		log.Println("err in GetEncryptedStruct, in hex.DecodeString, err:", err)
		return nil, err
	}
	pubKey, err := secp256k1.ParsePubKey(pubKeyBytes)
	if err != nil {
		log.Println("err in GetEncryptedStruct, in secp256k1.ParsePubKey, err:", err)
		return nil, err
	}
	ephemeralPrivKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		log.Println("err in GetEncryptedStruct, in secp256k1.GeneratePrivateKey(),"+
			"err:", err)
		return nil, err
	}
	ephemeralPubKey := ephemeralPrivKey.PubKey().SerializeCompressed()
	cipherKey := sha256.Sum256(secp256k1.GenerateSharedSecret(ephemeralPrivKey, pubKey))
	aead, err := NewAEAD(cipherKey[:])
	if err != nil {
		log.Println("err in GetEncryptedStruct, in NewAEAD, err:", err)
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	ciphertext := make([]byte, 4+len(ephemeralPubKey))
	binary.LittleEndian.PutUint32(ciphertext, uint32(len(ephemeralPubKey)))
	copy(ciphertext[4:], ephemeralPubKey)
	data = EncodeToBytes(message)
	ciphertext = aead.Seal(ciphertext, nonce, data, ephemeralPubKey)
	return ciphertext, err
}

func GetDecryptedStruct(pvtKeyHex string, msgEncrypted []byte, message interface{}) (err error) {
	// Decode the hex-encoded private key.
	pkBytes, err := hex.DecodeString(pvtKeyHex)
	if err != nil {
		log.Println("error in GetDecryptedStruct, in hex.DecodeString, err", err)
		return
	}

	privKey := secp256k1.PrivKeyFromBytes(pkBytes)
	pubKeyLen := binary.LittleEndian.Uint32(msgEncrypted[:4])
	senderPubKeyBytes := msgEncrypted[4 : 4+pubKeyLen]
	senderPubKey, err := secp256k1.ParsePubKey(senderPubKeyBytes)
	if err != nil {
		log.Println("error in GetDecryptedStruct, in hex.DecodeString, err", err)
		return
	}
	recoveredCipherKey := sha256.Sum256(secp256k1.GenerateSharedSecret(privKey, senderPubKey))
	// Open the sealed message.
	aead, err := NewAEAD(recoveredCipherKey[:])
	if err != nil {
		log.Println("error in GetDecryptedStruct, in NewAEAD, err", err)
		return
	}
	nonce := make([]byte, aead.NonceSize())
	recoveredData, err := aead.Open(nil, nonce, msgEncrypted[4+pubKeyLen:], senderPubKeyBytes)
	if err != nil {
		log.Println("error in GetDecryptedStruct, in aead.Open, err", err)
		return
	}
	err = DecodeToStruct(message, recoveredData)
	if err != nil {
		log.Println("error in GetDecryptedStruct, in DecodeToStruct, err", err)
		return
	}
	return err
}