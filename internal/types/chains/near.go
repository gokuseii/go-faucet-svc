package chains

import (
	"context"
	"encoding/json"
	"errors"
	"faucet-svc/internal/types"
	uint128 "github.com/eteu-technologies/golang-uint128"
	"github.com/eteu-technologies/near-api-go/pkg/client"
	"github.com/eteu-technologies/near-api-go/pkg/client/block"
	types2 "github.com/eteu-technologies/near-api-go/pkg/types"
	"github.com/eteu-technologies/near-api-go/pkg/types/action"
	"github.com/eteu-technologies/near-api-go/pkg/types/transaction"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"math/big"
	"regexp"
)

type nearChain struct {
	client      *client.Client
	signer      types.NearSigner
	id          string
	name        string
	kind        string
	decimals    float64
	nativeToken string
	rpc         string
}

func NewNearChain(client *client.Client, signer types.NearSigner, id, rpc, nativeToken string, decimals float64) Chain {
	return &nearChain{
		client:      client,
		signer:      signer,
		id:          id,
		name:        "Near " + id,
		kind:        "near",
		decimals:    decimals,
		nativeToken: nativeToken,
		rpc:         rpc,
	}
}

func (c *nearChain) ID() string {
	return c.id
}

func (c *nearChain) Name() string {
	return c.name
}

func (c *nearChain) Kind() string {
	return c.kind
}

func (c *nearChain) NativeToken() string {
	return c.nativeToken
}

func (c *nearChain) Decimals() float64 {
	return c.decimals
}

func (c *nearChain) GetBalance(address string, _ *string) (balance *big.Int, err error) {
	account, err := c.getAccountInfo(address)
	if err != nil {
		return
	}
	balance = &account.Amount.Int
	return
}

func (c *nearChain) Send(to string, amount *big.Int, _ *string) (txHash string, err error) {
	tx, err := c.buildTx(to, amount)
	if err != nil {
		return
	}
	txRes, err := c.client.RPCTransactionSendAwait(
		context.Background(),
		tx,
	)
	if err != nil {
		return
	}

	if failMessage := string(txRes.Status.Failure); failMessage != "" {
		err = errors.New(failMessage)
		return
	}
	txHash = txRes.Transaction.Hash.String()
	return
}

func (c *nearChain) getAccountInfo(id string) (acc types.AccountInfo, err error) {
	res, err := c.client.AccountView(context.Background(), id, block.FinalityFinal())
	if err != nil {
		return
	}
	err = json.Unmarshal(res.Result, &acc)
	return
}

func (c *nearChain) buildTx(receiverId string, amount *big.Int) (serializedTx string, err error) {
	pubKey := c.signer.KeyPair().PublicKey

	accessKey, err := c.client.AccessKeyView(context.Background(), c.signer.ID(), pubKey, block.FinalityFinal())
	if err != nil {
		return
	}

	blockDetails, err := c.client.BlockDetails(context.Background(), block.FinalityFinal())
	if err != nil {
		return
	}

	txn := transaction.Transaction{
		PublicKey:  pubKey.ToPublicKey(),
		SignerID:   c.signer.ID(),
		Nonce:      accessKey.Nonce + 1,
		ReceiverID: receiverId,
		Actions: []action.Action{
			action.NewTransfer(
				types2.Balance(uint128.FromBig(amount)),
			),
		},
		BlockHash: blockDetails.Header.Hash,
	}

	signedTx, err := transaction.NewSignedTransaction(c.signer.KeyPair(), txn)
	if err != nil {
		return
	}

	serializedTx, err = signedTx.Serialize()
	return
}

func ValidateNearAddress(value interface{}) error {
	return validation.Validate(
		value.(string),
		validation.Length(2, 64),
		validation.Match(
			regexp.MustCompile("^[a-z-_0-9]{2,56}(.testnet|.near)?$"),
		),
	)
}
