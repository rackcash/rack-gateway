package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Prod_env bool
	Nats     struct {
		TomlServers []string `toml:"servers"`
		Servers     string
	}
	RackCurrency struct {
		GlobalTestnet bool   `toml:"global_testnet"`
		CoinmarketApi string `toml:"coinmarket_api"`
		Ton           struct {
			Testnet bool
		}
		Sol struct {
			Testnet bool
		}
		Eth struct {
			Testnet bool
			ApiKey  string `toml:"api_key"`
		}
	} `toml:"rack_currency"`
}

func ReadConfig() *Config {

	byte_config, err := os.ReadFile(os.Getenv("CONFIG"))
	if err != nil {
		panic(err)
	}

	var config Config
	_, err = toml.Decode(string(byte_config), &config)
	if err != nil {
		panic(err)
	}

	user, err := os.ReadFile(os.Getenv("SECRETS") + "/nats-user.txt")
	if err != nil {
		panic(err)
	}

	pass, err := os.ReadFile(os.Getenv("SECRETS") + "/nats-password.txt")
	if err != nil {
		panic(err)
	}

	var formatedServers string
	for _, x := range config.Nats.TomlServers {
		connectUrl := fmt.Sprintf("nats://%s:%s@%s,", user, pass, string(x))
		formatedServers += connectUrl
	}

	config.Nats.Servers = formatedServers

	return &config
}
