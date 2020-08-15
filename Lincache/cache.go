package Lincache

import (
	"main/Lincache/LRU"
	"sync"
)

type cache struct {
	mu          sync.Mutex
	lru         *LRU.Cache
	maxmemosize int64
}

func (c *cache) add(key string, value ByteView) {
	// 并发写
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = LRU.New(c.maxmemosize, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	// 并发读
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
