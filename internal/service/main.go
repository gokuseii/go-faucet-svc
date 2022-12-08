package service

import (
	"faucet-svc/doorman"
	types2 "faucet-svc/internal/types"
	"gitlab.com/distributed_lab/kit/pgdb"
	"net"
	"net/http"

	"faucet-svc/internal/config"
	"gitlab.com/distributed_lab/kit/copus/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type service struct {
	log      *logan.Entry
	copus    types.Copus
	listener net.Listener
	chains   config.Chains
	signers  config.Signers
	tokens   types2.EvmTokens
	doorman  doorman.Connector
	db       *pgdb.DB
}

func (s *service) run() error {
	s.log.Info("Service started")
	r := s.router()

	if err := s.copus.RegisterChi(r); err != nil {
		return errors.Wrap(err, "cop failed")
	}

	return http.Serve(s.listener, r)
}

func newService(cfg config.Config) *service {
	return &service{
		log:      cfg.Log(),
		copus:    cfg.Copus(),
		listener: cfg.Listener(),
		chains:   cfg.Chains(),
		signers:  cfg.Signers(),
		tokens:   cfg.EvmTokens(),
		doorman:  cfg.DoormanConnector(),
		db:       cfg.DB(),
	}
}

func Run(cfg config.Config) {
	if err := newService(cfg).run(); err != nil {
		panic(err)
	}
}
