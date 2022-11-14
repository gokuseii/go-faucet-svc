package config

import (
	"context"
	"faucet-svc/internal/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Chainer interface {
	EvmChains() types.EvmChains
}

type chainer struct {
	once   comfig.Once
	getter kv.Getter
}

func NewChainer(getter kv.Getter) Chainer {
	return &chainer{getter: getter}
}

type evmChain struct {
	ID          int64  `fig:"chain_id,required"`
	Name        string `fig:"name,required"`
	RPC         string `fig:"rpc,required"`
	NativeToken string `fig:"native_token,required"`
}

func (c *chainer) EvmChains() types.EvmChains {
	return c.once.Do(func() interface{} {

		var cfg struct {
			Chains []evmChain `fig:"chains,required"`
		}

		err := figure.
			Out(&cfg).
			With(figure.BaseHooks, evmChainHook).
			From(kv.MustGetStringMap(c.getter, "evm")).
			Please()

		if err != nil {
			panic(errors.Wrap(err, "failed to figure out chains"))
		}

		validator := newDuplicationChainValidator()
		chains := types.EvmChains{}
		for _, conf := range cfg.Chains {
			if err := validator.validate(conf); err != nil {
				panic(err)
			}

			client, err := ethclient.Dial(conf.RPC)
			if err != nil {
				panic(errors.Wrap(err, "failed to dial rpc", logan.F{"chain_id": conf.ID}))
			}

			if id, err := client.ChainID(context.Background()); err == nil {
				if conf.ID != id.Int64() {
					panic(errors.Errorf("%s has different rpc and conf chain id", conf.Name))
				}
			}

			ch := types.NewEvmChain(client, conf.ID, conf.Name, conf.RPC, conf.NativeToken)
			chains.Set(conf.ID, ch)
		}

		return chains
	}).(types.EvmChains)
}

type duplicationChainsValidator struct {
	rpcMap   map[string]struct{}
	idsMap   map[int64]struct{}
	namesMap map[string]struct{}
}

func newDuplicationChainValidator() *duplicationChainsValidator {
	return &duplicationChainsValidator{
		rpcMap:   make(map[string]struct{}),
		idsMap:   make(map[int64]struct{}),
		namesMap: make(map[string]struct{}),
	}
}

func (v *duplicationChainsValidator) validate(conf evmChain) error {
	if _, ok := v.rpcMap[conf.RPC]; ok {
		return errors.Errorf("rpc %s url is duplicated", conf.RPC)
	}

	if _, ok := v.idsMap[conf.ID]; ok {
		return errors.Errorf("chain_id %d is duplicated", conf.ID)
	}

	if _, ok := v.namesMap[conf.Name]; ok {
		return errors.Errorf("name %s is duplicated", conf.Name)
	}

	v.idsMap[conf.ID] = struct{}{}
	v.namesMap[conf.Name] = struct{}{}
	v.rpcMap[conf.RPC] = struct{}{}

	return nil

}
