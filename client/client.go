package client

import (
	"github.com/gertjaap/vertcoin/client/ui"
)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Stop() {

}

func (c *Client) Run() {
	closeChan := make(chan bool, 1)
	ui.RunUI(closeChan)
}
