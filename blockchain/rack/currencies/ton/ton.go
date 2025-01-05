package ton

import (
	"context"
	"fmt"
	"infra/blockchain/rack/config"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
)

const (
	TESTNET_URL = "https://ton.org/testnet-global.config.json"
	DEFAULT_URL = "https://ton.org/global.config.json"

	// TESTNET_CACHE = "blockchain/rack/currencies/ton/testnet-global-config.json"
	// DEFAULT_CACHE = "blockchain/rack/currencies/ton/global-config.json"
)

func Init(config *config.Config, currencies *atomic.Int32) (ton.APIClientWrapped, *ton.BlockIDExt) {

	client := liteclient.NewConnectionPool()
	// https://ton-blockchain.github.io/global.config.json

	// connect to testnet lite server
	cfg, err := getTonConfig(config)
	if err != nil {
		panic("get config err: " + err.Error())
	}

	err = client.AddConnectionsFromConfig(context.Background(), cfg)
	if err != nil {
		panic("connection err: " + err.Error())
	}

	api := ton.NewAPIClient(client, ton.ProofCheckPolicyFast).WithRetry(2)
	api.SetTrustedBlockFromConfig(cfg)

	master, err := api.CurrentMasterchainInfo(context.Background()) // we fetch block just to trigger chain proof check
	if err != nil {
		panic("get masterchain info err: " + err.Error())
	}

	fmt.Printf("[%d] TON connected\n", currencies.Load())
	currencies.Add(1)

	return api, master
}

func ConnectTest(url string) (ton.APIClientWrapped, *ton.BlockIDExt) {
	client := liteclient.NewConnectionPool()

	err := client.AddConnectionsFromConfigUrl(context.Background(), url)
	if err != nil {
		panic("connection err: " + err.Error())
	}

	api := ton.NewAPIClient(client, ton.ProofCheckPolicyFast).WithRetry(2)
	// api.SetTrustedBlockFromConfig(cfg)

	master, err := api.CurrentMasterchainInfo(context.Background()) // we fetch block just to trigger chain proof check
	if err != nil {
		panic("get masterchain info err: " + err.Error())
	}

	return api, master

}

func getTonConfig(config *config.Config) (*liteclient.GlobalConfig, error) {

	// TODO: mainnet config
	if config.RackCurrency.Ton.Testnet {
		return liteclient.GetConfigFromUrl(context.Background(), TESTNET_URL)

		// r, err := http.Get(TESTNET_URL)
		// if err != nil {
		// 	fmt.Println("Can't connect to testnet")
		// 	return liteclient.GetConfigFromUrl(context.Background(), TESTNET_URL)
		// }
		// defer r.Body.Close()

		// body, err := io.ReadAll(r.Body)
		// if err != nil {
		// 	panic(err)
		// }

		// _ = os.WriteFile(TESTNET_CACHE, body, os.ModePerm)

		// return liteclient.GetConfigFromFile(TESTNET_CACHE)
	}
	return liteclient.GetConfigFromUrl(context.Background(), DEFAULT_URL)

}

func RunUpdateBlock(client ton.APIClientWrapped, oldBlock *atomic.Pointer[ton.BlockIDExt]) {
	for {
		err := UpdateBlockOnce(client, oldBlock)
		if err != nil {
			slog.Debug(err.Error())
			time.Sleep(2 * time.Second)
			continue
		}
		time.Sleep(5 * time.Second)
	}

}

func UpdateBlockOnce(client ton.APIClientWrapped, oldBlock *atomic.Pointer[ton.BlockIDExt]) error {
	master, err := client.CurrentMasterchainInfo(context.Background()) // we fetch block just to trigger chain proof check
	if err != nil {
		return err
	}

	// log.Println("master proof checks are completed successfully, now communication is 100% safe!")
	oldBlock.Store(master)
	return nil
}
