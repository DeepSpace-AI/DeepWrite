package storages

import (
	"context"
	"io"
	"strings"
	"time"

	appconfig "github.com/deepwrite/serivces/gateway/pkg/config"
)

const (
	ProviderAWS        = "aws-s3"
	ProviderMinIO      = "minio"
	ProviderR2         = "cloudflare-r2"
	ProviderAliyunOSS  = "aliyun-oss"
	ProviderTencentCOS = "tencent-cos"
	ProviderSSO        = "sso"
	ProviderCustom     = "custom"
)

type Config = appconfig.StorageConfig

func normalizeConfig(c Config) Config {
	c.Provider = strings.TrimSpace(strings.ToLower(c.Provider))
	c.Region = strings.TrimSpace(c.Region)
	c.Endpoint = strings.TrimSpace(c.Endpoint)
	c.Bucket = strings.TrimSpace(c.Bucket)
	c.AccessKeyID = strings.TrimSpace(c.AccessKeyID)
	c.SecretAccessKey = strings.TrimSpace(c.SecretAccessKey)
	c.SessionToken = strings.TrimSpace(c.SessionToken)
	c.CustomDomain = strings.TrimSpace(c.CustomDomain)
	if c.Provider == "" {
		c.Provider = ProviderAWS
	}
	if c.Region == "" {
		c.Region = "us-east-1"
	}
	return c
}

func validateConfig(c Config) error {
	normalized := normalizeConfig(c)
	if normalized.Bucket == "" {
		return ErrBucketRequired
	}
	if requiresEndpoint(normalized.Provider) && normalized.Endpoint == "" {
		return ErrEndpointRequired
	}
	return nil
}

type ObjectInfo struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	ETag         string            `json:"etag"`
	LastModified time.Time         `json:"last_modified"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type Storage interface {
	PutObject(ctx context.Context, key string, body io.Reader, contentType string, metadata map[string]string) error
	GetObject(ctx context.Context, key string) (io.ReadCloser, error)
	HeadObject(ctx context.Context, key string) (ObjectInfo, error)
	DeleteObject(ctx context.Context, key string) error
	PresignGetURL(ctx context.Context, key string, expiresIn time.Duration) (string, error)
	PresignPutURL(ctx context.Context, key string, expiresIn time.Duration, contentType string) (string, error)
}

func requiresEndpoint(provider string) bool {
	switch provider {
	case ProviderAWS:
		return false
	default:
		return true
	}
}
