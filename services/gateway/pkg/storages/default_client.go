package storages

import "sync"

var (
	defaultStorage Storage
	defaultMu      sync.RWMutex
)

func SetDefault(client Storage) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultStorage = client
}

func GetDefault() Storage {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultStorage
}

func InitDefault(cfg Config) error {
	client, err := New(cfg)
	if err != nil {
		return err
	}
	SetDefault(client)
	return nil
}
