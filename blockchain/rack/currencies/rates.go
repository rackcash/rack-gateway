package currencies

import (
	"encoding/json"
	"errors"
	"fmt"
	"infra/blockchain/rack/config"
	"infra/pkg/nats/natsdomain"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type Response struct {
	Data map[string]Data `json:"data"`
}

type Data struct {
	Quote map[string]Quote `json:"quote"`
}

type Quote struct {
	Price float64 `json:"price"`
}

var (
	ErrRateLimit    = errors.New("rate limit")
	ErrInvalidRates = errors.New("invalid rates") // when some currency rate is equal to 0 (error on the side of the currency rate api)
)
var RatesCache sync.Map // Key = currency, Value = Rates

// currency - USD/RUB/EUR
func sendRequest(config *config.Config, currency string) (*natsdomain.Rates, error) {
	fmt.Println("CURRENCY", currency)

	var url = fmt.Sprintf("https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?convert=%s&symbol=ETH,LTC,SOL,TON", currency)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accepts", "application/json")
	req.Header.Set("X-CMC_PRO_API_KEY", config.RackCurrency.CoinmarketApi)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	var rates = &natsdomain.Rates{
		ToCurrency: currency,
		Eth:        decimal.NewFromFloat(math.Round(response.Data["ETH"].Quote[currency].Price*100) / 100),
		Ltc:        decimal.NewFromFloat(math.Round(response.Data["LTC"].Quote[currency].Price*100) / 100),
		Sol:        decimal.NewFromFloat(math.Round(response.Data["SOL"].Quote[currency].Price*100) / 100),
		Ton:        decimal.NewFromFloat(math.Round(response.Data["TON"].Quote[currency].Price*100) / 100),
	}

	if rates.Eth.Equal(decimal.Zero) || rates.Ltc.Equal(decimal.Zero) || rates.Sol.Equal(decimal.Zero) || rates.Ton.Equal(decimal.Zero) {
		return nil, ErrInvalidRates
	}

	fmt.Println("ETH", rates.Eth, currency)
	RatesCache.Store(currency, rates)

	if strings.Contains(string(body), "You've exceeded your API Key's HTTP request rate limit") {
		return nil, ErrRateLimit
	}

	return rates, nil
}

// update rates every 1800 secs (30 mins)
func UpdateRates(config *config.Config) {
	ticker := time.NewTicker(1800 * time.Second)
	defer ticker.Stop()

	currs := []string{
		"RUB",
		"EUR",
		"USD",
	}

	go func() {
		for {
			for _, i := range currs {
				_, err := sendRequest(config, i)
				if err != nil {
					if errors.Is(err, ErrRateLimit) {
						fmt.Println("Rate limit. Sleep")
						time.Sleep(30 * time.Second)
						continue
					}
				}
				time.Sleep(1 * time.Second)
			}
			<-ticker.C
		}
	}()
}

// tries to get from the cache, or tries to send a request to the api
func GetRates(config *config.Config, currency string) (*natsdomain.Rates, error) {
	// return nil, fmt.Errorf("TEST ERROR")

	v, ok := RatesCache.Load(currency)
	if ok {
		fmt.Println("FROM CACHE")
		return v.(*natsdomain.Rates), nil
	}

	rates, err := sendRequest(config, currency)
	if err != nil {
		return nil, err
	}

	return rates, nil
}
