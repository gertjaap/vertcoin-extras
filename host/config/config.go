package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/gertjaap/vertcoin/logging"

	"github.com/gertjaap/vertcoin/util"
)

type Config struct {
	EnabledDaemons []EnabledDaemonConfig
}

type EnabledDaemonConfig struct {
	CoinId    string
	NetworkId string
}

func InitConfig() (*Config, error) {
	err := WriteDefaultsIfNotExist()
	if err != nil {
		return nil, err
	}
	return Read()
}

func DefaultConfig() *Config {
	var cfg = new(Config)
	cfg.EnabledDaemons = make([]EnabledDaemonConfig, 3)
	cfg.EnabledDaemons[0] = EnabledDaemonConfig{CoinId: "vtc", NetworkId: "regtest"}
	cfg.EnabledDaemons[1] = EnabledDaemonConfig{CoinId: "ltc", NetworkId: "regtest"}
	cfg.EnabledDaemons[2] = EnabledDaemonConfig{CoinId: "btc", NetworkId: "regtest"}
	return cfg
}

func ConfigFile() string {
	return path.Join(util.DataDirectory(), fmt.Sprintf("%s.conf", strings.ToLower(util.APP_NAME)))
}

func WriteDefaultsIfNotExist() error {
	if _, err := os.Stat(ConfigFile()); os.IsNotExist(err) {
		logging.Debug("Config file does not exist, creating default")
		err = DefaultConfig().Save()
		if err != nil {
			return err
		}
	}
	return nil
}

func Read() (*Config, error) {
	logging.Debugf("Reading config file: %s", ConfigFile())
	jsonConfig, err := ioutil.ReadFile(ConfigFile())
	if err != nil {
		return nil, err
	}

	c := new(Config)
	logging.Debug("Parsing config file")
	err = json.Unmarshal(jsonConfig, &c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) Save() error {
	logging.Debug("Serializing config")
	configJson, err := json.Marshal(c)
	if err != nil {
		return err
	}
	logging.Debugf("Saving config to %s", ConfigFile())
	ioutil.WriteFile(ConfigFile(), configJson, 0700)
	return nil
}

func (c *Config) CheckValid() error {

	return nil
}
