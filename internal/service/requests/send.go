package requests

import (
	"encoding/json"
	"faucet-svc/resources"
	validation "github.com/go-ozzo/ozzo-validation"
	"net/http"

	"gitlab.com/distributed_lab/logan/v3/errors"
)

type CreateSendRequest struct {
	Data resources.Send
}

func NewCreateSendRequest(r *http.Request) (CreateSendRequest, error) {
	var request CreateSendRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return request, errors.Wrap(err, "failed to unmarshal")
	}

	return request, request.validate()
}

func (r *CreateSendRequest) validate() error {
	return validation.Errors{
		"/data/":                  validation.Validate(&r.Data, validation.Required),
		"/data/id":                validation.Validate(&r.Data.ID, validation.Required),
		"/data/type":              validation.Validate(&r.Data.Type, validation.Required),
		"/data/attributes/to":     validation.Validate(&r.Data.Attributes.To, validation.Required),
		"/data/attributes/amount": validation.Validate(&r.Data.Attributes.Amount, validation.Required),
	}.Filter()
}
