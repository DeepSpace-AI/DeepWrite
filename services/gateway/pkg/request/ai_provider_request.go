package request

type CreateAIProviderModelRequest struct {
	Model                       string            `json:"model" binding:"required,max=120"`
	Provider                    string            `json:"provider" binding:"required,max=64"`
	RequestModel                string            `json:"request_model" binding:"omitempty,max=120"`
	BaseURL                     string            `json:"base_url" binding:"required,url,max=1024"`
	APIKey                      string            `json:"api_key" binding:"required,max=8192"`
	Organization                string            `json:"organization" binding:"omitempty,max=255"`
	ChatCompletionsPath         string            `json:"chat_completions_path" binding:"omitempty,max=255"`
	ChatResponsesPath           string            `json:"chat_responses_path" binding:"omitempty,max=255"`
	EmbeddingsPath              string            `json:"embeddings_path" binding:"omitempty,max=255"`
	RerankPath                  string            `json:"rerank_path" binding:"omitempty,max=255"`
	AudioSpeechPath             string            `json:"audio_speech_path" binding:"omitempty,max=255"`
	AudioTranscriptionsPath     string            `json:"audio_transcriptions_path" binding:"omitempty,max=255"`
	ModelsPath                  string            `json:"models_path" binding:"omitempty,max=255"`
	ExtraHeaders                map[string]string `json:"extra_headers"`
	SupportsChatCompletions     *bool             `json:"supports_chat_completions"`
	SupportsChatResponses       *bool             `json:"supports_chat_responses"`
	SupportsEmbeddings          *bool             `json:"supports_embeddings"`
	SupportsRerank              *bool             `json:"supports_rerank"`
	SupportsAudioSpeech         *bool             `json:"supports_audio_speech"`
	SupportsAudioTranscriptions *bool             `json:"supports_audio_transcriptions"`
	SupportsModels              *bool             `json:"supports_models"`
	Enabled                     *bool             `json:"enabled"`
}

type UpdateAIProviderModelRequest struct {
	Model                       *string            `json:"model" binding:"omitempty,max=120"`
	Provider                    *string            `json:"provider" binding:"omitempty,max=64"`
	RequestModel                *string            `json:"request_model" binding:"omitempty,max=120"`
	BaseURL                     *string            `json:"base_url" binding:"omitempty,url,max=1024"`
	APIKey                      *string            `json:"api_key" binding:"omitempty,max=8192"`
	Organization                *string            `json:"organization" binding:"omitempty,max=255"`
	ChatCompletionsPath         *string            `json:"chat_completions_path" binding:"omitempty,max=255"`
	ChatResponsesPath           *string            `json:"chat_responses_path" binding:"omitempty,max=255"`
	EmbeddingsPath              *string            `json:"embeddings_path" binding:"omitempty,max=255"`
	RerankPath                  *string            `json:"rerank_path" binding:"omitempty,max=255"`
	AudioSpeechPath             *string            `json:"audio_speech_path" binding:"omitempty,max=255"`
	AudioTranscriptionsPath     *string            `json:"audio_transcriptions_path" binding:"omitempty,max=255"`
	ModelsPath                  *string            `json:"models_path" binding:"omitempty,max=255"`
	ExtraHeaders                *map[string]string `json:"extra_headers"`
	SupportsChatCompletions     *bool              `json:"supports_chat_completions"`
	SupportsChatResponses       *bool              `json:"supports_chat_responses"`
	SupportsEmbeddings          *bool              `json:"supports_embeddings"`
	SupportsRerank              *bool              `json:"supports_rerank"`
	SupportsAudioSpeech         *bool              `json:"supports_audio_speech"`
	SupportsAudioTranscriptions *bool              `json:"supports_audio_transcriptions"`
	SupportsModels              *bool              `json:"supports_models"`
	Enabled                     *bool              `json:"enabled"`
}

type CreateAIProviderVendorRequest struct {
	Provider                string            `json:"provider" binding:"required,max=64"`
	BaseURL                 string            `json:"base_url" binding:"required,url,max=1024"`
	APIKey                  string            `json:"api_key" binding:"required,max=8192"`
	Organization            string            `json:"organization" binding:"omitempty,max=255"`
	ChatCompletionsPath     string            `json:"chat_completions_path" binding:"omitempty,max=255"`
	ChatResponsesPath       string            `json:"chat_responses_path" binding:"omitempty,max=255"`
	EmbeddingsPath          string            `json:"embeddings_path" binding:"omitempty,max=255"`
	RerankPath              string            `json:"rerank_path" binding:"omitempty,max=255"`
	AudioSpeechPath         string            `json:"audio_speech_path" binding:"omitempty,max=255"`
	AudioTranscriptionsPath string            `json:"audio_transcriptions_path" binding:"omitempty,max=255"`
	ModelsPath              string            `json:"models_path" binding:"omitempty,max=255"`
	ExtraHeaders            map[string]string `json:"extra_headers"`
	Enabled                 *bool             `json:"enabled"`
}

type CreateAIProviderModelByVendorRequest struct {
	Model                       string `json:"model" binding:"required,max=120"`
	RequestModel                string `json:"request_model" binding:"omitempty,max=120"`
	SupportsChatCompletions     *bool  `json:"supports_chat_completions"`
	SupportsChatResponses       *bool  `json:"supports_chat_responses"`
	SupportsEmbeddings          *bool  `json:"supports_embeddings"`
	SupportsRerank              *bool  `json:"supports_rerank"`
	SupportsAudioSpeech         *bool  `json:"supports_audio_speech"`
	SupportsAudioTranscriptions *bool  `json:"supports_audio_transcriptions"`
	SupportsModels              *bool  `json:"supports_models"`
	Enabled                     *bool  `json:"enabled"`
}

// UpdateAIProviderVendorEnabledRequest 用于启用或停用指定 AI Provider 厂商。
type UpdateAIProviderVendorEnabledRequest struct {
	// Enabled 为 true 时启用厂商，为 false 时停用厂商。
	Enabled *bool `json:"enabled" binding:"required" example:"true"`
}
