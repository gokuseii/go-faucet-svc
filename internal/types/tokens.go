package types

type EvmToken interface {
	Name() string
	Symbol() string
	Address() string
	Kind() string
	Chains() []int64
}

type evmToken struct {
	name    string
	symbol  string
	address string
	kind    string
	chains  []int64
}

func NewEvmToken(name, symbol, address, kind string, chains []int64) EvmToken {
	return &evmToken{
		name:    name,
		symbol:  symbol,
		address: address,
		kind:    kind,
		chains:  chains,
	}
}

func (t *evmToken) Name() string {
	return t.name
}

func (t *evmToken) Symbol() string {
	return t.symbol
}

func (t *evmToken) Address() string {
	return t.address
}

func (t *evmToken) Kind() string {
	return t.address
}

func (t *evmToken) Chains() []int64 {
	return t.chains
}

type EvmTokens map[string]EvmToken

func (tokens EvmTokens) Get(key string) (EvmToken, bool) {
	val, ok := tokens[key]
	return val, ok
}

func (tokens EvmTokens) Set(key string, val EvmToken) bool {
	if _, ok := tokens[key]; ok {
		return false
	}

	tokens[key] = val
	return true
}