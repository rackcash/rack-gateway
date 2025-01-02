package rr

import (
	"sync"
	"sync/atomic"
)

type RoundRobin interface {
	Next() (string, bool)
	GetProxyCount() int
}

type rr struct {
	data  *atomic.Pointer[[]string]
	mu    *sync.Mutex
	index *atomic.Uint32
}

func New(data *atomic.Pointer[[]string]) *rr {
	return &rr{
		data:  data,
		mu:    &sync.Mutex{},
		index: new(atomic.Uint32),
	}

}

func (rr *rr) Next() (string, bool) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	servers := *rr.data.Load()

	if len(servers) == 0 {
		return "", false
	}

	n := rr.index.Add(1)
	target := servers[(int(n)-1)%len(servers)]

	return target, true
}

func (rr *rr) GetProxyCount() int {
	servers := *rr.data.Load()
	return len(servers)
}
