package config

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"reflect"
)

var evmChainHook = figure.Hooks{
	"[]config.evmChain": func(value interface{}) (reflect.Value, error) {

		if value == nil {
			return reflect.Value{}, nil
		}

		switch s := value.(type) {
		case []interface{}:
			chains := make([]evmChain, 0, len(s))
			for _, val := range s {
				value := val.(map[interface{}]interface{})
				params := make(map[string]interface{})
				for k, v := range value {
					params[k.(string)] = v
				}
				var ch evmChain
				err := figure.
					Out(&ch).
					With(figure.BaseHooks).
					From(params).
					Please()

				if err != nil {
					return reflect.Value{}, errors.Wrap(err, "failed to figure out chain")
				}

				chains = append(chains, ch)
			}
			return reflect.ValueOf(chains), nil
		default:
			return reflect.Value{}, errors.New("unexpected type while figure []evmChainCfg")
		}
	},
}

var signerHook = figure.Hooks{
	"*ecdsa.PrivateKey": func(value interface{}) (reflect.Value, error) {
		switch v := value.(type) {
		case string:
			privKey, err := crypto.HexToECDSA(v)
			if err != nil {
				return reflect.Value{}, errors.Wrap(err, "invalid hex private key")
			}
			return reflect.ValueOf(privKey), nil
		default:
			return reflect.Value{}, fmt.Errorf("unsupported conversion from %T", value)
		}
	},
}
