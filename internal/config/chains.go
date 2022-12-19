package config

import (
	"context"
	"faucet-svc/internal/types"
	client2 "github.com/eteu-technologies/near-api-go/pkg/client"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/portto/solana-go-sdk/client"
	types3 "github.com/portto/solana-go-sdk/types"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Chainer interface {
	Chains(signers Signers) types.Chains
}

type chainer struct {
	once   comfig.Once
	getter kv.Getter
}

func NewChainer(getter kv.Getter) Chainer {
	return &chainer{getter: getter}
}

type evmChain struct {
	ID          string  `fig:"id,required"`
	Name        string  `fig:"name,required"`
	RPC         string  `fig:"rpc,required"`
	NativeToken string  `fig:"native_token,required"`
	Decimals    float64 `fig:"decimals,required"`
}

type solanaChain struct {
	ID       string  `fig:"id,required"`
	RPC      string  `fig:"rpc,required"`
	Decimals float64 `fig:"decimals,required"`
}

func (c *chainer) Evm(chains *types.Chains, signer types.EvmSigner) {

	var cfg struct {
		Chains []evmChain `fig:"chains,required"`
	}

	err := figure.
		Out(&cfg).
		With(figure.BaseHooks).
		From(kv.MustGetStringMap(c.getter, "evm")).
		Please()

	if err != nil {
		panic(errors.Wrap(err, "failed to figure out evm chains"))
	}

	validator := newDuplicationEvmChainValidator()
	for _, conf := range cfg.Chains {
		if err := validator.validate(conf); err != nil {
			panic(err)
		}

		cli, err := ethclient.Dial(conf.RPC)
		if err != nil {
			panic(errors.Wrap(err, "failed to dial rpc", logan.F{"chain_id": conf.ID}))
		}

		id, err := cli.ChainID(context.Background())
		if err != nil {
			panic(errors.Wrap(err, "chain has broken rpc", logan.F{"chain_id": conf.ID, "chain_rpc": conf.RPC}))
		}

		if conf.ID != id.String() {
			panic(errors.Errorf("%s has different rpc and conf chain id", conf.Name))
		}

		ch := types.NewEvmChain(cli, signer, conf.ID, conf.Name, conf.NativeToken, conf.RPC, conf.Decimals)
		chains.Set(ch.ID(), ch.Kind(), ch)
	}
	return
}

func (c *chainer) Solana(chains *types.Chains, signer types3.Account) {
	var cfg struct {
		Chains []solanaChain `fig:"chains,required"`
	}

	err := figure.
		Out(&cfg).
		From(kv.MustGetStringMap(c.getter, "solana")).
		Please()

	if err != nil {
		panic(errors.Wrap(err, "failed to figure out solana chains"))
	}

	validator := newDuplicationSolanaChainsValidator()
	for _, conf := range cfg.Chains {
		if err := validator.validate(conf); err != nil {
			panic(err)
		}
		cli := client.NewClient(conf.RPC)
		if _, err := cli.GetVersion(context.TODO()); err != nil {
			panic(errors.Errorf("failed to get solana chain version, chain %s", conf.ID))
		}

		ch := types.NewSolanaChain(cli, signer, conf.ID, "SOL", conf.RPC, conf.Decimals)
		chains.Set(ch.ID(), ch.Kind(), ch)
	}
	return
}

func (c *chainer) Near(chains *types.Chains, signer types.NearSigner) {
	var cfg struct {
		ID       string  `fig:"id,required"`
		RPC      string  `fig:"rpc,required"`
		Decimals float64 `fig:"decimals,required"`
	}

	err := figure.
		Out(&cfg).
		With(figure.BaseHooks).
		From(kv.MustGetStringMap(c.getter, "near")).
		Please()

	if err != nil {
		panic(errors.Wrap(err, "failed to figure out near chain"))
	}

	cli, err := client2.NewClient(cfg.RPC)
	if err != nil {
		panic(errors.Wrap(err, "failed to dial near rpc"))
	}
	ch := types.NewNearChain(&cli, signer, cfg.ID, cfg.RPC, "NEAR", cfg.Decimals)
	chains.Set(ch.ID(), ch.Kind(), ch)
	return
}

func (c *chainer) Chains(signers Signers) types.Chains {
	return c.once.Do(func() interface{} {
		chains := types.Chains{}
		c.Evm(&chains, signers.Evm())
		c.Solana(&chains, signers.Solana())
		c.Near(&chains, signers.Near())
		return chains
	}).(types.Chains)
}

type duplicationEvmChainsValidator struct {
	rpcMap   map[string]struct{}
	idsMap   map[string]struct{}
	namesMap map[string]struct{}
}

func newDuplicationEvmChainValidator() *duplicationEvmChainsValidator {
	return &duplicationEvmChainsValidator{
		rpcMap:   make(map[string]struct{}),
		idsMap:   make(map[string]struct{}),
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
	rpcMap map[string]struct{}
	idsMap map[string]struct{}
}

func newDuplicationSolanaChainsValidator() *duplicationSolanaChainsValidator {
	return &duplicationSolanaChainsValidator{
		rpcMap: make(map[string]struct{}),
		idsMap: make(map[string]struct{}),
	}
}

func (v *duplicationSolanaChainsValidator) validate(conf solanaChain) error {
	if _, ok := v.rpcMap[conf.RPC]; ok {
		return errors.Errorf("rpc %s url is duplicated", conf.RPC)
	}

	if _, ok := v.idsMap[conf.ID]; ok {
		return errors.Errorf("id %s is duplicated", conf.ID)
	}

	v.idsMap[conf.ID] = struct{}{}
	v.rpcMap[conf.RPC] = struct{}{}

	return nil
}
