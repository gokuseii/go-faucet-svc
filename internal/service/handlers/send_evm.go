package handlers

import (
	"context"
	"crypto/ecdsa"
	"faucet-svc/internal/contracts"
	"faucet-svc/internal/service/requests"
	"faucet-svc/internal/service/responses"
	types2 "faucet-svc/internal/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	validation "github.com/go-ozzo/ozzo-validation"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"golang.org/x/exp/slices"
	"math/big"
	"net/http"
	"strings"
)

var (
	emptyAddress = common.Address{}
)

func SendEvm(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewCreateSendEvmRequest(r)
	if err != nil {
		Log(r).WithError(err).Error("invalid request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	chains := Chains(r).Evm()
	chain, ok := chains.Get(request.Data.Attributes.ChainId)
	if !ok {
		Log(r).Error("unsupported chain")
		ape.RenderErr(w, problems.NotFound())
		return
	}

	signer := Signers(r).Evm()
	if !common.IsHexAddress(request.Data.Attributes.To) || request.Data.Attributes.To == signer.Address().String() {
		Log(r).Error("invalid receiver address")
		ape.RenderErr(w, problems.BadRequest(
			validation.Errors{"/data/attributes/to": errors.New("invalid receiver address")},
		)...)
		return
	}
	receiver := common.HexToAddress(request.Data.Attributes.To)

	amount := request.Data.Attributes.Amount
	z := types2.Amount{Int: amount}
	if z.IsLessOrEq(types2.ZeroValue.Int) {
		Log(r).Error("invalid amount for sending")
		ape.RenderErr(w, problems.BadRequest(
			validation.Errors{"/data/attributes/amount": errors.New("invalid amount for sending")},
		)...)
		return
	}

	cli := chain.Client()
	var tx *types.Transaction
	if request.Data.Attributes.TokenAddress != nil {
		tk, ok := Tokens(r).Get(strings.ToLower(*request.Data.Attributes.TokenAddress))
		if !ok || !slices.Contains(tk.Chains(), chain.ID()) {
			Log(r).Error("unsupported token on this chain")
			ape.RenderErr(w, problems.NotFound())
			return
		}

		tokenAddress := common.HexToAddress(tk.Address())
		contract, err := contracts.NewErc20(tokenAddress, cli)
		if err != nil {
			Log(r).WithError(err).Error("failed to create instance")
			ape.RenderErr(w, problems.InternalError())
			return
		}

		symbol, err := contract.Symbol(&bind.CallOpts{})
		if err != nil {
			Log(r).WithError(err).Errorf("failed to get symbol contract %s", tokenAddress.String())
			ape.RenderErr(w, problems.InternalError())
			return
		}

		if symbol != request.Data.Attributes.Symbol {
			Log(r).Error("invalid symbol for this token")
			ape.RenderErr(w, problems.BadRequest(
				validation.Errors{"/data/attributes/symbol": errors.New("invalid symbol")},
			)...)
			return
		}

		res, err := contract.BalanceOf(&bind.CallOpts{}, signer.Address())
		if err != nil {
			Log(r).WithError(err).Errorf("failed to get balance in contract %s", tokenAddress.String())
			ape.RenderErr(w, problems.InternalError())
			return
		}

		balance := types2.Amount{Int: *res}
		if balance.IsLessOrEq(amount) {
			Log(r).Error("insufficient balance")
			ape.RenderErr(w, problems.InternalError())
			return
		}

		tx, err = buildEvmTx(cli, signer, receiver, amount, tokenAddress)
	} else {
		if chain.NativeToken() != request.Data.Attributes.Symbol {
			Log(r).Error("unsupported symbol for this chain")
			ape.RenderErr(w, problems.NotFound())
			return
		}

		res, err := cli.BalanceAt(context.Background(), signer.Address(), nil)
		if err != nil {
			Log(r).WithError(err).Error("failed to get balance")
			ape.RenderErr(w, problems.InternalError())
			return
		}

		balance := types2.Amount{Int: *res}
		if balance.IsLessOrEq(amount) {
			Log(r).Error("insufficient balance")
			ape.RenderErr(w, problems.InternalError())
			return
		}

		tx, err = buildEvmTx(cli, signer, receiver, amount, emptyAddress)
	}

	if err != nil {
		Log(r).WithError(err).Error("failed to build tx")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	signedTx, err := signEvmTx(big.NewInt(chain.ID()), tx, signer.PrivKey())
	if err != nil {
		Log(r).WithError(err).Error("failed to sign tx")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	err = cli.SendTransaction(context.Background(), signedTx)
	if err != nil {
		Log(r).WithError(err).Error("failed to send transaction")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	response := responses.NewTransactionResponse(signedTx.Hash().String())
	w.WriteHeader(200)
	ape.Render(w, response)
}

func signEvmTx(chainId *big.Int, tx *types.Transaction, privateKey *ecdsa.PrivateKey) (*types.Transaction, error) {
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), privateKey)
	if err != nil {
		return &types.Transaction{}, err
	}

	return signedTx, nil
}

func getGasPrice(
	client *ethclient.Client,
	to common.Address,
	data []byte,
) (gasPrice *big.Int, gasLimit uint64, err error) {
	gasPrice, err = client.SuggestGasPrice(context.Background())
	if err != nil {
		return
	}

	gasLimit, err = client.EstimateGas(context.Background(), ethereum.CallMsg{
		To:   &to,
		Data: data,
	})
	return
}

func buildEvmTx(
	client *ethclient.Client,
	signer types2.EvmSigner,
	to common.Address,
	amount big.Int,
	tokenAddress common.Address,
) (tx *types.Transaction, err error) {
	nonce, err := client.PendingNonceAt(context.Background(), signer.Address())
	if err != nil {
		return
	}

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := crypto.NewKeccakState()
	_, err = hash.Write(transferFnSignature)
	if err != nil {
		return
	}

	methodID := hash.Sum(nil)[:4]
	paddedAddress := common.LeftPadBytes(to.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	gasPrice, gasLimit, err := getGasPrice(client, to, data)
	if err != nil {
		return
	}

	txData := types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &to,
		Value:    &amount,
		Data:     data,
	}

	if tokenAddress != emptyAddress {
		txData.To = &tokenAddress
		txData.Value = big.NewInt(0)
		txData.Gas = gasLimit * 4
	}

	tx = types.NewTx(&txData)
	return
}
