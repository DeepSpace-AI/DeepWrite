package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/deepwrite/serivces/gateway/pkg/database"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CreateProviderModelInput struct {
	Model                       string
	Provider                    string
	RequestModel                string
	BaseURL                     string
	APIKey                      string
	Organization                string
	ChatCompletionsPath         string
	ChatResponsesPath           string
	EmbeddingsPath              string
	RerankPath                  string
	AudioSpeechPath             string
	AudioTranscriptionsPath     string
	ModelsPath                  string
	ExtraHeaders                map[string]string
	SupportsChatCompletions     *bool
	SupportsChatResponses       *bool
	SupportsEmbeddings          *bool
	SupportsRerank              *bool
	SupportsAudioSpeech         *bool
	SupportsAudioTranscriptions *bool
	SupportsModels              *bool
	Enabled                     *bool
}

type UpdateProviderModelInput struct {
	Model                       *string
	Provider                    *string
	RequestModel                *string
	BaseURL                     *string
	APIKey                      *string
	Organization                *string
	ChatCompletionsPath         *string
	ChatResponsesPath           *string
	EmbeddingsPath              *string
	RerankPath                  *string
	AudioSpeechPath             *string
	AudioTranscriptionsPath     *string
	ModelsPath                  *string
	ExtraHeaders                *map[string]string
	SupportsChatCompletions     *bool
	SupportsChatResponses       *bool
	SupportsEmbeddings          *bool
	SupportsRerank              *bool
	SupportsAudioSpeech         *bool
	SupportsAudioTranscriptions *bool
	SupportsModels              *bool
	Enabled                     *bool
}

func CreateProviderModel(ctx context.Context, input CreateProviderModelInput) (ProviderModel, error) {
	if err := validateCreateInput(input); err != nil {
		return ProviderModel{}, err
	}

	var record ProviderModel
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		provider, err := findOrCreateProvider(tx, input.Provider, input, nil)
		if err != nil {
			return err
		}
		record = buildModelRecord(input, provider)
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		return nil
	})
	return record, err
}

func GetProviderModelByModel(ctx context.Context, model string) (ProviderModel, error) {
	var record ProviderModel
	err := database.DB.WithContext(ctx).
		Where("model = ?", strings.TrimSpace(model)).
		First(&record).Error
	return record, err
}

func ListProviderModels(ctx context.Context, enabled *bool, limit, offset int) ([]ProviderModel, error) {
	query := database.DB.WithContext(ctx).Model(&ProviderModel{}).Order("updated_at desc")
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

func UpdateProviderModel(ctx context.Context, model string, input UpdateProviderModelInput) (ProviderModel, error) {
	var record ProviderModel
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("model = ?", strings.TrimSpace(model)).First(&record).Error; err != nil {
			return err
		}

		var provider Provider
		if strings.TrimSpace(record.ProviderID) == "" {
			migratedProvider, err := findOrCreateProvider(tx, record.Provider, toCreateInputFromRecord(record), nil)
			if err != nil {
				return err
			}
			record.ProviderID = migratedProvider.ID
			initialSnapshot := map[string]any{}
			applyProviderSnapshot(&initialSnapshot, migratedProvider)
			if err := tx.Model(&ProviderModel{}).Where("id = ?", record.ID).Updates(initialSnapshot).Error; err != nil {
				return err
			}
			provider = migratedProvider
		} else if err := tx.Where("id = ?", record.ProviderID).First(&provider).Error; err != nil {
			return err
		}

		nextProviderName := provider.Name
		if input.Provider != nil && strings.TrimSpace(*input.Provider) != "" {
			nextProviderName = strings.TrimSpace(*input.Provider)
		}
		targetProvider, err := findOrCreateProvider(tx, nextProviderName, toCreateInput(input), &provider)
		if err != nil {
			return err
		}
		if targetProvider.ID != record.ProviderID {
			record.ProviderID = targetProvider.ID
		}

		providerUpdates := buildProviderUpdates(input)
		if len(providerUpdates) > 0 {
			if err := tx.Model(&Provider{}).Where("id = ?", targetProvider.ID).Updates(providerUpdates).Error; err != nil {
				return err
			}
			if err := tx.Where("id = ?", targetProvider.ID).First(&targetProvider).Error; err != nil {
				return err
			}
		}

		modelUpdates := buildModelUpdates(record, input)
		applyProviderSnapshot(&modelUpdates, targetProvider)
		if len(modelUpdates) > 0 {
			if err := tx.Model(&ProviderModel{}).Where("id = ?", record.ID).Updates(modelUpdates).Error; err != nil {
				return err
			}
		}
		if err := syncModelsByProvider(tx, targetProvider); err != nil {
			return err
		}
		return tx.Where("id = ?", record.ID).First(&record).Error
	})

	return record, err
}

func DeleteProviderModel(ctx context.Context, model string) (bool, error) {
	result := database.DB.WithContext(ctx).
		Where("model = ?", strings.TrimSpace(model)).
		Delete(&ProviderModel{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func BackfillProviderRelations(ctx context.Context) error {
	return database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var records []ProviderModel
		if err := tx.Where("provider_id IS NULL").Find(&records).Error; err != nil {
			return err
		}
		for _, record := range records {
			provider, err := findOrCreateProvider(tx, record.Provider, toCreateInputFromRecord(record), nil)
			if err != nil {
				return err
			}
			updates := map[string]any{}
			applyProviderSnapshot(&updates, provider)
			if err := tx.Model(&ProviderModel{}).Where("id = ?", record.ID).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func normalizeProvider(provider string) string {
	value := strings.TrimSpace(provider)
	if value == "" {
		return DefaultProvider
	}
	return value
}

func normalizeBaseURL(rawURL string) string {
	return strings.TrimRight(strings.TrimSpace(rawURL), "/")
}

func normalizePath(rawPath, fallback string) string {
	value := strings.TrimSpace(rawPath)
	if value == "" {
		value = fallback
	}
	if !strings.HasPrefix(value, "/") {
		return "/" + value
	}
	return value
}

func normalizeHeaders(headers map[string]string) datatypes.JSONMap {
	if len(headers) == 0 {
		return datatypes.JSONMap{}
	}
	out := datatypes.JSONMap{}
	for key, value := range headers {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		out[trimmedKey] = strings.TrimSpace(value)
	}
	return out
}

func validateCreateInput(input CreateProviderModelInput) error {
	if strings.TrimSpace(input.Model) == "" {
		return errors.New("model is required")
	}
	if strings.TrimSpace(input.Provider) == "" {
		return errors.New("provider is required")
	}
	if strings.TrimSpace(input.BaseURL) == "" {
		return errors.New("base_url is required")
	}
	if strings.TrimSpace(input.APIKey) == "" {
		return errors.New("api_key is required")
	}
	return nil
}

func buildModelRecord(input CreateProviderModelInput, provider Provider) ProviderModel {
	modelValue := strings.TrimSpace(input.Model)
	requestModel := strings.TrimSpace(input.RequestModel)
	if requestModel == "" {
		requestModel = modelValue
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	return ProviderModel{
		ProviderID:                  provider.ID,
		Model:                       modelValue,
		RequestModel:                requestModel,
		SupportsChatCompletions:     boolOrDefault(input.SupportsChatCompletions, true),
		SupportsChatResponses:       boolOrDefault(input.SupportsChatResponses, true),
		SupportsEmbeddings:          boolOrDefault(input.SupportsEmbeddings, true),
		SupportsRerank:              boolOrDefault(input.SupportsRerank, true),
		SupportsAudioSpeech:         boolOrDefault(input.SupportsAudioSpeech, true),
		SupportsAudioTranscriptions: boolOrDefault(input.SupportsAudioTranscriptions, true),
		SupportsModels:              boolOrDefault(input.SupportsModels, true),
		Enabled:                     enabled,
		Provider:                    provider.Name,
		BaseURL:                     provider.BaseURL,
		APIKey:                      provider.APIKey,
		Organization:                provider.Organization,
		ChatCompletionsPath:         provider.ChatCompletionsPath,
		ChatResponsesPath:           provider.ChatResponsesPath,
		EmbeddingsPath:              provider.EmbeddingsPath,
		RerankPath:                  provider.RerankPath,
		AudioSpeechPath:             provider.AudioSpeechPath,
		AudioTranscriptionsPath:     provider.AudioTranscriptionsPath,
		ModelsPath:                  provider.ModelsPath,
		ExtraHeaders:                provider.ExtraHeaders,
	}
}

func toCreateInput(input UpdateProviderModelInput) CreateProviderModelInput {
	out := CreateProviderModelInput{}
	if input.Provider != nil {
		out.Provider = strings.TrimSpace(*input.Provider)
	}
	if input.BaseURL != nil {
		out.BaseURL = strings.TrimSpace(*input.BaseURL)
	}
	if input.APIKey != nil {
		out.APIKey = strings.TrimSpace(*input.APIKey)
	}
	if input.Organization != nil {
		out.Organization = strings.TrimSpace(*input.Organization)
	}
	if input.ChatCompletionsPath != nil {
		out.ChatCompletionsPath = *input.ChatCompletionsPath
	}
	if input.ChatResponsesPath != nil {
		out.ChatResponsesPath = *input.ChatResponsesPath
	}
	if input.EmbeddingsPath != nil {
		out.EmbeddingsPath = *input.EmbeddingsPath
	}
	if input.RerankPath != nil {
		out.RerankPath = *input.RerankPath
	}
	if input.AudioSpeechPath != nil {
		out.AudioSpeechPath = *input.AudioSpeechPath
	}
	if input.AudioTranscriptionsPath != nil {
		out.AudioTranscriptionsPath = *input.AudioTranscriptionsPath
	}
	if input.ModelsPath != nil {
		out.ModelsPath = *input.ModelsPath
	}
	if input.ExtraHeaders != nil {
		out.ExtraHeaders = *input.ExtraHeaders
	}
	if input.Enabled != nil {
		out.Enabled = input.Enabled
	}
	return out
}

func toCreateInputFromRecord(record ProviderModel) CreateProviderModelInput {
	headers := map[string]string{}
	for key, value := range record.ExtraHeaders {
		headers[key] = fmt.Sprint(value)
	}
	enabled := record.Enabled
	return CreateProviderModelInput{
		Provider:                record.Provider,
		BaseURL:                 record.BaseURL,
		APIKey:                  record.APIKey,
		Organization:            record.Organization,
		ChatCompletionsPath:     record.ChatCompletionsPath,
		ChatResponsesPath:       record.ChatResponsesPath,
		EmbeddingsPath:          record.EmbeddingsPath,
		RerankPath:              record.RerankPath,
		AudioSpeechPath:         record.AudioSpeechPath,
		AudioTranscriptionsPath: record.AudioTranscriptionsPath,
		ModelsPath:              record.ModelsPath,
		ExtraHeaders:            headers,
		Enabled:                 &enabled,
	}
}

func findOrCreateProvider(tx *gorm.DB, providerName string, input CreateProviderModelInput, fallback *Provider) (Provider, error) {
	name := normalizeProvider(providerName)
	var provider Provider
	err := tx.Where("name = ?", name).First(&provider).Error
	if err == nil {
		return provider, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return Provider{}, err
	}
	baseURL := normalizeBaseURL(input.BaseURL)
	apiKey := strings.TrimSpace(input.APIKey)
	organization := strings.TrimSpace(input.Organization)
	chatCompletionsPath := normalizePath(input.ChatCompletionsPath, DefaultChatCompletionsPath)
	chatResponsesPath := normalizePath(input.ChatResponsesPath, DefaultChatResponsesPath)
	embeddingsPath := normalizePath(input.EmbeddingsPath, DefaultEmbeddingsPath)
	rerankPath := normalizePath(input.RerankPath, DefaultRerankPath)
	audioSpeechPath := normalizePath(input.AudioSpeechPath, DefaultAudioSpeechPath)
	audioTranscriptionsPath := normalizePath(input.AudioTranscriptionsPath, DefaultAudioTranscriptions)
	modelsPath := normalizePath(input.ModelsPath, DefaultModelsPath)
	headers := normalizeHeaders(input.ExtraHeaders)
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	if fallback != nil {
		if baseURL == "" {
			baseURL = fallback.BaseURL
		}
		if apiKey == "" {
			apiKey = fallback.APIKey
		}
		if organization == "" {
			organization = fallback.Organization
		}
		if strings.TrimSpace(input.ChatCompletionsPath) == "" {
			chatCompletionsPath = fallback.ChatCompletionsPath
		}
		if strings.TrimSpace(input.ChatResponsesPath) == "" {
			chatResponsesPath = fallback.ChatResponsesPath
		}
		if strings.TrimSpace(input.EmbeddingsPath) == "" {
			embeddingsPath = fallback.EmbeddingsPath
		}
		if strings.TrimSpace(input.RerankPath) == "" {
			rerankPath = fallback.RerankPath
		}
		if strings.TrimSpace(input.AudioSpeechPath) == "" {
			audioSpeechPath = fallback.AudioSpeechPath
		}
		if strings.TrimSpace(input.AudioTranscriptionsPath) == "" {
			audioTranscriptionsPath = fallback.AudioTranscriptionsPath
		}
		if strings.TrimSpace(input.ModelsPath) == "" {
			modelsPath = fallback.ModelsPath
		}
		if len(headers) == 0 {
			headers = fallback.ExtraHeaders
		}
		enabled = fallback.Enabled
		if input.Enabled != nil {
			enabled = *input.Enabled
		}
	}
	if baseURL == "" || apiKey == "" {
		return Provider{}, fmt.Errorf("base_url and api_key are required for provider %s", name)
	}
	provider = Provider{
		Name:                    name,
		BaseURL:                 baseURL,
		APIKey:                  apiKey,
		Organization:            organization,
		ChatCompletionsPath:     chatCompletionsPath,
		ChatResponsesPath:       chatResponsesPath,
		EmbeddingsPath:          embeddingsPath,
		RerankPath:              rerankPath,
		AudioSpeechPath:         audioSpeechPath,
		AudioTranscriptionsPath: audioTranscriptionsPath,
		ModelsPath:              modelsPath,
		ExtraHeaders:            headers,
		Enabled:                 enabled,
	}
	if err := tx.Create(&provider).Error; err != nil {
		return Provider{}, err
	}
	return provider, nil
}

func buildProviderUpdates(input UpdateProviderModelInput) map[string]any {
	updates := map[string]any{}
	if input.BaseURL != nil {
		next := normalizeBaseURL(*input.BaseURL)
		if next != "" {
			updates["base_url"] = next
		}
	}
	if input.APIKey != nil {
		next := strings.TrimSpace(*input.APIKey)
		if next != "" {
			updates["api_key"] = next
		}
	}
	if input.Organization != nil {
		updates["organization"] = strings.TrimSpace(*input.Organization)
	}
	if input.ChatCompletionsPath != nil {
		updates["chat_completions_path"] = normalizePath(*input.ChatCompletionsPath, DefaultChatCompletionsPath)
	}
	if input.ChatResponsesPath != nil {
		updates["chat_responses_path"] = normalizePath(*input.ChatResponsesPath, DefaultChatResponsesPath)
	}
	if input.EmbeddingsPath != nil {
		updates["embeddings_path"] = normalizePath(*input.EmbeddingsPath, DefaultEmbeddingsPath)
	}
	if input.RerankPath != nil {
		updates["rerank_path"] = normalizePath(*input.RerankPath, DefaultRerankPath)
	}
	if input.AudioSpeechPath != nil {
		updates["audio_speech_path"] = normalizePath(*input.AudioSpeechPath, DefaultAudioSpeechPath)
	}
	if input.AudioTranscriptionsPath != nil {
		updates["audio_transcriptions_path"] = normalizePath(*input.AudioTranscriptionsPath, DefaultAudioTranscriptions)
	}
	if input.ModelsPath != nil {
		updates["models_path"] = normalizePath(*input.ModelsPath, DefaultModelsPath)
	}
	if input.ExtraHeaders != nil {
		updates["extra_headers"] = normalizeHeaders(*input.ExtraHeaders)
	}
	if input.Enabled != nil {
		updates["enabled"] = *input.Enabled
	}
	return updates
}

func buildModelUpdates(record ProviderModel, input UpdateProviderModelInput) map[string]any {
	updates := map[string]any{}
	if input.Model != nil {
		next := strings.TrimSpace(*input.Model)
		if next != "" {
			updates["model"] = next
		}
	}
	if input.RequestModel != nil {
		next := strings.TrimSpace(*input.RequestModel)
		if next != "" {
			updates["request_model"] = next
		}
	}
	if input.SupportsChatCompletions != nil {
		updates["supports_chat_completions"] = *input.SupportsChatCompletions
	}
	if input.SupportsChatResponses != nil {
		updates["supports_chat_responses"] = *input.SupportsChatResponses
	}
	if input.SupportsEmbeddings != nil {
		updates["supports_embeddings"] = *input.SupportsEmbeddings
	}
	if input.SupportsRerank != nil {
		updates["supports_rerank"] = *input.SupportsRerank
	}
	if input.SupportsAudioSpeech != nil {
		updates["supports_audio_speech"] = *input.SupportsAudioSpeech
	}
	if input.SupportsAudioTranscriptions != nil {
		updates["supports_audio_transcriptions"] = *input.SupportsAudioTranscriptions
	}
	if input.SupportsModels != nil {
		updates["supports_models"] = *input.SupportsModels
	}
	if input.Enabled != nil {
		updates["enabled"] = *input.Enabled
	}
	if _, ok := updates["request_model"]; !ok {
		if nextModel, updated := updates["model"]; updated && strings.TrimSpace(record.RequestModel) == strings.TrimSpace(record.Model) {
			updates["request_model"] = nextModel
		}
	}
	return updates
}

func applyProviderSnapshot(updates *map[string]any, provider Provider) {
	target := *updates
	target["provider_id"] = provider.ID
	target["provider"] = provider.Name
	target["base_url"] = provider.BaseURL
	target["api_key"] = provider.APIKey
	target["organization"] = provider.Organization
	target["chat_completions_path"] = provider.ChatCompletionsPath
	target["chat_responses_path"] = provider.ChatResponsesPath
	target["embeddings_path"] = provider.EmbeddingsPath
	target["rerank_path"] = provider.RerankPath
	target["audio_speech_path"] = provider.AudioSpeechPath
	target["audio_transcriptions_path"] = provider.AudioTranscriptionsPath
	target["models_path"] = provider.ModelsPath
	target["extra_headers"] = provider.ExtraHeaders
	*updates = target
}

func syncModelsByProvider(tx *gorm.DB, provider Provider) error {
	updates := map[string]any{}
	applyProviderSnapshot(&updates, provider)
	return tx.Model(&ProviderModel{}).Where("provider_id = ?", provider.ID).Updates(updates).Error
}

func boolOrDefault(value *bool, defaultValue bool) bool {
	if value == nil {
		return defaultValue
	}
	return *value
}
