package helpers

import (
	"math"
	"math/big"
)

func IsLessOrEq(x, y *big.Int) bool {
	if cmp := x.Cmp(y); cmp <= 0 {
		return true
	}
	return false
}

func ToHumanBalance(amount *big.Int, decimals float64) float64 {
	x, _ := new(big.Float).SetInt(amount).Float64()
	return x / math.Pow(10, decimals)
}
