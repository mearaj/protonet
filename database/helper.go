package database

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"gioui.org/app"
	"github.com/libp2p/go-libp2p-core/crypto"
	log "github.com/sirupsen/logrus"
	"golang.org/x/image/colornames"
	"image/color"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const AppDir string = "protonet"

const (
	ContactsDir      = "contacts"
	ContactsFile     = "contacts.dat"
	TextMessagesDir  = "text_messages"
	TextMessagesFile = "text_messages.dat"
	AccountsDir      = "accounts"
	AccountDir       = "account"
	AccountsFile     = "accounts.dat"
	AccountFile      = "account.dat"
)

func GeneratePrivateKey(accType int) (pvtKey crypto.PrivKey, pvtKeyStr string, publicKeyHex string, err error) {
	pvtKey, publicKey, err := crypto.GenerateKeyPair(accType, 2048)
	if err != nil {
		log.Println("error in GeneratePrivateKey err:", err)
		return
	}

	publicKeyBytes, err := publicKey.Raw()
	if err != nil {
		log.Println("error in GeneratePrivateKey,publicKey.Raw", err)
		return
	}
	publicKeyHex = hex.EncodeToString(publicKeyBytes)
	pvtKeyStr = fmt.Sprintf("%x", pvtKey.(*crypto.Secp256k1PrivateKey).D)
	return pvtKey, pvtKeyStr, publicKeyHex, err
}

func GetPrivateKeyFromHex(hexString string) (pvtKey crypto.PrivKey, err error) {
	hexToBytes, err := hex.DecodeString(hexString)
	if err != nil {
		log.Println("error in GetPrivateKeyFromHex, err:", err)
		return
	}
	pvtKey, err = crypto.UnmarshalSecp256k1PrivateKey(hexToBytes)
	if err != nil {
		log.Println("error in GetPrivateKeyFromHex, UnmarshalSecp256k1PrivateKey err:", err)
		return
	}
	return
}

func GetRandomNRGBA(index int) color.NRGBA {
	colors := []color.NRGBA{
		{R: colornames.Red.R, G: colornames.Red.G, B: colornames.Red.B, A: colornames.Red.A},
		{R: colornames.Green.R, G: colornames.Green.G, B: colornames.Green.B, A: colornames.Green.A},
		{R: colornames.Blue.R, G: colornames.Blue.G, B: colornames.Blue.B, A: colornames.Blue.A},
		{R: colornames.Brown.R, G: colornames.Brown.G, B: colornames.Brown.B, A: colornames.Brown.A},
		{R: colornames.Purple.R, G: colornames.Purple.G, B: colornames.Purple.B, A: colornames.Purple.A},
	}
	return colors[index%len(colors)]
}

func GetInitialsFromName(name string) string {
	if len(name) > 2 {
		nameArr := strings.Split(name, " ")
		if len(nameArr) > 1 {
			return strings.ToUpper(string(nameArr[0][0])) + strings.ToUpper(string(nameArr[1][0]))
		} else {
			if len(nameArr[0]) > 1 {
				return strings.ToUpper(string(nameArr[0][0])) +
					strings.ToUpper(string(nameArr[0][1]))
			} else if len(nameArr[0]) == 1 {
				return strings.ToUpper(string(nameArr[0][0])) + "P"
			}
		}
	} else if len(name) == 1 {
		return strings.ToUpper(name) + "P"
	} else if len(name) == 2 {
		return strings.ToUpper(name)
	}
	return "PR"
}

func NewAEAD(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func DeleteFileIfExist(fileName string) bool {
	dirPath, err := app.DataDir()
	if err != nil {
		return false
	}
	fullPath := filepath.Join(dirPath, AppDir, fileName)
	err = os.Remove(fullPath)
	if err != nil {
		log.Println("error in DeleteFileIfExist, err:", err)
	}
	return err != nil
}
func CreateDirIfNotExist(dirName string, perm os.FileMode) (dirPath string, err error) {
	dirPath, err = app.DataDir()
	if err != nil {
		log.Println("error in CreateDirIfNotExist, err:", err)
		return
	}
	dirPath = filepath.Join(dirPath, AppDir, dirName)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, perm)
		if err != nil {
			return dirPath, err
		}
	}
	return
}

func DeleteDirIfExist(dirName string) bool {
	// dirPath is the root directory for our app dynamically allocated by the framework
	dirPath, err := app.DataDir()
	if err != nil {
		return false
	}
	// Everything resides inside dirPath/AppDir/
	// where AppDir is currently protonet.live
	fullPath := filepath.Join(dirPath, AppDir, dirName)
	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		err = os.RemoveAll(fullPath)
		if err != nil {
			return false
		}
	}
	return true
}

func CreateFileIfNotExist(dirPath string, filename string) (file *os.File, err error) {
	dirPath, err = CreateDirIfNotExist(dirPath, os.ModePerm)
	if err != nil {
		log.Println(err)
		return
	}
	file, err = os.OpenFile(
		filepath.Join(dirPath, filename),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		os.ModePerm,
	)
	if err != nil {
		log.Println(err)
	}
	return
}

func GetFileInfosFromDir(dirPath string) (allPaths []os.FileInfo, err error) {
	dirname, err := CreateDirIfNotExist(dirPath, os.ModePerm)
	if err != nil {
		log.Println("error in GetFileInfosFromDir, in CreateDirIfNotExist", dirPath, err)
		return
	}

	allPaths, err = ioutil.ReadDir(dirname)
	if err != nil {
		log.Println("error in GetFileInfosFromDir, in ioutil.ReadDir", err)
		return
	}
	return allPaths, err
}
