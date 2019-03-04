package config

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type Network struct {
	VtcAddressPrefix   string
	AssetAddressPrefix string
	StartHash          *chainhash.Hash
	StartHeight        int
	Name               string
}

func GetNetwork(name string) *Network {
	n := new(Network)
	if name == "regtest" {
		n.VtcAddressPrefix = "bcrt"
		n.AssetAddressPrefix = "rvtca"
		n.StartHash = &chainhash.Hash{}
		n.StartHeight = 0
		n.Name = "REGTEST"
		return n
	} else if name == "testnet" {
		n.VtcAddressPrefix = "tvtc"
		n.AssetAddressPrefix = "tvtca"
		n.StartHash, _ = chainhash.NewHashFromStr("cecdde91a6e53ead307ef615b78b6f47a0f5e4d3046e1a0df7501507ed28ffb6")
		n.StartHeight = 161205
		n.Name = "TESTNET"
		return n
	} else if name == "mainnet" {
		n.VtcAddressPrefix = "vtc"
		n.AssetAddressPrefix = "vtca"
		n.StartHash, _ = chainhash.NewHashFromStr("c9a400541bc579e1121c1990d94b96f4eef5bdc922fd5b763dbb4789cd28ce6d")
		n.StartHeight = 1048725
		n.Name = "MAINNET"
		return n
	}

	return nil
}
