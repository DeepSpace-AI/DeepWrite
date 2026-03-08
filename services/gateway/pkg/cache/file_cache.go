package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const noTTL time.Duration = 0

var ErrCacheMiss = errors.New("cache miss")

type fileCacheItem struct {
	Value    json.RawMessage `json:"value"`
	ExpireAt *time.Time      `json:"expire_at,omitempty"`
}

type fileCacheSnapshot struct {
	Items map[string]fileCacheItem `json:"items"`
}

type FileCache struct {
	path  string
	mu    sync.RWMutex
	items map[string]fileCacheItem
}

func NewFileCache(opts FileOptions) (*FileCache, error) {
	path := opts.Path
	if path == "" {
		path = "./tmp/cache/cache.json"
	}

	cache := &FileCache{
		path:  path,
		items: make(map[string]fileCacheItem),
	}

	if err := cache.load(); err != nil {
		return nil, err
	}

	return cache, nil
}

func (f *FileCache) Set(key string, value interface{}, expireTime time.Time) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal cache value failed: %w", err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if !expireTime.IsZero() && time.Now().After(expireTime) {
		delete(f.items, key)
		return f.saveLocked()
	}

	item := fileCacheItem{Value: valueBytes}
	if !expireTime.IsZero() {
		expire := expireTime
		item.ExpireAt = &expire
	}

	f.items[key] = item
	return f.saveLocked()
}

func (f *FileCache) Get(key string) (interface{}, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	item, ok := f.items[key]
	if !ok {
		return nil, ErrCacheMiss
	}

	if isExpired(item) {
		delete(f.items, key)
		if err := f.saveLocked(); err != nil {
			return nil, err
		}
		return nil, ErrCacheMiss
	}

	var value interface{}
	if err := json.Unmarshal(item.Value, &value); err != nil {
		return nil, fmt.Errorf("unmarshal cache value failed: %w", err)
	}

	return value, nil
}

func (f *FileCache) Delete(key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.items, key)
	return f.saveLocked()
}

func (f *FileCache) Incr(key string, delta int64) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	current, item, err := f.getInt64Locked(key)
	if err != nil {
		return 0, err
	}
	current += delta

	b, _ := json.Marshal(current)
	item.Value = b
	f.items[key] = item

	if err := f.saveLocked(); err != nil {
		return 0, err
	}

	return current, nil
}

func (f *FileCache) Decr(key string, delta int64) (int64, error) {
	return f.Incr(key, -delta)
}

func (f *FileCache) Exists(key string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	item, ok := f.items[key]
	if !ok {
		return false, nil
	}

	if isExpired(item) {
		delete(f.items, key)
		if err := f.saveLocked(); err != nil {
			return false, err
		}
		return false, nil
	}

	return true, nil
}

func (f *FileCache) Flush() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.items = make(map[string]fileCacheItem)
	return f.saveLocked()
}

func (f *FileCache) getInt64Locked(key string) (int64, fileCacheItem, error) {
	item, ok := f.items[key]
	if !ok || isExpired(item) {
		newItem := fileCacheItem{}
		f.items[key] = newItem
		return 0, newItem, nil
	}

	var asInt int64
	if err := json.Unmarshal(item.Value, &asInt); err == nil {
		return asInt, item, nil
	}

	var asNumber json.Number
	if err := json.Unmarshal(item.Value, &asNumber); err == nil {
		v, parseErr := strconv.ParseInt(asNumber.String(), 10, 64)
		if parseErr == nil {
			return v, item, nil
		}
	}

	return 0, fileCacheItem{}, fmt.Errorf("cache value for key %s is not integer", key)
}

func (f *FileCache) load() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
		return fmt.Errorf("create cache dir failed: %w", err)
	}

	if _, err := os.Stat(f.path); errors.Is(err, os.ErrNotExist) {
		return f.saveLocked()
	}

	b, err := os.ReadFile(f.path)
	if err != nil {
		return fmt.Errorf("read cache file failed: %w", err)
	}
	if len(b) == 0 {
		return nil
	}

	snapshot := fileCacheSnapshot{}
	if err := json.Unmarshal(b, &snapshot); err != nil {
		return fmt.Errorf("parse cache file failed: %w", err)
	}

	if snapshot.Items == nil {
		snapshot.Items = make(map[string]fileCacheItem)
	}

	f.items = snapshot.Items
	for key, item := range f.items {
		if isExpired(item) {
			delete(f.items, key)
		}
	}

	return f.saveLocked()
}

func (f *FileCache) saveLocked() error {
	snapshot := fileCacheSnapshot{Items: f.items}
	b, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("marshal cache snapshot failed: %w", err)
	}

	tmp := f.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return fmt.Errorf("write cache temp file failed: %w", err)
	}

	if err := os.Rename(tmp, f.path); err != nil {
		return fmt.Errorf("replace cache file failed: %w", err)
	}

	return nil
}

func isExpired(item fileCacheItem) bool {
	if item.ExpireAt == nil {
		return false
	}
	return time.Now().After(*item.ExpireAt)
}
