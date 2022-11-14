package config

import (
	"crypto/ecdsa"
	"faucet-svc/internal/types"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Signerer interface {
	Signer() types.Signer
}

type signerer struct {
	once   comfig.Once
	getter kv.Getter
}

func NewSignerer(getter kv.Getter) Signerer {
	return &signerer{
		getter: getter,
	}
}

func (s *signerer) Signer() types.Signer {
	return s.once.Do(func() interface{} {
		var cfg struct {
			PrivKey *ecdsa.PrivateKey `fig:"signer,required"`
		}

		err := figure.
			Out(&cfg).
			With(figure.BaseHooks, signerHook).
			From(kv.MustGetStringMap(s.getter, "evm")).
			Please()

		if err != nil {
			panic(errors.Wrap(err, "failed to figure out signer"))
		}

		return types.NewSigner(cfg.PrivKey)
	}).(types.Signer)
}
