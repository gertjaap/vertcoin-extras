package main

import (
	"log"

	"github.com/gertjaap/vertcoin-openassets/blockprocessor"
	"github.com/gertjaap/vertcoin-openassets/server"
)

func main() {
	go blockprocessor.Loop()
	log.Fatal(server.RunHttpServer())
}
