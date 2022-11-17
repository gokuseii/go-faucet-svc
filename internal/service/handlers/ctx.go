package handlers

import (
	"context"
	"faucet-svc/internal/config"
	"net/http"

	"gitlab.com/distributed_lab/logan/v3"
)

type ctxKey int

const (
	logCtxKey ctxKey = iota
	chainerCtxKey
	signererCtxKey
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
