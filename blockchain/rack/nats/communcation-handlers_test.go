package nats

import (
	"fmt"
	"infra/blockchain/rack/currencies/eth"
	"infra/blockchain/rack/currencies/ton"
	"infra/pkg/nats/natsdomain"
	"testing"

	"github.com/nats-io/nats.go"
)

func TestGetTxStatusHandler_Eth(t *testing.T) {

	txHash := "0x"
	txTempWallet := "0x"
	balanceAddress := ""
	searchBy := natsdomain.SearchByHash
	cryptocurrency := "eth"

	data := fmt.Sprintf(`{"TxHash": "%s", "TxTempWallet": "%s", "BalanceAddress": "%s", "SearchBy": %d, "Cryptocurrency": "%s"}`, txHash, txTempWallet, balanceAddress, searchBy, cryptocurrency)

	msg := &nats.Msg{
		Data: []byte(data),
	}

	app := App{
		EthClient: eth.Connect("https://holesky.infura.io/v3/" /* infura api key */),
	}

	app.GetTxStatusHandler(msg)

}

func TestGetTxStatusHandler_Ton(t *testing.T) {

	txHash := ""
	txTempWallet := "0x"
	balanceAddress := "0x"
	searchBy := natsdomain.SearchByAddress
	cryptocurrency := "ton"

	data := fmt.Sprintf(`{"TxHash": "%s", "TxTempWallet": "%s", "BalanceAddress": "%s", "SearchBy": %d, "Cryptocurrency": "%s"}`, txHash, txTempWallet, balanceAddress, searchBy, cryptocurrency)

	msg := &nats.Msg{
		Data: []byte(data),
	}

	tonClient, blockid := ton.ConnectTest(ton.TESTNET_URL)

	tonConfig := Ton{Client: tonClient}
	tonConfig.BlockID.Store(blockid)

	app := App{
		Ton: &tonConfig,
	}

	app.GetTxStatusHandler(msg)

}
