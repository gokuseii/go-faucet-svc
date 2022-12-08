package handlers

import (
	"context"
	"faucet-svc/internal/contracts"
	"faucet-svc/internal/service/helpers"
	"faucet-svc/internal/service/requests"
	"faucet-svc/internal/service/responses"
	types2 "faucet-svc/internal/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	validation "github.com/go-ozzo/ozzo-validation"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"golang.org/x/exp/slices"
	"math/big"
	"net/http"
	"strings"
)

func SendEvm(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewCreateSendEvmRequest(r)
	if err != nil {
		helpers.Log(r).WithError(err).Error("invalid request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	chains := helpers.Chains(r).Evm()
	chain, ok := chains.Get(request.Data.ID)
	if !ok {
		helpers.Log(r).Error("unsupported chain")
		ape.RenderErr(w, problems.NotFound())
		return
	}

	signer := helpers.Signers(r).Evm()
	if !common.IsHexAddress(request.Data.Attributes.To) || request.Data.Attributes.To == signer.Address().String() {
		helpers.Log(r).Error("invalid receiver address")
		ape.RenderErr(w, problems.BadRequest(
			validation.Errors{"/data/attributes/to": errors.New("invalid receiver address")},
		)...)
		return
	}
	receiver := common.HexToAddress(request.Data.Attributes.To)

	amount := request.Data.Attributes.Amount
	z := types2.Amount{Int: amount}
	if z.IsLessOrEq(types2.ZeroValue.Int) {
		helpers.Log(r).Error("invalid amount for sending")
		ape.RenderErr(w, problems.BadRequest(
			validation.Errors{"/data/attributes/amount": errors.New("invalid amount for sending")},
		)...)
		return
	}

	cli := chain.Client()
	var decimals float64
	var tx *types.Transaction
	if request.Data.Attributes.TokenAddress != nil {
		tk, ok := helpers.Tokens(r).Get(strings.ToLower(*request.Data.Attributes.TokenAddress))
		if !ok || !slices.Contains(tk.Chains(), chain.ID()) {
			helpers.Log(r).Error("unsupported token on this chain")
			ape.RenderErr(w, problems.NotFound())
			return
		}

		tokenAddress := common.HexToAddress(tk.Address())
		contract, err := contracts.NewErc20(tokenAddress, cli)
		if err != nil {
			helpers.Log(r).WithError(err).Error("failed to create instance")
			ape.RenderErr(w, problems.InternalError())
			return
		}

		symbol, err := contract.Symbol(&bind.CallOpts{})
		if err != nil {
			helpers.Log(r).WithError(err).Errorf("failed to get symbol contract %s", tokenAddress.String())
			ape.RenderErr(w, problems.InternalError())
			return
		}

		if symbol != request.Data.Attributes.Symbol {
			helpers.Log(r).Error("invalid symbol for this token")
			ape.RenderErr(w, problems.BadRequest(
				validation.Errors{"/data/attributes/symbol": errors.New("invalid symbol")},
			)...)
			return
		}

		res, err := contract.BalanceOf(&bind.CallOpts{}, signer.Address())
		if err != nil {
			helpers.Log(r).WithError(err).Errorf("failed to get balance in contract %s", tokenAddress.String())
			ape.RenderErr(w, problems.InternalError())
			return
		}

		balance := types2.Amount{Int: *res}
		if balance.IsLessOrEq(amount) {
			helpers.Log(r).Error("insufficient balance")
			ape.RenderErr(w, problems.InternalError())
			return
		}

		d, err := contract.Decimals(&bind.CallOpts{})
		if err != nil {
			helpers.Log(r).WithError(err).Errorf("failed to get decimals of token %s", tk.Address())
			ape.RenderErr(w, problems.InternalError())
			return
		}
		decimals = float64(d)

		tx, err = chain.BuildTx(signer, receiver, amount, &tokenAddress)
	} else {
		if chain.NativeToken() != request.Data.Attributes.Symbol {
			helpers.Log(r).Error("unsupported symbol for this chain")
			ape.RenderErr(w, problems.NotFound())
			return
		}

		res, err := cli.BalanceAt(context.Background(), signer.Address(), nil)
		if err != nil {
			helpers.Log(r).WithError(err).Error("failed to get balance")
			ape.RenderErr(w, problems.InternalError())
			return
		}

		balance := types2.Amount{Int: *res}
		if balance.IsLessOrEq(amount) {
			helpers.Log(r).Error("insufficient balance")
			ape.RenderErr(w, problems.InternalError())
			return
		}
		decimals = chain.Decimals()

		tx, err = chain.BuildTx(signer, receiver, amount, nil)
	}

	if err != nil {
		helpers.Log(r).WithError(err).Error("failed to build tx")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	cid := big.NewInt(0)
	cid.SetString(chain.ID(), 10)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(cid), signer.PrivKey())
	if err != nil {
		helpers.Log(r).WithError(err).Error("failed to sign tx")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	err = cli.SendTransaction(context.Background(), signedTx)
	if err != nil {
		helpers.Log(r).WithError(err).Error("failed to send transaction")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	err = helpers.UpdateBalance(r, chain.ID(), "evm", helpers.ToHumanBalance(&amount, decimals), request.Data.Attributes.TokenAddress)
	if err != nil {
		helpers.Log(r).WithError(err).Error("failed to update balance")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	response := responses.NewTransactionResponse(signedTx.Hash().String())
	w.WriteHeader(200)
	ape.Render(w, response)
}
