package ai

import (
	"time"

	"gorm.io/datatypes"
)

const (
	DefaultChatCompletionsPath  = "/chat/completions"
	DefaultChatResponsesPath    = "/responses"
	DefaultEmbeddingsPath       = "/embeddings"
	DefaultRerankPath           = "/rerank"
	DefaultAudioSpeechPath      = "/audio/speech"
	DefaultAudioTranscriptions  = "/audio/transcriptions"
	DefaultModelsPath           = "/models"
	DefaultProvider             = "openai-compatible"
	DefaultModelConfigTableName = "dw_ai_models"
	DefaultProviderTableName    = "dw_ai_providers"
)

type Provider struct {
	ID                      string            `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name                    string            `json:"name" gorm:"type:varchar(64);not null;uniqueIndex:uk_dw_ai_providers_name"`
	BaseURL                 string            `json:"base_url" gorm:"type:varchar(1024);not null"`
	APIKey                  string            `json:"-" gorm:"column:api_key;type:text;not null"`
	Organization            string            `json:"organization" gorm:"type:varchar(255)"`
	ChatCompletionsPath     string            `json:"chat_completions_path" gorm:"type:varchar(255);not null;default:'/chat/completions'"`
	ChatResponsesPath       string            `json:"chat_responses_path" gorm:"type:varchar(255);not null;default:'/responses'"`
	EmbeddingsPath          string            `json:"embeddings_path" gorm:"type:varchar(255);not null;default:'/embeddings'"`
	RerankPath              string            `json:"rerank_path" gorm:"type:varchar(255);not null;default:'/rerank'"`
	AudioSpeechPath         string            `json:"audio_speech_path" gorm:"type:varchar(255);not null;default:'/audio/speech'"`
	AudioTranscriptionsPath string            `json:"audio_transcriptions_path" gorm:"type:varchar(255);not null;default:'/audio/transcriptions'"`
	ModelsPath              string            `json:"models_path" gorm:"type:varchar(255);not null;default:'/models'"`
	ExtraHeaders            datatypes.JSONMap `json:"extra_headers" gorm:"type:jsonb;not null;default:'{}'::jsonb"`
	Enabled                 bool              `json:"enabled" gorm:"not null;default:true"`
	CreatedAt               time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt               time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
}

type ProviderModel struct {
	ID                          string            `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ProviderID                  string            `json:"provider_id" gorm:"type:uuid;index"`
	Model                       string            `json:"model" gorm:"type:varchar(120);not null;uniqueIndex:uk_dw_ai_models_model"`
	RequestModel                string            `json:"request_model" gorm:"type:varchar(120);not null"`
	SupportsChatCompletions     bool              `json:"supports_chat_completions" gorm:"not null;default:true"`
	SupportsChatResponses       bool              `json:"supports_chat_responses" gorm:"not null;default:true"`
	SupportsEmbeddings          bool              `json:"supports_embeddings" gorm:"not null;default:true"`
	SupportsRerank              bool              `json:"supports_rerank" gorm:"not null;default:true"`
	SupportsAudioSpeech         bool              `json:"supports_audio_speech" gorm:"not null;default:true"`
	SupportsAudioTranscriptions bool              `json:"supports_audio_transcriptions" gorm:"not null;default:true"`
	SupportsModels              bool              `json:"supports_models" gorm:"not null;default:true"`
	Provider                    string            `json:"provider" gorm:"type:varchar(64);not null;default:'openai-compatible'"`
	BaseURL                     string            `json:"base_url" gorm:"type:varchar(1024);not null"`
	APIKey                      string            `json:"-" gorm:"column:api_key;type:text;not null"`
	Organization                string            `json:"organization" gorm:"type:varchar(255)"`
	ChatCompletionsPath         string            `json:"chat_completions_path" gorm:"type:varchar(255);not null;default:'/chat/completions'"`
	ChatResponsesPath           string            `json:"chat_responses_path" gorm:"type:varchar(255);not null;default:'/responses'"`
	EmbeddingsPath              string            `json:"embeddings_path" gorm:"type:varchar(255);not null;default:'/embeddings'"`
	RerankPath                  string            `json:"rerank_path" gorm:"type:varchar(255);not null;default:'/rerank'"`
	AudioSpeechPath             string            `json:"audio_speech_path" gorm:"type:varchar(255);not null;default:'/audio/speech'"`
	AudioTranscriptionsPath     string            `json:"audio_transcriptions_path" gorm:"type:varchar(255);not null;default:'/audio/transcriptions'"`
	ModelsPath                  string            `json:"models_path" gorm:"type:varchar(255);not null;default:'/models'"`
	ExtraHeaders                datatypes.JSONMap `json:"extra_headers" gorm:"type:jsonb;not null;default:'{}'::jsonb"`
	Enabled                     bool              `json:"enabled" gorm:"not null;default:true"`
	CreatedAt                   time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                   time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Provider) TableName() string {
	return DefaultProviderTableName
}

func (ProviderModel) TableName() string {
	return DefaultModelConfigTableName
}
