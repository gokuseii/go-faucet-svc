package types

type AccountInfo struct {
	Amount        Amount `json:"amount"`
	BlockHash     string `json:"block_hash"`
	BlockHeight   uint   `json:"block_height"`
	CodeHash      string `json:"code_hash"`
	Locked        Amount `json:"locked"`
	StoragePaidAt uint   `json:"storage_paid_at"`
	StorageUsage  uint   `json:"storage_usage"`
}
