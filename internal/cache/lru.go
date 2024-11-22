package cache

import (
	"container/list"
	"sync"
	"time"
)

type CacheItem struct {
	Key        string
	Value      any
	Expiration time.Time
}

type LRUCache struct {
	capacity int
	items    map[string]*list.Element
	order    *list.List
	mu       sync.RWMutex
	ttl      time.Duration
}

func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	cache := &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
		ttl:      ttl,
	}
	go cache.cleanupExpired()
	return cache
}

func NewLRUStorage(c *LRUCache) *Storage {
	return &Storage{
		Users: &UserLRUCache{c},
	}
}

func (c *LRUCache) Get(key string) (any, bool) {
	c.mu.RLock()
	el, found := c.items[key]
	c.mu.RUnlock()

	if found {
		item := el.Value.(*CacheItem)
		if time.Now().After(item.Expiration) {
			c.mu.Lock()
			c.Delete(key)
			c.mu.Unlock()
			return nil, false
		}
		c.mu.Lock()
		c.order.MoveToFront(el)
		c.mu.Unlock()
		return item.Value, true
	}
	return nil, false
}

func (c *LRUCache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// exists, update it
	if el, found := c.items[key]; found {
		c.order.MoveToFront(el)
		el.Value.(*CacheItem).Value = value
		el.Value.(*CacheItem).Expiration = time.Now().Add(c.ttl)
		return
	}

	// capacity full, evict oldest
	if len(c.items) >= c.capacity {
		evict := c.order.Back()
		if evict != nil {
			c.order.Remove(evict)
			delete(c.items, evict.Value.(*CacheItem).Key)
		}
	}

	// add new item
	item := &CacheItem{
		Key:        key,
		Value:      value,
		Expiration: time.Now().Add(c.ttl),
	}
	el := c.order.PushFront(item)
	c.items[key] = el
}

func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, found := c.items[key]; found {
		c.order.Remove(el)
		delete(c.items, key)
	}
}

func (c *LRUCache) cleanupExpired() {
	for {
		time.Sleep(c.ttl / 2) // Cleanup frequency

		c.mu.RLock()
		now := time.Now()

		toDelete := []string{}

		for el := c.order.Back(); el != nil; el = el.Prev() {
			item := el.Value.(*CacheItem)
			if now.After(item.Expiration) {
				toDelete = append(toDelete, item.Key)
			}
		}
		c.mu.RUnlock()

		// Delete expired items with a write lock
		c.mu.Lock()
		for _, key := range toDelete {
			if el, found := c.items[key]; found {
				c.order.Remove(el)
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}
