package main

import (
	"fmt"
	"log"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gertjaap/vertcoin-openassets/blockprocessor"
	"github.com/gertjaap/vertcoin-openassets/config"
	"github.com/gertjaap/vertcoin-openassets/server"
	"github.com/gertjaap/vertcoin-openassets/wallet"
	"github.com/skratchdot/open-golang/open"
)

func main() {
	cfg, err := config.InitConfig()
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

	_, err = client.GetBestBlockHash()
	if err != nil {
		log.Fatal(fmt.Errorf("Error connecting to Vertcoin Core RPC: %s\n", err.Error()))
	}

	defer client.Shutdown()

	w := wallet.NewWallet(client, cfg)
	bp := blockprocessor.NewBlockProcessor(w, client, cfg)
	srv := server.NewHttpServer(w, cfg)

	err = w.InitKey()
	if err != nil {
		log.Fatal(err)
	}

	go bp.Loop()

	go func() {
		time.Sleep(time.Second)
		open.Run(fmt.Sprintf("http://localhost:%d/", cfg.Port))
	}()

	log.Fatal(srv.Run())
}
