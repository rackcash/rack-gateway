package sol

import (
	"context"
	"fmt"
	"infra/blockchain/rack/config"
	"sync/atomic"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

const (
	SOL_COMISSION = 5000
)

var rpcUrl string
var wsUrl string

func Init(config *config.Config, currencies *atomic.Int32) (*rpc.Client, *ws.Client) {

	if config.RackCurrency.Sol.Testnet {
		rpcUrl = rpc.TestNet_RPC
		wsUrl = rpc.TestNet_WS
	} else {
		rpcUrl = rpc.MainNetBeta_RPC
		wsUrl = rpc.MainNetBeta_WS
	}

	client := rpc.New(rpcUrl)
	ws, err := ws.Connect(context.Background(), wsUrl)
	if err != nil {
		panic(err)
	}

	fmt.Printf("[%d] SOL connected: %s\n", currencies.Load(), rpcUrl)

	currencies.Add(1)

	return client, ws
}

func UpdateWs(oldWs *atomic.Pointer[ws.Client]) error {
	newWs, err := ws.Connect(context.Background(), wsUrl)
	if err != nil {
		return err
	}

	oldWsPtr := oldWs.Load()

	if oldWsPtr != nil {
		oldWsPtr.Close()
	}

	oldWs.Store(newWs)

	return nil
}

// func RunUpdateWs(oldWs *atomic.Pointer[ws.Client]) {
// /* 	for {
// 		err := UpdateWs(oldWs)
// 		if err != nil {
// 			slog.Debug(err.Error())
// 			time.Sleep(2 * time.Second)
// 			continue
// 		}
// 		time.Sleep(60 * time.Second)
// 	}

// }
