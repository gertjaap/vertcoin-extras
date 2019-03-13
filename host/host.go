package host

import (
	"os"
	"path"
	"runtime"

	"github.com/gertjaap/vertcoin/host/coinparams"
	"github.com/gertjaap/vertcoin/host/daemon"
	"github.com/gertjaap/vertcoin/logging"
	"github.com/gertjaap/vertcoin/util"
)

type Host struct {
	daemonManager *daemon.DaemonManager
}

func NewHost() *Host {
	return &Host{daemonManager: daemon.NewDaemonManager()}
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

	h.daemonManager.Loop()

	return nil
}
