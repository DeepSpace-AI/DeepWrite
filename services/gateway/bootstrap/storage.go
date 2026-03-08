package bootstrap

import (
	"strings"

	"github.com/deepwrite/serivces/gateway/pkg/config"
	"github.com/deepwrite/serivces/gateway/pkg/storages"
)

var Storage storages.Storage

func SetupStorage(cfg config.Config) error {
	if strings.TrimSpace(cfg.Storage.Bucket) == "" {
		return nil
	}

	if err := storages.InitDefault(cfg.Storage); err != nil {
		return err
	}

	Storage = storages.GetDefault()
	return nil
}

func GetStorage() storages.Storage {
	return Storage
}
