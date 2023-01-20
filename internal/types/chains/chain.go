package chains

import (
	"math/big"
)

type Chain interface {
	ID() string
	Name() string
	Kind() string
	NativeToken() string
	Decimals() float64
	GetBalance(address string, tokenAddress *string) (*big.Int, error)
	Send(to string, amount *big.Int, tokenAddress *string) (txHash string, err error)
}

type Chains map[string]Chain

func (chains Chains) Get(id, kind string) (Chain, bool) {
	val, ok := chains[kind+":"+id]
	return val, ok
}

func (chains Chains) Set(id, kind string, val Chain) bool {
	if _, ok := chains[kind+":"+id]; ok {
		return false
	}

	chains[kind+":"+id] = val
	return true
}
