package requests

import (
	"encoding/json"
	"faucet-svc/resources"
	validation "github.com/go-ozzo/ozzo-validation"
	"net/http"

	"gitlab.com/distributed_lab/logan/v3/errors"
)

type CreateSendEvmRequest struct {
	Data resources.SendEvm
}

func NewCreateSendEvmRequest(r *http.Request) (CreateSendEvmRequest, error) {
	var request CreateSendEvmRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return request, errors.Wrap(err, "failed to unmarshal")
	}

	return request, request.validate()
}

func (r *CreateSendEvmRequest) validate() error {

	return validation.Errors{
		"/data/":                  validation.Validate(&r.Data, validation.Required),
		"/data/id":                validation.Validate(&r.Data.ID, validation.Required),
		"/data/attributes/to":     validation.Validate(&r.Data.Attributes.To, validation.Required, validation.Length(40, 42)),
		"/data/attributes/symbol": validation.Validate(&r.Data.Attributes.Symbol, validation.Required),
		"/data/attributes/amount": validation.Validate(&r.Data.Attributes.Amount, validation.Required),
	}.Filter()
}
