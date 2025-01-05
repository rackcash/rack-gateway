package config

import (
	"sync"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Prod_env bool `envconfig:"PROD_ENV"          required:"true"`
	Nats     struct {
		// TomlServers []string `toml:"servers"`
		Servers string `envconfig:"NATS_SERVERS"          required:"true"`
	}
	RackCurrency struct {
		CoinmarketApi string `envconfig:"COINMARKET_API"          required:"true"`
		Ton           struct {
			Testnet bool `envconfig:"TON_TESTNET"          required:"true"`
		}
		Sol struct {
			Testnet bool `envconfig:"SOL_TESTNET"          required:"true"`
		}
		Eth struct {
			Testnet bool   `envconfig:"ETH_TESTNET"          required:"true"`
			ApiKey  string `envconfig:"ETH_RPC_KEY"          required:"true"`
		}
	}
}

var once sync.Once

func ReadConfig() *Config {

	var config Config

	once.Do(func() {
		if err := envconfig.Process("", &config); err != nil {
			panic(err)
		}
	})

	// user, err := os.ReadFile(os.Getenv("SECRETS") + "/nats-user.txt")
	// if err != nil {
	// 	panic(err)
	// }

	// pass, err := os.ReadFile(os.Getenv("SECRETS") + "/nats-password.txt")
	// if err != nil {
	// 	panic(err)
	// }

	// var formatedServers string
	// for _, x := range config.Nats.TomlServers {
	// 	connectUrl := fmt.Sprintf("nats://%s:%s@%s,", user, pass, string(x))
	// 	formatedServers += connectUrl
	// }

	// config.Nats.Servers = formatedServers

	return &config
}
