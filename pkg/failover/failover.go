package failover

import (
	"errors"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Failover interface {
	AddUrl(url *url.URL) error      // adds one url
	AddUrls(urls ...*url.URL) error // adds multiple urls

	Request(requestFunc func(url *url.URL) error, localOpts ...func(*Options)) error // sends a request to the url using your requestFunc

	MustParseURL(url string) *url.URL      // panics if url is invalid
	ParseURL(url string) (*url.URL, error) // returns error if url is invalid

	MustParseURLS(urls ...string) []*url.URL
	ParseURLS(urls ...string) ([]*url.URL, error)
}

type onErrType uint8

const (
	ReqOnErrRemoveAndReconnect onErrType = iota // remove url and reconnect (default attempts - 3)
	ReqOnErrIgnore                              // ignore error (returns nil)
	ReqOnErrReturnErr                           // return error
	ReqOnErrReconnectNext                       // reconnect (to next url)
	ReqOnErrReconnectCurrent                    // reconnect (to current url)
)

type Options struct {
	CheckConn            func(url *url.URL) error // function for checking url
	CheckUrlBeforeAdding bool                     // check url before adding (using CheckConn)
	CheckUrlDelay        time.Duration            // delay before checking url (defualt 30s)
	ReqOnErr             onErrType                // what to do on request error
	MaxAttempts          uint16                   // max attempts (default 3)

}
type storage struct {
	activeUrls    *atomic.Pointer[[]*url.URL]
	badUrls       *atomic.Pointer[[]*url.URL]
	roundRobin    roundRobin
	options       *Options
	mu            *sync.Mutex
	startCronChan chan bool
}

func New(checkConnection func(url *url.URL) error, options ...func(*Options)) Failover {
	activeUrls := &atomic.Pointer[[]*url.URL]{}
	badUrls := &atomic.Pointer[[]*url.URL]{}

	opts := &Options{ // default options
		CheckUrlBeforeAdding: true,
		CheckUrlDelay:        time.Second * 30,
		CheckConn:            checkConnection,
	}

	for _, opt := range options {
		opt(opts)
	}

	s := &storage{roundRobin: newRoundRobin(activeUrls), badUrls: badUrls, activeUrls: activeUrls, options: opts, mu: &sync.Mutex{}, startCronChan: make(chan bool)}

	// go printUrls(s.activeUrls, s.badUrls)
	go s.runCronUrlCheck()

	return s

}

func (s *storage) AddUrls(urls ...*url.URL) error {
	for _, url := range urls {
		err := s.AddUrl(url)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *storage) AddUrl(url__ *url.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if url__ == nil {
		return errors.New("url is nil")
	}

	if s.options.CheckUrlBeforeAdding {
		err := s.options.CheckConn(url__)
		if err != nil {
			return fmt.Errorf("check connection: %w", err)
		}
	}

	urlsPtr := s.activeUrls.Load()
	if urlsPtr == nil {
		s.activeUrls.Store(&[]*url.URL{url__})
		return nil
	}
	urls := *urlsPtr

	urls = removeUrl(urls, url__) // remove url if already exists
	urls = append(urls, url__)    // append unique url

	s.activeUrls.Store(&urls)

	return nil
}

func (s *storage) Request(requestFunc func(url *url.URL) error, localOpts ...func(*Options)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	urlsPtr := s.activeUrls.Load()
	if urlsPtr == nil {
		return errors.New("no urls found")
	}

	urls := *urlsPtr

	url, found := s.roundRobin.Next()
	if !found {
		return errors.New("round robin: no urls found")
	}

	optsCopy := *s.options // we need to copy opts because we need to change it (locally)

	for _, opt := range localOpts {
		opt(&optsCopy)
	}

	return s.request(requestFunc, &optsCopy, urls, url, 1)
}

func (s *storage) MustParseURL(url__ string) *url.URL {
	u, err := url.Parse(url__)
	if err != nil {
		panic(err)
	}
	return u
}

func (s *storage) ParseURL(url__ string) (*url.URL, error) {
	return url.Parse(url__)
}

func (s *storage) MustParseURLS(urls ...string) []*url.URL {
	var parsed []*url.URL

	for _, i := range urls {
		parsed = append(parsed, s.MustParseURL(i))
	}

	return parsed
}

func (s *storage) ParseURLS(urls ...string) ([]*url.URL, error) {
	var parsed []*url.URL

	for _, i := range urls {
		p, err := s.ParseURL(i)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, p)
	}

	return parsed, nil
}
