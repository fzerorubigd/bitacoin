package config

import (
	"github.com/fzerorubigd/bitacoin/helper"
)

var Config BlockChainConfig

type BlockChainConfig struct {
	IP           string
	Port         string
	InitialNodes []string
}

func ReadConfigFile(path string) error {
	return helper.ReadJSON(path, &Config)
}
