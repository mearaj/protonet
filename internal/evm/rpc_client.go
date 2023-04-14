package evm

import (
	"context"
	"errors"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mearaj/protonet/internal/db"
	"math"
	"math/big"
	"sync"
)

var ErrAlreadyConnected = errors.New("already connected")
var ErrNotConnected = errors.New("not connected")

type RPCClient struct {
	*RPCClients
	RPC
	client      *ethclient.Client
	clientMutex sync.RWMutex
	apiKey      string
	apiKeyMutex sync.RWMutex
}

func (c *RPCClient) getClient() *ethclient.Client {
	c.clientMutex.RLock()
	defer c.clientMutex.RUnlock()
	return c.client
}

func (c *RPCClient) setClient(cl *ethclient.Client) {
	c.clientMutex.Lock()
	defer c.clientMutex.Unlock()
	c.client = cl
}

func (c *RPCClient) APIKey() string {
	c.apiKeyMutex.RLock()
	defer c.apiKeyMutex.RUnlock()
	return c.apiKey
}

func (c *RPCClient) SetAPIKey(apiKey string) {
	c.apiKeyMutex.Lock()
	defer c.apiKeyMutex.Unlock()
	c.apiKey = apiKey
}

func (c *RPCClient) IsConnected() bool {
	return c.getClient() != nil
}

func (c *RPCClient) Connect() (err error) {
	if c.IsConnected() {
		return ErrAlreadyConnected
	}
	url, err := c.RPC.GetURL(c.apiKey)
	if err != nil {
		return err
	}
	conn, err := ethclient.Dial(url)
	if err != nil {
		if conn != nil {
			conn.Close()
			c.setClient(nil)
		}
		return err
	}
	c.setClient(conn)
	return nil
}

func (c *RPCClient) Disconnect() error {
	if !c.IsConnected() {
		return ErrNotConnected
	}
	client := c.getClient()
	if client != nil {
		client.Close()
	}
	c.setClient(nil)
	return nil
}
func (c *RPCClient) ShowBalance(acc db.Account) (string, error) {
	if !c.IsConnected() {
		return "", ErrNotConnected
	}
	client := c.getClient()
	if client == nil {
		return "", ErrNotConnected
	}
	bal, err := client.BalanceAt(context.Background(), common2.HexToAddress(acc.EthAddress), nil)
	if err != nil {
		return "", err
	}
	fBalance := new(big.Float)
	fBalance.SetString(bal.String())
	val := new(big.Float).Quo(fBalance, big.NewFloat(math.Pow10(18)))
	return val.String(), nil
}

// RpcClients wraps multiple RpcClients for a Chain
type RPCClients struct {
	Chain      Chain
	RPCClients []*RPCClient
}

func (c *RPCClients) IsConnected() bool {
	for _, state := range c.RPCClients {
		if state.IsConnected() {
			return true
		}
	}
	return false
}

func GetAllRPCClients() []*RPCClients {
	rpcClients := make([]*RPCClients, len(ChainsSlice()))
	for i, ch := range ChainsSlice() {
		rpcClient := RPCClients{
			Chain:      ch,
			RPCClients: make([]*RPCClient, 0),
		}
		for _, rpc := range ch.RPC {
			rpcClient.RPCClients = append(rpcClient.RPCClients, &RPCClient{RPC: rpc, client: nil, RPCClients: &rpcClient})
		}
		rpcClients[i] = &rpcClient
	}
	return rpcClients
}
