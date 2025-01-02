package service

import (
	"infra/api/internal/config"
	"infra/api/internal/domain"
	"infra/api/internal/logger"
	"strconv"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestGetProxy(t *testing.T) {
	var proxies = []string{
		// "boss:boss@127.0.0.1:1080",
		"login:password@ip:port",
	}

	logger := logger.Init(&config.Config{Prod_env: false})

	s := NewWebhookSenderService(proxies, logger)

	var response = domain.ResponseInvoiceInfo{
		Id:             gofakeit.UUID(),
		Amount:         strconv.Itoa(gofakeit.Int()),
		Cryptocurrency: gofakeit.RandomString([]string{"ETH", "SOL", "TON"}),
		IsPaid:         gofakeit.Bool(),
		Status:         "paid",
		// CreatedAt:     ,
	}

	// if time.Now().Unix() > invoice.EndTimestamp && invoice.Status.IsNotPaid() {
	// 	response.Status = "end"
	// }

	t.Log(s.Send("http://0.0.0.0:9999", response))

}

func TestParseProxy(t *testing.T) {
	proxies := []struct {
		str   string
		valid bool
	}{
		{"login:password@ip:port", true},
		{"login:password:ip:port", false},
		{"login", false},
		{"login:password:", false},
		{"login:password:127.0.0.1:1234:", false},
		{"login:password@127.0.0.1:1234", true},
		{"", false},
		{" ", false},
	}

	s := WebhookSenderService{}

	for _, i := range proxies {
		_, err := s.parseProxy(i.str)
		if err != nil && i.valid {
			t.Fatal(err)
		}
	}

}

func TestSendWithProxy(t *testing.T) {
	var proxies = []string{
		"boss:boss@127.0.0.1:1080",
		// "login:password@ip:port",
	}

	logger := logger.Init(&config.Config{Prod_env: false})

	s := NewWebhookSenderService(proxies, logger)

	// connect: connection refused
	err := s.sendWithProxy("http://127.0.0.1:9999", "s:s@127.0.0.1:1080", []byte(`{"test": "true"}`))
	t.Log(err)
}
