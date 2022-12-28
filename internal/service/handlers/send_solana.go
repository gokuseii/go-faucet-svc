package handlers

import (
	"context"
	"faucet-svc/internal/service/requests"
	"faucet-svc/internal/service/responses"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/portto/solana-go-sdk/common"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"net/http"
)

func SendSolana(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewCreateSendSolanaRequest(r)
	if err != nil {
		Log(r).WithError(err).Error("invalid request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	chains := Chains(r).Solana()
	chain, ok := chains.Get(request.Data.ID)
	if !ok {
		Log(r).Error("unsupported chain")
		ape.RenderErr(w, problems.NotFound())
		return
	}

	signer := Signers(r).Solana()
	signerAddress := signer.PublicKey.String()
	receiver := common.PublicKeyFromString(request.Data.Attributes.To)
	if receiver.String() == signerAddress || receiver.String() == "11111111111111111111111111111111" {
		Log(r).Error("invalid receiver address")
		ape.RenderErr(w, problems.BadRequest(
			validation.Errors{"/data/attributes/to": errors.New("invalid receiver address")},
		)...)
		return
	}

	balance, err := chain.Client().GetBalance(context.TODO(), signer.PublicKey.String())
	if err != nil {
		Log(r).WithError(err).Error("failed to get balance")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	amount := request.Data.Attributes.Amount
	if balance < amount {
		Log(r).Error("insufficient balance")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	tx, err := chain.BuildTx(signer, receiver, amount)
	if err != nil {
		Log(r).WithError(err).Error("failed to build tx")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	txHash, err := chain.Client().SendTransaction(context.TODO(), tx)
	if err != nil {
		Log(r).WithError(err).Error("failed to send transaction")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	response := responses.NewTransactionResponse(txHash)
	w.WriteHeader(200)
	ape.Render(w, response)
}
