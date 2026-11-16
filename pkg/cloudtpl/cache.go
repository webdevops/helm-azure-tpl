package cloudtpl

import (
	"github.com/webdevops/helm-azure-tpl/pkg/cache"
)

var (
	globalCache *cache.Cache
)

func init() {
	globalCache = cache.New()
}
