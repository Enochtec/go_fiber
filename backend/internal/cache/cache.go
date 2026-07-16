package cache

import (
	"sync"
	"time"
)

type item struct {
	value  any
	expiry time.Time
}

type Cache struct {
	mu       sync.RWMutex
	items    map[string]item
	defaultTTL time.Duration
	stopCh   chan struct{}
}

func New(defaultTTL time.Duration) *Cache {
	c := &Cache{
		items:      make(map[string]item),
		defaultTTL: defaultTTL,
		stopCh:     make(chan struct{}),
	}
	go c.evictLoop()
	return c
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	it, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(it.expiry) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return nil, false
	}
	return it.value, true
}

func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	c.items[key] = item{value: value, expiry: time.Now().Add(c.defaultTTL)}
	c.mu.Unlock()
}

func (c *Cache) SetWithTTL(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	c.items[key] = item{value: value, expiry: time.Now().Add(ttl)}
	c.mu.Unlock()
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

func (c *Cache) Clear() {
	c.mu.Lock()
	clear(c.items)
	c.mu.Unlock()
}

func (c *Cache) Stop() {
	close(c.stopCh)
}

func (c *Cache) evictLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for k, it := range c.items {
				if now.After(it.expiry) {
					delete(c.items, k)
				}
			}
			c.mu.Unlock()
		case <-c.stopCh:
			return
		}
	}
}
