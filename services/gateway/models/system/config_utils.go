package system

import (
	"context"
	"errors"
	"strings"

	"github.com/deepwrite/serivces/gateway/pkg/database"
	"gorm.io/gorm"
)

func GetConfig(ctx context.Context, key string) (string, error) {
	var config SystemConfig
	err := database.DB.WithContext(ctx).
		Where("key = ?", strings.TrimSpace(key)).
		First(&config).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	return config.Value, err
}

func GetConfigsByPrefix(ctx context.Context, prefix string) (map[string]string, error) {
	var configs []SystemConfig
	err := database.DB.WithContext(ctx).
		Where("key LIKE ?", prefix+"%").
		Find(&configs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, c := range configs {
		result[c.Key] = c.Value
	}
	return result, nil
}

func GetAllConfigs(ctx context.Context) ([]SystemConfig, error) {
	var configs []SystemConfig
	err := database.DB.WithContext(ctx).
		Order("key ASC").
		Find(&configs).Error
	return configs, err
}

func SetConfig(ctx context.Context, key, value, description string, updatedBy *string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("config key cannot be empty")
	}

	config := SystemConfig{
		Key:         key,
		Value:       value,
		Description: description,
		UpdatedBy:   updatedBy,
	}

	return database.DB.WithContext(ctx).
		Save(&config).Error
}

func DeleteConfig(ctx context.Context, key string) error {
	return database.DB.WithContext(ctx).
		Delete(&SystemConfig{}, "key = ?", strings.TrimSpace(key)).Error
}

func GetConfigWithDefault(ctx context.Context, key, defaultValue string) string {
	value, err := GetConfig(ctx, key)
	if err != nil || value == "" {
		return defaultValue
	}
	return value
}
