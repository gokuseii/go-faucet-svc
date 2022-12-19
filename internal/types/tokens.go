package types

type EvmToken interface {
	Name() string
	Symbol() string
	Address() string
	Kind() string
	Chains() []string
	Decimals() float64
}

type evmToken struct {
	name     string
	symbol   string
	address  string
	kind     string
	chains   []string
	decimals float64
}

func NewEvmToken(name, symbol, address, kind string, chains []string, decimals float64) EvmToken {
	return &evmToken{
		name:     name,
		symbol:   symbol,
		address:  address,
		kind:     kind,
		chains:   chains,
		decimals: decimals,
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
	return t.kind
}

func (t *evmToken) Chains() []string {
	return t.chains
}

func (t *evmToken) Decimals() float64 {
	return t.decimals
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
