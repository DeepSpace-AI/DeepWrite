package handler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	aimodel "github.com/deepwrite/serivces/gateway/models/ai"
	"github.com/deepwrite/serivces/gateway/pkg/request"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AIProviderHandler struct {
}

// @Summary      AI 模型配置列表
// @Description  获取 AI 提供商模型配置列表，支持按 enabled 过滤
// @Tags         AI
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        enabled query bool false "是否启用"
// @Param        limit query int false "分页大小"
// @Param        offset query int false "偏移量"
// @Success      200 {object} response.Response "获取成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限访问"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /ai/providers [get]
func (h *AIProviderHandler) List(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	limit, offset := request.ParsePagination(c)
	enabled, ok := parseOptionalBool(c.Query("enabled"))
	if !ok {
		response.Failed(c, response.ErrorBadRequestCode, "enabled must be true or false")
		return
	}

	records, err := aimodel.ListProviderModels(c.Request.Context(), enabled, limit, offset)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "获取 AI 模型配置失败")
		return
	}

	items := make([]map[string]any, 0, len(records))
	for _, record := range records {
		items = append(items, toProviderModelView(record))
	}

	response.Success(c, response.SuccessCode, gin.H{
		"items":  items,
		"limit":  limit,
		"offset": offset,
	})
}

// @Summary      获取 AI 模型配置
// @Description  根据模型标识获取单个 AI 提供商配置
// @Tags         AI
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        model path string true "模型标识"
// @Success      200 {object} response.Response "获取成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限访问"
// @Failure      404 {object} response.Response "模型配置不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /ai/providers/{model} [get]
func (h *AIProviderHandler) GetByModel(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	model := strings.TrimSpace(c.Param("model"))
	if model == "" {
		response.Failed(c, response.ErrorBadRequestCode, "model is required")
		return
	}

	record, err := aimodel.GetProviderModelByModel(c.Request.Context(), model)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "模型配置不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取 AI 模型配置失败")
		return
	}

	response.Success(c, response.SuccessCode, toProviderModelView(record))
}

// @Summary      创建 AI 模型配置
// @Description  创建一个新的 AI 提供商模型配置（模型、密钥、端点等）
// @Tags         AI
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body request.CreateAIProviderModelRequest true "AI 模型配置"
// @Success      201 {object} response.Response "创建成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限访问"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /ai/providers [post]
func (h *AIProviderHandler) Create(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	var req request.CreateAIProviderModelRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	record, err := aimodel.CreateProviderModel(c.Request.Context(), aimodel.CreateProviderModelInput{
		Model:                       req.Model,
		Provider:                    req.Provider,
		RequestModel:                req.RequestModel,
		BaseURL:                     req.BaseURL,
		APIKey:                      req.APIKey,
		Organization:                req.Organization,
		ChatCompletionsPath:         req.ChatCompletionsPath,
		ChatResponsesPath:           req.ChatResponsesPath,
		EmbeddingsPath:              req.EmbeddingsPath,
		RerankPath:                  req.RerankPath,
		AudioSpeechPath:             req.AudioSpeechPath,
		AudioTranscriptionsPath:     req.AudioTranscriptionsPath,
		ModelsPath:                  req.ModelsPath,
		ExtraHeaders:                req.ExtraHeaders,
		SupportsChatCompletions:     req.SupportsChatCompletions,
		SupportsChatResponses:       req.SupportsChatResponses,
		SupportsEmbeddings:          req.SupportsEmbeddings,
		SupportsRerank:              req.SupportsRerank,
		SupportsAudioSpeech:         req.SupportsAudioSpeech,
		SupportsAudioTranscriptions: req.SupportsAudioTranscriptions,
		SupportsModels:              req.SupportsModels,
		Enabled:                     req.Enabled,
	})
	if err != nil {
		if isDuplicateModelErr(err) {
			response.Failed(c, response.ErrorBadRequestCode, "model already exists")
			return
		}
		if isInvalidInputErr(err) {
			response.Failed(c, response.ErrorBadRequestCode, err.Error())
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "创建 AI 模型配置失败")
		return
	}

	response.Success(c, response.SuccessCreatedCode, toProviderModelView(record))
}

// @Summary      更新 AI 模型配置
// @Description  根据模型标识更新 AI 提供商模型配置（可部分更新）
// @Tags         AI
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        model path string true "模型标识"
// @Param        request body request.UpdateAIProviderModelRequest true "AI 模型配置更新参数"
// @Success      200 {object} response.Response "更新成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限访问"
// @Failure      404 {object} response.Response "模型配置不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /ai/providers/{model} [put]
func (h *AIProviderHandler) Update(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	model := strings.TrimSpace(c.Param("model"))
	if model == "" {
		response.Failed(c, response.ErrorBadRequestCode, "model is required")
		return
	}

	var req request.UpdateAIProviderModelRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	record, err := aimodel.UpdateProviderModel(c.Request.Context(), model, aimodel.UpdateProviderModelInput{
		Model:                       req.Model,
		Provider:                    req.Provider,
		RequestModel:                req.RequestModel,
		BaseURL:                     req.BaseURL,
		APIKey:                      req.APIKey,
		Organization:                req.Organization,
		ChatCompletionsPath:         req.ChatCompletionsPath,
		ChatResponsesPath:           req.ChatResponsesPath,
		EmbeddingsPath:              req.EmbeddingsPath,
		RerankPath:                  req.RerankPath,
		AudioSpeechPath:             req.AudioSpeechPath,
		AudioTranscriptionsPath:     req.AudioTranscriptionsPath,
		ModelsPath:                  req.ModelsPath,
		ExtraHeaders:                req.ExtraHeaders,
		SupportsChatCompletions:     req.SupportsChatCompletions,
		SupportsChatResponses:       req.SupportsChatResponses,
		SupportsEmbeddings:          req.SupportsEmbeddings,
		SupportsRerank:              req.SupportsRerank,
		SupportsAudioSpeech:         req.SupportsAudioSpeech,
		SupportsAudioTranscriptions: req.SupportsAudioTranscriptions,
		SupportsModels:              req.SupportsModels,
		Enabled:                     req.Enabled,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "模型配置不存在")
			return
		}
		if isDuplicateModelErr(err) {
			response.Failed(c, response.ErrorBadRequestCode, "model already exists")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "更新 AI 模型配置失败")
		return
	}

	response.Success(c, response.SuccessCode, toProviderModelView(record))
}

// @Summary      删除 AI 模型配置
// @Description  根据模型标识删除 AI 提供商配置
// @Tags         AI
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        model path string true "模型标识"
// @Success      200 {object} response.Response "删除成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限访问"
// @Failure      404 {object} response.Response "模型配置不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /ai/providers/{model} [delete]
func (h *AIProviderHandler) Delete(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	model := strings.TrimSpace(c.Param("model"))
	if model == "" {
		response.Failed(c, response.ErrorBadRequestCode, "model is required")
		return
	}

	deleted, err := aimodel.DeleteProviderModel(c.Request.Context(), model)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "删除 AI 模型配置失败")
		return
	}
	if !deleted {
		response.Failed(c, 404, "模型配置不存在")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{"model": model, "deleted": true})
}

func ensureAdmin(c *gin.Context) bool {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return false
	}
	userRole := strings.TrimSpace(c.GetString("user_role"))
	if !strings.EqualFold(userRole, "admin") {
		response.Failed(c, response.ErrorForbiddenCode, "需要管理员权限")
		return false
	}
	return true
}

func parseOptionalBool(raw string) (*bool, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, true
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, false
	}
	return &parsed, true
}

func toProviderModelView(record aimodel.ProviderModel) map[string]any {
	headers := map[string]string{}
	for key, value := range record.ExtraHeaders {
		headers[key] = fmt.Sprint(value)
	}
	return map[string]any{
		"id":                            record.ID,
		"provider_id":                   record.ProviderID,
		"model":                         record.Model,
		"provider":                      record.Provider,
		"request_model":                 record.RequestModel,
		"supports_chat_completions":     record.SupportsChatCompletions,
		"supports_chat_responses":       record.SupportsChatResponses,
		"supports_embeddings":           record.SupportsEmbeddings,
		"supports_rerank":               record.SupportsRerank,
		"supports_audio_speech":         record.SupportsAudioSpeech,
		"supports_audio_transcriptions": record.SupportsAudioTranscriptions,
		"supports_models":               record.SupportsModels,
		"base_url":                      record.BaseURL,
		"organization":                  record.Organization,
		"chat_completions_path":         record.ChatCompletionsPath,
		"chat_responses_path":           record.ChatResponsesPath,
		"embeddings_path":               record.EmbeddingsPath,
		"rerank_path":                   record.RerankPath,
		"audio_speech_path":             record.AudioSpeechPath,
		"audio_transcriptions_path":     record.AudioTranscriptionsPath,
		"models_path":                   record.ModelsPath,
		"extra_headers":                 headers,
		"enabled":                       record.Enabled,
		"has_api_key":                   strings.TrimSpace(record.APIKey) != "",
		"api_key_masked":                maskSecret(record.APIKey),
		"created_at":                    record.CreatedAt,
		"updated_at":                    record.UpdatedAt,
	}
}

func maskSecret(value string) string {
	secret := strings.TrimSpace(value)
	if secret == "" {
		return ""
	}
	if len(secret) <= 6 {
		return strings.Repeat("*", len(secret))
	}
	return secret[:3] + strings.Repeat("*", len(secret)-6) + secret[len(secret)-3:]
}

func isDuplicateModelErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}

func isInvalidInputErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "required")
}
