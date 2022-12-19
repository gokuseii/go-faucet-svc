package handlers

import (
	"faucet-svc/internal/service/helpers"
	"faucet-svc/internal/types"
	"faucet-svc/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"math/big"
	"net/http"
)

func GetTokenList(w http.ResponseWriter, r *http.Request) {

	chains := helpers.Chains(r)
	signer := helpers.Signers(r).Evm()
	tokens := helpers.Tokens(r)

	var tokenList []resources.Token
	for _, token := range tokens {
		for _, chainId := range token.Chains() {
			chain, ok := chains.Get(chainId, "evm")
			if !ok {
				continue
			}

			tokenAddress := token.Address()
			balance, err := chain.GetBalance(signer.Address().String(), &tokenAddress)
			if err != nil {
				helpers.Log(r).WithError(err).Errorf("failed to get balance of token %s", token.Address())
				ape.RenderErr(w, problems.InternalError())
				return
			}

			tokenList = append(tokenList, newToken(
				token.Address(),
				token.Kind(),
				token.Name(),
				token.Symbol(),
				balance,
				token.Decimals(),
				chain,
			))
		}
	}

	response := resources.TokenListResponse{
		Data: tokenList,
	}

	ape.Render(w, response)
}

func newToken(id, kind, name, symbol string, balance *big.Int, decimals float64, chain types.Chain) resources.Token {
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
		Relationships: newRelation(chain),
	}
}

func newRelation(chain types.Chain) *resources.TokenRelationships {
	return &resources.TokenRelationships{
		Chain: newChain(chain.ID(), chain.Kind(), chain.Name(), chain.NativeToken(), big.NewInt(0), 0),
	}
}
