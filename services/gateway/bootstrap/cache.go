package bootstrap

import (
	"fmt"

	"github.com/deepwrite/serivces/gateway/pkg/cache"
	"github.com/deepwrite/serivces/gateway/pkg/config"
)

var Cache cache.Cache

func SetupCache(cfg config.Config) error {
	addr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	cacheInstance, err := cache.New(cache.Options{
		Backend: cfg.Cache.Backend,
		Redis: cache.RedisOptions{
			Addr:     addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
			MaxRetry: cfg.Redis.MaxRetry,
			Prefix:   cfg.Cache.RedisPrefix,
		},
		File: cache.FileOptions{
			Path: cfg.Cache.FilePath,
		},
	})
	if err != nil {
		return err
	}

	Cache = cacheInstance
	cache.SetDefault(cacheInstance)
	return nil
}

func GetCache() cache.Cache {
	return Cache
}
