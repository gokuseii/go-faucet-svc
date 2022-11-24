package near

import (
	"faucet-svc/internal/types"
)

type AccountInfo struct {
	Amount        types.Amount `json:"amount"`
	BlockHash     string       `json:"block_hash"`
	BlockHeight   uint         `json:"block_height"`
	CodeHash      string       `json:"code_hash"`
	Locked        types.Amount `json:"locked"`
	StoragePaidAt uint         `json:"storage_paid_at"`
	StorageUsage  uint         `json:"storage_usage"`
}
