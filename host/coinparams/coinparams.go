package coinparams

import (
	"encoding/json"

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
	Name             string `json:"name"`
	Id               string `json:"id"`
	DaemonParameters string `json:"daemonParameters"`
	Generate         bool   `json:"generate"`
	DaemonPort       int    `json:"daemonPort"`
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
