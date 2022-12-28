package requests

import (
	"encoding/json"
	"faucet-svc/resources"
	validation "github.com/go-ozzo/ozzo-validation"
	"net/http"

	"gitlab.com/distributed_lab/logan/v3/errors"
)

type CreateSendSolanaRequest struct {
	Data resources.SendSolana
}

func NewCreateSendSolanaRequest(r *http.Request) (CreateSendSolanaRequest, error) {
	var request CreateSendSolanaRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return request, errors.Wrap(err, "failed to unmarshal")
	}

	return request, request.validate()
}

func (r *CreateSendSolanaRequest) validate() error {
	return validation.Errors{
		"/data/":                  validation.Validate(&r.Data, validation.Required),
		"/data/id":                validation.Validate(&r.Data.ID, validation.Required),
		"/data/attributes/to":     validation.Validate(&r.Data.Attributes.To, validation.Required, validation.Length(44, 44)),
		"/data/attributes/amount": validation.Validate(&r.Data.Attributes.Amount, validation.Required, validation.Min(uint(1))),
	}.Filter()
}
