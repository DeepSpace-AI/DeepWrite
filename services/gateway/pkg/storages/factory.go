package storages

import "strings"

func New(cfg Config) (Storage, error) {
	normalized := normalizeConfig(cfg)
	if err := validateConfig(normalized); err != nil {
		return nil, err
	}

	switch normalized.Provider {
	case ProviderAWS, ProviderMinIO, ProviderR2, ProviderAliyunOSS, ProviderTencentCOS, ProviderSSO, ProviderCustom:
		return newS3Storage(normalized)
	default:
		if strings.TrimSpace(normalized.Endpoint) != "" {
			return newS3Storage(normalized)
		}
		return nil, ErrUnsupportedProvider
	}
}
