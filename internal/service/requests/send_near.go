package requests

import (
	"encoding/json"
	"faucet-svc/resources"
	validation "github.com/go-ozzo/ozzo-validation"
	"net/http"
	"regexp"

	"gitlab.com/distributed_lab/logan/v3/errors"
)

type CreateSendNearRequest struct {
	Data resources.SendNear
}

func NewCreateSendNearRequest(r *http.Request) (CreateSendNearRequest, error) {
	var request CreateSendNearRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return request, errors.Wrap(err, "failed to unmarshal")
	}

	return request, request.validate()
}

func (r *CreateSendNearRequest) validate() error {
	return validation.Errors{
		"/data/": validation.Validate(&r.Data, validation.Required),
		"/data/attributes/to": validation.Validate(&r.Data.Attributes.To,
			validation.Required,
			validation.Length(2, 64),
			validation.Match(
				regexp.MustCompile("^[a-z-_0-9]+(.testnet)?$"),
			),
		),
		"/data/attributes/amount": validation.Validate(&r.Data.Attributes.Amount, validation.Required),
	}.Filter()
}
