/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "math/big"

type SendEvmAttributes struct {
	Amount       big.Int `json:"amount"`
	ChainId      int64   `json:"chain_id"`
	Symbol       string  `json:"symbol"`
	To           string  `json:"to"`
	TokenAddress *string `json:"token_address,omitempty"`
}
