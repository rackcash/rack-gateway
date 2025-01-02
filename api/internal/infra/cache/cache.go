package cache

import (
	"sync"
	"time"
)

func InitStorage() *Cache {
	return &Cache{
		Storage: sync.Map{},
	}
}

func (c *Cache) Set(k any, v any, expiration time.Duration) {
	c.Storage.Store(k, v)
	go c.delByExp(k, v, expiration)
}

// sets value without expiration
func (c *Cache) SetNoExp(k any, v any) {
	c.Storage.Store(k, v)
}

func (c *Cache) Del(k any) {
	c.Storage.Delete(k)
}

func (c *Cache) Load(k any) any {
	v, _ := c.Storage.Load(k)
	return v
}

func (c *Cache) LoadOrSet(k any, v any, expiration time.Duration) any {
	act, _ := c.Storage.LoadOrStore(k, v)
	go c.delByExp(k, act, expiration)
	return act
}
func (c *Cache) LoadOrSetNoExp(k any, v any, expiration time.Duration) any {
	act, _ := c.Storage.LoadOrStore(k, v)
	return act
}

func (c *Cache) delByExp(k any, v any, expiration time.Duration) {
	time.Sleep(expiration)
	cacheValue, ok := c.Storage.Load(k)
	if !ok {
		return
	}
	if cacheValue != v { // value changed
		return
	}
	c.Storage.Delete(k)
}
