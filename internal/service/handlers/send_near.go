package handlers

import (
	"context"
	"faucet-svc/internal/service/requests"
	"faucet-svc/internal/service/responses"
	types2 "faucet-svc/internal/types"
	validation "github.com/go-ozzo/ozzo-validation"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"math/big"
	"net/http"
)

func SendNear(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewCreateSendNearRequest(r)
	if err != nil {
		Log(r).WithError(err).Error("invalid request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	signer := Signers(r).Near()
	receiverId := request.Data.Attributes.To
	if receiverId == signer.ID() {
		Log(r).Error("invalid receiver address")
		ape.RenderErr(w, problems.BadRequest(
			validation.Errors{"/data/attributes/to": errors.New("invalid receiver address")},
		)...)
		return
	}

	amount := request.Data.Attributes.Amount
	z := types2.Amount{Int: amount}
	if z.IsLessOrEq(*big.NewInt(0)) {
		Log(r).Error("invalid amount for sending")
		ape.RenderErr(w, problems.BadRequest(
			validation.Errors{"/data/attributes/amount": errors.New("invalid amount for sending")},
		)...)
		return
	}

	chain := Chains(r).Near()
	cli := chain.Client()

	_, err = chain.GetAccountInfo(receiverId)
	if err != nil {
		Log(r).WithError(err).Error("failed to get receiver account")
		ape.RenderErr(w, problems.NotFound())
		return
	}

	signerInfo, err := chain.GetAccountInfo(signer.ID())
	if err != nil {
		Log(r).WithError(err).Error("failed to get signer account")
		ape.RenderErr(w, problems.NotFound())
		return
	}

	if signerInfo.Amount.IsLessOrEq(amount) {
		Log(r).Error("insufficient balance")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	tx, err := chain.BuildTx(signer, receiverId, amount)
	if err != nil {
		Log(r).WithError(err).Error("failed to build tx")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	serializedTx, err := chain.SignAndSerializeTx(signer, tx)
	if err != nil {
		Log(r).WithError(err).Error("failed to sign transaction")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if err != nil {
		Log(r).WithError(err).Error("failed to serialize transaction")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	txRes, err := cli.RPCTransactionSendAwait(
		context.Background(),
		serializedTx,
	)
	if err != nil {
		Log(r).WithError(err).Error("failed to send transaction")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	response := responses.NewTransactionResponse(txRes.Transaction.Hash.String())
	w.WriteHeader(200)
	ape.Render(w, response)
}
