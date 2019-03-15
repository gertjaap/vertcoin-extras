package coinparams

import (
	"encoding/json"

	"github.com/btcsuite/btcd/btcec"

	"github.com/adiabat/bech32"

	"github.com/gertjaap/vertcoin/util"

	"github.com/gobuffalo/packr/v2"
)

type Coin struct {
	Id                 string        `json:"id"`
	Name               string        `json:"name"`
	NodeExecutableName string        `json:"nodeExecutableName"`
	Nodes              []CoinNode    `json:"nodes"`
	Networks           []CoinNetwork `json:"networks"`
}

type CoinNode struct {
	Version string `json:"version"`
	Url     string `json:"url"`
	Hash    []byte `json:"sha256"`
	Os      string `json:"os"`
	Arch    string `json:"arch"`
}

type CoinNetwork struct {
	Name                          string `json:"name"`
	Id                            string `json:"id"`
	DaemonParameters              string `json:"daemonParameters"`
	Generate                      bool   `json:"generate"`
	DaemonPort                    int    `json:"daemonPort"`
	GenesisHash                   []byte `json:"genesisHash"`
	Bip44CoinIndex                uint32 `json:"bip44"`
	Base58PrefixPubKeyAddress     byte   `json:"base58PrefixPubKeyAddress"`
	Base58PrefixScriptAddress     byte   `json:"base58PrefixScriptAddress"`
	Base58PrefixSecretKey         byte   `json:"base58PrefixSecretKey"`
	Base58PrefixExtendedPublicKey []byte `json:"base58PrefixExtendedPublicKey"`
	Base58PrefixExtendedSecretKey []byte `json:"base58PrefixExtendedSecretKey"`
	Bech32Prefix                  string `json:"bech32Prefix"`
	Bech32PrefixStealth           string `json:"bech32PrefixStealth"`
	Bech32PrefixAssets            string `json:"bech32PrefixAssets"`
}

var coins []Coin

func ensureCoinsLoaded() error {
	if len(coins) == 0 {
		box := packr.New("coinparams", "./data")

		coinsJson, err := box.Find("coinparams.json")
		if err != nil {
			return err
		}
		err = json.Unmarshal(coinsJson, &coins)
		if err != nil {
			return err
		}
	}
	return nil
}

func Coins() ([]Coin, error) {
	err := ensureCoinsLoaded()
	if err != nil {
		return []Coin{}, err
	}
	return coins, nil
}

func (cn CoinNetwork) Address(pubKey *btcec.PublicKey) string {
	pkh := util.KeyHashFromPubKey(pubKey)
	s, _ := bech32.SegWitV0Encode(cn.Bech32Prefix, pkh[:])
	return s
}

func (cn CoinNetwork) AssetsAddress(pubKey *btcec.PublicKey) string {
	pkh := util.KeyHashFromPubKey(pubKey)
	s, _ := bech32.SegWitV0Encode(cn.Bech32PrefixAssets, pkh[:])
	return s
}

func (cn CoinNetwork) StealthAddress(pubKey *btcec.PublicKey) string {
	return bech32.Encode(cn.Bech32PrefixStealth, pubKey.SerializeCompressed())
}
