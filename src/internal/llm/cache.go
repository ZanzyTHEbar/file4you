package llm

import (
	"sync"
	"time"
)

type Cache struct {
	mu      sync.RWMutex
	items   map[string]CacheItem
	maxSize int
}

type CacheItem struct {
	Decision DestinationDecision
	ExpireAt time.Time
}

func (c *Cache) Get(key string) (DestinationDecision, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if item, exists := c.items[key]; exists && time.Now().Before(item.ExpireAt) {
		return item.Decision, true
	}
	return DestinationDecision{}, false
}
