package ipinfo

import (
	"sync"
	"time"
)

// cacheItem represents a cached item with expiration
type cacheItem struct {
	data   *IPInfo
	expiry time.Time
}

// isExpired checks if the cache item has expired
func (c *cacheItem) isExpired() bool {
	return time.Now().After(c.expiry)
}

// memoryCache implements an in-memory cache with TTL
type memoryCache struct {
	items map[string]*cacheItem
	mutex sync.RWMutex
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache() Cache {
	cache := &memoryCache{
		items: make(map[string]*cacheItem),
	}
	
	// Start cleanup goroutine
	go cache.cleanup()
	
	return cache
}

// Get retrieves an item from cache
func (c *memoryCache) Get(key string) (*IPInfo, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	item, exists := c.items[key]
	if !exists || item.isExpired() {
		return nil, false
	}
	
	return item.data, true
}

// Set stores an item in cache with TTL
func (c *memoryCache) Set(key string, value *IPInfo, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.items[key] = &cacheItem{
		data:   value,
		expiry: time.Now().Add(ttl),
	}
}

// Delete removes an item from cache
func (c *memoryCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	delete(c.items, key)
}

// Clear removes all items from cache
func (c *memoryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.items = make(map[string]*cacheItem)
}

// cleanup periodically removes expired items
func (c *memoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		c.mutex.Lock()
		for key, item := range c.items {
			if item.isExpired() {
				delete(c.items, key)
			}
		}
		c.mutex.Unlock()
	}
}