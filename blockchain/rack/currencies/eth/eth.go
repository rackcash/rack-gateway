package eth

import (
	"context"
	"fmt"
	"infra/blockchain/rack/config"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/ethclient"
)

func Init(config *config.Config, currencies *atomic.Int32) *ethclient.Client {
	var url string
	if config.RackCurrency.Eth.Testnet {
		url = "https://holesky.infura.io/v3/"
	} else {
		url = "https://mainnet.infura.io/v3/"
	}

	// url = "http://127.0.0.1:8545/"

	fmt.Printf("[%d] ETH connected: %s\n", currencies.Load(), url)

	client := Connect(url + config.RackCurrency.Eth.ApiKey)

	currencies.Add(1)

	return client
}

func Connect(url string) *ethclient.Client {

	client, err := ethclient.Dial(url)
	if err != nil {
		panic("Can't connect: " + err.Error())
	}

	_, err = client.ChainID(context.Background())
	if err != nil {
		panic("Can't connect: " + err.Error())
	}

	return client

}
