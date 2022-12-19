package helpers

import (
	"errors"
	"faucet-svc/doorman"
	"faucet-svc/internal/types/pg"
	"math"
	"math/big"
	"net/http"
)

func IsLessOrEq(x, y big.Int) bool {
	if cmp := x.Cmp(&y); cmp <= 0 {
		return true
	}
	return false
}

func ToHumanBalance(amount *big.Int, decimals float64) float64 {
	x, _ := new(big.Float).SetInt(amount).Float64()
	return x / math.Pow(10, decimals)
}

func UpdateBalance(r *http.Request, chainId, chainType string, amount float64, tokenAddress *string) error {
	userId, ok := doorman.GetHeader(r, "User-Id")
	if !ok {
		return errors.New("failed to get user id from header")
	}

	tknAddr := ""
	if tokenAddress != nil {
		tknAddr = *tokenAddress
	}

	balanceQ := BalancesQ(r)
	balance, err := balanceQ.FilterByUserID(userId).
		FilterByChainID(chainId).
		FilterByChainType(chainType).
		FilterByTokenAddress(tknAddr).
		Get()
	if err != nil {
		return err
	}

	if balance == nil {
		balance = &pg.Balance{
			UserId:       userId,
			ChainId:      chainId,
			ChainType:    chainType,
			TokenAddress: tknAddr,
			Amount:       amount,
		}
		err := balanceQ.Create(balance)
		return err
	}

	err = balanceQ.Update(*balance, amount)
	return err
}
