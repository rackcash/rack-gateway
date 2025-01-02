package nats

import (
	"encoding/json"
	"fmt"
	"infra/blockchain/rack/config"
	"infra/blockchain/rack/currencies"
	"infra/pkg/nats/natsdomain"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func (app *App) natsCoreHandler(msg *nats.Msg) {

	switch msg.Subject {
	case natsdomain.SubjGetRates.String():
		dataJson, err := Unmarshal[natsdomain.ReqGetRates](msg.Data)
		if err != nil {
			msg.Respond([]byte("error: " + err.Error()))
			return
		}
		r, err := currencies.GetRates(app.Config, dataJson.ToCurrency)

		data, err := json.Marshal(RatesFormat(r, err))
		if err != nil {
			msg.Respond([]byte("error: " + err.Error()))
			return
		}

		fmt.Println(string(data))
		err = msg.Respond(data)
		if err != nil {
			msg.Respond([]byte("error: " + err.Error()))
			return
		}

	case natsdomain.SubjIsPayed.String():
		app.IsPaidHandler(msg)
	case natsdomain.SubjGetBalance.String():
		app.GetBalanceHandler(msg)
	case natsdomain.SubjGetTxStatus.String():
		app.GetTxStatusHandler(msg)
	case natsdomain.SubjNewWallet.String():
		app.NewWalletHandler(msg)
	case natsdomain.SubjPing.String():
		msg.Respond([]byte("pong"))
	}

}

func (app *App) consumerHandler(msg jetstream.Msg) {

	meta, _ := msg.Metadata()
	if meta != nil {
		if meta.NumDelivered > 6 {
			fmt.Println("Too many deliveries", meta.NumDelivered)
			msg.Ack()
			return
		}
	}

	switch msg.Subject() {
	case natsdomain.SubjJsWithdraw.String():
		fmt.Println("subject: ", msg.Subject())
		app.ServerWithdrawHandler(msg)
	case natsdomain.SubjJsMerchantWithdrawal.String():
		fmt.Println("subject: ", msg.Subject())
		app.MerchantWithdrawalHandler(msg)

	default:
		fmt.Println("invalid subject: " + msg.Subject())
	}
}

const WORKERS = 10

func (app *App) Run(c *config.Config, ns *natsdomain.Ns) {

	_, err := app.C.Consume(app.consumerHandler)
	if err != nil {
		fmt.Println("Consume error: ", err)
		return
	}

	//  nats core

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for range WORKERS {
			_, err := ns.Nc.QueueSubscribe("currencies.core.*", "currency_workers", app.natsCoreHandler)
			if err != nil {
				fmt.Println("QueueSubscribe error: ", err)
				wg.Done()
				break
			}
		}
	}()
	wg.Wait()

	// jetstream
}
