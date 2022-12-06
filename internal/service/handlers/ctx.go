package handlers

import (
	"context"
	"faucet-svc/doorman"
	"faucet-svc/internal/config"
	"faucet-svc/internal/types"
	"net/http"

	"gitlab.com/distributed_lab/logan/v3"
)

type ctxKey int

const (
	logCtxKey ctxKey = iota
	chainerCtxKey
	signererCtxKey
	tokensCtxKey
	doormanConnectorCtxKey
)

func CtxLog(entry *logan.Entry) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, logCtxKey, entry)
	}
}

func Log(r *http.Request) *logan.Entry {
	return r.Context().Value(logCtxKey).(*logan.Entry)
}

func CtxSigners(v config.Signers) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, signererCtxKey, v)
	}
}

func Signers(r *http.Request) config.Signers {
	return r.Context().Value(signererCtxKey).(config.Signers)
}

func CtxChains(entry config.Chains) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, chainerCtxKey, entry)
	}
}

func Chains(r *http.Request) config.Chains {
	return r.Context().Value(chainerCtxKey).(config.Chains)
}

func CtxTokens(entry types.EvmTokens) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, tokensCtxKey, entry)
	}
}

func Tokens(r *http.Request) types.EvmTokens {
	return r.Context().Value(tokensCtxKey).(types.EvmTokens)
}

func CtxDoormanConnector(entry doorman.Connector) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, doormanConnectorCtxKey, entry)
	}
}
func DoormanConnector(r *http.Request) doorman.Connector {
	return r.Context().Value(doormanConnectorCtxKey).(doorman.Connector)
}
