package main

import (
	"os"
	"os/signal"

	"github.com/gertjaap/vertcoin/client"
	"github.com/gertjaap/vertcoin/host"
)

func main() {
	host := host.NewHost()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			host.Stop()
		}
	}()

	go func() {
		err := host.Start()
		if err != nil {
			panic(err)
		}
	}()

	cli := client.NewClient()
	cli.Run()
	host.Stop()
}
