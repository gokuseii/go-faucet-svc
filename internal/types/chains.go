package types

import (
	client2 "github.com/eteu-technologies/near-api-go/pkg/client"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/portto/solana-go-sdk/client"
)

type EvmChain interface {
	Client() *ethclient.Client
	ID() int64
	Name() string
	RPC() string
	NativeToken() string
}

type evmChain struct {
	client      *ethclient.Client
	id          int64
	name        string
	rpc         string
	nativeToken string
}

func NewEvmChain(client *ethclient.Client, id int64, name string, rpc string, nativeToken string) EvmChain {
	return &evmChain{
		client:      client,
		id:          id,
		name:        name,
		rpc:         rpc,
		nativeToken: nativeToken,
	}
}

func (chain *evmChain) ID() int64 {
	return chain.id
}

func (chain *evmChain) Name() string {
	return chain.name
}

func (chain *evmChain) RPC() string {
	return chain.rpc
}

func (chain *evmChain) NativeToken() string {
	return chain.nativeToken
}

func (chain *evmChain) Client() *ethclient.Client {
	return chain.client
}

type EvmChains map[int64]EvmChain

func (chains EvmChains) Get(key int64) (EvmChain, bool) {
	val, ok := chains[key]
	return val, ok
}

func (chains EvmChains) Set(key int64, val EvmChain) bool {
	if _, ok := chains[key]; ok {
		return false
	}

	chains[key] = val
	return true
}

type SolanaChain interface {
	Client() *client.Client
	Name() string
	RPC() string
}

type solanaChain struct {
	client *client.Client
	name   string
	rpc    string
}

func NewSolanaChain(client *client.Client, name string, rpc string) SolanaChain {
	return &solanaChain{
		client: client,
		name:   name,
		rpc:    rpc,
	}
}

func (chain *solanaChain) Client() *client.Client {
	return chain.client
}

func (chain *solanaChain) Name() string {
	return chain.name
}

func (chain *solanaChain) RPC() string {
	return chain.rpc
}

type SolanaChains map[string]SolanaChain

func (chains SolanaChains) Get(key string) (SolanaChain, bool) {
	val, ok := chains[key]
	return val, ok
}

func (chains SolanaChains) Set(key string, val SolanaChain) bool {
	if _, ok := chains[key]; ok {
		return false
	}

	chains[key] = val
	return true
}

type NearChain interface {
	Client() *client2.Client
	RPC() string
}

type nearChain struct {
	client *client2.Client
	rpc    string
}

func NewNearChain(client *client2.Client, rpc string) NearChain {
	return &nearChain{
		client: client,
		rpc:    rpc,
	}
}

func (chain *nearChain) Client() *client2.Client {
	return chain.client
}

func (chain *nearChain) RPC() string {
	return chain.rpc
}
