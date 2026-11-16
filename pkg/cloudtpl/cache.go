package cloudtpl

import (
	"github.com/webdevops/helm-azure-tpl/pkg/cache"
)

var (
	globalCache *cache.Cache
)

// init creates a new global cache
func init() {
	globalCache = cache.New()
}
