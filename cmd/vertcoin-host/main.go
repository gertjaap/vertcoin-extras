package main

import (
	"os"
	"os/signal"

	"github.com/gertjaap/vertcoin/logging"

	"github.com/gertjaap/vertcoin/host"
)

func main() {
	host := host.NewHost(nil)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			host.Stop()
		}
	}()

	err := host.Start()
	if err != nil {
		logging.Errorf("Error starting host: %s", err.Error())
		os.Exit(1)
	}
}
