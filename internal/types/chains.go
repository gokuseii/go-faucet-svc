package types

import (
	"context"
	"encoding/json"
	"errors"
	"faucet-svc/internal/contracts"
	uint128 "github.com/eteu-technologies/golang-uint128"
	client2 "github.com/eteu-technologies/near-api-go/pkg/client"
	"github.com/eteu-technologies/near-api-go/pkg/client/block"
	types2 "github.com/eteu-technologies/near-api-go/pkg/types"
	"github.com/eteu-technologies/near-api-go/pkg/types/action"
	"github.com/eteu-technologies/near-api-go/pkg/types/transaction"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/portto/solana-go-sdk/client"
	common2 "github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/program/sysprog"
	types3 "github.com/portto/solana-go-sdk/types"
	"math/big"
	"regexp"
)

type Chain interface {
	ID() string
	Name() string
	Kind() string
	NativeToken() string
	Decimals() float64
	ValidateAddress(address string) bool
	GetBalance(address string, tokenAddress *string) (*big.Int, error)
	Send(to string, amount *big.Int, tokenAddress *string) (txHash string, err error)
}

type evmChain struct {
	client      *ethclient.Client
	signer      EvmSigner
	id          string
	name        string
	kind        string
	decimals    float64
	nativeToken string
	rpc         string
}

func NewEvmChain(client *ethclient.Client, signer EvmSigner, id, name, nativeToken, rpc string, decimals float64) Chain {
	return &evmChain{
		client:      client,
		signer:      signer,
		id:          id,
		name:        name,
		kind:        "evm",
		decimals:    decimals,
		nativeToken: nativeToken,
		rpc:         rpc,
	}
}

func (c *evmChain) ID() string {
	return c.id
}

func (c *evmChain) Name() string {
	return c.name
}

func (c *evmChain) Kind() string {
	return c.kind
}

func (c *evmChain) NativeToken() string {
	return c.nativeToken
}

func (c *evmChain) Decimals() float64 {
	return c.decimals
}

func (c *evmChain) ValidateAddress(address string) bool {
	return common.IsHexAddress(address)
}

func (c *evmChain) GetBalance(address string, tokenAddress *string) (balance *big.Int, err error) {
	addr := common.HexToAddress(address)
	if tokenAddress != nil {
		contract, err := contracts.NewErc20(common.HexToAddress(*tokenAddress), c.client)
		if err != nil {
			return &big.Int{}, err
		}
		balance, err = contract.BalanceOf(&bind.CallOpts{}, addr)
		return balance, err
	}
	balance, err = c.client.BalanceAt(context.Background(), addr, nil)
	return
}

func (c *evmChain) Send(to string, amount *big.Int, tokenAddress *string) (txHash string, err error) {
	var tknAddr *common.Address
	if tokenAddress != nil {
		addr := common.HexToAddress(*tokenAddress)
		tknAddr = &addr
	}

	signedTx, err := c.buildTx(common.HexToAddress(to), *amount, tknAddr)
	if err != nil {
		return
	}
	err = c.client.SendTransaction(context.Background(), signedTx)
	return signedTx.Hash().String(), err
}

func (c *evmChain) getGasPrice(to common.Address, data []byte) (gasPrice *big.Int, gasLimit uint64, err error) {
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

func (c *evmChain) buildTx(to common.Address, amount big.Int, tokenAddress *common.Address) (signedTx *types.Transaction, err error) {
	nonce, err := c.client.PendingNonceAt(context.Background(), c.signer.Address())
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

	gasPrice, gasLimit, err := c.getGasPrice(to, data)
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

	tx := types.NewTx(&txData)

	cid := big.NewInt(0)
	cid.SetString(c.ID(), 10)
	signedTx, err = types.SignTx(tx, types.NewEIP155Signer(cid), c.signer.PrivKey())
	return
}

type solanaChain struct {
	client      *client.Client
	signer      types3.Account
	id          string
	name        string
	kind        string
	decimals    float64
	nativeToken string
	rpc         string
}

func NewSolanaChain(client *client.Client, signer types3.Account, id, nativeToken, rpc string, decimals float64) Chain {
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

func (c *solanaChain) ValidateAddress(address string) bool {
	return len(address) == 44 && common2.PublicKeyFromString(address).String() != "11111111111111111111111111111111"
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
	tx, err := c.buildTx(common2.PublicKeyFromString(to), amount.Uint64())
	if err != nil {
		return
	}
	txHash, err = c.client.SendTransaction(context.TODO(), tx)
	return
}

func (c *solanaChain) buildTx(receiver common2.PublicKey, amount uint64) (tx types3.Transaction, err error) {
	response, err := c.client.GetLatestBlockhash(context.TODO())
	if err != nil {
		return
	}

	message := types3.NewMessage(
		types3.NewMessageParam{
			FeePayer: c.signer.PublicKey, // public key of the transaction signer
			Instructions: []types3.Instruction{
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
	tx, err = types3.NewTransaction(
		types3.NewTransactionParam{
			Message: message,
			Signers: []types3.Account{c.signer},
		},
	)
	return
}

type nearChain struct {
	client      *client2.Client
	signer      NearSigner
	id          string
	name        string
	kind        string
	decimals    float64
	nativeToken string
	rpc         string
}

func NewNearChain(client *client2.Client, signer NearSigner, id, rpc, nativeToken string, decimals float64) Chain {
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

func (c *nearChain) ValidateAddress(address string) bool {
	matched, err := regexp.MatchString("^[a-z-_0-9]{2,64}(.testnet|.near)?$", address)
	return matched && err == nil
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

func (c *nearChain) getAccountInfo(id string) (acc AccountInfo, err error) {
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

type Chains map[string]Chain

func (chains Chains) Get(id, kind string) (Chain, bool) {
	val, ok := chains[kind+":"+id]
	return val, ok
}

func (chains Chains) Set(id, kind string, val Chain) bool {
	if _, ok := chains[kind+":"+id]; ok {
		return false
	}

	chains[kind+":"+id] = val
	return true
}
