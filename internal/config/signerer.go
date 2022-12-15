package config

import (
	"crypto/ecdsa"
	"faucet-svc/internal/types"
	types3 "github.com/eteu-technologies/near-api-go/pkg/types"
	types2 "github.com/portto/solana-go-sdk/types"
	"gitlab.com/distributed_lab/figure/v3"
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
		With(signerHook).
		From(kv.MustGetStringMap(s.getter, "evm")).
		Please()

	if err != nil {
		panic(errors.Wrap(err, "failed to figure out evm signer"))
	}

	signer, err := types.NewEvmSigner(cfg.PrivKey)
	if err != nil {
		panic(errors.Wrap(err, "failed to get evm signer"))
	}

	return signer
}

func (s *signerer) Solana() types2.Account {
	var cfg struct {
		PrivKey string `fig:"signer,required"`
	}

	err := figure.
		Out(&cfg).
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

func (s *signerer) Near() types.NearSigner {
	var cfg struct {
		ID      types3.AccountID `fig:"signer_id,required"`
		PrivKey string           `fig:"signer,required"`
	}

	err := figure.
		Out(&cfg).
		From(kv.MustGetStringMap(s.getter, "near")).
		Please()

	if err != nil {
		panic(errors.Wrap(err, "failed to figure out near signer"))
	}

	signer, err := types.NewNearSigner(cfg.ID, cfg.PrivKey)
	if err != nil {
		panic(errors.Wrap(err, "failed to get near signer"))
	}
	return signer
}

func (s *signerer) Signers() Signers {
	return s.once.Do(func() interface{} {
		evmSigner := s.Evm()
		solanaSigner := s.Solana()
		near := s.Near()
		return NewSigners(evmSigner, solanaSigner, near)
	}).(Signers)
}

type Signers interface {
	Evm() types.EvmSigner
	Solana() types2.Account
	Near() types.NearSigner
}

type signers struct {
	evm    types.EvmSigner
	solana types2.Account
	near   types.NearSigner
}

func NewSigners(evm types.EvmSigner, solana types2.Account, near types.NearSigner) Signers {
	return &signers{
		evm:    evm,
		solana: solana,
		near:   near,
	}
}

func (s *signers) Evm() types.EvmSigner {
	return s.evm
}

func (s *signers) Solana() types2.Account {
	return s.solana
}

func (s *signers) Near() types.NearSigner {
	return s.near
}
