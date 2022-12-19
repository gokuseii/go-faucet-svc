package handlers

import (
	"faucet-svc/internal/config"
	"faucet-svc/internal/service/helpers"
	"faucet-svc/internal/service/requests"
	"faucet-svc/internal/service/responses"
	"faucet-svc/internal/types"
	validation "github.com/go-ozzo/ozzo-validation"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"math/big"
	"net/http"
	"strings"
)

func Send(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewCreateSendRequest(r)
	if err != nil {
		helpers.Log(r).WithError(err).Error("invalid request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	chains := helpers.Chains(r)
	chain, ok := chains.Get(request.Data.ID, string(request.Data.Type))
	if !ok {
		helpers.Log(r).Error("chain not found")
		ape.RenderErr(w, problems.NotFound())
		return
	}

	signers := helpers.Signers(r)
	signerAddress := GetSignerAddress(chain, signers)
	receiver := request.Data.Attributes.To
	if !chain.ValidateAddress(receiver) || receiver == signerAddress {
		helpers.Log(r).Error("invalid receiver address")
		ape.RenderErr(w, problems.BadRequest(
			validation.Errors{"/data/attributes/to": errors.New("invalid receiver address")},
		)...)
		return
	}

	tokenAddress := request.Data.Attributes.TokenAddress
	signerBalance, err := chain.GetBalance(signerAddress, tokenAddress)
	if err != nil {
		helpers.Log(r).WithError(err).Errorf("failed to get signer balance")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	amount := request.Data.Attributes.Amount
	if helpers.IsLessOrEq(*signerBalance, amount) {
		helpers.Log(r).Error("insufficient balance")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if helpers.IsLessOrEq(amount, *big.NewInt(0)) {
		helpers.Log(r).Error("invalid amount for sending")
		ape.RenderErr(w, problems.BadRequest(
			validation.Errors{"/data/attributes/amount": errors.New("invalid amount for sending")},
		)...)
		return
	}

	txHash, err := chain.Send(receiver, &amount, tokenAddress)
	if err != nil {
		helpers.Log(r).WithError(err).Error("failed to send transaction")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	decimals := chain.Decimals()
	if tokenAddress != nil {
		token, ok := helpers.Tokens(r).Get(strings.ToLower(*tokenAddress))
		if !ok {
			helpers.Log(r).Error("not found token")
			ape.RenderErr(w, problems.InternalError())
			return
		}
		decimals = token.Decimals()
	}

	humanBalance := helpers.ToHumanBalance(&amount, decimals)
	err = helpers.UpdateBalance(r, chain.ID(), chain.Kind(), humanBalance, tokenAddress)
	if err != nil {
		helpers.Log(r).WithError(err).Error("failed to update balance")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	response := responses.NewTransactionResponse(txHash)
	w.WriteHeader(200)
	ape.Render(w, response)
}

func GetSignerAddress(chain types.Chain, signers config.Signers) (signerAddress string) {
	switch chain.Kind() {
	case "evm":
		signerAddress = signers.Evm().Address().String()
	case "solana":
		signerAddress = signers.Solana().PublicKey.String()
	case "near":
		signerAddress = signers.Near().ID()
	}
	return
}
