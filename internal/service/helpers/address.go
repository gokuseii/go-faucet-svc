package helpers

import (
	"faucet-svc/internal/config"
)

func GetSignerAddress(chainType string, signers config.Signers) (signerAddress string) {
	switch chainType {
	case "evm":
		signerAddress = signers.Evm().Address().String()
	case "solana":
		signerAddress = signers.Solana().PublicKey.String()
	case "near":
		signerAddress = signers.Near().ID()
	}
	return
}
