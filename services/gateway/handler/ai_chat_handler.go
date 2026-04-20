package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/models/ai"
	"github.com/deepwrite/serivces/gateway/pkg/crypto"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AIChatHandler struct{}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature *float64      `json:"temperature,omitempty"`
	TopP        *float64      `json:"top_p,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream"`
}

type ChatResponse struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []ChatResponseChoice `json:"choices"`
	Usage   *ChatResponseUsage   `json:"usage,omitempty"`
}

type ChatResponseChoice struct {
	Index        int                  `json:"index"`
	Message      *ChatResponseMessage `json:"message,omitempty"`
	Delta        *ChatResponseDelta   `json:"delta,omitempty"`
	FinishReason string               `json:"finish_reason"`
}

type ChatResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponseDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type ChatResponseUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func (h *AIChatHandler) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AIChat] Failed to bind JSON: %v", err)
		response.Failed(c, response.ErrorBadRequestCode, "Invalid request body")
		return
	}

	log.Printf("[AIChat] Request: model=%s, messages=%d, stream=%v", req.Model, len(req.Messages), req.Stream)

	if req.Model == "" {
		response.Failed(c, response.ErrorBadRequestCode, "model is required")
		return
	}

	if len(req.Messages) == 0 {
		response.Failed(c, response.ErrorBadRequestCode, "messages is required")
		return
	}

	record, err := ai.GetProviderModelByModel(c.Request.Context(), req.Model)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[AIChat] Model not found: %s", req.Model)
			response.Failed(c, 404, "Model not found")
			return
		}
		log.Printf("[AIChat] Failed to get model config: %v", err)
		response.Failed(c, response.ErrorUnknownCode, "Failed to get model config")
		return
	}

	log.Printf("[AIChat] Found model: %s, baseURL: %s, enabled: %v", record.Model, record.BaseURL, record.Enabled)

	if !record.Enabled {
		response.Failed(c, 400, "Model is disabled")
		return
	}

	apiKey, err := crypto.DecryptAPIKey(record.APIKey)
	if err != nil {
		log.Printf("[AIChat] Failed to decrypt API key: %v", err)
		response.Failed(c, response.ErrorUnknownCode, "Failed to decrypt API key")
		return
	}

	chatURL := strings.TrimSuffix(record.BaseURL, "/") + record.ChatCompletionsPath

	reqBody := map[string]interface{}{
		"model":    record.RequestModel,
		"messages": req.Messages,
		"stream":   req.Stream,
	}
	if req.Temperature != nil {
		reqBody["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		reqBody["top_p"] = *req.TopP
	}
	if req.MaxTokens != nil {
		reqBody["max_tokens"] = *req.MaxTokens
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to build request")
		return
	}

	log.Printf("[AIChat] Sending request to: %s", chatURL)
	log.Printf("[AIChat] Request body: %s", string(reqJSON))

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, chatURL, bytes.NewReader(reqJSON))
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to create request")
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	if record.Organization != "" {
		httpReq.Header.Set("OpenAI-Organization", record.Organization)
	}
	for key, value := range record.ExtraHeaders {
		if key != "" && value != "" {
			httpReq.Header.Set(key, strings.TrimSpace(value.(string)))
		}
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("[AIChat] Request failed: %v", err)
		response.Failed(c, response.ErrorUnknownCode, "AI service request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	log.Printf("[AIChat] Response status: %d", resp.StatusCode)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[AIChat] AI service error: %d - %s", resp.StatusCode, string(body))

		c.JSON(resp.StatusCode, gin.H{
			"error": map[string]interface{}{
				"message": string(body),
				"status":  resp.StatusCode,
			},
		})
		return
	}

	if req.Stream {
		h.handleStreamResponse(c, resp)
		return
	}

	h.handleNonStreamResponse(c, resp)
}

func (h *AIChatHandler) handleStreamResponse(c *gin.Context, resp *http.Response) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	reader := bufio.NewReader(resp.Body)
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		log.Printf("[AIChat] Streaming not supported")
		response.Failed(c, response.ErrorUnknownCode, "Streaming not supported")
		return
	}

	lineCount := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("[AIChat] Stream EOF, total lines: %d", lineCount)
				break
			}
			log.Printf("[AIChat] Stream read error: %v", err)
			break
		}

		lineCount++
		if lineCount <= 5 {
			log.Printf("[AIChat] Line %d: %s", lineCount, strings.TrimSpace(line))
		}

		c.Writer.WriteString(line)
		flusher.Flush()
	}

	log.Printf("[AIChat] Stream completed, total lines: %d", lineCount)
}

func (h *AIChatHandler) handleNonStreamResponse(c *gin.Context, resp *http.Response) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to read response")
		return
	}

	c.Data(resp.StatusCode, "application/json", body)
}
