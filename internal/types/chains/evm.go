package chains

import (
	"context"
	"errors"
	"faucet-svc/internal/contracts"
	types2 "faucet-svc/internal/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"math/big"
)

type evmChain struct {
	client      *ethclient.Client
	signer      types2.EvmSigner
	id          string
	name        string
	kind        string
	decimals    float64
	nativeToken string
	rpc         string
}

func NewEvmChain(client *ethclient.Client, signer types2.EvmSigner, id, name, nativeToken, rpc string, decimals float64) Chain {
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

func ValidateEvmAddress(value interface{}) error {
	err := validation.Validate(value.(string), validation.Length(40, 42))
	if err != nil {
		return err
	}
	if !common.IsHexAddress(value.(string)) {
		return errors.New("invalid address")
	}
	return nil
}
