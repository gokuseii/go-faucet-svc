package types

import (
	"crypto/ecdsa"
	"errors"
	"github.com/eteu-technologies/near-api-go/pkg/types/key"
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

func NewEvmSigner(privKey *ecdsa.PrivateKey) (EvmSigner, error) {
	publicKeyECDSA, ok := privKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	signer := evmSigner{
		address: address,
		privKey: privKey,
	}
	return &signer, nil
}

func (s *evmSigner) Address() common.Address {
	return s.address
}

func (s *evmSigner) PrivKey() *ecdsa.PrivateKey {
	return s.privKey
}

type NearSigner interface {
	ID() string
	KeyPair() key.KeyPair
}

type nearSigner struct {
	id      string
	keyPair key.KeyPair
}

func NewNearSigner(id, privKey string) (NearSigner, error) {
	keyPair, err := key.NewBase58KeyPair(privKey)
	if err != nil {
		return nil, err
	}

	signer := nearSigner{
		id:      id,
		keyPair: keyPair,
	}

	return &signer, nil
}

func (s *nearSigner) ID() string {
	return s.id
}

func (s *nearSigner) KeyPair() key.KeyPair {
	return s.keyPair
}
