package types

import (
	"context"
	"encoding/json"
	uint128 "github.com/eteu-technologies/golang-uint128"
	client2 "github.com/eteu-technologies/near-api-go/pkg/client"
	"github.com/eteu-technologies/near-api-go/pkg/client/block"
	types2 "github.com/eteu-technologies/near-api-go/pkg/types"
	"github.com/eteu-technologies/near-api-go/pkg/types/action"
	"github.com/eteu-technologies/near-api-go/pkg/types/transaction"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/portto/solana-go-sdk/client"
	common2 "github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/program/sysprog"
	types3 "github.com/portto/solana-go-sdk/types"
	"math/big"
)

type EvmChain interface {
	Client() *ethclient.Client
	ID() string
	Name() string
	RPC() string
	NativeToken() string
	Decimals() float64
	GetGasPrice(to common.Address, data []byte) (*big.Int, uint64, error)
	BuildTx(signer EvmSigner, to common.Address, amount big.Int, tokenAddress *common.Address) (*types.Transaction, error)
}

type evmChain struct {
	client      *ethclient.Client
	id          string
	name        string
	rpc         string
	nativeToken string
	decimals    float64
}

func NewEvmChain(client *ethclient.Client, id, name, rpc, nativeToken string, decimals float64) EvmChain {
	return &evmChain{
		client:      client,
		id:          id,
		name:        name,
		rpc:         rpc,
		nativeToken: nativeToken,
		decimals:    decimals,
	}
}

func (c *evmChain) ID() string {
	return c.id
}

func (c *evmChain) Name() string {
	return c.name
}

func (c *evmChain) RPC() string {
	return c.rpc
}

func (c *evmChain) NativeToken() string {
	return c.nativeToken
}

func (c *evmChain) Decimals() float64 {
	return c.decimals
}

func (c *evmChain) Client() *ethclient.Client {
	return c.client
}

func (c *evmChain) GetGasPrice(to common.Address, data []byte) (gasPrice *big.Int, gasLimit uint64, err error) {
	gasPrice, err = c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return
	}

	gasLimit, err = c.client.EstimateGas(context.Background(), ethereum.CallMsg{
		To:   &to,
		Data: data,
	})
	return
}

func (c *evmChain) BuildTx(signer EvmSigner, to common.Address, amount big.Int, tokenAddress *common.Address) (tx *types.Transaction, err error) {
	nonce, err := c.client.PendingNonceAt(context.Background(), signer.Address())
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

	gasPrice, gasLimit, err := c.GetGasPrice(to, data)
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

	if tokenAddress != nil {
		txData.To = tokenAddress
		txData.Value = big.NewInt(0)
		txData.Gas = gasLimit * 4
	}

	tx = types.NewTx(&txData)
	return
}

type EvmChains map[string]EvmChain

func (chains EvmChains) Get(key string) (EvmChain, bool) {
	val, ok := chains[key]
	return val, ok
}

func (chains EvmChains) Set(key string, val EvmChain) bool {
	if _, ok := chains[key]; ok {
		return false
	}

	chains[key] = val
	return true
}

type SolanaChain interface {
	Client() *client.Client
	ID() string
	RPC() string
	Decimals() float64
	BuildTx(signer types3.Account, receiver common2.PublicKey, amount uint64) (types3.Transaction, error)
}

type solanaChain struct {
	client   *client.Client
	id       string
	rpc      string
	decimals float64
}

func NewSolanaChain(client *client.Client, id, rpc string, decimals float64) SolanaChain {
	return &solanaChain{
		client:   client,
		id:       id,
		rpc:      rpc,
		decimals: decimals,
	}
}

func (c *solanaChain) Client() *client.Client {
	return c.client
}

func (c *solanaChain) ID() string {
	return c.id
}

func (c *solanaChain) RPC() string {
	return c.rpc
}

func (c *solanaChain) Decimals() float64 {
	return c.decimals
}

func (c *solanaChain) BuildTx(signer types3.Account, receiver common2.PublicKey, amount uint64) (tx types3.Transaction, err error) {
	response, err := c.client.GetLatestBlockhash(context.TODO())
	if err != nil {
		return
	}

	message := types3.NewMessage(
		types3.NewMessageParam{
			FeePayer: signer.PublicKey, // public key of the transaction signer
			Instructions: []types3.Instruction{
				sysprog.Transfer(
					sysprog.TransferParam{
						From:   signer.PublicKey, // public key of the transaction sender
						To:     receiver,         // wallet address of the transaction receiver
						Amount: amount,           // transaction amount
					},
				),
			},
			RecentBlockhash: response.Blockhash, // recent block hash
		},
	)

	// create a transaction with the message and TX signer
	tx, err = types3.NewTransaction(
		types3.NewTransactionParam{
			Message: message,
			Signers: []types3.Account{signer},
		},
	)
	return
}

type SolanaChains map[string]SolanaChain

func (chains SolanaChains) Get(key string) (SolanaChain, bool) {
	val, ok := chains[key]
	return val, ok
}

func (chains SolanaChains) Set(key string, val SolanaChain) bool {
	if _, ok := chains[key]; ok {
		return false
	}

	chains[key] = val
	return true
}

type NearChain interface {
	Client() *client2.Client
	ID() string
	RPC() string
	Decimals() float64
	GetAccountInfo(id string) (AccountInfo, error)
	BuildTx(signer NearSigner, receiverId string, amount big.Int) (transaction.Transaction, error)
	SignAndSerializeTx(signer NearSigner, tx transaction.Transaction) (string, error)
}

type nearChain struct {
	client   *client2.Client
	id       string
	rpc      string
	decimals float64
}

func NewNearChain(client *client2.Client, id, rpc string, decimals float64) NearChain {
	return &nearChain{
		client:   client,
		id:       id,
		rpc:      rpc,
		decimals: decimals,
	}
}

func (c *nearChain) Client() *client2.Client {
	return c.client
}

func (c *nearChain) ID() string {
	return c.id
}

func (c *nearChain) RPC() string {
	return c.rpc
}

func (c *nearChain) Decimals() float64 {
	return c.decimals
}

func (c *nearChain) GetAccountInfo(id string) (acc AccountInfo, err error) {
	res, err := c.client.AccountView(context.Background(), id, block.FinalityFinal())
	if err != nil {
		return
	}
	err = json.Unmarshal(res.Result, &acc)
	return
}

func (c *nearChain) BuildTx(signer NearSigner, receiverId string, amount big.Int) (txn transaction.Transaction, err error) {
	pubKey := signer.KeyPair().PublicKey

	accessKey, err := c.client.AccessKeyView(context.Background(), signer.ID(), pubKey, block.FinalityFinal())
	if err != nil {
		return
	}

	blockDetails, err := c.client.BlockDetails(context.Background(), block.FinalityFinal())
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
				types2.Balance(uint128.FromBig(&amount)),
			),
		},
		BlockHash: blockDetails.Header.Hash,
	}
	return
}

func (c *nearChain) SignAndSerializeTx(signer NearSigner, tx transaction.Transaction) (serTx string, err error) {
	signedTx, err := transaction.NewSignedTransaction(signer.KeyPair(), tx)
	if err != nil {
		return
	}
	serTx, err = signedTx.Serialize()
	return
}
