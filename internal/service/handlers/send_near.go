package handlers

import (
	"context"
	"encoding/json"
	"faucet-svc/internal/service/requests"
	"faucet-svc/internal/service/responses"
	types2 "faucet-svc/internal/types"
	"faucet-svc/internal/types/near"
	uint128 "github.com/eteu-technologies/golang-uint128"
	"github.com/eteu-technologies/near-api-go/pkg/client"
	"github.com/eteu-technologies/near-api-go/pkg/client/block"
	"github.com/eteu-technologies/near-api-go/pkg/types"
	"github.com/eteu-technologies/near-api-go/pkg/types/action"
	"github.com/eteu-technologies/near-api-go/pkg/types/transaction"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/jsonapi"
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

	_, err, problem := getAccountInfo(cli, receiverId)
	if err != nil {
		Log(r).WithError(err).Error("failed to get receiver account")
		ape.RenderErr(w, problem)
		return
	}

	signerInfo, err, problem := getAccountInfo(cli, signer.ID())
	if err != nil {
		Log(r).WithError(err).Error("failed to get signer account")
		ape.RenderErr(w, problem)
		return
	}

	if signerInfo.Amount.IsLessOrEq(amount) {
		Log(r).Error("insufficient balance")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	tx, err := buildTx(cli, signer, receiverId, amount)
	if err != nil {
		Log(r).WithError(err).Error("failed to build tx")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	serializedTx, err := signAndSerializeTx(signer, tx)
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

func getAccountInfo(cli *client.Client, id string) (near.AccountInfo, error, *jsonapi.ErrorObject) {

	res, err := cli.AccountView(context.Background(), id, block.FinalityFinal())
	if err != nil {
		return near.AccountInfo{}, err, problems.NotFound()
	}

	var acc near.AccountInfo
	err = json.Unmarshal(res.Result, &acc)
	if err != nil {
		return near.AccountInfo{}, err, problems.InternalError()
	}

	return acc, nil, nil
}

func buildTx(
	cli *client.Client,
	signer types2.NearSigner,
	receiverId string,
	amount big.Int,
) (txn transaction.Transaction, err error) {
	pubKey := signer.KeyPair().PublicKey

	accessKey, err := cli.AccessKeyView(context.Background(), signer.ID(), pubKey, block.FinalityFinal())
	if err != nil {
		return
	}

	blockDetails, err := cli.BlockDetails(context.Background(), block.FinalityFinal())
	if err != nil {
		return
	}

	txn = transaction.Transaction{
		PublicKey:  pubKey.ToPublicKey(),
		SignerID:   signer.ID(),
		Nonce:      accessKey.Nonce + 1,
		ReceiverID: receiverId,
		Actions: []action.Action{
			action.NewTransfer(
				types.Balance(uint128.FromBig(&amount)),
			),
		},
		BlockHash: blockDetails.Header.Hash,
	}
	return
}

func signAndSerializeTx(
	signer types2.NearSigner,
	tx transaction.Transaction,
) (serTx string, err error) {
	signedTx, err := transaction.NewSignedTransaction(signer.KeyPair(), tx)
	if err != nil {
		return
	}

	serTx, err = signedTx.Serialize()
	return
}
