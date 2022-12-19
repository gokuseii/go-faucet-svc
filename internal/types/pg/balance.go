package pg

type Balance struct {
	ID           uint64  `db:"id"`
	UserId       string  `db:"user_id"`
	ChainId      string  `db:"chain_id"`
	ChainType    string  `db:"chain_type"`
	TokenAddress string  `db:"token_address"`
	Amount       float64 `db:"amount"`
}
