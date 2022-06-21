package ttllruerrcache

import (
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/simplelru"
)

func mustLRU(lru *simplelru.LRU, err error) *simplelru.LRU {
	if err != nil {
		panic(err)
	}
	return lru
}

type cacheItem struct {
	val   interface{}
	ttlAt time.Time
}

// Cache can cache items with both a TTL and a LRU
type Cache struct {
	// This does not need locking.  Just use simplelru.LRU (rather than the locked variant)
	// You can leave this empty and use the Size and OnEviction parameters and Cache will make one for you
	LRUCache   simplelru.LRUCache
	ItemTTL    time.Duration
	Size       int
	OnEviction func(key interface{}, value interface{})
	Now        func() time.Time

	lruMu sync.RWMutex
}

// Set key to value in the cache
func (c *Cache) Set(key interface{}, value interface{}) {
	c.SetFull(key, value, c.now(), c.ItemTTL)
}

// Get a value from the cache if it is still valid.  Returns the value and true/false if the value was in the cache
func (c *Cache) Get(key interface{}) (interface{}, bool) {
	return c.GetFull(key, c.now())
}

// Clean removes old TTL items from the cache
func (c *Cache) Clean() {
	c.CleanFull(c.now())
}

func (c *Cache) onEviction(key interface{}, value interface{}) {
	if c.OnEviction != nil {
		c.OnEviction(key, value)
	}
}

func (c *Cache) now() time.Time {
	if c.Now == nil {
		return time.Now()
	}
	return time.Now()
}

func (c *Cache) size() int {
	if c.Size == 0 {
		return 1024
	}
	return c.Size
}

func (c *Cache) getTTLAt(now time.Time, itemTTL time.Duration) time.Time {
	if itemTTL == 0 {
		return time.Time{}
	}
	return now.Add(itemTTL)
}

// CleanFull is like clean, but at an explicit time
func (c *Cache) CleanFull(now time.Time) {
	c.lruMu.RLock()
	if c.LRUCache == nil {
		c.lruMu.RUnlock()
		return
	}
	allKeys := c.LRUCache.Keys()
	c.lruMu.RUnlock()
	for _, key := range allKeys {
		// A lot of locking here, but won't hold the lock for a much longer amount of time
		c.PeekFull(key, now)
	}
}

// SetFull is like Set, but with an explicit time and TTL
func (c *Cache) SetFull(key interface{}, val interface{}, now time.Time, itemTTL time.Duration) {
	// Key is not valid in cache, try to make it
	c.lruMu.Lock()
	defer c.lruMu.Unlock()
	ci := cacheItem{
		val:   val,
		ttlAt: c.getTTLAt(now, itemTTL),
	}
	if itemTTL < 0 {
		return
	}
	if c.LRUCache == nil {
		c.LRUCache = mustLRU(simplelru.NewLRU(c.size(), c.onEviction))
	}
	c.LRUCache.Add(key, &ci)
}

// GetFull is like Get, but with an explicit time
func (c *Cache) GetFull(key interface{}, now time.Time) (interface{}, bool) {
	// Key is not valid in cache, try to make it
	c.lruMu.RLock()
	if c.LRUCache == nil {
		c.lruMu.RUnlock()
		return nil, false
	}
	item, existing := c.LRUCache.Get(key)
	c.lruMu.RUnlock()
	if !existing {
		return nil, false
	}
	asCI := item.(*cacheItem)
	if asCI.ttlAt.IsZero() || now.Before(asCI.ttlAt) {
		return asCI.val, true
	}

	// This item needs to be TTL
	// Repeat above lock with a write lock
	c.lruMu.Lock()
	defer c.lruMu.Unlock()
	item, existing = c.LRUCache.Get(key)
	if !existing {
		return nil, false
	}
	asCI = item.(*cacheItem)
	if asCI.ttlAt.IsZero() || now.Before(asCI.ttlAt) {
		return asCI.val, true
	}
	c.LRUCache.Remove(key)
	c.onEviction(key, asCI.val)
	return nil, false
}

// PeekFull is like an LRU Peek (does not adjust the LRU), but at an explicit time
func (c *Cache) PeekFull(key interface{}, now time.Time) (interface{}, bool) {
	// Key is not valid in cache, try to make it
	c.lruMu.RLock()
	if c.LRUCache == nil {
		c.lruMu.RUnlock()
		return nil, false
	}
	item, existing := c.LRUCache.Peek(key)
	c.lruMu.RUnlock()
	if !existing {
		return nil, false
	}
	asCI := item.(*cacheItem)
	if asCI.ttlAt.IsZero() || now.Before(asCI.ttlAt) {
		return asCI.val, true
	}

	// This item needs to be TTL
	// Repeat above lock with a write lock
	c.lruMu.Lock()
	defer c.lruMu.Unlock()
	item, existing = c.LRUCache.Peek(key)
	if !existing {
		return nil, false
	}
	asCI = item.(*cacheItem)
	if asCI.ttlAt.IsZero() || now.Before(asCI.ttlAt) {
		return asCI.val, true
	}
	c.LRUCache.Remove(key)
	c.onEviction(key, asCI.val)
	return nil, false
}
