package handlers

import (
	"faucet-svc/doorman"
	"faucet-svc/internal/service/helpers"
	"faucet-svc/internal/service/requests"
	"faucet-svc/internal/service/responses"
	"faucet-svc/internal/types/pg"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
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
	signerAddress := helpers.GetSignerAddress(string(request.Data.Type), signers)
	tokenAddress := request.Data.Attributes.TokenAddress
	signerBalance, err := chain.GetBalance(signerAddress, tokenAddress)
	if err != nil {
		helpers.Log(r).WithError(err).Errorf("failed to get balance")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	amount := request.Data.Attributes.Amount
	if helpers.IsLessOrEq(signerBalance, &amount) {
		helpers.Log(r).Error("insufficient balance")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	receiver := request.Data.Attributes.To
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

	userId, ok := doorman.GetHeader(r, "User-Id")
	if !ok {
		helpers.Log(r).Error("failed to get user id from header")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	humanBalance := helpers.ToHumanBalance(&amount, decimals)
	balance := pg.NewBalance(userId, chain.ID(), chain.Kind(), humanBalance, tokenAddress)
	balanceQ := helpers.BalancesQ(r)
	err = balanceQ.Update(&balance)
	if err != nil {
		helpers.Log(r).WithError(err).Error("failed to update balance")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	response := responses.NewTransactionResponse(txHash)
	w.WriteHeader(200)
	ape.Render(w, response)
}
