package nats

import (
	"context"
	"fmt"
	"infra/blockchain/rack/currencies/eth"
	"infra/blockchain/rack/currencies/sol"
	"infra/blockchain/rack/currencies/ton"
	"time"

	"github.com/shopspring/decimal"
	"github.com/xssnick/tonutils-go/address"
)

func (app *App) GetBalance(cryptocurrency string, addr string) (decimal.Decimal, error) {

	var balance decimal.Decimal
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// TODO: Add more cryptocurrencies
	// get balance
	switch cryptocurrency {
	case "sol":
		balance, err = sol.GetBalanceSol(ctx, app.Sol.Client, addr)
	case "ton":
		addr, err_ := address.ParseAddr(addr)
		if err_ != nil {
			return decimal.Decimal{}, err
		}
		balance, err = ton.GetBalance(app.Ton.Client, &app.Ton.BlockID, addr)
	case "eth":
		balance, err = eth.GetBalanceEther(app.EthClient, addr)
	default:
		return decimal.Decimal{}, fmt.Errorf("invalid cryptocurrency: " + cryptocurrency)
	}

	// handle balance
	if err != nil {
		return decimal.Decimal{}, err
	}

	return balance, nil

}

func (app *App) NewWallet(cryptocurrency string) (address string, privateKey string, err error) {
	// TODO: add more cryptocurrencies
	switch cryptocurrency {
	case "sol":
		address, privateKey, err = sol.NewWallet()
		if err != nil {
			return "", "", err
		}

	case "eth":
		address, privateKey = eth.NewWallet()
	case "ton":
		wallet, err := ton.NewWallet(&app.Ton.Client, app.Ton.BlockID.Load())
		if err != nil {
			return "", "", err
		}
		address = wallet.Address
		privateKey = wallet.Seed
	default:
		return "", "", fmt.Errorf("invalid cryptocurrency: " + cryptocurrency)
	}

	return address, privateKey, nil
}
