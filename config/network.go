package config

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type Network struct {
	VtcAddressPrefix     string
	AssetAddressPrefix   string
	StealthAddressPrefix string
	StartHash            *chainhash.Hash
	StartHeight          int
	Name                 string
	DonationAddress      string
}

func GetNetwork(name string) *Network {
	n := new(Network)
	if name == "regtest" {
		n.VtcAddressPrefix = "rvtc"
		n.AssetAddressPrefix = "rvtca"
		n.StealthAddressPrefix = "rvtcs"
		n.StartHash = &chainhash.Hash{}
		n.StartHeight = 0
		n.Name = "REGTEST"
		n.DonationAddress = "rvtc1q9tyr0dmqfphm7hg9vaeh0yw9lttpfc436vgcwy"
		return n
	} else if name == "testnet" {
		n.VtcAddressPrefix = "tvtc"
		n.AssetAddressPrefix = "tvtca"
		n.StealthAddressPrefix = "tvtcs"
		n.StartHash, _ = chainhash.NewHashFromStr("cecdde91a6e53ead307ef615b78b6f47a0f5e4d3046e1a0df7501507ed28ffb6")
		n.StartHeight = 161205
		n.Name = "TESTNET"
		n.DonationAddress = "tvtc1qkh87j422ntkwd5pg8pkg9edgav8k4rfs757hln"
		return n
	} else if name == "mainnet" {
		n.VtcAddressPrefix = "vtc"
		n.AssetAddressPrefix = "vtca"
		n.StealthAddressPrefix = "vtcs"
		n.StartHash, _ = chainhash.NewHashFromStr("c9a400541bc579e1121c1990d94b96f4eef5bdc922fd5b763dbb4789cd28ce6d")
		n.StartHeight = 1048725
		n.Name = "MAINNET"
		n.DonationAddress = "vtc1qddgd7hg5hy8xl3a0fvlqc2zw0xr9fs69838u6w"
		return n
	}

	return nil
}
