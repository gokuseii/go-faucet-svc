package config

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"reflect"
)

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
