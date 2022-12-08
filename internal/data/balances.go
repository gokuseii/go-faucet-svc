package data

import (
	"faucet-svc/internal/types/pg"
)

type BalancesQ interface {
	New() BalancesQ
	Create(balance *pg.Balance) error
	Get() (*pg.Balance, error)
	Update(balance pg.Balance, amount float64) error
	FilterByUserID(userId string) BalancesQ
	FilterByChainID(chainId string) BalancesQ
	FilterByChainType(chainType string) BalancesQ
	FilterByTokenAddress(tokenAddress string) BalancesQ
}
