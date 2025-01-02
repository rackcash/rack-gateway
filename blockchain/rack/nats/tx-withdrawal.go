package nats

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"infra/blockchain/rack/currencies/eth"
	"infra/blockchain/rack/currencies/sol"
	"infra/blockchain/rack/currencies/ton"
	"infra/pkg/nats/natsdomain"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/shopspring/decimal"
	"github.com/xssnick/tonutils-go/address"
)

func (app *App) ServerWithdraw(msg jetstream.Msg, withdrawMsg *natsdomain.ReqWithdrawal, finalAmount decimal.Decimal) {
	defer msg.Ack()

	var attemptsLimit int = 2

	md, _ := msg.Metadata()
	if md != nil {
		fmt.Println("NUM DELIVERD", md.NumDelivered)

		if md.NumDelivered > 1 { //  if the server goes down, give it one try
			attemptsLimit = 0
		}
	}

	// TODO: add more cryptocurrencies
	switch withdrawMsg.Crypto {
	case "sol":
		app.withdrawSol(withdrawMsg, finalAmount)
	case "ton":
		app.withdrawTon(withdrawMsg, finalAmount)
	case "eth":
		app.withdrawEth(withdrawMsg, finalAmount, attemptsLimit)
	default:
		fmt.Println("error: invalid currency")
	}
}

func (app *App) withdrawEth(withdrawMsg *natsdomain.ReqWithdrawal, finalAmount decimal.Decimal, attemptsLimit int) {
	var withdrawResponse natsdomain.ResWithdrawal

	finalAmountWei := eth.EtherToWei(finalAmount)

	fmt.Println("FINAL AMOUNT ETHER", finalAmount)
	fmt.Println("FINAL AMOUNT WEI", finalAmountWei)

	tx, err := eth.CreateTx(app.EthClient, withdrawMsg.Address, withdrawMsg.Private, finalAmountWei, attemptsLimit)
	if err != nil {
		slog.Debug(err.Error())
		app.serverWithdrawTxError(err, withdrawMsg, finalAmount, "")
		return
	}
	fmt.Println("TX HASH WAAAA", tx.Hash().String())

	app.serverWithdrawInfo(withdrawMsg, finalAmount, tx.Hash().String())

	go func() {
		withdrawResponse.MerchantId = withdrawMsg.MerchantId
		withdrawResponse.InvoiceId = withdrawMsg.InvoiceId
		withdrawResponse.Crypto = withdrawMsg.Crypto
		withdrawResponse.Address = withdrawMsg.Address

		sent, err := eth.SendTx(app.EthClient, tx)
		if err != nil { // error
			slog.Debug(err.Error())
			app.serverWithdrawHandleError(err, withdrawResponse, natsdomain.NewMsgId(withdrawMsg.InvoiceId, natsdomain.MsgActionError))
			return
		}

		fromAddress := eth.PrivateToAddress(withdrawMsg.Private)
		fmt.Println("BALANCE AFTER")
		fmt.Println(eth.GetBalanceEther(app.EthClient, fromAddress.Hex()))

		app.serverWithdrawTxSuccess(withdrawResponse, withdrawMsg.Status, *sent, tx.Hash().String(), withdrawMsg.TxTempWallet)
	}()

}

func (app *App) withdrawTon(withdrawMsg *natsdomain.ReqWithdrawal, finalAmount decimal.Decimal) {
	var withdrawResponse natsdomain.ResWithdrawal

	wallet, err := ton.GetWalletBySeed(app.Ton.Client, withdrawMsg.Private)
	if err != nil {
		slog.Debug(err.Error())
		app.serverWithdrawTxError(err, withdrawMsg, finalAmount, "")
		return
	}

	toAddr, err := address.ParseAddr(withdrawMsg.Address)
	if err != nil {
		slog.Debug(err.Error())
		app.serverWithdrawTxError(err, withdrawMsg, finalAmount, "")
		return
	}

	// final amount 0, because the mode will be set to send the entire balance
	// finalAmount - balance
	tx, err := ton.CreateTX(wallet, toAddr, decimal.NewFromInt(0), finalAmount)
	if err != nil {
		slog.Debug(err.Error())
		app.serverWithdrawTxError(err, withdrawMsg, finalAmount, "")
		return
	}

	app.serverWithdrawInfo(withdrawMsg, finalAmount, "")

	go func() {
		withdrawResponse.MerchantId = withdrawMsg.MerchantId
		withdrawResponse.InvoiceId = withdrawMsg.InvoiceId
		withdrawResponse.Crypto = withdrawMsg.Crypto
		withdrawResponse.Address = withdrawMsg.Address

		tx, err := ton.SendTx(wallet, tx)
		if err != nil {
			slog.Debug(err.Error())
			app.serverWithdrawTxError(err, withdrawMsg, finalAmount, "")
			return
		}

		fmt.Println("SENT. Hash: ", base64.StdEncoding.EncodeToString(tx.Hash))

		app.serverWithdrawTxSuccess(withdrawResponse, withdrawMsg.Status, finalAmount, base64.StdEncoding.EncodeToString(tx.Hash), withdrawMsg.TxTempWallet)
	}()
}

func (app *App) withdrawSol(withdrawMsg *natsdomain.ReqWithdrawal, finalAmount decimal.Decimal) {
	var withdrawResponse natsdomain.ResWithdrawal

	pbc, err := sol.StringToPBC(withdrawMsg.Address)
	if err != nil {
		slog.Debug(err.Error())
		app.serverWithdrawTxError(err, withdrawMsg, finalAmount, "")
		return
	}

	private, err := sol.StringToPriv(withdrawMsg.Private)
	if err != nil {
		slog.Debug(err.Error())
		app.serverWithdrawTxError(err, withdrawMsg, finalAmount, "")
		return
	}

	tx, txHash, err := sol.CreateTx(app.Sol.Client, finalAmount.BigInt().Uint64(), private, pbc)
	if err != nil {
		slog.Debug(err.Error())
		app.serverWithdrawTxError(err, withdrawMsg, finalAmount, "")
		return
	}

	app.serverWithdrawInfo(withdrawMsg, finalAmount, txHash)

	go func() {
		withdrawResponse.MerchantId = withdrawMsg.MerchantId
		withdrawResponse.InvoiceId = withdrawMsg.InvoiceId
		withdrawResponse.Crypto = withdrawMsg.Crypto
		withdrawResponse.Address = withdrawMsg.Address

		sig, err := sol.SendTx(app.Sol.Client, &app.Sol.Ws, tx)
		if err != nil {
			slog.Debug(err.Error())
			app.serverWithdrawTxError(err, withdrawMsg, finalAmount, txHash)
			return
		}

		fmt.Println("SENT. Hash: ", sig.String())

		app.serverWithdrawTxSuccess(withdrawResponse, withdrawMsg.Status, sol.LamportsToSol(finalAmount.BigInt().Int64()-sol.SOL_COMISSION) /* (sol.CreateTx -> fee) */, sig.String(), withdrawMsg.TxTempWallet)
	}()
}

// responses

func (app *App) serverWithdrawInfo(withdrawMsg *natsdomain.ReqWithdrawal, finalAmount decimal.Decimal, txHash string) {
	kvData, err := json.Marshal(natsdomain.ResWithdrawal{
		// Error: ,
		MerchantId:   withdrawMsg.MerchantId,
		InvoiceId:    withdrawMsg.InvoiceId,
		Crypto:       withdrawMsg.Crypto,
		Address:      withdrawMsg.Address,
		Amount:       finalAmount,
		Status:       withdrawMsg.Status,
		TxStatus:     natsdomain.WithdrawalTxStatusProcessing,
		TxHash:       txHash,
		TxTempWallet: withdrawMsg.TxTempWallet,
	})

	if err != nil {
		fmt.Println("error: ", err)
		return
	}

	_, err = app.Ns.Js.Publish(context.Background(), natsdomain.SubjResWithdrawal.String(), kvData, jetstream.WithMsgID(natsdomain.NewMsgId(withdrawMsg.InvoiceId, natsdomain.MsgActionInfo)))
	if err != nil {
		fmt.Println("error: ", err)
		return
	}

}

func (app *App) serverWithdrawTxError(err error, withdrawMsg *natsdomain.ReqWithdrawal, finalAmount decimal.Decimal, txHash string) {
	res := natsdomain.ResWithdrawal{
		// Error:      ,
		MerchantId:   withdrawMsg.MerchantId,
		InvoiceId:    withdrawMsg.InvoiceId,
		Crypto:       withdrawMsg.Crypto,
		Address:      withdrawMsg.Address,
		Amount:       finalAmount,
		Status:       withdrawMsg.Status,
		TxStatus:     natsdomain.WithdrawalTxStatusError,
		TxHash:       txHash,
		TxTempWallet: withdrawMsg.TxTempWallet,
	}
	res.Error.Msg = err.Error()

	data, err := json.Marshal(res)

	if err != nil {
		fmt.Println("marshal error: ", err)
		return
	}

	_, err = app.Ns.Js.Publish(context.Background(), natsdomain.SubjResWithdrawal.String(), data, jetstream.WithMsgID(natsdomain.NewMsgId(withdrawMsg.InvoiceId, natsdomain.MsgActionError)))
	if err != nil {
		fmt.Println("publish error: ", err)
		return
	}

}

func (app *App) serverWithdrawHandleError(err error, withdrawResponse natsdomain.ResWithdrawal, msgId string) {
	withdrawResponse.Status = "withdraw_error"
	withdrawResponse.TxStatus = natsdomain.WithdrawalTxStatusError
	withdrawResponse.Amount = decimal.NewFromInt(0)
	withdrawResponse.Error.Msg = err.Error()

	data, err := json.Marshal(withdrawResponse)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	_, err = app.Ns.Js.Publish(context.Background(), natsdomain.SubjResWithdrawal.String(), data, jetstream.WithMsgID(msgId))
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
}

func (app *App) serverWithdrawTxSuccess(withdrawResponse natsdomain.ResWithdrawal, withdrawMsgStatus string, sent decimal.Decimal, txHash string, txTempWallet string) {
	withdrawResponse.TxStatus = natsdomain.WithdrawalTxStatusSent
	withdrawResponse.Status = withdrawMsgStatus
	withdrawResponse.Amount = sent
	withdrawResponse.TxHash = txHash
	withdrawResponse.TxTempWallet = txTempWallet

	data, err := json.Marshal(withdrawResponse)
	if err != nil {
		slog.Debug(err.Error())
		return
	}

	_, err = app.Ns.Js.Publish(context.Background(), natsdomain.SubjResWithdrawal.String(), data, jetstream.WithMsgID(natsdomain.NewMsgId(withdrawResponse.InvoiceId, natsdomain.MsgActionSuccess)))
	if err != nil {
		slog.Debug(err.Error())
		return
	}
	fmt.Println("BROADCAsst")
}
