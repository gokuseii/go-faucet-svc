package chains

import (
	"context"
	"errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/program/sysprog"
	"github.com/portto/solana-go-sdk/types"
	"math/big"
)

type solanaChain struct {
	client      *client.Client
	signer      types.Account
	id          string
	name        string
	kind        string
	decimals    float64
	nativeToken string
	rpc         string
}

func NewSolanaChain(client *client.Client, signer types.Account, id, nativeToken, rpc string, decimals float64) Chain {
	return &solanaChain{
		client:      client,
		signer:      signer,
		id:          id,
		name:        "Solana " + id,
		kind:        "solana",
		decimals:    decimals,
		nativeToken: nativeToken,
		rpc:         rpc,
	}
}

func (c *solanaChain) ID() string {
	return c.id
}

func (c *solanaChain) Name() string {
	return c.name
}

func (c *solanaChain) Kind() string {
	return c.kind
}

func (c *solanaChain) NativeToken() string {
	return c.nativeToken
}

func (c *solanaChain) Decimals() float64 {
	return c.decimals
}

func (c *solanaChain) GetBalance(address string, _ *string) (balance *big.Int, err error) {
	bal, err := c.client.GetBalance(context.TODO(), address)
	if err != nil {
		return
	}
	balance = big.NewInt(int64(bal))
	return
}

func (c *solanaChain) Send(to string, amount *big.Int, _ *string) (txHash string, err error) {
	tx, err := c.buildTx(common.PublicKeyFromString(to), amount.Uint64())
	if err != nil {
		return
	}
	txHash, err = c.client.SendTransaction(context.TODO(), tx)
	return
}

func (c *solanaChain) buildTx(receiver common.PublicKey, amount uint64) (tx types.Transaction, err error) {
	response, err := c.client.GetLatestBlockhash(context.TODO())
	if err != nil {
		return
	}

	message := types.NewMessage(
		types.NewMessageParam{
			FeePayer: c.signer.PublicKey, // public key of the transaction signer
			Instructions: []types.Instruction{
				sysprog.Transfer(
					sysprog.TransferParam{
						From:   c.signer.PublicKey, // public key of the transaction sender
						To:     receiver,           // wallet address of the transaction receiver
						Amount: amount,             // transaction amount
					},
				),
			},
			RecentBlockhash: response.Blockhash, // recent block hash
		},
	)

	// create a transaction with the message and TX signer
	tx, err = types.NewTransaction(
		types.NewTransactionParam{
			Message: message,
			Signers: []types.Account{c.signer},
		},
	)
	return
}

func ValidateSolanaAddress(value interface{}) error {
	err := validation.Validate(value.(string), validation.Length(32, 44))
	if err != nil {
		return err
	}
	if common.PublicKeyFromString(value.(string)).String() == "11111111111111111111111111111111" {
		return errors.New("invalid receiver address")
	}
	return nil
}
