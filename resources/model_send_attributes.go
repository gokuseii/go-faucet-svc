/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "math/big"

type SendAttributes struct {
	Amount       big.Int `json:"amount"`
	To           string  `json:"to"`
	TokenAddress *string `json:"token_address,omitempty"`
}
