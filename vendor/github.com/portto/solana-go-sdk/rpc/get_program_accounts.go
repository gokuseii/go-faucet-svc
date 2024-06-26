package rpc

import (
	"context"
)

type GetProgramAccountsResponse JsonRpcResponse[GetProgramAccounts]

type GetProgramAccountsWithContextResponse JsonRpcResponse[GetProgramAccountsWithContext]

type GetProgramAccount struct {
	Pubkey  string      `json:"pubkey"`
	Account AccountInfo `json:"account"`
}

type GetProgramAccounts []GetProgramAccount

type GetProgramAccountsWithContext struct {
	Context Context            `json:"context"`
	Value   GetProgramAccounts `json:"value"`
}

// GetProgramAccountsConfig is a option config for `getProgramAccounts`
type GetProgramAccountsConfig struct {
	Encoding   AccountEncoding                  `json:"encoding,omitempty"`
	Commitment Commitment                       `json:"commitment,omitempty"`
	DataSlice  *DataSlice                       `json:"dataSlice,omitempty"`
	Filters    []GetProgramAccountsConfigFilter `json:"filters,omitempty"`
}

type getProgramAccountsConfig struct {
	GetProgramAccountsConfig
	WithContext bool `json:"withContext,omitempty"`
}

// GetProgramAccountsConfigFilter you can set either MemCmp or DataSize but can be both, if needed, separate them into two
type GetProgramAccountsConfigFilter struct {
	MemCmp   *GetProgramAccountsConfigFilterMemCmp `json:"memcmp,omitempty"`
	DataSize uint64                                `json:"dataSize,omitempty"`
}

type GetProgramAccountsConfigFilterMemCmp struct {
	Offset uint64 `json:"offset"`
	Bytes  string `json:"bytes"`
}

func (c *RpcClient) GetProgramAccounts(ctx context.Context, programId string) (JsonRpcResponse[GetProgramAccounts], error) {
	return c.processGetProgramAccounts(c.Call(ctx, "getProgramAccounts", programId))
}

func (c *RpcClient) GetProgramAccountsWithConfig(ctx context.Context, programId string, cfg GetProgramAccountsConfig) (JsonRpcResponse[GetProgramAccounts], error) {
	return c.processGetProgramAccounts(c.Call(ctx, "getProgramAccounts", programId, c.toInternalGetProgramAccountsConfig(cfg, false)))
}

func (c *RpcClient) processGetProgramAccounts(body []byte, rpcErr error) (res JsonRpcResponse[GetProgramAccounts], err error) {
	err = c.processRpcCall(body, rpcErr, &res)
	return
}

func (c *RpcClient) GetProgramAccountsWithContext(ctx context.Context, programId string) (JsonRpcResponse[GetProgramAccountsWithContext], error) {
	return c.processGetProgramAccountsWithContext(c.Call(ctx, "getProgramAccounts", programId, c.toInternalGetProgramAccountsConfig(GetProgramAccountsConfig{}, true)))
}

func (c *RpcClient) GetProgramAccountsWithContextAndConfig(ctx context.Context, programId string, cfg GetProgramAccountsConfig) (JsonRpcResponse[GetProgramAccountsWithContext], error) {
	return c.processGetProgramAccountsWithContext(c.Call(ctx, "getProgramAccounts", programId, c.toInternalGetProgramAccountsConfig(cfg, true)))
}

func (c *RpcClient) processGetProgramAccountsWithContext(body []byte, rpcErr error) (res JsonRpcResponse[GetProgramAccountsWithContext], err error) {
	err = c.processRpcCall(body, rpcErr, &res)
	return
}

func (c *RpcClient) toInternalGetProgramAccountsConfig(cfg GetProgramAccountsConfig, withContext bool) getProgramAccountsConfig {
	return getProgramAccountsConfig{
		GetProgramAccountsConfig: cfg,
		WithContext:              withContext,
	}
}
