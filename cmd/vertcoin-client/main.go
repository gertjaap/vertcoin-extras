package main

import (
	"github.com/gertjaap/vertcoin/client"
)

func main() {
	cli := client.NewClient()
	cli.Run()
}
