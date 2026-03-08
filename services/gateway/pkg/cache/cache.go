package cache

import "time"

type Cache interface {
	Set(key string, value interface{}, expireTime time.Time) error
	Get(key string) (interface{}, error)
	Delete(key string) error
	Incr(key string, delta int64) (int64, error)
	Decr(key string, delta int64) (int64, error)
	Exists(key string) (bool, error)
	Flush() error
}

const (
	BackendRedis = "redis"
	BackendFile  = "file"
)

type Options struct {
	Backend string
	Redis   RedisOptions
	File    FileOptions
}

type RedisOptions struct {
	Addr     string
	Password string
	DB       int
	MaxRetry int
	Prefix   string
}

type FileOptions struct {
	Path string
}
