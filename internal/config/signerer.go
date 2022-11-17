package config

import (
	"crypto/ecdsa"
	"faucet-svc/internal/types"
	types2 "github.com/portto/solana-go-sdk/types"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Signerer interface {
	Signers() Signers
}

type signerer struct {
	once   comfig.Once
	getter kv.Getter
}

func NewSignerer(getter kv.Getter) Signerer {
	return &signerer{getter: getter}
}

func (s *signerer) Evm() types.EvmSigner {
	var cfg struct {
		PrivKey *ecdsa.PrivateKey `fig:"signer,required"`
	}

	err := figure.
		Out(&cfg).
		With(figure.BaseHooks, signerHook).
		From(kv.MustGetStringMap(s.getter, "evm")).
		Please()

	if err != nil {
		panic(errors.Wrap(err, "failed to figure out evm signer"))
	}

	return types.NewEvmSigner(cfg.PrivKey)
}

func (s *signerer) Solana() types2.Account {
	var cfg struct {
		PrivKey string `fig:"signer,required"`
	}

	err := figure.
		Out(&cfg).
		With(figure.BaseHooks).
		From(kv.MustGetStringMap(s.getter, "solana")).
		Please()

	if err != nil {
		panic(errors.Wrap(err, "failed to figure out solana signer"))
	}

	account, err := types2.AccountFromBase58(cfg.PrivKey)
	if err != nil {
		panic(errors.Wrap(err, "failed to get solana account from private key"))
	}

	return account
}

func (s *signerer) Signers() Signers {
	return s.once.Do(func() interface{} {
		evmSigner := s.Evm()
		solanaSigner := s.Solana()
		return NewSigners(evmSigner, solanaSigner)
	}).(Signers)
}

type Signers interface {
	Evm() types.EvmSigner
	Solana() types2.Account
}

type signers struct {
	evm    types.EvmSigner
	solana types2.Account
}

func NewSigners(evm types.EvmSigner, solana types2.Account) Signers {
	return &signers{
		evm:    evm,
		solana: solana,
	}
}
func (c *signers) Evm() types.EvmSigner {
	return c.evm
}

func (c *signers) Solana() types2.Account {
	return c.solana
}
