package failover

import (
	"fmt"
	"net/url"
	"sync/atomic"
	"testing"
)

func TestRoundRobin(t *testing.T) {

	tests := []*url.URL{
		{Host: "127.0.0.1"},
		{Host: "127.0.0.2"},
		{Host: "127.0.0.3"},
		{Host: "127.0.0.4"},
		{Host: "127.0.0.5"},
		{Host: "127.0.0.6"},
		{Host: "127.0.0.7"},
		{Host: "127.0.0.8"},
		{Host: "127.0.0.9"},
	}

	var x atomic.Pointer[[]*url.URL]
	x.Store(&[]*url.URL{})

	arr := *x.Load()
	arr = append(arr, tests...)

	x.Store(&arr)
	rr := newRoundRobin(&x)

	for range 10 {
		go func() {
			c, ok := rr.Next()
			fmt.Printf("c: %v, ok: %v\n", c, ok)

			c, ok = rr.Next()
			fmt.Printf("c: %v, ok: %v\n", c, ok)

			c, ok = rr.Next()
			fmt.Printf("c: %v, ok: %v\n", c, ok)

			c, ok = rr.Next()
			fmt.Printf("c: %v, ok: %v\n", c, ok)
		}()
	}

}
