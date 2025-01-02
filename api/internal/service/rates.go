package service

import (
	"encoding/json"
	"fmt"
	"infra/api/internal/infra/cache"
	"infra/api/internal/infra/nats"
	"infra/pkg/nats/natsdomain"
	"infra/pkg/utils"
	"time"
)

type RatesService struct {
	cache *cache.Cache
	ns    *natsdomain.Ns
}

func NewRatesService(cache *cache.Cache, ns *natsdomain.Ns) *RatesService {
	return &RatesService{cache: cache, ns: ns}
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
