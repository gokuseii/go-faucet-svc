package types

import "github.com/ethereum/go-ethereum/ethclient"

type EvmChain interface {
	ID() int64
	Name() string
	RPC() string
	NativeToken() string
	Client() *ethclient.Client
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
