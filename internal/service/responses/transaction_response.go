package responses

import (
	"faucet-svc/resources"
)

type TransactionResponse struct {
	Data resources.Transaction `json:"data"`
}

func NewTransactionResponse(id string) TransactionResponse {
	return TransactionResponse{
		Data: resources.Transaction{
			Id:   id,
			Type: "transaction",
		},
	}
}
