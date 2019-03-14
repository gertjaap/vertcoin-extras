package wallet

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gertjaap/vertcoin/host/coinparams"
)

type WalletManager struct {
	wallets []*Wallet
}

func NewWalletManager() *WalletManager {
	return &WalletManager{wallets: []*Wallet{}}
}

func (w *WalletManager) AddWallet(c *rpcclient.Client, coinNetwork coinparams.CoinNetwork, coin coinparams.Coin, key *Key) error {
	wal, err := NewWallet(c, coinNetwork, coin, key)
	if err != nil {
		return err
	}
	w.wallets = append(w.wallets, wal)
	return nil
}
