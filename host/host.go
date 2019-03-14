package host

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/howeyc/gopass"

	"github.com/gertjaap/vertcoin/host/coinparams"
	"github.com/gertjaap/vertcoin/host/daemon"
	"github.com/gertjaap/vertcoin/host/wallet"
	"github.com/gertjaap/vertcoin/logging"
	"github.com/gertjaap/vertcoin/util"
)

type Host struct {
	daemonManager *daemon.DaemonManager
	passChan      chan wallet.PasswordPrompt
}

func NewHost(passChan chan wallet.PasswordPrompt) *Host {
	return &Host{daemonManager: daemon.NewDaemonManager(), passChan: passChan}
}

func (h *Host) Stop() {
	h.daemonManager.Stop()
}

func (h *Host) Start() error {
	os.Mkdir(util.DataDirectory(), 0700)
	logging.SetLogLevel(int(logging.LogLevelDebug))
	logFilePath := path.Join(util.DataDirectory(), "vertcoin.log")
	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer logFile.Close()
	logging.SetLogFile(logFile)

	if h.passChan == nil {
		// nil passchan means host handles it in console
		h.passChan = make(chan wallet.PasswordPrompt)
		go func() {
			for {
				passRequest := <-h.passChan

				for {
					fmt.Printf("Please enter passphrase [%s]: ", passRequest.Reason)
					pass, err := gopass.GetPasswd()
					if err != nil {
						passRequest.ResponseChannel <- ""
						close(passRequest.ResponseChannel)
						break
					}

					if passRequest.Confirm {
						fmt.Printf("\nPlease confirm passphrase [%s]: ", passRequest.Reason)
						pass2, err := gopass.GetPasswd()
						if err != nil {
							passRequest.ResponseChannel <- ""
							close(passRequest.ResponseChannel)
							break
						}
						if !bytes.Equal(pass, pass2) {
							fmt.Printf("\nPasswords do not match, please try again.\n\n")
							continue
						}
					}
					passRequest.ResponseChannel <- string(pass)
					close(passRequest.ResponseChannel)
					break
				}
			}
		}()
	}
	key, err := wallet.NewKey(path.Join(util.DataDirectory(), "privkey.hex"), h.passChan)
	if err != nil {
		return err
	}

	coins, err := coinparams.Coins()
	if err != nil {
		return err
	}
	for _, c := range coins {
		node := coinparams.CoinNode{Hash: nil}
		for _, n := range c.Nodes {
			if n.Arch == runtime.GOARCH && n.Os == runtime.GOOS {
				node = n
				break
			}
		}

		if node.Hash != nil {
			for _, n := range c.Networks {
				h.daemonManager.AddDaemon(c, node, n)
			}
		}
	}

	wm := wallet.NewWalletManager()

	daemons := h.daemonManager.Daemons()
	for _, d := range daemons {
		rpc, err := d.RpcClient()
		if err != nil {
			return err
		}
		err = wm.AddWallet(rpc, d.Network, d.Coin, key)
		if err != nil {
			return err
		}
	}

	h.daemonManager.Loop()
	return nil
}
