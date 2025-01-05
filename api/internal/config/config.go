package config

import (
	"os"
	"strings"
	"sync"

	"github.com/kelseyhightower/envconfig"
	"gorm.io/gorm"
)

type Config struct {
	DB *gorm.DB `ignored:"true"`

	// TODO: remove
	// ProxyPath string   `toml:"proxy_path"` // used in webhook-sender
	// ProxyList []string `toml:"-"`          // reads proxies from ProxyPath and fills it with

	Prod_env bool `envconfig:"PROD_ENV"          required:"true"`

	// Telegram struct {
	// 	Token string `toml:"token"`
	// }

	Testing struct {
		Enabled              bool `envconfig:"TESTING"          required:"true"`
		TxConfirmDelay       int  `envconfig:"TESTING_TX_CONFIRM_DELAY"          required:"true"`
		TxFinProcessingDelay int  `envconfig:"TX_FIN_PROCESSING_DELAY"          required:"true"`
		TxFinDelay           int  `envconfig:"TX_FIN_DELAY"          required:"true"`
	}

	Postgres struct {
		Host     string `envconfig:"DB_HOST"          required:"true"`
		User     string `envconfig:"DB_USER"          required:"true"`
		Password string `envconfig:"DB_PASSWORD"          required:"true"`
		Db_name  string `envconfig:"DB_NAME"          required:"true"`
		Port     uint16 `envconfig:"DB_PORT"          required:"true"`
		Ssl_mode string `envconfig:"DB_SSL_MODE"          required:"true"`
	}
	Nats struct {
		Servers string `envconfig:"NATS_SERVERS"          required:"true"`
		// TomlServers []string `toml:"servers"`
	}

	PrivateKey string `envconfig:"API_ADMIN_KEY"          required:"true"`
	Api        struct {
		Ipv4  string `envconfig:"API_IPV4"          required:"true"`
		Proto string `envconfig:"API_PROTO"          required:"true"`
	}
}

type ConfigS struct {
	DB *gorm.DB `ignored:"true"`

	// TODO: remove
	// ProxyPath string   `toml:"proxy_path"` // used in webhook-sender
	// ProxyList []string `toml:"-"`          // reads proxies from ProxyPath and fills it with

	Prod_env bool `envconfig:"PROD_ENV"          required:"true"`

	// Telegram struct {
	// 	Token string `toml:"token"`
	// }

	Testing struct {
		Enabled              bool `envconfig:"TESTING"          required:"true"`
		TxConfirmDelay       int  `envconfig:"TESTING_TX_CONFIRM_DELAY"          required:"true"`
		TxFinProcessingDelay int  `envconfig:"TX_FIN_PROCESSING_DELAY"          required:"true"`
		TxFinDelay           int  `envconfig:"TX_FIN_DELAY"          required:"true"`
	}

	// PrivateKey string `toml:"private_key"`
	Postgres struct {
		Host     string `envconfig:"DB_HOST"          required:"true"`
		User     string `envconfig:"DB_USER"          required:"true"`
		Password string `envconfig:"DB_PASSWORD"          required:"true"`
		Db_name  string `envconfig:"DB_NAME"          required:"true"`
		Port     uint16 `envconfig:"DB_PORT"          required:"true"`
		Ssl_mode string `envconfig:"DB_SSL_MODE"          required:"true"`
	}
	Nats struct {
		Servers string `envconfig:"NATS_SERVERS"          required:"true"`
		// TomlServers []string `toml:"servers"`
	}
	Api struct {
		Ipv4  string `envconfig:"API_IPV4"          required:"true"`
		Proto string `envconfig:"API_PROTO"          required:"true"`
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

	// webhook proxies
	// config.ProxyList = GetProxyList(config.ProxyPath)

	if config.Prod_env && config.Testing.Enabled {
		panic("cannot use testing in prod")
	}

	return &config
}

func GetProxyList(path string) []string {
	proxyList, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	proxyListArray := strings.Split(string(proxyList), "\n")
	return proxyListArray
}
