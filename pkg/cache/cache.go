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

func New() *Cache {
	return &Cache{
		cache:  cache.New(15*time.Minute, 1*time.Minute),
		prefix: "",
	}
}

func (c *Cache) key(k string) string {
	if c.prefix != "" {
		k = c.prefix + CacheSeparator + k
	}
	return k
}

func (c *Cache) Get(k string) (interface{}, bool) {
	return c.cache.Get(c.key(k))
}

func (c *Cache) Set(k string, x interface{}, d time.Duration) {
	c.cache.Set(c.key(k), x, d)
}

func (c *Cache) SetDefault(k string, val interface{}) {
	c.cache.SetDefault(c.key(k), val)
}

func (c *Cache) Sub(prefix string) *Cache {
	if c.prefix != "" {
		prefix = c.prefix + CacheSeparator + prefix
	}

	return &Cache{
		cache:  c.cache,
		prefix: prefix,
	}
}
