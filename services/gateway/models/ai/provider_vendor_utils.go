package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/pkg/crypto"
	"github.com/deepwrite/serivces/gateway/pkg/database"
	"gorm.io/gorm"
)

type CreateProviderInput struct {
	Provider                string
	BaseURL                 string
	APIKey                  string
	Organization            string
	ChatCompletionsPath     string
	ChatResponsesPath       string
	EmbeddingsPath          string
	RerankPath              string
	AudioSpeechPath         string
	AudioTranscriptionsPath string
	ModelsPath              string
	ExtraHeaders            map[string]string
	Enabled                 *bool
}

type CreateProviderModelByVendorInput struct {
	Model                       string
	RequestModel                string
	SupportsChatCompletions     *bool
	SupportsChatResponses       *bool
	SupportsEmbeddings          *bool
	SupportsRerank              *bool
	SupportsAudioSpeech         *bool
	SupportsAudioTranscriptions *bool
	SupportsModels              *bool
	Enabled                     *bool
}

type DiscoveredProviderModel struct {
	Model   string `json:"model"`
	Name    string `json:"name"`
	OwnedBy string `json:"owned_by"`
}

func ListProviders(ctx context.Context, enabled *bool, limit, offset int) ([]Provider, error) {
	query := database.DB.WithContext(ctx).Model(&Provider{}).Order("updated_at desc")
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	var providers []Provider
	err := query.Find(&providers).Error
	return providers, err
}

func GetProviderByName(ctx context.Context, providerName string) (Provider, error) {
	var provider Provider
	err := database.DB.WithContext(ctx).
		Where("name = ?", normalizeProvider(providerName)).
		First(&provider).Error
	return provider, err
}

func GetProviderByID(ctx context.Context, providerID string) (Provider, error) {
	var provider Provider
	err := database.DB.WithContext(ctx).
		Where("id = ?", providerID).
		First(&provider).Error
	return provider, err
}

func UpdateProviderEnabledByID(ctx context.Context, providerID string, enabled bool) (Provider, error) {
	providerID = strings.TrimSpace(providerID)
	if providerID == "" {
		return Provider{}, errors.New("provider_id is required")
	}

	var provider Provider
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", providerID).First(&provider).Error; err != nil {
			return err
		}

		if err := tx.Model(&Provider{}).Where("id = ?", providerID).Update("enabled", enabled).Error; err != nil {
			return err
		}

		if err := tx.Model(&ProviderModel{}).Where("provider_id = ?", providerID).Update("enabled", enabled).Error; err != nil {
			return err
		}

		return tx.Where("id = ?", providerID).First(&provider).Error
	})

	return provider, err
}

func UpsertProvider(ctx context.Context, input CreateProviderInput) (Provider, error) {
	if strings.TrimSpace(input.Provider) == "" {
		return Provider{}, errors.New("provider is required")
	}
	if strings.TrimSpace(input.BaseURL) == "" {
		return Provider{}, errors.New("base_url is required")
	}
	if strings.TrimSpace(input.APIKey) == "" {
		return Provider{}, errors.New("api_key is required")
	}

	var provider Provider
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		createInput := CreateProviderModelInput{
			Provider:                input.Provider,
			BaseURL:                 input.BaseURL,
			APIKey:                  input.APIKey,
			Organization:            input.Organization,
			ChatCompletionsPath:     input.ChatCompletionsPath,
			ChatResponsesPath:       input.ChatResponsesPath,
			EmbeddingsPath:          input.EmbeddingsPath,
			RerankPath:              input.RerankPath,
			AudioSpeechPath:         input.AudioSpeechPath,
			AudioTranscriptionsPath: input.AudioTranscriptionsPath,
			ModelsPath:              input.ModelsPath,
			ExtraHeaders:            input.ExtraHeaders,
			Enabled:                 input.Enabled,
		}
		current, err := findOrCreateProvider(tx, input.Provider, createInput, nil)
		if err != nil {
			return err
		}

		updates := buildProviderUpdates(UpdateProviderModelInput{
			BaseURL:                 &input.BaseURL,
			APIKey:                  &input.APIKey,
			Organization:            &input.Organization,
			ChatCompletionsPath:     toStringPtr(input.ChatCompletionsPath),
			ChatResponsesPath:       toStringPtr(input.ChatResponsesPath),
			EmbeddingsPath:          toStringPtr(input.EmbeddingsPath),
			RerankPath:              toStringPtr(input.RerankPath),
			AudioSpeechPath:         toStringPtr(input.AudioSpeechPath),
			AudioTranscriptionsPath: toStringPtr(input.AudioTranscriptionsPath),
			ModelsPath:              toStringPtr(input.ModelsPath),
			ExtraHeaders:            toMapPtr(input.ExtraHeaders),
			Enabled:                 input.Enabled,
		})
		if len(updates) > 0 {
			if err := tx.Model(&Provider{}).Where("id = ?", current.ID).Updates(updates).Error; err != nil {
				return err
			}
			if err := tx.Model(&ProviderModel{}).Where("provider_id = ?", current.ID).Updates(updates).Error; err != nil {
				return err
			}
		}
		return tx.Where("id = ?", current.ID).First(&provider).Error
	})
	return provider, err
}

func ListModelsByProvider(ctx context.Context, providerName string, enabled *bool, limit, offset int) ([]ProviderModel, error) {
	query := database.DB.WithContext(ctx).
		Model(&ProviderModel{}).
		Where("provider = ?", normalizeProvider(providerName)).
		Order("updated_at desc")
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	var records []ProviderModel
	err := query.Find(&records).Error
	return records, err
}

func ListModelsByProviderID(ctx context.Context, providerID string, enabled *bool, limit, offset int) ([]ProviderModel, error) {
	query := database.DB.WithContext(ctx).
		Model(&ProviderModel{}).
		Where("provider_id = ?", strings.TrimSpace(providerID)).
		Order("updated_at desc")
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	var records []ProviderModel
	err := query.Find(&records).Error
	return records, err
}

func CreateProviderModelByVendor(ctx context.Context, providerName string, input CreateProviderModelByVendorInput) (ProviderModel, error) {
	modelValue := strings.TrimSpace(input.Model)
	if modelValue == "" {
		return ProviderModel{}, errors.New("model is required")
	}

	var record ProviderModel
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var provider Provider
		if err := tx.Where("name = ?", normalizeProvider(providerName)).First(&provider).Error; err != nil {
			return err
		}

		record = buildModelRecord(CreateProviderModelInput{
			Model:                       modelValue,
			RequestModel:                strings.TrimSpace(input.RequestModel),
			SupportsChatCompletions:     input.SupportsChatCompletions,
			SupportsChatResponses:       input.SupportsChatResponses,
			SupportsEmbeddings:          input.SupportsEmbeddings,
			SupportsRerank:              input.SupportsRerank,
			SupportsAudioSpeech:         input.SupportsAudioSpeech,
			SupportsAudioTranscriptions: input.SupportsAudioTranscriptions,
			SupportsModels:              input.SupportsModels,
			Enabled:                     input.Enabled,
		}, provider)
		return tx.Create(&record).Error
	})
	return record, err
}

func CreateProviderModelByVendorID(ctx context.Context, providerID string, input CreateProviderModelByVendorInput) (ProviderModel, error) {
	modelValue := strings.TrimSpace(input.Model)
	if modelValue == "" {
		return ProviderModel{}, errors.New("model is required")
	}

	var record ProviderModel
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var provider Provider
		if err := tx.Where("id = ?", strings.TrimSpace(providerID)).First(&provider).Error; err != nil {
			return err
		}

		record = buildModelRecord(CreateProviderModelInput{
			Model:                       modelValue,
			RequestModel:                strings.TrimSpace(input.RequestModel),
			SupportsChatCompletions:     input.SupportsChatCompletions,
			SupportsChatResponses:       input.SupportsChatResponses,
			SupportsEmbeddings:          input.SupportsEmbeddings,
			SupportsRerank:              input.SupportsRerank,
			SupportsAudioSpeech:         input.SupportsAudioSpeech,
			SupportsAudioTranscriptions: input.SupportsAudioTranscriptions,
			SupportsModels:              input.SupportsModels,
			Enabled:                     input.Enabled,
		}, provider)
		return tx.Create(&record).Error
	})
	return record, err
}

func DiscoverProviderModels(ctx context.Context, provider Provider) ([]DiscoveredProviderModel, error) {
	baseURL := normalizeBaseURL(provider.BaseURL)
	if baseURL == "" {
		return nil, errors.New("base_url is required")
	}
	modelsPath := normalizePath(provider.ModelsPath, DefaultModelsPath)
	targetURL, err := joinURL(baseURL, modelsPath)
	if err != nil {
		return nil, err
	}

	apiKey, err := crypto.DecryptAPIKey(provider.APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt api_key: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	if strings.TrimSpace(provider.Organization) != "" {
		req.Header.Set("OpenAI-Organization", strings.TrimSpace(provider.Organization))
	}
	for key, value := range provider.ExtraHeaders {
		headerKey := strings.TrimSpace(key)
		if headerKey == "" {
			continue
		}
		req.Header.Set(headerKey, strings.TrimSpace(fmt.Sprint(value)))
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("failed to discover models")
	}

	models := parseDiscoveredModels(bodyBytes)
	if len(models) == 0 {
		return nil, errors.New("no models discovered")
	}
	return models, nil
}

func parseDiscoveredModels(body []byte) []DiscoveredProviderModel {
	type modelItem struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		OwnedBy string `json:"owned_by"`
	}
	type wrappedResponse struct {
		Data []modelItem `json:"data"`
	}

	var wrapped wrappedResponse
	if err := json.Unmarshal(body, &wrapped); err == nil && len(wrapped.Data) > 0 {
		items := make([]DiscoveredProviderModel, 0, len(wrapped.Data))
		for _, item := range wrapped.Data {
			modelID := strings.TrimSpace(item.ID)
			if modelID == "" {
				continue
			}
			items = append(items, DiscoveredProviderModel{
				Model:   modelID,
				Name:    strings.TrimSpace(item.Name),
				OwnedBy: strings.TrimSpace(item.OwnedBy),
			})
		}
		return uniqueDiscoveredModels(items)
	}

	var plainList []string
	if err := json.Unmarshal(body, &plainList); err == nil && len(plainList) > 0 {
		items := make([]DiscoveredProviderModel, 0, len(plainList))
		for _, item := range plainList {
			modelID := strings.TrimSpace(item)
			if modelID == "" {
				continue
			}
			items = append(items, DiscoveredProviderModel{Model: modelID})
		}
		return uniqueDiscoveredModels(items)
	}

	return nil
}

func uniqueDiscoveredModels(items []DiscoveredProviderModel) []DiscoveredProviderModel {
	seen := make(map[string]struct{}, len(items))
	out := make([]DiscoveredProviderModel, 0, len(items))
	for _, item := range items {
		modelID := strings.TrimSpace(item.Model)
		if modelID == "" {
			continue
		}
		if _, ok := seen[modelID]; ok {
			continue
		}
		seen[modelID] = struct{}{}
		out = append(out, item)
	}
	return out
}

func joinURL(baseURL, path string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + normalizePath(path, DefaultModelsPath)
	return parsed.String(), nil
}

func toStringPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func toMapPtr(value map[string]string) *map[string]string {
	if len(value) == 0 {
		return nil
	}
	return &value
}
