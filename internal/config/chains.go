package config

import (
	"context"
	"faucet-svc/internal/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/portto/solana-go-sdk/client"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Chainer interface {
	Chains() Chains
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

type solanaChain struct {
	Name string `fig:"name,required"`
	RPC  string `fig:"rpc,required"`
}

func (c *chainer) Evm() types.EvmChains {

	var cfg struct {
		Chains []evmChain `fig:"chains,required"`
	}

	err := figure.
		Out(&cfg).
		With(figure.BaseHooks, evmChainHook).
		From(kv.MustGetStringMap(c.getter, "evm")).
		Please()

	if err != nil {
		panic(errors.Wrap(err, "failed to figure out evm chains"))
	}

	validator := newDuplicationEvmChainValidator()
	chs := types.EvmChains{}
	for _, conf := range cfg.Chains {
		if err := validator.validate(conf); err != nil {
			panic(err)
		}

		cli, err := ethclient.Dial(conf.RPC)
		if err != nil {
			panic(errors.Wrap(err, "failed to dial rpc", logan.F{"chain_id": conf.ID}))
		}

		if id, err := cli.ChainID(context.Background()); err == nil {
			if conf.ID != id.Int64() {
				panic(errors.Errorf("%s has different rpc and conf chain id", conf.Name))
			}
		}

		ch := types.NewEvmChain(cli, conf.ID, conf.Name, conf.RPC, conf.NativeToken)
		chs.Set(conf.ID, ch)
	}
	return chs
}

func (c *chainer) Solana() types.SolanaChains {
	var cfg struct {
		Chains []solanaChain `fig:"chains,required"`
	}

	err := figure.
		Out(&cfg).
		With(figure.BaseHooks, solanaChainHook).
		From(kv.MustGetStringMap(c.getter, "solana")).
		Please()

	if err != nil {
		panic(errors.Wrap(err, "failed to figure out solana chains"))
	}

	validator := newDuplicationSolanaChainsValidator()
	chs := types.SolanaChains{}
	for _, conf := range cfg.Chains {
		if err := validator.validate(conf); err != nil {
			panic(err)
		}
		cli := client.NewClient(conf.RPC)
		if _, err := cli.GetVersion(context.TODO()); err != nil {
			panic(errors.Errorf("failed to get solana chain version, chain %s", conf.Name))
		}

		ch := types.NewSolanaChain(cli, conf.Name, conf.RPC)
		chs.Set(conf.Name, ch)
	}
	return chs
}

func (c *chainer) Chains() Chains {
	return c.once.Do(func() interface{} {
		evmChains := c.Evm()
		solanaChains := c.Solana()
		return NewChains(evmChains, solanaChains)
	}).(Chains)
}

type duplicationEvmChainsValidator struct {
	rpcMap   map[string]struct{}
	idsMap   map[int64]struct{}
	namesMap map[string]struct{}
}

func newDuplicationEvmChainValidator() *duplicationEvmChainsValidator {
	return &duplicationEvmChainsValidator{
		rpcMap:   make(map[string]struct{}),
		idsMap:   make(map[int64]struct{}),
		namesMap: make(map[string]struct{}),
	}
}

func (v *duplicationEvmChainsValidator) validate(conf evmChain) error {
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

type duplicationSolanaChainsValidator struct {
	rpcMap   map[string]struct{}
	namesMap map[string]struct{}
}

func newDuplicationSolanaChainsValidator() *duplicationSolanaChainsValidator {
	return &duplicationSolanaChainsValidator{
		rpcMap:   make(map[string]struct{}),
		namesMap: make(map[string]struct{}),
	}
}

func (v *duplicationSolanaChainsValidator) validate(conf solanaChain) error {
	if _, ok := v.rpcMap[conf.RPC]; ok {
		return errors.Errorf("rpc %s url is duplicated", conf.RPC)
	}

	if _, ok := v.namesMap[conf.Name]; ok {
		return errors.Errorf("name %s is duplicated", conf.Name)
	}

	v.namesMap[conf.Name] = struct{}{}
	v.rpcMap[conf.RPC] = struct{}{}

	return nil
}

type Chains interface {
	Evm() types.EvmChains
	Solana() types.SolanaChains
}

type chains struct {
	evm    types.EvmChains
	solana types.SolanaChains
}

func NewChains(evm types.EvmChains, solana types.SolanaChains) Chains {
	return &chains{
		evm:    evm,
		solana: solana,
	}
}

func (c *chains) Evm() types.EvmChains {
	return c.evm
}

func (c *chains) Solana() types.SolanaChains {
	return c.solana
}
