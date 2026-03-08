package cache

import (
	"fmt"
	"strings"
)

func New(opts Options) (Cache, error) {
	backend := strings.TrimSpace(strings.ToLower(opts.Backend))
	if backend == "" {
		backend = BackendRedis
	}

	switch backend {
	case BackendRedis:
		return NewRedisCache(opts.Redis)
	case BackendFile:
		return NewFileCache(opts.File)
	default:
		return nil, fmt.Errorf("unsupported cache backend: %s", backend)
	}
}
