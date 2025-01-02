package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"infra/api/internal/domain"
	"infra/api/internal/infra/cache"
	"infra/api/internal/logger"
	"infra/pkg/rr"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-playground/validator/v10"
	"golang.org/x/net/proxy"
)

type WebhookSenderService struct {
	rr    rr.RoundRobin
	list  *atomic.Pointer[[]string]
	l     logger.Logger
	cache *cache.Cache
}

func NewWebhookSenderService(proxyList []string, l logger.Logger) *WebhookSenderService {
	var list atomic.Pointer[[]string]
	list.Store(&proxyList)

	return &WebhookSenderService{rr: rr.New(&list), list: &list, l: l, cache: cache.InitStorage()}
}

type MyRoundTripper struct {
	r http.RoundTripper
}

func (mrt MyRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	// r.Header.Add("Connection", "keep-alive")
	r.Header.Add("Sec-Fetch-Dest", "empty")
	r.Header.Add("Sec-Fetch-Mode", "cors")
	r.Header.Add("Sec-Fetch-Site", "same-origin")
	r.Header.Add("TE", "trailers")
	r.Header.Add("User-Agent", "rack-webhook")
	return mrt.r.RoundTrip(r)

}

func (s *WebhookSenderService) sendWithoutProxy(url string, payload []byte) error {

	client := http.Client{
		Timeout: time.Second * 5,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)

	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	return nil
}

const PROXY_REFUSED = "connect: connection refused"

func (s *WebhookSenderService) sendWithProxy(url string, stringProxy string, payload []byte) error {

	socks, err := s.parseProxy(stringProxy)
	if err != nil {
		return fmt.Errorf("cant' parse proxy: " + err.Error())
	}

	fmt.Println("PROXY:", socks)

	auth := proxy.Auth{
		User:     socks.user,
		Password: socks.pass,
	}

	dialer, err := proxy.SOCKS5("tcp", socks.ip+":"+socks.port, &auth, proxy.Direct)
	if err != nil {
		return err
	}

	dialContext := func(ctx context.Context, network, address string) (net.Conn, error) {
		conn, err := dialer.Dial(network, address)
		if err != nil {
			fmt.Println("dialer.Dial error", err)
			return nil, err
		}

		return conn, err
	}

	transport := &http.Transport{
		DialContext:       dialContext,
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: MyRoundTripper{r: transport},
		Timeout:   5 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	return nil

}
func (s *WebhookSenderService) Send(url string, info domain.ResponseInvoiceInfo) error {
	var MAX_ATTEMPTS = s.rr.GetProxyCount()
	var err error

	if exists := s.cache.Load(info.Id); exists != nil {
		return fmt.Errorf("webhook already sent")
	}

	payload, err := json.Marshal(info)
	if err != nil {
		return err
	}

	fmt.Println("trying to send webhook")

	// err = s.sendWithoutProxy(url, payload)
	// if err != nil {
	// 	attempts++
	// 	s.l.TemplWebhookErr("send without proxy error: "+err.Error(), url, attempts, logger.NA, payload)
	// 	// time.Sleep((RECONNECT_FIRST_NUM<<attempts + 2) */*  */ time.Second)
	// 	return nil
	// }

	stringProxy, ok := s.rr.Next()
	err = func() error {
		var attempts int

	sendReq:
		attempts++

		if attempts > MAX_ATTEMPTS {
			return fmt.Errorf("max attempts exceeded")
		}

		if !ok {
			s.l.Debug("Can't get proxy. sending without proxy")
			err = s.sendWithoutProxy(url, payload)
			if err != nil {
				s.l.TemplWebhookErr("send without proxy error: "+err.Error(), url, attempts, logger.NA, payload)
				return err
			}
			return nil
		}

		err = s.sendWithProxy(url, stringProxy, payload)
		if err != nil {
			s.l.TemplWebhookErr("send with proxy error: "+err.Error(), url, attempts, stringProxy, payload)

			stringProxy, ok = s.rr.Next()
			time.Sleep(5 * time.Second)
			goto sendReq
		}
		return nil
	}()
	if err == nil {
		s.cache.SetNoExp(info.Id, true)
	}

	return err
}

type parsedProxy struct {
	user string `validate:"required,gte=2"`
	pass string `validate:"required,gte=2"`
	ip   string `validate:"required,gte=2"`
	port string `validate:"required,gte=2"`
}

// login:password@ip:port
func (s *WebhookSenderService) parseProxy(str string) (parsedProxy, error) {
	splitA := strings.Split(str, ":") //  to [user pass@ip port]

	if len(splitA) <= 1 {
		return parsedProxy{}, fmt.Errorf("invalid proxy format: given: " + str)
	}

	splitB := strings.Split(splitA[1], "@") // to [ pass ip]

	if len(splitB) != 2 {
		return parsedProxy{}, fmt.Errorf("invalid proxy format: given: " + str)
	}

	var pp = parsedProxy{}

	if len(splitA) != 3 {
		return parsedProxy{}, fmt.Errorf("invalid proxy format: given: " + str)
	}
	pp.user = splitA[0]
	pp.pass = splitB[0]

	pp.ip = splitB[1]
	pp.port = splitA[2]

	validator := validator.New()
	err := validator.Struct(pp)
	if err != nil {
		return parsedProxy{}, err
	}

	return pp, nil
}

func (s *WebhookSenderService) UpdateList(proxies []string) {

	var validProxies []string

	for _, proxy := range proxies {
		_, err := s.parseProxy(proxy)
		if err != nil {
			fmt.Printf("invalid proxy: %s\n", proxy)
			continue
		}
		validProxies = append(validProxies, proxy)
	}

	s.list.Store(&validProxies)
}

func (s *WebhookSenderService) GetList() []string {
	listPtr := s.list.Load()
	if listPtr == nil {
		return []string{}
	}

	return *listPtr
}
