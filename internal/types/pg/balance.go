package pg

type Balance struct {
	ID           uint64  `db:"id"`
	UserId       string  `db:"user_id"`
	ChainId      string  `db:"chain_id"`
	ChainType    string  `db:"chain_type"`
	TokenAddress string  `db:"token_address"`
	Amount       float64 `db:"amount"`
}

func NewBalance(userId, chainId, chainType string, amount float64, tokenAddress *string) Balance {
	tknAddr := ""
	if tokenAddress != nil {
		tknAddr = *tokenAddress
	}
	return Balance{
		UserId:       userId,
		ChainId:      chainId,
		ChainType:    chainType,
		TokenAddress: tknAddr,
		Amount:       amount,
	}
}
