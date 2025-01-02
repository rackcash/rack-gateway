package failover

import (
	"net/url"
	"sync"
	"sync/atomic"
)

type roundRobin interface {
	Next() (*url.URL, bool)
}

type rr struct {
	mu      *sync.Mutex
	index   *atomic.Uint32
	servers *atomic.Pointer[[]*url.URL]
}

func newRoundRobin(servers *atomic.Pointer[[]*url.URL]) roundRobin {

	// fmt.Println(index.Load())

	return rr{
		index:   new(atomic.Uint32),
		servers: servers,
		mu:      &sync.Mutex{},
	}
}

func (rr rr) Next() (*url.URL, bool) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	servers := *rr.servers.Load()

	if len(servers) == 0 {
		return nil, false
	}

	n := rr.index.Add(1)
	conn := servers[(int(n)-1)%len(servers)]

	if conn.String() != "" {
		return conn, true
	}

	return nil, false
}
