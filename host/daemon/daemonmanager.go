package daemon

import (
	"sync"
	"time"

	"github.com/gertjaap/vertcoin/host/coinparams"
	"github.com/gertjaap/vertcoin/logging"
)

type DaemonManager struct {
	daemons     []*Daemon
	daemonsLock sync.Mutex
	stop        chan bool
}

func NewDaemonManager() *DaemonManager {
	return &DaemonManager{daemons: []*Daemon{}, daemonsLock: sync.Mutex{}, stop: make(chan bool)}
}

func (dm *DaemonManager) Stop() {
	dm.stop <- true
}

func (dm *DaemonManager) AddDaemon(coin coinparams.Coin, coinNode coinparams.CoinNode, coinNetwork coinparams.CoinNetwork) {
	dm.daemonsLock.Lock()
	d := NewDaemon(coin, coinNode, coinNetwork, 56200+len(dm.daemons))
	dm.daemons = append(dm.daemons, d)
	dm.daemonsLock.Unlock()
}

func (dm *DaemonManager) Loop() {
	for {
		for _, d := range dm.daemons {
			err := d.StartIfNecessary()
			if err != nil {
				logging.Errorf("Error in daemon: %s/%s - %s\n", d.Coin.Id, d.Network.Id, err.Error())
			}

			err = d.GenerateIfNecessary()
			if err != nil {
				logging.Errorf("Error generating block: %s/%s - %s\n", d.Coin.Id, d.Network.Id, err.Error())
			}

		}

		select {
		case <-dm.stop:
			for _, d := range dm.daemons {
				err := d.Stop()
				if err != nil {
					logging.Errorf("Error stopping daemon: %s/%s - %s\n", d.Coin.Id, d.Network.Id, err.Error())
				}
			}
			return
		default:
			time.Sleep(time.Second * 5)
		}
	}
}
