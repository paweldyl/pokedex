package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	cacheMap   map[string]cacheEntry
	cacheMutex *sync.Mutex
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(interval time.Duration) Cache {
	cache := Cache{
		cacheMap:   map[string]cacheEntry{},
		cacheMutex: &sync.Mutex{},
	}
	go cache.reapLoop(interval)
	return cache
}

func (c *Cache) Add(key string, val []byte) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.cacheMap[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	ce, ok := c.cacheMap[key]
	if !ok {
		return nil, false
	}
	return ce.val, true
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		t := <-ticker.C
		for mapKey, mapVal := range c.cacheMap {
			if mapVal.createdAt.Before(t) {
				delete(c.cacheMap, mapKey)
			}
		}
	}
}
