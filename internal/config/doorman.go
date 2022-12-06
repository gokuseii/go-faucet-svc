package config

import (
	"faucet-svc/doorman"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type DoormanConfiger interface {
	DoormanConfig() *DoormanConfig
	DoormanConnector() doorman.Connector
}

type DoormanConfig struct {
	ServiceUrl string `fig:"service_url,required"`
}

func NewDoormanConfiger(getter kv.Getter) DoormanConfiger {
	return &doormanConfig{
		getter: getter,
	}
}

type doormanConfig struct {
	getter kv.Getter
	once   comfig.Once
}

func (c *doormanConfig) DoormanConfig() *DoormanConfig {
	return c.once.Do(func() interface{} {
		raw := kv.MustGetStringMap(c.getter, "doorman")
		config := DoormanConfig{}
		err := figure.Out(&config).From(raw).Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out"))
		}
		return &config
	}).(*DoormanConfig)
}

func (c *doormanConfig) DoormanConnector() doorman.Connector {
	return doorman.NewConnector(c.DoormanConfig().ServiceUrl)
}
