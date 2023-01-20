package requests

import (
	"encoding/json"
	"faucet-svc/internal/service/helpers"
	"faucet-svc/internal/types/chains"
	"faucet-svc/resources"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"math/big"
	"net/http"
	"strings"
)

type CreateSendRequest struct {
	Data resources.Send
}

func NewCreateSendRequest(r *http.Request) (CreateSendRequest, error) {
	var request CreateSendRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return request, errors.Wrap(err, "failed to unmarshal")
	}

	return request, request.validate(r)
}

func (r *CreateSendRequest) validate(req *http.Request) error {
	return validation.Errors{
		"/data/":     validation.Validate(r.Data, validation.Required),
		"/data/id":   validation.Validate(r.Data.ID, validation.Required),
		"/data/type": validation.Validate(r.Data.Type, validation.Required),
		"/data/attributes/to": validation.Validate(
			r.Data.Attributes.To, validation.Required,
			validation.When(r.Data.Type == "evm", validation.By(chains.ValidateEvmAddress)),
			validation.When(r.Data.Type == "near", validation.By(chains.ValidateNearAddress)),
			validation.When(r.Data.Type == "solana", validation.By(chains.ValidateSolanaAddress)),
			validation.When(r.Data.Attributes.To != "",
				validation.By(func(value interface{}) error {
					signerAddress := helpers.GetSignerAddress(string(r.Data.Type), helpers.Signers(req))
					if strings.ToLower(r.Data.Attributes.To) == strings.ToLower(signerAddress) {
						return errors.New("cant be equal to signer address")
					}
					return nil
				}),
			),
		),
		"/data/attributes/amount": validation.Validate(
			r.Data.Attributes.Amount, validation.Required,
			validation.By(func(value interface{}) error {
				amount := value.(big.Int)
				if helpers.IsLessOrEq(&amount, big.NewInt(0)) {
					return errors.New("must be greater than 0")
				}
				return nil
			}),
		),
		"/data/attributes/token_address": validation.Validate(
			&r.Data.Attributes.TokenAddress,
			validation.When(
				r.Data.Type == "evm" && r.Data.Attributes.TokenAddress != nil,
				validation.NilOrNotEmpty,
				validation.By(func(value interface{}) error {
					return chains.ValidateEvmAddress(*r.Data.Attributes.TokenAddress)
				}),
			),
		),
	}.Filter()
}
