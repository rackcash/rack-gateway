package nats

import (
	"encoding/json"
	"infra/pkg/nats/natsdomain"
	"time"
)

func Unmarshal[T any](data []byte) (*T, error) {
	var unm T
	err := json.Unmarshal(data, &unm)
	if err != nil {
		return nil, err
	}
	return &unm, nil
}

// converts *currencies.Rates to nats.ReqGetRates
func RatesFormat(rates *natsdomain.Rates, err error) natsdomain.ResGetRates {
	if err != nil {
		return natsdomain.ResGetRates{
			Error: natsdomain.Error{
				IsError:   true,
				Message:   err.Error(),
				Timestamp: time.Now(),
			},
		}
	}

	return natsdomain.ResGetRates{
		Error: natsdomain.Error{
			IsError: false,
		},
		Rates: *rates,
	}
}

// func (app *App) ClientWithdraw(msg *nats.Msg, withdrawMsg *ReqClientWithdraw, finalAmountEther decimal.Decimal) {

// 	fmt.Println(withdrawMsg)
// 	switch withdrawMsg.Currency {
// 	case "eth":
// 		finalAmountWei := eth.EtherToWei(finalAmountEther)

// 		fmt.Println("FINAL AMOUNT ETHER", finalAmountEther)
// 		fmt.Println("FINAL AMOUNT WEI", finalAmountWei)

// 		tx, err := eth.CreateTx(app.EthClient, withdrawMsg.Address, withdrawMsg.Private, finalAmountWei, 3)
// 		if err != nil {
// 			fmt.Println(err)
// 			msg.Respond([]byte("error: " + err.Error()))
// 			return
// 		}

// 		fmt.Println("MERCHANT ID:", withdrawMsg.MerchantId)
// 		fmt.Println("PRIVATE:", withdrawMsg.Private)

// 		go func() {
// 			var withdrawResponse ResClientWithdraw
// 			withdrawResponse.Address = withdrawMsg.Address
// 			// withdrawResponse.Amount = finalAmountEther
// 			withdrawResponse.Comission = withdrawMsg.Comission
// 			withdrawResponse.Currency = withdrawMsg.Currency
// 			withdrawResponse.MerchantId = withdrawMsg.MerchantId
// 			withdrawResponse.UserId = withdrawMsg.UserId
// 			withdrawResponse.Private = withdrawMsg.Private

// 			sent, err := eth.SendTx(app.EthClient, tx)
// 			if err != nil { // error
// 				app.clientWithdrawHandleError(err, withdrawResponse)
// 				return
// 			}

// 			fromAddress := eth.PrivateToAddress(withdrawMsg.Private)
// 			fmt.Println("BALANCE AFTER")
// 			fmt.Println(eth.GetBalanceEther(app.EthClient, fromAddress.Hex()))

// 			// success

// 			withdrawResponse.Status = "sent"
// 			withdrawResponse.Amount = *sent
// 			withdrawResponse.TxHash = tx.Hash().String()
// 			// withdrawResponse.Error.Msg = err.Error()

// 			data, err := json.Marshal(withdrawResponse)
// 			if err != nil {
// 				fmt.Println("Error: ", err)
// 				return
// 			}

// 			err = app.Ns.Nc.Publish("client.withdraw.response", data)
// 			if err != nil {
// 				fmt.Println("Error: ", err)
// 				return
// 			}
// 		}()

// 		msg.Respond([]byte("ok"))

// 	default:
// 		msg.Respond([]byte("error: invalid currency"))

// 	}
// }

// func (app *App) clientWithdrawHandleError(err error, withdrawResponse ResClientWithdraw) {
// 	withdrawResponse.Status = "withdraw_error"
// 	withdrawResponse.Amount = decimal.NewFromInt(0)
// 	withdrawResponse.Error.Msg = err.Error()

// 	data, err := json.Marshal(withdrawResponse)
// 	if err != nil {
// 		fmt.Println("Error: ", err)
// 		return
// 	}

// 	err = app.Ns.Nc.Publish("server.withdraw.response", data)
// 	if err != nil {
// 		fmt.Println("Error: ", err)
// 		return
// 	}
// }

// var withdrawLock sync.Map

// func IsWithdrawLocked(invoice *MsgServerWithdraw) bool {
// 	// withdrawLock

// 	return false

// }

// func WithrawLock(privateKey string) {
// 	withdrawLock.Store(privateKey, true)
// }

// func WithdrawUnlock() {
// 	// FIXME: fix
// 	// withdrawLock.Store(privat/*  */eKey, false)

// }
