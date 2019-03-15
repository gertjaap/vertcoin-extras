package host

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/gertjaap/vertcoin/host/config"

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
	logging.SetLogLevel(int(logging.LogLevelDebug))

	if _, err := os.Stat(util.DataDirectory()); os.IsNotExist(err) {
		logging.Debugf("Creating data directory")
		os.MkdirAll(util.DataDirectory(), 0700)
	}

	logFilePath := path.Join(util.DataDirectory(), "vertcoin.log")
	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer logFile.Close()
	logging.SetLogFile(logFile)

	logging.Debugf("Initializing config")
	conf, err := config.InitConfig()
	if err != nil {
		return err
	}

	if h.passChan == nil {
		// nil passchan means host handles it in console
		h.passChan = make(chan wallet.PasswordPrompt)
		go h.passwordLoop()
	}

	coins, err := coinparams.Coins()
	if err != nil {
		return err
	}

	usedBipPaths := []uint32{}
	for _, ed := range conf.EnabledDaemons {
		for _, c := range coins {
			for _, n := range c.Networks {
				if n.Id == ed.NetworkId && c.Id == ed.CoinId {
					usedBipPaths = append(usedBipPaths, n.Bip44CoinIndex)
				}
			}
		}
	}

	key, err := wallet.NewKey(path.Join(util.DataDirectory(), "privkey.hex"), h.passChan, usedBipPaths)
	if err != nil {
		return err
	}

	for _, ed := range conf.EnabledDaemons {
		started := false
		nobinary := false
		var node coinparams.CoinNode
		for _, c := range coins {
			if c.Id == ed.CoinId {
				for _, n := range c.Nodes {
					if n.Arch == runtime.GOARCH && n.Os == runtime.GOOS {
						node = n
						break
					}
				}

				if node.Hash != nil {
					for _, n := range c.Networks {
						if n.Id == ed.NetworkId {
							h.daemonManager.AddDaemon(c, node, n)
							started = true
						}
					}
				} else {
					nobinary = true
				}
			}
		}
		if nobinary {
			logging.Warnf("Cannot start node for network %s/%s: No compatible binary found", ed.CoinId, ed.NetworkId)
		} else if !started {
			logging.Warnf("Cannot start node for network %s/%s: Unknown coin or network", ed.CoinId, ed.NetworkId)
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

func (h *Host) passwordLoop() {

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

}
