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

func (app *App) MerchantWithdrawal(msg jetstream.Msg, withdrawMsg *natsdomain.ReqMerchantWithdrawal, finalAmount decimal.Decimal) {
	defer msg.Ack()

	var attemptsLimit int = 2

	md, _ := msg.Metadata()
	if md != nil {
		fmt.Println("NUM DELIVERD", md.NumDelivered)

		if md.NumDelivered > 1 { // if the server goes down, give it one try
			attemptsLimit = 0
		}
	}

	fmt.Println("WITHDRAWAL TIMESTAMP", withdrawMsg.WithdrawalTimestamp)
	// return

	// TODO: add more cryptocurrencies
	switch withdrawMsg.Crypto {
	case "sol":
		app.withdrawMerchantSol(withdrawMsg, finalAmount)
	case "ton":
		app.withdrawMerchantTon(withdrawMsg, finalAmount)
	case "eth":
		app.withdrawMerchantEth(withdrawMsg, finalAmount, attemptsLimit)
	default:
		fmt.Println("error: invalid crypto")
	}

}

func (app *App) withdrawMerchantEth(withdrawMsg *natsdomain.ReqMerchantWithdrawal, finalAmount decimal.Decimal, attemptsLimit int) {
	var withdrawResponse natsdomain.ResMerchantWithdrawal

	withdrawResponse.WithdrawalTimestamp = withdrawMsg.WithdrawalTimestamp
	withdrawResponse.WithdrawalID = withdrawMsg.WithdrawalID
	// withdrawResponse.UserId = withdrawMsg.UserId

	finalAmountWei := eth.EtherToWei(finalAmount)

	fmt.Println("FINAL AMOUNT ETHER", finalAmount)
	fmt.Println("FINAL AMOUNT WEI", finalAmountWei)

	tx, err := eth.CreateTx(app.EthClient, withdrawMsg.ToAddress, withdrawMsg.Private, finalAmountWei, attemptsLimit)
	if err != nil {
		slog.Debug(err.Error())
		app.merchantWithdrawalTxError(err, withdrawMsg, finalAmount, "")
		return
	}
	fmt.Println("TX HASH WAAAA", tx.Hash().String())

	// app.merchantWithdrawalInfo(withdrawMsg, finalAmount, tx.Hash().String())

	go func() {
		withdrawResponse.MerchantId = withdrawMsg.MerchantId
		// withdrawResponse.InvoiceId = withdrawMsg.InvoiceId
		withdrawResponse.Crypto = withdrawMsg.Crypto
		withdrawResponse.ToAddress = withdrawMsg.ToAddress
		withdrawResponse.FromAddress = withdrawMsg.FromAddress

		sent, err := eth.SendTx(app.EthClient, tx)
		if err != nil { // error
			slog.Debug(err.Error())
			app.merchantWithdrawalHandleError(err, withdrawResponse, natsdomain.NewMsgId(withdrawMsg.WithdrawalTimestamp+withdrawMsg.MerchantId, natsdomain.MsgActionError))
			return
		}

		fromAddress := eth.PrivateToAddress(withdrawMsg.Private)
		fmt.Println("BALANCE AFTER")
		fmt.Println(eth.GetBalanceEther(app.EthClient, fromAddress.Hex()))

		app.merchantWithdrawalTxSuccess(withdrawResponse, *sent, tx.Hash().String())
	}()

}

func (app *App) withdrawMerchantTon(withdrawMsg *natsdomain.ReqMerchantWithdrawal, finalAmount decimal.Decimal) {
	var withdrawResponse natsdomain.ResMerchantWithdrawal

	withdrawResponse.WithdrawalTimestamp = withdrawMsg.WithdrawalTimestamp
	withdrawResponse.WithdrawalID = withdrawMsg.WithdrawalID
	// withdrawResponse.UserId = withdrawMsg.UserId

	wallet, err := ton.GetWalletBySeed(app.Ton.Client, withdrawMsg.Private)
	if err != nil {
		slog.Debug(err.Error())
		app.merchantWithdrawalTxError(err, withdrawMsg, finalAmount, "")
		return
	}

	toAddr, err := address.ParseAddr(withdrawMsg.ToAddress)
	if err != nil {
		slog.Debug(err.Error())
		app.merchantWithdrawalTxError(err, withdrawMsg, finalAmount, "")
		return
	}

	// final amount 0, because the mode will be set to send the entire balance
	// finalAmount - balance
	tx, err := ton.CreateTX(wallet, toAddr, decimal.NewFromInt(0), finalAmount)
	if err != nil {
		slog.Debug(err.Error())
		app.merchantWithdrawalTxError(err, withdrawMsg, finalAmount, "")
		return
	}

	// app.merchantWithdrawalInfo(withdrawMsg, finalAmount, "")

	go func() {
		withdrawResponse.MerchantId = withdrawMsg.MerchantId
		withdrawResponse.Crypto = withdrawMsg.Crypto
		withdrawResponse.ToAddress = withdrawMsg.ToAddress
		withdrawResponse.FromAddress = withdrawMsg.FromAddress

		tx, err := ton.SendTx(wallet, tx)
		if err != nil {
			slog.Debug(err.Error())
			app.merchantWithdrawalTxError(err, withdrawMsg, finalAmount, "")
			return
		}

		fmt.Println("SENT. Hash: ", base64.StdEncoding.EncodeToString(tx.Hash))

		app.merchantWithdrawalTxSuccess(withdrawResponse, finalAmount, base64.StdEncoding.EncodeToString(tx.Hash))
	}()
}

func (app *App) withdrawMerchantSol(withdrawMsg *natsdomain.ReqMerchantWithdrawal, finalAmount decimal.Decimal) {
	var withdrawResponse natsdomain.ResMerchantWithdrawal

	withdrawResponse.WithdrawalTimestamp = withdrawMsg.WithdrawalTimestamp
	withdrawResponse.WithdrawalID = withdrawMsg.WithdrawalID
	// withdrawResponse.UserId = withdrawMsg.UserId

	pbc, err := sol.StringToPBC(withdrawMsg.ToAddress)
	if err != nil {
		slog.Debug(err.Error())
		app.merchantWithdrawalTxError(err, withdrawMsg, finalAmount, "")
		return
	}

	private, err := sol.StringToPriv(withdrawMsg.Private)
	if err != nil {
		slog.Debug(err.Error())
		app.merchantWithdrawalTxError(err, withdrawMsg, finalAmount, "")
		return
	}

	tx, txHash, err := sol.CreateTx(app.Sol.Client, finalAmount.BigInt().Uint64(), private, pbc)
	if err != nil {
		slog.Debug(err.Error())
		app.merchantWithdrawalTxError(err, withdrawMsg, finalAmount, "")
		return
	}

	// app.merchantWithdrawalInfo(withdrawMsg, finalAmount, txHash)

	go func() {
		withdrawResponse.MerchantId = withdrawMsg.MerchantId
		// withdrawResponse.InvoiceId = withdrawMsg.InvoiceId
		withdrawResponse.Crypto = withdrawMsg.Crypto
		withdrawResponse.ToAddress = withdrawMsg.ToAddress
		withdrawResponse.FromAddress = withdrawMsg.FromAddress

		sig, err := sol.SendTx(app.Sol.Client, &app.Sol.Ws, tx)
		if err != nil {
			slog.Debug(err.Error())
			app.merchantWithdrawalTxError(err, withdrawMsg, finalAmount, txHash)
			return
		}

		fmt.Println("SENT. Hash: ", sig.String())

		app.merchantWithdrawalTxSuccess(withdrawResponse, sol.LamportsToSol(finalAmount.BigInt().Int64()-sol.SOL_COMISSION) /*  (sol.CreateTx -> fee) */, sig.String())
	}()
}

// responses

// func (app *App) merchantWithdrawalInfo(withdrawMsg *natsdomain.ReqMerchantWithdrawal, finalAmount decimal.Decimal, txHash string) {
// 	kvData, err := json.Marshal(natsdomain.ResMerchantWithdrawal{
// 		// Error: ,
// 		// MerchantId:   withdrawMsg.MerchantId,
// 		// InvoiceId:    withdrawMsg.InvoiceId,
// 		// Crypto:       withdrawMsg.Crypto,
// 		// ToAddress:      withdrawMsg.ToAddress,
// 		// Amount:       finalAmount,
// 		Status: natsdomain.MerchantWithdrawalStatus(),
// 		// TxStatus:     natsdomain.WithdrawalTxStatusProcessing,
// 		TxHash: txHash,
// 		// TxTempWallet: withdrawMsg.TxTempWallet,
// 	})

// 	if err != nil {
// 		fmt.Println("error: ", err)
// 		return
// 	}

// 	_, err = app.Ns.Js.Publish(context.Background(), natsdomain.SubjResMerchantWithdrawal.String(), kvData, jetstream.WithMsgID(natsdomain.NewMsgId(withdrawMsg.WithdrawalTimestamp+withdrawMsg.UserId, natsdomain.MsgActionInfo)))
// 	if err != nil {
// 		fmt.Println("error: ", err)
// 		return
// 	}

// }

func (app *App) merchantWithdrawalTxError(err error, withdrawMsg *natsdomain.ReqMerchantWithdrawal, finalAmount decimal.Decimal, txHash string) {
	res := natsdomain.ResMerchantWithdrawal{
		ReqMerchantWithdrawal: natsdomain.ReqMerchantWithdrawal{
			Withdrawal: natsdomain.Withdrawal{
				FromAddress: withdrawMsg.FromAddress,
				MerchantId:  withdrawMsg.MerchantId,
				ToAddress:   withdrawMsg.ToAddress,
				Private:     withdrawMsg.Private,
				Crypto:      withdrawMsg.Crypto,
				Amount:      withdrawMsg.Amount,
			},
			WithdrawalID:        withdrawMsg.WithdrawalID,
			WithdrawalTimestamp: withdrawMsg.WithdrawalTimestamp,
		},
		// Error:      ,

		// MerchantId: withdrawMsg.MerchantId,
		// InvoiceId:  withdrawMsg.InvoiceId,
		// Crypto:     withdrawMsg.Crypto,
		// ToAddress:    withdrawMsg.ToAddress,
		// Amount:     finalAmount,
		Status: natsdomain.MerchantWithdrawalStatusError,
		// TxStatus:     natsdomain.WithdrawalTxStatusError,
		TxHash: txHash,
		// TxTempWallet: withdrawMsg.TxTempWallet,
	}

	// FromAddress: jsonRes.FromAddress,
	// MerchantId:  jsonRes.MerchantId,
	// ToAddress:   jsonRes.ToAddress,
	// Private:     jsonRes.Private,
	// Crypto:      jsonRes.Crypto,
	// Amount:      jsonRes.Amount,

	res.Error.Msg = err.Error()

	data, err := json.Marshal(res)

	if err != nil {
		fmt.Println("marshal error: ", err)
		return
	}

	_, err = app.Ns.Js.Publish(context.Background(), natsdomain.SubjResMerchantWithdrawal.String(), data, jetstream.WithMsgID(natsdomain.NewMsgId(withdrawMsg.WithdrawalTimestamp+withdrawMsg.MerchantId, natsdomain.MsgActionError)))
	if err != nil {
		fmt.Println("publish error: ", err)
		return
	}

}

func (app *App) merchantWithdrawalHandleError(err error, withdrawResponse natsdomain.ResMerchantWithdrawal, msgId string) {
	// BUG: invalid struct (merchantWithdrawalTxError)
	withdrawResponse.Status = natsdomain.MerchantWithdrawalStatusError
	// withdrawResponse.TxStatus = natsdomain.WithdrawalTxStatusError
	withdrawResponse.Amount = decimal.NewFromInt(0)
	withdrawResponse.Error.Msg = err.Error()

	data, err := json.Marshal(withdrawResponse)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	_, err = app.Ns.Js.Publish(context.Background(), natsdomain.SubjResMerchantWithdrawal.String(), data, jetstream.WithMsgID(msgId))
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
}

func (app *App) merchantWithdrawalTxSuccess(withdrawResponse natsdomain.ResMerchantWithdrawal, sent decimal.Decimal, txHash string) {
	// withdrawResponse.TxStatus = natsdomain.WithdrawalTxStatusSent
	withdrawResponse.Status = natsdomain.MerchantWithdrawalStatusSent
	withdrawResponse.Amount = sent
	withdrawResponse.TxHash = txHash
	// withdrawResponse.TxTempWallet = txTempWallet

	data, err := json.Marshal(withdrawResponse)
	if err != nil {
		slog.Debug(err.Error())
		return
	}

	var msgId = natsdomain.NewMsgId(withdrawResponse.WithdrawalTimestamp+withdrawResponse.MerchantId, natsdomain.MsgActionSuccess)

	fmt.Println("MSG ID " + msgId)
	_, err = app.Ns.Js.Publish(context.Background(), natsdomain.SubjResMerchantWithdrawal.String(), data, jetstream.WithMsgID(msgId))

	if err != nil {
		slog.Debug(err.Error())
		return
	}
	fmt.Println("BROADCAsst")
}
