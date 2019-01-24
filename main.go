package main

import (
	"fmt"
	"log"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gertjaap/vertcoin-openassets/blockprocessor"
	"github.com/gertjaap/vertcoin-openassets/config"
	"github.com/gertjaap/vertcoin-openassets/server"
	"github.com/gertjaap/vertcoin-openassets/wallet"
	"github.com/skratchdot/open-golang/open"
)

func main() {
	configChanged := make(chan bool, 0)
	cfg, err := config.InitConfig(configChanged)
	if err != nil {
		log.Fatal(err)
	}

	connCfg := &rpcclient.ConnConfig{
		Host:         cfg.RpcHost,
		User:         cfg.RpcUser,
		Pass:         cfg.RpcPassword,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}

	w := wallet.NewWallet(client, cfg)
	bp := blockprocessor.NewBlockProcessor(w, client, cfg)
	srv := server.NewHttpServer(w, cfg, bp)

	go func(wal *wallet.Wallet, blp *blockprocessor.BlockProcessor, config *rpcclient.ConnConfig) {
		for {
			<-configChanged
			config.Host = cfg.RpcHost
			config.User = cfg.RpcUser
			config.Pass = cfg.RpcPassword
		}
	}(w, bp, connCfg)

	err = w.InitKey()
	if err != nil {
		log.Fatal(err)
	}

	go bp.Loop()

	open.Run(fmt.Sprintf("http://localhost:%d/", cfg.Port))

	log.Fatal(srv.Run())
}
