package config

import (
	"faucet-svc/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"strings"
)

type Tokens interface {
	EvmTokens() types.EvmTokens
}

type tokens struct {
	once   comfig.Once
	getter kv.Getter
}

func NewTokens(getter kv.Getter) Tokens {
	return &tokens{getter: getter}
}

type token struct {
	Name     string   `fig:"name,required"`
	Symbol   string   `fig:"symbol,required"`
	Address  string   `fig:"address,required"`
	Kind     string   `fig:"type,required"`
	Chains   []string `fig:"chains,required"`
	Decimals float64  `fig:"decimals,required"`
}

func (c *tokens) EvmTokens() types.EvmTokens {

	var cfg struct {
		Tokens []token `fig:"external_tokens,required"`
	}

	err := figure.
		Out(&cfg).
		From(kv.MustGetStringMap(c.getter, "evm")).
		Please()

	if err != nil {
		panic(errors.Wrap(err, "failed to figure out evm tokens"))
	}

	tkns := types.EvmTokens{}
	for _, conf := range cfg.Tokens {
		if _, ok := tkns.Get(conf.Address); ok {
			panic(errors.Errorf("Token address duplicated %s", conf.Address))
		}

		if !common.IsHexAddress(conf.Address) {
			panic(errors.Errorf("Invalid token address %s", conf.Address))
		}

		if conf.Kind != "ERC20" {
			panic(errors.Errorf("%s not supported contract type", conf.Kind))
		}

		if len(conf.Chains) == 0 {
			panic(errors.Errorf("Not found supported chains %s", conf.Address))
		}

		tk := types.NewEvmToken(conf.Name, conf.Symbol, conf.Address, conf.Kind, conf.Chains, conf.Decimals)
		tkns.Set(strings.ToLower(conf.Address), tk)
	}
	return tkns
}
