package helpers

import (
	"math"
	"math/big"
)

func ToHumanBalance(amount *big.Int, decimals float64) float64 {
	x, _ := new(big.Float).SetInt(amount).Float64()
	return x / math.Pow(10, decimals)
}
