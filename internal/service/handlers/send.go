package handlers

import (
	"context"
	"crypto/ecdsa"
	"faucet-svc/internal/service/requests"
	"faucet-svc/internal/service/responses"
	types2 "faucet-svc/internal/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	validation "github.com/go-ozzo/ozzo-validation"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"math/big"
	"net/http"
)

func SendEvmToken(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewCreateSendEvmTokenRequest(r)
	if err != nil {
		Log(r).WithError(err).Error("invalid request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	chains := EvmChains(r)
	chain, ok := chains.Get(request.Data.Attributes.ChainId)
	if !ok {
		Log(r).Error("unsupported chain")
		ape.RenderErr(w, problems.NotFound())
		return
	}

	if chain.NativeToken() != *request.Data.Attributes.Symbol {
		Log(r).Error("unsupported symbol for this chain")
		ape.RenderErr(w, problems.NotFound())
		return
	}

	signer := Signer(r)
	if !common.IsHexAddress(request.Data.Attributes.To) || request.Data.Attributes.To == signer.Address().String() {
		Log(r).Error("invalid receiver address")
		ape.RenderErr(w, problems.BadRequest(
			validation.Errors{"/data/attributes/to": errors.New("invalid receiver address")},
		)...)
		return
	}

	amount := request.Data.Attributes.Amount
	if isLessOrEq(amount, *big.NewInt(0)) {
		Log(r).Error("invalid amount for sending")
		ape.RenderErr(w, problems.BadRequest(
			validation.Errors{"/data/attributes/amount": errors.New("invalid amount for sending")},
		)...)
		return
	}

	balance, err := chain.Client().BalanceAt(context.Background(), signer.Address(), nil)
	if err != nil {
		Log(r).WithError(err).Error("failed to get balance")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if isLessOrEq(*balance, amount) {
		Log(r).Error("insufficient balance")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	receiver := common.HexToAddress(request.Data.Attributes.To)
	tx, err := buildTx(chain.Client(), signer, receiver, amount)
	if err != nil {
		Log(r).WithError(err).Error("failed to build tx")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	signedTx, err := signTx(big.NewInt(chain.ID()), &tx, signer.PrivKey())
	if err != nil {
		Log(r).WithError(err).Error("failed to sign tx")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	err = chain.Client().SendTransaction(context.Background(), signedTx)
	if err != nil {
		Log(r).WithError(err).Error("failed to send transaction")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	response := responses.NewTransactionResponse(signedTx.Hash().String())
	w.WriteHeader(200)
	ape.Render(w, response)
}

func isLessOrEq(x, y big.Int) bool {
	if cmp := x.Cmp(&y); cmp <= 0 {
		return true
	}
	return false
}

func signTx(chainId *big.Int, tx *types.Transaction, privateKey *ecdsa.PrivateKey) (*types.Transaction, error) {
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), privateKey)
	if err != nil {
		return &types.Transaction{}, err
	}

	return signedTx, nil
}

func buildTx(
	client *ethclient.Client,
	signer types2.Signer,
	to common.Address,
	amount big.Int,
) (types.Transaction, error) {
	nonce, err := client.PendingNonceAt(context.Background(), signer.Address())
	if err != nil {
		return types.Transaction{}, err
	}

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := crypto.NewKeccakState()
	_, err = hash.Write(transferFnSignature)
	if err != nil {
		return types.Transaction{}, err
	}

	methodID := hash.Sum(nil)[:4]
	paddedAddress := common.LeftPadBytes(to.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return types.Transaction{}, err
	}

	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		To:   &to,
		Data: data,
	})
	if err != nil {
		return types.Transaction{}, err
	}

	txData := types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &to,
		Value:    &amount,
		Data:     data,
	}

	tx := types.NewTx(&txData)
	return *tx, nil
}
