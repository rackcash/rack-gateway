package ton

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync/atomic"
	"time"

	"github.com/shopspring/decimal"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

type Wallet struct {
	Address string
	Seed    string // private key
}

func NewWallet(tonApi *ton.APIClientWrapped, master *ton.BlockIDExt) (*Wallet, error) {
	seed := wallet.NewSeed()

	wallet, err := wallet.FromSeed(*tonApi, seed, wallet.V4R2)
	if err != nil {
		return nil, err
	}

	var structSeed string

	for _, i := range seed {
		structSeed += i + " "
	}

	return &Wallet{
		Address: wallet.Address().String(),
		Seed:    structSeed,
	}, nil
}

func GetBalance(client ton.APIClientWrapped, blockId *atomic.Pointer[ton.BlockIDExt], addr *address.Address) (decimal.Decimal, error) {
	err := getBalanceUpdateBlock(client, blockId)
	if err != nil {
		return decimal.Decimal{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	account, err := client.WaitForBlock(blockId.Load().SeqNo).GetAccount(ctx, blockId.Load(), addr)
	if err != nil {
		return decimal.Decimal{}, err
	}

	fmt.Println(addr.String())
	if account == nil {
		fmt.Println("account == nil")
		return decimal.Zero, nil
	}
	if account.State == nil {
		fmt.Println("account.State == nil")
		return decimal.Zero, nil
	}

	fmt.Println("account.State.Balance: ", account.State.Balance)
	return decimal.NewFromString(account.State.Balance.String())
}

func CreateTX(w *wallet.Wallet, toAddress *address.Address, amount decimal.Decimal, balance decimal.Decimal) (*wallet.Message, error) {

	log.Println("sending transaction and waiting for confirmation...")

	fmt.Println("TON AMOUNT", tlb.MustFromTON(amount.String()))
	fmt.Println("STRING AMOUNT", amount.String())
	fmt.Println("AMOUNT", amount.String())

	tlbBalance, err := tlb.FromTON(balance.String())
	if err != nil {
		return nil, err
	}

	fmt.Println("BALANCE", tlbBalance.Nano().Uint64())

	if tlbBalance.Nano().Uint64() <= 3000000 /* 0.003 TON */ {
		return nil, errors.New("balance is too small")
	}

	tx, err := w.BuildTransfer(toAddress, tlb.MustFromTON(amount.String()), false, "")
	if err != nil {
		log.Fatalln("Transfer err:", err.Error())
		return nil, err
	}

	tx.Mode = 128 // send entire balance

	return tx, nil
}

func SendTx(w *wallet.Wallet, tx *wallet.Message) (*tlb.Transaction, error) {
	fmt.Println("START SENDING")
	trans, block, err := w.SendWaitTransaction(context.Background(), tx)
	if err != nil {
		return nil, err
	}

	balance, err := w.GetBalance(context.Background(), block)
	if err != nil {
		log.Fatalln("GetBalance err:", err.Error())
		return nil, err
	}

	log.Printf("transaction confirmed at block %d, hash: %s balance left: %s", block.SeqNo,
		base64.StdEncoding.EncodeToString(trans.Hash), balance.String())

	// 0 - equals
	// 1 - greater
	cmp := balance.Nano().Cmp(big.NewInt(0))
	if cmp > 0 {
		return nil, errors.New("balance > 0 (needs to be 0)")
	}

	return trans, nil
}

func GetWalletBySeed(client ton.APIClientWrapped, seedString string) (*wallet.Wallet, error) {
	seedArray := strings.Fields(seedString)

	wallet, err := wallet.FromSeed(client, seedArray, wallet.V4R2)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

func GetTxByAddress(client ton.APIClientWrapped, block *ton.BlockIDExt, address *address.Address) (*Tx, error) {
	account, err := client.WaitForBlock(block.SeqNo).GetAccount(context.Background(), block, address)
	if err != nil {
		panic(err)
	}

	if account.LastTxHash == nil {
		fmt.Println("account.LastTxHash == nil")
		return nil, err
	}

	txs, err := client.ListTransactions(context.Background(), address, 1, account.LastTxLT, account.LastTxHash)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if txs == nil {
		return nil, fmt.Errorf("tx == nil")
	}

	tx, err := isOutTx(txs[0])
	if err != nil {
		return nil, err
	}

	// if !tx.IsOut {
	// 	fmt.Println("tx.IsOut == false")
	// 	return
	// }
	return &tx, nil

}

type Tx struct {
	Amount decimal.Decimal
	IsOut  bool
	ToAddr string
}

func isOutTx(t *tlb.Transaction) (Tx, error) {
	var destinations string

	var tx = Tx{
		Amount: decimal.Zero,
		IsOut:  false,
		ToAddr: "null",
	}

	_, out := new(big.Int), new(big.Int)

	if t.IO.Out != nil {
		listOut, err := t.IO.Out.ToSlice()
		if err != nil {
			return tx, err
		}

		for _, m := range listOut {
			destinations = m.Msg.DestAddr().String()
			if m.MsgType == tlb.MsgTypeInternal {
				out.Add(out, m.AsInternal().Amount.Nano())
			}
		}
	}

	amountDecimal, err := decimal.NewFromString(tlb.FromNanoTON(out).String())
	if err != nil {
		return tx, err
	}

	if out.Cmp(big.NewInt(0)) != 0 {
		tx.Amount = amountDecimal
		tx.IsOut = true
		tx.ToAddr = destinations
	}

	return tx, nil
}

func getBalanceUpdateBlock(client ton.APIClientWrapped, blockId *atomic.Pointer[ton.BlockIDExt]) error {
	var attempts = 3
	var err error

	for attempts > 0 {
		err = UpdateBlockOnce(client, blockId)
		if err == nil {
			return nil
		}
		attempts--
	}
	return err
}
