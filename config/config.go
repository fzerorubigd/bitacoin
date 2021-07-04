package config

import (
	"github.com/fzerorubigd/bitacoin/helper"
)

var Config BlockChainConfig

type BlockChainConfig struct {
	Host         string
	InitialNodes []string
	PubKeyPath   string
}

func ReadConfigFile(path string) error {
	return helper.ReadJSON(path, &Config)
}
