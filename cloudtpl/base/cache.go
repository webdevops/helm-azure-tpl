package base

import (
	"log/slog"
	"strings"
)

// cacheResult caches template function results (eg. Azure REST API resource information)
func (e *BaseEngine) CacheResult(cacheKey string, callback func() (interface{}, error)) (interface{}, error) {
	if val, ok := e.Cache.Get(cacheKey); ok {
		e.Logger.Info("found in cache", slog.String("cacheKey", cacheKey))
		return val, nil
	}

	ret, err := callback()
	if err != nil {
		return nil, err
	}

	e.Cache.SetDefault(cacheKey, ret)

	return ret, nil
}

func (e *BaseEngine) GenerateCacheKey(val ...string) string {
	return strings.Join(val, ":")
}
