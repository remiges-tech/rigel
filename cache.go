package rigel

import "sync"

type InMemoryCache struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		data: make(map[string]interface{}),
	}
}

func (c *InMemoryCache) Get(key string) (value interface{}, found bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, found = c.data[key]
	return
}

func (c *InMemoryCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}
