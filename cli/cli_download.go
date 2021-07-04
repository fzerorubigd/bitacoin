package cli

import (
	"flag"
	"github.com/fzerorubigd/bitacoin/config"
	"github.com/fzerorubigd/bitacoin/downloader"
	"github.com/fzerorubigd/bitacoin/interactor"
	"github.com/fzerorubigd/bitacoin/storege"
	"log"
)

func download(store storege.Store, args ...string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	var configFilePath string
	fs.StringVar(&configFilePath, "config", "config.json", "config file path")
	err := fs.Parse(args[1:])
	if err != nil {
		return err
	}

	err = config.ReadConfigFile(configFilePath)
	if err != nil {
		log.Fatalf("read config file err:%s\n", err.Error())
	}

	interactor.Init()
	downloader.DownloadBlockChainData(store)

	return nil
}

func init() {
	addCommand("download", "download blockchain from other nodes", download)
}
