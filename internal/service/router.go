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
			handlers.CtxEvmChains(s.evmChains),
			handlers.CtxSigner(s.signer),
		),
	)
	r.Route("/faucet", func(r chi.Router) {
		r.Post("/send/evm", handlers.SendEvmToken)
	})

	return r
}
