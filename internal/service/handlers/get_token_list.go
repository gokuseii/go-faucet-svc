package handlers

import (
	"faucet-svc/internal/contracts"
	"faucet-svc/internal/service/helpers"
	"faucet-svc/resources"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"math/big"
	"net/http"
)

func GetTokenList(w http.ResponseWriter, r *http.Request) {

	chains := Chains(r).Evm()
	signer := Signers(r).Evm()
	tokens := Tokens(r)

	var tokenList []resources.Token
	for _, token := range tokens {
		for _, chainId := range token.Chains() {
			chain, ok := chains.Get(chainId)
			if !ok {
				continue
			}

			contract, err := contracts.NewErc20(common.HexToAddress(token.Address()), chain.Client())
			if err != nil {
				Log(r).WithError(err).Errorf("failed to create token instance %s", token.Address())
				ape.RenderErr(w, problems.InternalError())
				return
			}

			balance, err := contract.BalanceOf(&bind.CallOpts{}, signer.Address())
			if err != nil {
				Log(r).WithError(err).Errorf("failed to get balance of token %s", token.Address())
				ape.RenderErr(w, problems.InternalError())
				return
			}

			decimals, err := contract.Decimals(&bind.CallOpts{})
			if err != nil {
				Log(r).WithError(err).Errorf("failed to get decimals of token %s", token.Address())
				ape.RenderErr(w, problems.InternalError())
				return
			}

			tokenList = append(tokenList, newToken(
				token.Address(),
				"ERC20",
				token.Name(),
				token.Symbol(),
				chainId,
				"evm",
				balance,
				float64(decimals),
			))
		}
	}

	response := resources.TokenListResponse{
		Data: tokenList,
	}

	ape.Render(w, response)
}

func newToken(id, kind, name, symbol, chainId, chainType string, balance *big.Int, decimals float64) resources.Token {
	return resources.Token{
		Key: resources.Key{
			ID:   id,
			Type: resources.ResourceType(kind),
		},
		Attributes: resources.TokenAttributes{
			Name:    name,
			Balance: helpers.ToHumanBalance(balance, decimals),
			Symbol:  symbol,
		},
		Relationships: newRelation(chainId, chainType),
	}
}

func newRelation(id, kind string) resources.TokenRelationships {
	return resources.TokenRelationships{
		Chain: resources.Relation{
			Data: &resources.Key{
				ID:   id,
				Type: resources.ResourceType(kind),
			},
		},
	}
}
