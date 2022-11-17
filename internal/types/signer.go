package types

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type EvmSigner interface {
	Address() common.Address
	PrivKey() *ecdsa.PrivateKey
}

type evmSigner struct {
	address common.Address
	privKey *ecdsa.PrivateKey
}

func NewEvmSigner(privKey *ecdsa.PrivateKey) EvmSigner {
	publicKeyECDSA, ok := privKey.Public().(*ecdsa.PublicKey)
	if !ok {
		panic("error casting public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return &evmSigner{
		address: address,
		privKey: privKey,
	}
}

func (s *evmSigner) Address() common.Address {
	return s.address
}

func (s *evmSigner) PrivKey() *ecdsa.PrivateKey {
	return s.privKey
}
