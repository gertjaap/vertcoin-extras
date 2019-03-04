package config

import (
	"fmt"
	"os"

	"github.com/go-ini/ini"
)

type Config struct {
	rpcChanged  chan bool
	Network     *Network
	RpcHost     string
	RpcUser     string
	RpcPassword string
	Port        uint16
	Cors        bool
	Donate      bool
}

func InitConfig(rpcChanged chan bool) (*Config, error) {
	c := new(Config)
	c.rpcChanged = rpcChanged
	err := c.Initialize()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) DefaultConfig() string {
	defaultConfig := ""
	defaultConfig += "network=testnet\n"
	defaultConfig += "rpchost=localhost:15888\n"
	defaultConfig += "rpcuser=vtc\n"
	defaultConfig += "rpcpassword=vtc\n"
	defaultConfig += "port=27888\n"
	defaultConfig += "cors=false\n"

	return defaultConfig
}

func (c *Config) WriteDefaultsIfNotExist() error {
	if _, err := os.Stat("vertcoin-extras.conf"); os.IsNotExist(err) {
		f, err := os.Create("vertcoin-extras.conf")
		if err != nil {
			return err
		}

		_, err = f.WriteString(c.DefaultConfig())
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) Read() error {
	cfg, err := ini.Load("vertcoin-extras.conf")
	if err != nil {
		return err
	}

	c.Network = GetNetwork(cfg.Section("").Key("network").String())
	c.RpcHost = cfg.Section("").Key("rpchost").String()
	c.RpcUser = cfg.Section("").Key("rpcuser").String()
	c.RpcPassword = cfg.Section("").Key("rpcpassword").String()
	c.Port = uint16(cfg.Section("").Key("port").MustInt(27888))
	c.Cors = cfg.Section("").Key("cors").MustBool(false)
	c.Donate = false

	return nil
}

func (c *Config) Initialize() error {
	err := c.WriteDefaultsIfNotExist()
	if err != nil {
		return err
	}

	err = c.Read()
	if err != nil {
		return err
	}

	return c.CheckValid()
}

func (c *Config) SetRpcCredentials(host, user, password string) error {
	c.RpcHost = host
	c.RpcUser = user
	c.RpcPassword = password
	c.rpcChanged <- true
	return c.Save()
}
func (c *Config) Save() error {
	cfg, err := ini.Load("vertcoin-extras.conf")
	if err != nil {
		return err
	}

	if c.Donate {
		cfg.Section("").Key("donate").SetValue("true")
	} else {
		cfg.Section("").Key("donate").SetValue("false")
	}

	cfg.Section("").Key("rpchost").SetValue(c.RpcHost)
	cfg.Section("").Key("rpcuser").SetValue(c.RpcUser)
	cfg.Section("").Key("rpcpassword").SetValue(c.RpcPassword)

	return cfg.SaveTo("vertcoin-extras.conf")
}

func (c *Config) CheckValid() error {
	if c.Network == nil {
		return fmt.Errorf("Network is not configured")
	}
	if c.RpcHost == "" {
		return fmt.Errorf("RPC host is not configured")
	}
	if c.RpcUser == "" {
		return fmt.Errorf("RPC user is not configured")
	}
	if c.RpcPassword == "" {
		return fmt.Errorf("RPC password is not configured")
	}
	return nil
}
