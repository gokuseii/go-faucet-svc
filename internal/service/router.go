package service

import (
	"faucet-svc/internal/data/pg"
	"faucet-svc/internal/service/handlers"
	"faucet-svc/internal/service/helpers"
	"faucet-svc/internal/service/middlewares"
	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
)

func (s *service) router() chi.Router {
	r := chi.NewRouter()

	r.Use(
		ape.RecoverMiddleware(s.log),
		ape.LoganMiddleware(s.log),
		ape.CtxMiddleware(
			helpers.CtxLog(s.log),
			helpers.CtxChains(s.chains),
			helpers.CtxSigners(s.signers),
			helpers.CtxTokens(s.tokens),
			helpers.CtxDoormanConnector(s.doorman),
			helpers.CtxBalancesQ(pg.NewBalancesQ(s.db)),
		),
	)

	r.Route("/faucet", func(r chi.Router) {
		r.Get("/chains", handlers.GetChainList)
		r.Get("/tokens", handlers.GetTokenList)
		r.With(middlewares.CheckAuthorization).Route("/send", func(r chi.Router) {
			r.Post("/evm", handlers.SendEvm)
			r.Post("/solana", handlers.SendSolana)
			r.Post("/near", handlers.SendNear)
		})
	})

	// TODO: delete
	r.Get("/doorman/authenticate", handlers.Authenticate)

	return r
}
