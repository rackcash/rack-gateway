package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

func TestSet(t *testing.T) {
	var keys []string
	c := InitStorage()

	// test set
	for range 10000 {
		k := gofakeit.BuzzWord()
		keys = append(keys, k)

		go c.Set(k, gofakeit.BuzzWord(), time.Second*time.Duration(gofakeit.IntRange(1, 5)))
	}

	// time.Sleep(5 * time.Second)

	// test del
	// for _, i := range keys {
	// 	go c.Del(i)
	// }

	time.Sleep(6 * time.Second)

	// test load
	for _, i := range keys {
		v := c.Load(i)
		if v != nil {
			fmt.Println(v)
		}
	}

}
