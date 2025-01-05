package service

import (
	"encoding/json"
	"fmt"
	"infra/api/internal/domain"
	"infra/api/internal/infra/cache"
	"infra/api/internal/infra/nats"
	"infra/pkg/nats/natsdomain"
	"infra/pkg/utils"
	"time"

	"github.com/shopspring/decimal"
)

type RatesService struct {
	cache    *cache.Cache
	invoices *InvoicesService
	ns       *natsdomain.Ns
}

func NewRatesService(cache *cache.Cache, ns *natsdomain.Ns, invoices *InvoicesService) *RatesService {
	return &RatesService{cache: cache, ns: ns, invoices: invoices}
}

func (s *RatesService) Get(amountCurrency string) (*natsdomain.Rates, error) {
	var rates *natsdomain.Rates

	rates, ok := s.cache.Load(amountCurrency).(*natsdomain.Rates)
	if rates != nil && ok {
		return rates, nil
	}

	rates, err := getRatesFromNats(s.ns, amountCurrency)
	if err != nil {
		return nil, err
	}
	if rates == nil {
		return nil, fmt.Errorf("rates is nil, but no error")
	}

	s.cache.Set(amountCurrency, rates, time.Minute*5)
	return rates, nil
}

func (s *RatesService) Convert(amount decimal.Decimal, crypto domain.Crypto, rates *natsdomain.Rates) (decimal.Decimal, decimal.Decimal, error) {
	// TODO: add more cryptocurrencies
	var converted decimal.Decimal
	var rate decimal.Decimal

	switch crypto {
	case domain.CRYPTO_SOL:
		fmt.Println("RATES SOL", rates.Sol.String())
		rate = rates.Sol
		converted = s.invoices.CalculateFinAmount(amount, rate, 5)
	case domain.CRYPTO_ETH:
		fmt.Println("RATES ETH", rates.Eth.String())
		rate = rates.Eth
		converted = s.invoices.CalculateFinAmount(amount, rate)
	case domain.CRYPTO_TON:
		fmt.Println("RATES TON", rates.Ton.String())
		rate = rates.Ton
		converted = s.invoices.CalculateFinAmount(amount, rate, 3)
	default:
		return decimal.Zero, decimal.Zero, fmt.Errorf(domain.ErrMsgInvalidCrypto)
	}
	return converted, rate, nil
}

func getRatesFromNats(ns *natsdomain.Ns, currency string) (*natsdomain.Rates, error) {

	jsonMsg, err := json.Marshal(natsdomain.ReqGetRates{ToCurrency: currency})
	if err != nil {
		return nil, err
	}

	data, err := ns.ReqAndRecv(natsdomain.SubjGetRates, jsonMsg)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(data))

	is, errMsg := nats.HelpersIsError(data)
	if is {
		return nil, fmt.Errorf(errMsg)
	}

	res, err := utils.Unmarshal[natsdomain.ResGetRates](data)
	if err != nil {
		return nil, err
	}

	if res.IsError {
		return nil, fmt.Errorf(res.Message)
	}

	return &res.Rates, nil
}
