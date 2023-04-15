package wallet

import (
	"encoding/hex"
	"errors"
	libcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/mearaj/protonet/alog"
	"github.com/mearaj/protonet/internal/common"
	"github.com/mearaj/protonet/internal/db"
	"github.com/mearaj/protonet/internal/evm"
	_ "github.com/mearaj/protonet/internal/evm"
	"github.com/mearaj/protonet/internal/model"
	"github.com/mearaj/protonet/utils"
	"strings"
)

var GlobalWallet = New()
var _ Manager = &Wallet{}

// const defaultChainMediator = "https://mainnet.infura.io"
// const defaultChainMediator = "https://ethereum.publicnode.com"
const defaultChainMediator = "https://rpc.ntity.io"

type Manager interface {
	CreateAccount(privateKeyHex string) error
	AutoCreateAccount() error
	Connections() []*evm.RPCClients
}

type Wallet struct {
	connections []*evm.RPCClients
	*db.ProtoDB
	FavoriteChains utils.Map[string, struct{}]
	FavoriteRPCs   utils.Map[string, struct{}]
}

var _ Manager = &Wallet{}

func New() *Wallet {
	wa := &Wallet{}
	wa.connections = evm.GetAllRPCClients()
	wa.ProtoDB = db.New()
	wa.FavoriteChains = utils.NewMap[string, struct{}]()
	wa.FavoriteRPCs = utils.NewMap[string, struct{}]()
	return wa
}

func (w *Wallet) CreateAccount(pvtKeyHex string) (err error) {
	if strings.TrimSpace(pvtKeyHex) == "" {
		err = errors.New("private key is empty")
		return
	}
	if !w.ProtoDB.IsOpen() {
		return db.ErrDBNotOpened
	}
	defer func() {
		if err != nil {
			alog.Logger().Errorln(err)
		}
	}()
	algo := libcrypto.ECDSA
	pvtKey, err := common.GetPrivateKeyFromStr(pvtKeyHex, algo)
	if err != nil {
		return
	}
	pubKeyBytes, err := pvtKey.GetPublic().Raw()
	if err != nil {
		return
	}
	publicKeyStr := hex.EncodeToString(pubKeyBytes)
	pvtKeyBytes, err := hex.DecodeString(pvtKeyHex)
	if err != nil {
		return
	}
	pvtKeyEnc := hex.EncodeToString(pvtKeyBytes)
	ethAddress, err := common.GetEthAddress(pvtKeyEnc)
	if err != nil {
		return err
	}
	account := model.Account{
		PrivateKey: pvtKeyEnc,
		PublicKey:  publicKeyStr,
		EthAddress: ethAddress,
	}
	err = w.ProtoDB.AddUpdateAccount(&account)
	return err
}

func (w *Wallet) AutoCreateAccount() (err error) {
	defer func() {
		if err != nil {
			alog.Logger().Errorln(err)
		}
	}()
	_, pvtKeyStr, _, publicKeyStr, _, err := common.CreatePrivateKey(libcrypto.ECDSA)
	if err != nil {
		return err
	}
	ethAddress, err := common.GetEthAddress(pvtKeyStr)
	if err != nil {
		return
	}
	pvtKeyBytes, err := hex.DecodeString(pvtKeyStr)
	if err != nil {
		return
	}
	pvtKeyEnc := hex.EncodeToString(pvtKeyBytes)
	account := model.Account{
		PrivateKey: pvtKeyEnc,
		PublicKey:  publicKeyStr,
		EthAddress: ethAddress,
	}
	err = w.ProtoDB.AddUpdateAccount(&account)
	return err
}

func (w *Wallet) Connections() []*evm.RPCClients {
	return w.connections
}
