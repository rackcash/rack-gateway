package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"gorm.io/gorm"
)

type Config struct {
	DB *gorm.DB

	ProxyPath string   `toml:"proxy_path"` // used in webhook-sender
	ProxyList []string `toml:"-"`          // reads proxies from ProxyPath and fills it with

	Prod_env bool

	Telegram struct {
		Token string `toml:"token"`
	}

	Testing struct {
		Enabled              bool
		TxConfirmDelay       time.Duration `toml:"tx_confirm_delay"`
		TxFinProcessingDelay time.Duration `toml:"tx_fin_processing_delay"`
		TxFinDelay           time.Duration `toml:"tx_fin_delay"`
	} `toml:"testing"`

	PrivateKey string `toml:"private_key"`
	Postgres   struct {
		Host     string
		User     string
		Password string
		Db_name  string
		Port     uint16
		Ssl_mode string
	}
	Nats struct {
		Servers     string
		TomlServers []string `toml:"servers"`
	}
	Api struct {
		Ipv4  string
		Proto string
	} `toml:"rack_web"`
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

	fmt.Println("PATH", config.ProxyPath)

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

	// webhook proxies
	config.ProxyList = GetProxyList(config.ProxyPath)

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
