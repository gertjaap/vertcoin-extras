package main

import (
	"fmt"
	"log"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gertjaap/vertcoin-openassets/blockprocessor"
	"github.com/gertjaap/vertcoin-openassets/server"
	"github.com/gertjaap/vertcoin-openassets/wallet"
)

func main() {

	connCfg := &rpcclient.ConnConfig{
		Host:         "localhost:18443",
		User:         "vtc",
		Pass:         "vtc",
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Shutdown()

	w := wallet.NewWallet(client)
	err = w.InitKey()
	if err != nil {
		log.Fatal(err)
	}

	vtcAdr, err := w.VertcoinAddress()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Your VTC Address is: %s\n", vtcAdr)

	assAdr, err := w.AssetsAddress()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Your VTC Asset Address is: %s\n", assAdr)

	bp := blockprocessor.NewBlockProcessor(w, client)

	go bp.Loop()

	srv := server.NewHttpServer(w)

	log.Fatal(srv.Run())
}
