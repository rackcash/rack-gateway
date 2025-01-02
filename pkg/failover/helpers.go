package failover

import (
	"errors"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

func addToBadUrls(activeUrls__ *atomic.Pointer[[]*url.URL], badUrls__ *atomic.Pointer[[]*url.URL], url__ *url.URL) {
	activeUrlsPtr := activeUrls__.Load()
	if activeUrlsPtr == nil {
		return
	}

	activeUrls := *activeUrlsPtr

	activeUrls = removeUrl(activeUrls, url__)
	activeUrls__.Store(&activeUrls)

	badUrlsPtr := badUrls__.Load()
	if badUrlsPtr == nil {
		badUrls__.Store(&[]*url.URL{url__})
		return
	}

	badUrls := *badUrlsPtr
	badUrls = removeUrl(badUrls, url__)
	badUrls = append(badUrls, url__)
	badUrls__.Store(&badUrls)

}

func removeFromBadUrls(activeUrls__ *atomic.Pointer[[]*url.URL], badUrls__ *atomic.Pointer[[]*url.URL], url__ *url.URL) {
	badUrlsPtr := badUrls__.Load()
	if badUrlsPtr == nil {
		return
	}

	badUrls := *badUrlsPtr

	badUrls = removeUrl(badUrls, url__)
	badUrls__.Store(&badUrls)

	activeUrlsPtr := activeUrls__.Load()
	if activeUrlsPtr == nil {
		activeUrls__.Store(&[]*url.URL{url__})
		return
	}

	activeUrls := *activeUrlsPtr
	activeUrls = removeUrl(activeUrls, url__)
	activeUrls = append(activeUrls, url__)
	activeUrls__.Store(&activeUrls)

}

func removeUrl(slice []*url.URL, element *url.URL) []*url.URL {
	newSlice := make([]*url.URL, 0, len(slice))

	for _, i := range slice {
		if i.String() != element.String() {
			newSlice = append(newSlice, i)
		}

	}

	return newSlice
}

func (s *storage) request(requestFunc func(url *url.URL) error, optsCopy *Options, urls []*url.URL, url *url.URL, attempts uint16) error {
	err := requestFunc(url)
	if err == nil {
		return nil
	}

	if attempts >= optsCopy.MaxAttempts {
		return err
	}

	switch optsCopy.ReqOnErr {
	case ReqOnErrIgnore:
		return nil
	case ReqOnErrReturnErr:
		return err
	case ReqOnErrReconnectCurrent:
		return s.request(requestFunc, optsCopy, urls, url, attempts+1)
	case ReqOnErrReconnectNext:
		url, found := s.roundRobin.Next()
		if !found {
			return errors.New("round robin: no urls found")
		}
		return s.request(requestFunc, optsCopy, urls, url, attempts+1)
	case ReqOnErrRemoveAndReconnect:
		addToBadUrls(s.activeUrls, s.badUrls, url)
		// printUrlsOnce(s.activeUrls, s.badUrls)

		url, found := s.roundRobin.Next()
		if !found {
			return errors.New("round robin: no urls found")
		}
		return s.request(requestFunc, optsCopy, urls, url, attempts+1)
	}

	return err
}

func (s *storage) runCronUrlCheck() {

	ticker := time.NewTicker(s.options.CheckUrlDelay * time.Second)
	defer ticker.Stop()

	for {
		if s.activeUrls.Load() == nil && s.badUrls.Load() == nil {
			continue
		}

		if s.activeUrls.Load() == nil {
			continue
		}

		for range ticker.C {
			if cronUrlCheck(s.activeUrls, s.badUrls, s.options) != nil {
				break
			}
		}
	}

}

func cronUrlCheck(urls__ *atomic.Pointer[[]*url.URL], badUrls__ *atomic.Pointer[[]*url.URL], options *Options) error {
	var wg sync.WaitGroup

	wg.Add(2) // bad urls and active urls

	go func() { // check active urls
		defer func() {
			wg.Done()
		}()

		urlsPtr := urls__.Load()
		if urlsPtr == nil {
			return
		}

		urls := *urlsPtr

		for _, url := range urls {
			err := options.CheckConn(url)
			if err != nil {
				addToBadUrls(urls__, badUrls__, url)
				continue
			}
		}
	}()

	go func() {
		defer func() {
			wg.Done()
		}()

		badUrlsPtr := badUrls__.Load()
		if badUrlsPtr == nil {
			return
		}

		badUrls := *badUrlsPtr
		for _, url := range badUrls {
			err := options.CheckConn(url)
			if err == nil {
				removeFromBadUrls(urls__, badUrls__, url)
				continue
			}
		}

	}()

	wg.Wait()

	return nil

}

// func printUrls(activeUrls__ *atomic.Pointer[[]*url.URL], badUrls__ *atomic.Pointer[[]*url.URL]) {
// 	for {
// 		activeUrlsPtr := activeUrls__.Load()
// 		if activeUrlsPtr != nil {
// 			fmt.Printf("active urls: %v\n", *activeUrlsPtr)
// 		}

// 		badUrlsPtr := badUrls__.Load()
// 		if badUrlsPtr != nil {
// 			fmt.Printf("bad urls: %v\n", *badUrlsPtr)
// 		}
// 		time.Sleep(1 * time.Second)

// 	}
// }

// func printUrlsOnce(activeUrls__ *atomic.Pointer[[]*url.URL], badUrls__ *atomic.Pointer[[]*url.URL]) {
// 	activeUrlsPtr := activeUrls__.Load()
// 	if activeUrlsPtr != nil {
// 		fmt.Printf("active urls: %v\n", *activeUrlsPtr)
// 	}

// 	badUrlsPtr := badUrls__.Load()
// 	if badUrlsPtr != nil {
// 		fmt.Printf("bad urls: %v\n", *badUrlsPtr)
// 	}
// }
