package types

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Signer interface {
	Address() common.Address
	PrivKey() *ecdsa.PrivateKey
}

type signer struct {
	address common.Address
	privKey *ecdsa.PrivateKey
}

func NewSigner(privKey *ecdsa.PrivateKey) Signer {

	publicKeyECDSA, ok := privKey.Public().(*ecdsa.PublicKey)
	if !ok {
		panic("error casting public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return &signer{
		address: address,
		privKey: privKey,
	}
}

func (s *signer) Address() common.Address {
	return s.address
}

func (s *signer) PrivKey() *ecdsa.PrivateKey {
	return s.privKey
}
