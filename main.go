package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gertjaap/vertcoin/blockprocessor"
	"github.com/gertjaap/vertcoin/config"
	"github.com/gertjaap/vertcoin/daemon"
	"github.com/gertjaap/vertcoin/server"
	"github.com/gertjaap/vertcoin/ui"
	"github.com/gertjaap/vertcoin/util"
	"github.com/gertjaap/vertcoin/wallet"
)

func main() {
	os.Mkdir(util.DataDirectory(), 0700)

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

	defer client.Shutdown()
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
	go func() {
		err := daemon.StartDaemon()
		if err != nil {
			fmt.Printf("Error starting daemon: %s\n", err.Error())
		}
	}()
	go func() {
		log.Fatal(srv.Run())
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			daemon.StopDaemon()
		}
	}()

	closeChan := make(chan bool, 1)
	go func() {
		<-closeChan
		daemon.StopDaemon()
	}()
	ui.RunUI(closeChan)
}
