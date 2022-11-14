/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "math/big"

type SendEvmTokenAttributes struct {
	Amount  big.Int `json:"amount"`
	ChainId int64   `json:"chain_id"`
	Symbol  *string `json:"symbol,omitempty"`
	To      string  `json:"to"`
}
