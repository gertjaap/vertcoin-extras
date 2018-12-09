package config

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type Network struct {
	VtcAddressPrefix   string
	AssetAddressPrefix string
	StartHash          *chainhash.Hash
}

func GetNetwork(name string) *Network {
	n := new(Network)
	if name == "regtest" {
		n.VtcAddressPrefix = "bcrt"
		n.AssetAddressPrefix = "rvtca"
		n.StartHash = &chainhash.Hash{}
		return n
	} else if name == "testnet" {
		n.VtcAddressPrefix = "tvtc"
		n.AssetAddressPrefix = "tvtca"
		n.StartHash, _ = chainhash.NewHashFromStr("cecdde91a6e53ead307ef615b78b6f47a0f5e4d3046e1a0df7501507ed28ffb6")
		return n
	}

	return nil
}
