package handlers

import (
	"context"
	"faucet-svc/internal/service/helpers"
	"faucet-svc/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"math/big"
	"net/http"
	"strings"
)

func GetChainList(w http.ResponseWriter, r *http.Request) {

	chains := helpers.Chains(r)
	signers := helpers.Signers(r)

	var chainList []resources.Chain
	for _, chain := range chains.Evm() {
		signer := signers.Evm()
		balance, err := chain.Client().BalanceAt(context.Background(), signer.Address(), nil)
		if err != nil {
			helpers.Log(r).WithError(err).Error("failed to get evm balance")
			ape.RenderErr(w, problems.InternalError())
			return
		}

		chainList = append(chainList, newChain(
			chain.ID(),
			"evm",
			chain.Name(),
			chain.NativeToken(),
			balance,
			chain.Decimals(),
		))
	}

	for _, chain := range chains.Solana() {
		signer := signers.Solana()
		balance, err := chain.Client().GetBalance(context.Background(), signer.PublicKey.String())
		if err != nil {
			helpers.Log(r).WithError(err).Error("failed to get solana balance")
			ape.RenderErr(w, problems.InternalError())
			return
		}

		name := "Solana " + chain.ID()
		strings.ToTitle(name)
		chainList = append(chainList, newChain(
			chain.ID(),
			"solana",
			name,
			"SOL",
			big.NewInt(int64(balance)),
			chain.Decimals(),
		))
	}

	nearSigner := signers.Near()
	near := chains.Near()
	acc, err := near.GetAccountInfo(nearSigner.ID())
	if err != nil {
		helpers.Log(r).WithError(err).Error("failed to get near balance")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	chainList = append(chainList, newChain(
		near.ID(),
		"near",
		"Near Testnet",
		"NEAR",
		&acc.Amount.Int,
		near.Decimals(),
	))

	response := resources.ChainListResponse{
		Data: chainList,
	}

	ape.Render(w, response)
}

func newChain(id, kind, name, nativeToken string, balance *big.Int, decimals float64) resources.Chain {
	return resources.Chain{
		Key: resources.Key{
			ID:   id,
			Type: resources.ResourceType(kind),
		},
		Attributes: resources.ChainAttributes{
			Name:        name,
			Balance:     helpers.ToHumanBalance(balance, decimals),
			NativeToken: nativeToken,
		},
	}
}
