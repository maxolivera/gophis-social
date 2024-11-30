package lru

import (
	"container/list"
	"context"
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
	go cache.cleanup()
	return cache
}

func (c *LRUCache) Get(ctx context.Context, key string) (any, bool) {
	select {
	case <-ctx.Done():
		return nil, false
	default:
		c.mu.RLock()
		el, found := c.items[key]
		c.mu.RUnlock()

		if found {
			item := el.Value.(*CacheItem)
			if time.Now().After(item.Expiration) {
				c.Delete(ctx, key)
				return nil, false
			}

			c.mu.Lock()
			c.order.MoveToFront(el)
			c.mu.Unlock()

			return item.Value, true
		}
		return nil, false
	}
}

func (c *LRUCache) Set(ctx context.Context, key string, value any) {
	select {
	case <-ctx.Done():
		return
	default:
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
}

func (c *LRUCache) Delete(ctx context.Context, key string) {
	select {
	case <-ctx.Done():
		return
	default:
		c.mu.Lock()
		defer c.mu.Unlock()

		if el, found := c.items[key]; found {
			c.order.Remove(el)
			delete(c.items, key)
		}
	}
}

func (c *LRUCache) Len(ctx context.Context) int {
	select {
	case <-ctx.Done():
		return -1
	default:
		c.mu.RLock()
		defer c.mu.RUnlock()
		return len(c.items)
	}
}

func (c *LRUCache) cleanup() {
	for {
		time.Sleep(c.ttl / 2) // Frequency

		c.mu.Lock()
		now := time.Now()
		for el := c.order.Back(); el != nil; el = el.Prev() {
			item := el.Value.(*CacheItem)
			if now.After(item.Expiration) {
				c.order.Remove(el)
				delete(c.items, item.Key)
			}
		}
		c.mu.Unlock()
	}
}
