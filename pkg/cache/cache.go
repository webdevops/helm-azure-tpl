package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	CacheSeparator = ":"
)

type (
	Cache struct {
		cache  *cache.Cache
		prefix string
	}
)

// New creates a new cache
func New() *Cache {
	return &Cache{
		cache:  cache.New(15*time.Minute, 1*time.Minute),
		prefix: "",
	}
}

// key generates a key and adds the optional prefix
func (c *Cache) key(k string) string {
	if c.prefix != "" {
		k = c.prefix + CacheSeparator + k
	}
	return k
}

// Get an item from the cache. Returns the item or nil, and a bool indicating
// whether the key was found.
func (c *Cache) Get(k string) (interface{}, bool) {
	return c.cache.Get(c.key(k))
}

// Set add an item to the cache, replacing any existing item. If the duration is 0
// (DefaultExpiration), the cache's default expiration time is used. If it is -1
// (NoExpiration), the item never expires.
func (c *Cache) Set(k string, x interface{}, d time.Duration) {
	c.cache.Set(c.key(k), x, d)
}

// SetDefault add an item to the cache, replacing any existing item, using the default
// expiration.
func (c *Cache) SetDefault(k string, val interface{}) {
	c.cache.SetDefault(c.key(k), val)
}

// Sub creates a sub cache using an additional prefix
func (c *Cache) Sub(prefix string) *Cache {
	if c.prefix != "" {
		prefix = c.prefix + CacheSeparator + prefix
	}

	return &Cache{
		cache:  c.cache,
		prefix: prefix,
	}
}
