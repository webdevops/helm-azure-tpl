package azuretpl

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var (
	globalCache *cache.Cache
)

func init() {
	globalCache = cache.New(15*time.Minute, 1*time.Minute)
}
