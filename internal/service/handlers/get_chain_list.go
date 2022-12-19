package handlers

import (
	"faucet-svc/internal/service/helpers"
	"faucet-svc/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"math/big"
	"net/http"
)

func GetChainList(w http.ResponseWriter, r *http.Request) {
	chains := helpers.Chains(r)
	signers := helpers.Signers(r)
	var chainList []resources.Chain
	for _, chain := range chains {
		signerAddress := GetSignerAddress(chain, signers)
		balance, err := chain.GetBalance(signerAddress, nil)
		if err != nil {
			helpers.Log(r).WithError(err).Errorf("failed to get balance on %s chain, wallet address %s", chain.Kind(), signerAddress)
			ape.RenderErr(w, problems.InternalError())
			return
		}

		chainList = append(chainList, newChain(
			chain.ID(),
			chain.Kind(),
			chain.Name(),
			chain.NativeToken(),
			balance,
			chain.Decimals(),
		))
	}

	response := resources.ChainListResponse{
		Data: chainList,
	}

	ape.Render(w, response)
}

func newChain(id, kind, name, nativeToken string, balance *big.Int, decimals float64) resources.Chain {
	bal := helpers.ToHumanBalance(balance, decimals)
	return resources.Chain{
		Key: resources.Key{
			ID:   id,
			Type: resources.ResourceType(kind),
		},
		Attributes: resources.ChainAttributes{
			Name:        name,
			Balance:     &bal,
			NativeToken: nativeToken,
		},
	}
}
