package cache

import (
	"sync"
	"time"
)

type item[V any] struct {
	value      V
	expiration time.Time
}

// Cache 是一个简单的基于内存的 TTL 缓存。
type Cache[K comparable, V any] struct {
	ttl  time.Duration
	data map[K]item[V]
	mu   sync.RWMutex
}

// New 创建缓存实例。
func New[K comparable, V any](ttl time.Duration) *Cache[K, V] {
	if ttl <= 0 {
		ttl = 2 * time.Minute
	}
	return &Cache[K, V]{
		ttl:  ttl,
		data: make(map[K]item[V]),
	}
}

// Get 读取缓存项。
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if entry, ok := c.data[key]; ok {
		if entry.expiration.After(time.Now()) {
			return entry.value, true
		}
	}
	var zero V
	return zero, false
}

// Set 写入缓存。
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = item[V]{
		value:      value,
		expiration: time.Now().Add(c.ttl),
	}
}

// Purge 清空缓存。
func (c *Cache[K, V]) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[K]item[V])
}
