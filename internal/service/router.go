package service

import (
	"faucet-svc/internal/service/handlers"
	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
)

func (s *service) router() chi.Router {
	r := chi.NewRouter()

	r.Use(
		ape.RecoverMiddleware(s.log),
		ape.LoganMiddleware(s.log),
		ape.CtxMiddleware(
			handlers.CtxLog(s.log),
			handlers.CtxChains(s.chains),
			handlers.CtxSigners(s.signers),
		),
	)
	r.Route("/faucet", func(r chi.Router) {
		r.Route("/send", func(r chi.Router) {
			r.Post("/evm", handlers.SendEvm)
			r.Post("/solana", handlers.SendSolana)
		})
	})

	return r
}
