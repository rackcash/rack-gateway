package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"infra/blockchain/rack/currencies/eth"
	"infra/blockchain/rack/currencies/sol"
	"infra/blockchain/rack/currencies/ton"
	"infra/pkg/nats/natsdomain"
	"log"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/shopspring/decimal"
	"github.com/xssnick/tonutils-go/address"
)

func (app *App) ServerWithdrawHandler(msg jetstream.Msg) {

	withdrawMsg, err := Unmarshal[natsdomain.ReqWithdrawal](msg.Data())
	if err != nil {
		fmt.Println("error: " + err.Error())
		// msg.Respond([]byte("error: " + err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	switch withdrawMsg.Crypto {
	case "ton":
		wallet, err := ton.GetWalletBySeed(app.Ton.Client, withdrawMsg.Private)
		if err != nil {
			fmt.Println(withdrawMsg.Private)
			slog.Debug(err.Error())
			return
		}

		final, err := ton.GetBalance(app.Ton.Client, &app.Ton.BlockID, wallet.WalletAddress())
		if err != nil {
			slog.Debug(err.Error())
			return
		}
		app.ServerWithdraw(msg, withdrawMsg, final)
	case "eth":
		address := eth.PrivateToAddress(withdrawMsg.Private)

		final, err := eth.GetBalanceEther(app.EthClient, address.Hex())
		if err != nil {
			slog.Debug(err.Error())
			return
		}
		app.ServerWithdraw(msg, withdrawMsg, final)
	case "sol":
		final, err := sol.GetBalanceLamports(ctx, app.Sol.Client, withdrawMsg.TxTempWallet)
		if err != nil {
			slog.Debug(err.Error(), "address", withdrawMsg.Address)
			return
		}

		app.ServerWithdraw(msg, withdrawMsg, decimal.NewFromUint64(final))
	default:
		fmt.Println("invalid cryptocurrency: " + withdrawMsg.Crypto)
		return

	}
}

// withdrawal from tg
func (app *App) MerchantWithdrawalHandler(msg jetstream.Msg) {
	withdrawMsg, err := Unmarshal[natsdomain.ReqMerchantWithdrawal](msg.Data())
	if err != nil {
		fmt.Println("error: " + err.Error())
		// msg.Respond([]byte("error: " + err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	switch withdrawMsg.Crypto {
	case "ton":
		wallet, err := ton.GetWalletBySeed(app.Ton.Client, withdrawMsg.Private)
		if err != nil {
			fmt.Println(withdrawMsg.Private)
			slog.Debug(err.Error())
			return
		}

		final, err := ton.GetBalance(app.Ton.Client, &app.Ton.BlockID, wallet.WalletAddress())
		if err != nil {
			slog.Debug(err.Error())
			return
		}
		app.MerchantWithdrawal(msg, withdrawMsg, final)
	case "eth":
		address := eth.PrivateToAddress(withdrawMsg.Private)

		final, err := eth.GetBalanceEther(app.EthClient, address.Hex())
		if err != nil {
			slog.Debug(err.Error())
			return
		}
		app.MerchantWithdrawal(msg, withdrawMsg, final)
	case "sol":
		final, err := sol.GetBalanceLamports(ctx, app.Sol.Client, withdrawMsg.FromAddress)
		if err != nil {
			slog.Debug(err.Error(), "address", withdrawMsg.ToAddress)
			return
		}

		app.MerchantWithdrawal(msg, withdrawMsg, decimal.NewFromUint64(final))
	default:
		fmt.Println("invalid cryptocurrency: " + withdrawMsg.Crypto)
		return

	}
}

func (app *App) NewWalletHandler(msg *nats.Msg) {
	newWallet, err := Unmarshal[natsdomain.ReqNewWallet](msg.Data)
	if err != nil {
		fmt.Println("error: " + err.Error())
		msg.Respond([]byte("error: " + err.Error()))
		return
	}

	address, privateKey, err := app.NewWallet(newWallet.Cryptocurrency)
	if err != nil {
		fmt.Println("error: " + err.Error())
		msg.Respond([]byte("error: " + err.Error()))
		return
	}

	data, err := json.Marshal(natsdomain.ResNewWallet{Address: address, PrivateKey: privateKey})
	if err != nil {
		fmt.Println("error: " + err.Error())
		msg.Respond([]byte("error: " + err.Error()))
		return
	}

	err = msg.Respond(data)
	if err != nil {
		fmt.Println("error: " + err.Error())
		return
	}
}

func (app *App) IsPaidHandler(msg *nats.Msg) {
	var resIsPaid natsdomain.ResIsPaid

	dataJson, err := Unmarshal[natsdomain.ReqIsPaid](msg.Data)
	if err != nil {
		fmt.Println("error Unmarshal[natsdomain.ReqIsPaid](msg.Data): ", err)
		msg.Respond([]byte("error: " + err.Error()))
		return
	}

	balance, err := app.GetBalance(dataJson.Cryptocurrency, dataJson.Address)
	if err != nil {
		fmt.Println("GetBalance error: ", err)
		msg.Respond([]byte("error: " + err.Error()))
		return
	}

	resIsPaid.Amount = balance

	if resIsPaid.Amount.Equal(decimal.NewFromInt(0)) { // not paid
		resIsPaid.Status = "not_paid"
		resIsPaid.Paid = false
	}

	if resIsPaid.Amount.LessThan(dataJson.TargetAmount) && !resIsPaid.Amount.Equal(decimal.NewFromInt(0)) { // paid less
		resIsPaid.Status = "paid_less"
		resIsPaid.Paid = false
	}

	if resIsPaid.Amount.Equal(dataJson.TargetAmount) { // paid
		resIsPaid.Status = "paid"
		resIsPaid.Paid = true
	}

	if resIsPaid.Amount.GreaterThan(dataJson.TargetAmount) { // paid over
		resIsPaid.Status = "paid_over"
		resIsPaid.Paid = true
	}

	data, err := json.Marshal(&resIsPaid)
	if err != nil {
		log.Printf("Marshal Error %v: %s, %t, %d\n", err, resIsPaid.Status, resIsPaid.Paid, resIsPaid.Amount)
		msg.Respond([]byte("error: " + err.Error()))
		return
	}
	msg.Respond([]byte(data))

}

func (app *App) GetTxStatusHandler(msg *nats.Msg) {
	getTx, err := Unmarshal[natsdomain.ReqGetTxStatus](msg.Data)
	if err != nil {
		fmt.Println("Unmarshal[MsgGetTxStatus]: ", err)
		msg.Respond([]byte("error: " + err.Error()))
		return
	}

	var success bool
	var isPending bool
	var amount decimal.Decimal
	// TODO: add more cryptocurrencies
	switch getTx.Cryptocurrency {
	case "sol":
		if getTx.SearchBy != natsdomain.SearchByHash {
			fmt.Println("invalid 'search by': " + getTx.SearchBy.String())
			msg.Respond([]byte("error: invalid 'search by': " + getTx.SearchBy.String()))
			return
		}

		_amount, suc, err := sol.GetTxByHash(app.Sol.Client, getTx.TxHash)
		if err != nil {
			slog.Debug(err.Error())
			msg.Respond([]byte("error: " + err.Error()))
			return
		}
		amount = _amount
		success = suc
		isPending = false
	case "eth":
		if getTx.SearchBy != natsdomain.SearchByHash {
			fmt.Println("invalid 'search by': " + getTx.SearchBy.String())
			msg.Respond([]byte("error: invalid 'search by': " + getTx.SearchBy.String()))
			return
		}

		tx, suc, isPend, err := eth.GetTxByHash(app.EthClient, getTx.TxHash, getTx.TxTempWallet)
		if err != nil {
			slog.Debug(err.Error())
			msg.Respond([]byte("error: " + err.Error()))
			return
		}
		amount = *eth.WeiToEther(tx.Value())
		success = suc
		isPending = isPend
	case "ton":
		if getTx.SearchBy != natsdomain.SearchByAddress {
			fmt.Println("invalid 'search by': " + getTx.SearchBy.String())
			msg.Respond([]byte("error: invalid 'search by': " + getTx.SearchBy.String()))
			return
		}

		addr, err := address.ParseAddr(getTx.TxTempWallet)
		if err != nil {
			fmt.Println(err)
			msg.Respond([]byte("error: " + err.Error()))
			return
		}

		tx, err := ton.GetTxByAddress(app.Ton.Client, app.Ton.BlockID.Load(), addr)
		if err != nil {
			slog.Debug(err.Error())
			msg.Respond([]byte("error: " + err.Error()))
			return
		}

		success = tx.IsOut // cause true == sent
		if tx.ToAddr != getTx.BalanceAddress {
			fmt.Println("tx.ToAddr != getTx.BalanceAddress")
			success = false
		}

		amount = tx.Amount
		isPending = false
	default:
		msg.Respond([]byte("error: invalid cryptocurrency"))
		return
	}

	balance, err := app.GetBalance(getTx.Cryptocurrency, getTx.TxTempWallet)
	if err != nil {
		fmt.Println(err)
		msg.Respond([]byte("error: " + err.Error()))
		return
	}

	data, err := json.Marshal(natsdomain.ResGetTxStatus{TxHash: getTx.TxHash, IsPending: isPending, Balance: balance, Amount: amount, Success: success})
	if err != nil {
		fmt.Println(err)
		msg.Respond([]byte("error: " + err.Error()))
		return
	}
	msg.Respond(data)
	fmt.Println(string(data))

}

func (app *App) GetBalanceHandler(msg *nats.Msg) {
	getBalance, err := Unmarshal[natsdomain.ReqGetBalance](msg.Data)
	if err != nil {
		msg.Respond([]byte("error: " + err.Error()))
		return
	}

	balance, err := app.GetBalance(getBalance.Cryptocurrency, getBalance.Address)
	if err != nil {
		msg.Respond([]byte("error: " + err.Error()))
		return
	}

	fmt.Println("balance: ", balance)

	data, err := json.Marshal(natsdomain.ResGetBalance{Cryptocurrency: getBalance.Cryptocurrency, Balance: balance})
	if err != nil {
		msg.Respond([]byte("error: " + err.Error()))
		return
	}

	err = msg.Respond(data)
	if err != nil {
		msg.Respond([]byte("error: " + err.Error()))
		return
	}
}
