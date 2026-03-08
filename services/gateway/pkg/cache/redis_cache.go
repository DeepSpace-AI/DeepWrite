package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	prefix string
}

func NewRedisCache(opts RedisOptions) (*RedisCache, error) {
	addr := strings.TrimSpace(opts.Addr)
	if addr == "" {
		addr = "127.0.0.1:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:       addr,
		Password:   opts.Password,
		DB:         opts.DB,
		MaxRetries: opts.MaxRetry,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &RedisCache{client: client, prefix: strings.TrimSpace(opts.Prefix)}, nil
}

func (r *RedisCache) Set(key string, value interface{}, expireTime time.Time) error {
	ctx := context.Background()
	b, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal cache value failed: %w", err)
	}

	ttl := noTTL
	if !expireTime.IsZero() {
		ttl = time.Until(expireTime)
		if ttl <= 0 {
			return r.Delete(key)
		}
	}

	return r.client.Set(ctx, r.key(key), b, ttl).Err()
}

func (r *RedisCache) Get(key string) (interface{}, error) {
	ctx := context.Background()
	data, err := r.client.Get(ctx, r.key(key)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}
		return nil, err
	}

	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("unmarshal cache value failed: %w", err)
	}
	return value, nil
}

func (r *RedisCache) Delete(key string) error {
	ctx := context.Background()
	return r.client.Del(ctx, r.key(key)).Err()
}

func (r *RedisCache) Incr(key string, delta int64) (int64, error) {
	ctx := context.Background()
	return r.client.IncrBy(ctx, r.key(key), delta).Result()
}

func (r *RedisCache) Decr(key string, delta int64) (int64, error) {
	ctx := context.Background()
	return r.client.DecrBy(ctx, r.key(key), delta).Result()
}

func (r *RedisCache) Exists(key string) (bool, error) {
	ctx := context.Background()
	count, err := r.client.Exists(ctx, r.key(key)).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *RedisCache) Flush() error {
	ctx := context.Background()
	if r.prefix == "" {
		return r.client.FlushDB(ctx).Err()
	}

	iter := r.client.Scan(ctx, 0, r.prefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func (r *RedisCache) key(key string) string {
	if r.prefix == "" {
		return key
	}
	return r.prefix + key
}
