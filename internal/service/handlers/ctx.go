package handlers

import (
	"context"
	"faucet-svc/internal/types"
	"net/http"

	"gitlab.com/distributed_lab/logan/v3"
)

type ctxKey int

const (
	logCtxKey ctxKey = iota
	evmChainsCtxKey
	signerCtxKey
)

func CtxLog(entry *logan.Entry) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, logCtxKey, entry)
	}
}

func Log(r *http.Request) *logan.Entry {
	return r.Context().Value(logCtxKey).(*logan.Entry)
}

func CtxSigner(v types.Signer) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, signerCtxKey, v)
	}
}

func Signer(r *http.Request) types.Signer {
	return r.Context().Value(signerCtxKey).(types.Signer)
}

func CtxEvmChains(entry types.EvmChains) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, evmChainsCtxKey, entry)
	}
}

func EvmChains(r *http.Request) types.EvmChains {
	return r.Context().Value(evmChainsCtxKey).(types.EvmChains)
}
