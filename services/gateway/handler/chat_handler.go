package handler

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	agentmodel "github.com/deepwrite/serivces/gateway/models/agent"
	aimodel "github.com/deepwrite/serivces/gateway/models/ai"
	systemmodel "github.com/deepwrite/serivces/gateway/models/system"
	"github.com/deepwrite/serivces/gateway/pkg/crypto"
	"github.com/deepwrite/serivces/gateway/pkg/request"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ChatHandler struct{}

type SendMessageRequest struct {
	Content     string   `json:"content" binding:"required"`
	Stream      bool     `json:"stream"`
	ContextDocs []string `json:"context_docs"`
}

type AgentChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type titleGenerationTask struct {
	sessionID string
	model     aimodel.ProviderModel
	apiKey    string
	userID    string
}

var (
	titleTaskChan = make(chan titleGenerationTask, 100)
	titleWorkerWg sync.WaitGroup
)

func init() {
	startTitleGenerationWorker()
}

func startTitleGenerationWorker() {
	for i := 0; i < 3; i++ {
		titleWorkerWg.Add(1)
		go titleGenerationWorker()
	}
}

func titleGenerationWorker() {
	defer titleWorkerWg.Done()
	for task := range titleTaskChan {
		generateAndSaveTitle(task.sessionID, task.model, task.apiKey, task.userID)
	}
}

func generateAndSaveTitle(sessionID string, chatModel aimodel.ProviderModel, apiKey string, userID string) {
	log.Printf("[TitleWorker] Starting title generation for session %s", sessionID)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	messages, err := agentmodel.GetMessagesBySession(ctx, sessionID, 10, 0)
	if err != nil {
		log.Printf("[TitleWorker] Failed to get messages for session %s: %v", sessionID, err)
		return
	}

	log.Printf("[TitleWorker] Found %d messages for session %s", len(messages), sessionID)

	if len(messages) == 0 {
		log.Printf("[TitleWorker] No messages found for session %s", sessionID)
		return
	}

	model, modelAPIKey := getSystemAIModel(ctx, systemmodel.ConfigKeyTitleGenModel, chatModel, apiKey)
	log.Printf("[TitleWorker] Using model: %s (provider: %s, base_url: %s)", model.Model, model.Provider, model.BaseURL)

	title := generateSessionTitle(model, modelAPIKey, messages)
	log.Printf("[TitleWorker] AI generated title: '%s' for session %s", title, sessionID)

	if title == "" {
		title = extractSimpleTitle(messages)
		log.Printf("[TitleWorker] Using simple title: '%s' for session %s", title, sessionID)
	}

	if title == "" {
		log.Printf("[TitleWorker] Failed to generate any title for session %s", sessionID)
		return
	}

	if err := agentmodel.UpdateSessionTitle(ctx, sessionID, title); err != nil {
		log.Printf("[TitleWorker] Failed to update title for session %s: %v", sessionID, err)
		return
	}

	log.Printf("[TitleWorker] SUCCESS - Updated title for session %s: %s", sessionID, title)

	PublishSessionEvent(userID, sessionID, "title_updated", map[string]string{
		"title": title,
	})
}

func extractSimpleTitle(messages []agentmodel.Message) string {
	for _, msg := range messages {
		if msg.Role == "user" {
			content := strings.TrimSpace(msg.Content)
			if len(content) > 50 {
				return content[:47] + "..."
			}
			return content
		}
	}
	return "New Conversation"
}

func getSystemAIModel(ctx context.Context, configKey string, fallbackModel aimodel.ProviderModel, fallbackAPIKey string) (aimodel.ProviderModel, string) {
	modelID, err := systemmodel.GetConfig(ctx, configKey)
	if err != nil {
		log.Printf("[SystemConfig] Error reading config '%s': %v, using fallback model", configKey, err)
		return fallbackModel, fallbackAPIKey
	}

	if modelID == "" {
		modelID, err = systemmodel.GetConfig(ctx, systemmodel.ConfigKeyDefaultModel)
		if err != nil || modelID == "" {
			log.Printf("[SystemConfig] No default model configured, using fallback model")
			return fallbackModel, fallbackAPIKey
		}
	}

	model, err := aimodel.GetProviderModelByModel(ctx, modelID)
	if err != nil {
		log.Printf("[SystemConfig] Configured model '%s' not found: %v, using fallback model", modelID, err)
		return fallbackModel, fallbackAPIKey
	}

	if !model.Enabled {
		log.Printf("[SystemConfig] Configured model '%s' is disabled, using fallback model", modelID)
		return fallbackModel, fallbackAPIKey
	}

	apiKey, err := crypto.DecryptAPIKey(model.APIKey)
	if err != nil {
		log.Printf("[SystemConfig] Failed to decrypt API key for '%s': %v, using fallback model", modelID, err)
		return fallbackModel, fallbackAPIKey
	}

	log.Printf("[SystemConfig] Using configured model '%s' (provider: %s)", modelID, model.Provider)
	return model, apiKey
}

func submitTitleGenerationTask(sessionID string, model aimodel.ProviderModel, apiKey string, userID string) {
	log.Printf("[TitleTask] Submitting task for session %s, model: %s, user: %s", sessionID, model.Model, userID)
	select {
	case titleTaskChan <- titleGenerationTask{
		sessionID: sessionID,
		model:     model,
		apiKey:    apiKey,
		userID:    userID,
	}:
		log.Printf("[TitleTask] Task submitted successfully for session %s (queue size: %d)", sessionID, len(titleTaskChan))
	default:
		log.Printf("[TitleTask] Task queue full (%d), skipping title generation for session %s", len(titleTaskChan), sessionID)
	}
}

func (h *ChatHandler) GenerateTitle(c *gin.Context) {
	sessionID := strings.TrimSpace(c.Param("id"))
	if sessionID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "session_id is required")
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	session, err := agentmodel.GetSessionByID(c.Request.Context(), sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "Session not found")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "Failed to get session")
		return
	}

	if session.UserID != userID {
		response.Failed(c, 403, "You don't have permission to access this session")
		return
	}

	messages, err := agentmodel.GetMessagesBySession(c.Request.Context(), sessionID, 10, 0)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get messages")
		return
	}

	if len(messages) == 0 {
		response.Failed(c, response.ErrorBadRequestCode, "No messages to generate title from")
		return
	}

	agent, err := agentmodel.GetAgentByID(c.Request.Context(), session.AgentID)
	if err != nil {
		response.Failed(c, 404, "Agent not found")
		return
	}

	var modelRecord aimodel.ProviderModel
	if agent.DefaultModel != "" {
		modelRecord, err = aimodel.GetProviderModelByModel(c.Request.Context(), agent.DefaultModel)
	}
	if err != nil || agent.DefaultModel == "" {
		if errors.Is(err, gorm.ErrRecordNotFound) || agent.DefaultModel == "" {
			enabled := true
			models, listErr := aimodel.ListProviderModels(c.Request.Context(), &enabled, 1, 0)
			if listErr != nil || len(models) == 0 {
				response.Failed(c, response.ErrorUnknownCode, "No available AI model")
				return
			}
			modelRecord = models[0]
		} else {
			response.Failed(c, response.ErrorUnknownCode, "Failed to get model config")
			return
		}
	}

	apiKey, err := crypto.DecryptAPIKey(modelRecord.APIKey)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to decrypt API key")
		return
	}

	title := generateSessionTitle(modelRecord, apiKey, messages)
	if title == "" {
		response.Failed(c, response.ErrorUnknownCode, "Failed to generate title")
		return
	}

	if err := agentmodel.UpdateSessionTitle(c.Request.Context(), sessionID, title); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to update session title")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{"title": title})
}

func generateSessionTitle(model aimodel.ProviderModel, apiKey string, messages []agentmodel.Message) string {
	var conversationText strings.Builder
	for _, msg := range messages {
		if msg.Role == "user" {
			conversationText.WriteString("User: ")
		} else {
			conversationText.WriteString("Assistant: ")
		}
		if len(msg.Content) > 500 {
			conversationText.WriteString(msg.Content[:500])
		} else {
			conversationText.WriteString(msg.Content)
		}
		conversationText.WriteString("\n")
	}

	systemPrompt := `You are a title generator. Generate a concise, descriptive title (maximum 50 characters) for the following conversation. The title should summarize the main topic or question. Return ONLY the title, nothing else. Do not include quotes or punctuation at the end.`

	chatURL := strings.TrimSuffix(model.BaseURL, "/") + model.ChatCompletionsPath

	chatMessages := []AgentChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: "Generate a title for this conversation:\n\n" + conversationText.String()},
	}

	reqBody := map[string]interface{}{
		"model":       model.RequestModel,
		"messages":    chatMessages,
		"max_tokens":  20,
		"temperature": 0.3,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[GenerateTitle] Failed to marshal request: %v", err)
		return ""
	}

	log.Printf("[GenerateTitle] Request URL: %s", chatURL)
	log.Printf("[GenerateTitle] Request Body: %s", string(reqJSON))

	client := &http.Client{Timeout: 30 * time.Second}
	httpReq, err := http.NewRequest(http.MethodPost, chatURL, bytes.NewReader(reqJSON))
	if err != nil {
		log.Printf("[GenerateTitle] Failed to create request: %v", err)
		return ""
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	if model.Organization != "" {
		httpReq.Header.Set("OpenAI-Organization", model.Organization)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("[GenerateTitle] Request failed: %v", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("[GenerateTitle] Response Status: %d", resp.StatusCode)
	log.Printf("[GenerateTitle] Response Body: %s", string(body))

	if resp.StatusCode == 429 {
		log.Printf("[GenerateTitle] Rate limit hit (429), will use fallback title")
		return ""
	}

	if resp.StatusCode >= 400 {
		log.Printf("[GenerateTitle] AI service error: %d - %s", resp.StatusCode, string(body))
		return ""
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		log.Printf("[GenerateTitle] Failed to parse response: %v", err)
		return ""
	}

	if len(chatResp.Choices) == 0 {
		log.Printf("[GenerateTitle] No choices in response")
		return ""
	}

	title := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	if len(title) > 100 {
		title = title[:100] + "..."
	}

	log.Printf("[GenerateTitle] Generated title: %s", title)
	return title
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	h.handleChat(c, false)
}

func (h *ChatHandler) SendMessageStream(c *gin.Context) {
	h.handleChat(c, true)
}

func (h *ChatHandler) handleChat(c *gin.Context, stream bool) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	sessionID := strings.TrimSpace(c.Param("id"))
	if sessionID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "session_id is required")
		return
	}

	session, err := agentmodel.GetSessionByID(c.Request.Context(), sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "Session not found")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "Failed to get session")
		return
	}

	if session.UserID != userID {
		response.Failed(c, 403, "You don't have permission to access this session")
		return
	}

	var req SendMessageRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	agent, err := agentmodel.GetAgentByID(c.Request.Context(), session.AgentID)
	if err != nil {
		response.Failed(c, 404, "Agent not found")
		return
	}

	if !agent.Enabled {
		response.Failed(c, 400, "Agent is disabled")
		return
	}

	userMsg, err := agentmodel.CreateMessage(c.Request.Context(), agentmodel.CreateMessageInput{
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Content,
	})
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to create user message")
		return
	}

	messages, err := agentmodel.GetMessagesBySession(c.Request.Context(), sessionID, 50, 0)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get message history")
		return
	}

	promptCtx := agentmodel.BuildPromptContext(c.Request.Context(), userID, &session)

	if agent.InjectMemory {
		retriever := agentmodel.NewMemoryRetriever()
		retrievedMems, memErr := retriever.RetrieveRelevant(
			c.Request.Context(),
			agent.ID,
			userID,
			req.Content,
			agent.MemoryRetrievalCount,
			agent.MemoryMinRelevance,
		)
		if memErr == nil && len(retrievedMems) > 0 {
			promptCtx.RetrievedMemories = retrievedMems
		}
	}

	builder := agentmodel.NewPromptBuilder(&agent, &promptCtx)
	assembled := builder.Assemble()

	log.Printf("[AgentChat] Prompt assembled for agent %s:", agent.Name)
	log.Printf("[AgentChat] - Identity Layer: %d chars", len(assembled.Layers[agentmodel.LayerIdentity]))
	log.Printf("[AgentChat] - Full Prompt length: %d chars", len(assembled.FullPrompt))
	log.Printf("[AgentChat] - Full Prompt:\n%s", assembled.FullPrompt)

	chatMessages := buildChatMessages(assembled.FullPrompt, messages, req.Content)

	var modelRecord aimodel.ProviderModel
	if agent.DefaultModel != "" {
		modelRecord, err = aimodel.GetProviderModelByModel(c.Request.Context(), agent.DefaultModel)
	}
	if err != nil || agent.DefaultModel == "" {
		if errors.Is(err, gorm.ErrRecordNotFound) || agent.DefaultModel == "" {
			enabled := true
			models, listErr := aimodel.ListProviderModels(c.Request.Context(), &enabled, 1, 0)
			if listErr != nil || len(models) == 0 {
				response.Failed(c, response.ErrorUnknownCode, "No available AI model")
				return
			}
			modelRecord = models[0]
		} else {
			response.Failed(c, response.ErrorUnknownCode, "Failed to get model config")
			return
		}
	}

	if !modelRecord.Enabled {
		response.Failed(c, 400, "Model is disabled")
		return
	}

	apiKey, err := crypto.DecryptAPIKey(modelRecord.APIKey)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to decrypt API key")
		return
	}

	chatURL := strings.TrimSuffix(modelRecord.BaseURL, "/") + modelRecord.ChatCompletionsPath

	reqBody := map[string]interface{}{
		"model":    modelRecord.RequestModel,
		"messages": chatMessages,
		"stream":   stream,
	}

	if agent.Temperature > 0 {
		reqBody["temperature"] = agent.Temperature
	}
	if agent.MaxTokens > 0 {
		reqBody["max_tokens"] = agent.MaxTokens
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to build request")
		return
	}

	log.Printf("[AgentChat] Session: %s, Agent: %s, Model: %s", sessionID, agent.Name, modelRecord.Model)

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, chatURL, bytes.NewReader(reqJSON))
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to create request")
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	if modelRecord.Organization != "" {
		httpReq.Header.Set("OpenAI-Organization", modelRecord.Organization)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "AI service request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[AgentChat] AI service error: %d - %s", resp.StatusCode, string(body))
		response.Failed(c, resp.StatusCode, "AI service error: "+string(body))
		return
	}

	if stream {
		h.handleAgentStreamResponse(c, resp, sessionID, modelRecord, apiKey, session.Title == "", userID)
	} else {
		h.handleAgentNonStreamResponse(c, resp, sessionID, userMsg, modelRecord, apiKey, session.Title == "", userID)
	}
}

func (h *ChatHandler) handleAgentStreamResponse(c *gin.Context, resp *http.Response, sessionID string, model aimodel.ProviderModel, apiKey string, needTitle bool, userID string) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	reader := bufio.NewReader(resp.Body)
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.Failed(c, response.ErrorUnknownCode, "Streaming not supported")
		return
	}

	var fullContent strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("[AgentChat] Stream read error: %v", err)
			break
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		data = strings.TrimSpace(data)

		if data == "[DONE]" {
			c.Writer.WriteString("data: [DONE]\n\n")
			flusher.Flush()
			break
		}

		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		choices, ok := chunk["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			continue
		}

		choice, ok := choices[0].(map[string]interface{})
		if !ok {
			continue
		}

		delta, ok := choice["delta"].(map[string]interface{})
		if !ok {
			continue
		}

		content, _ := delta["content"].(string)
		if content != "" {
			fullContent.WriteString(content)
		}

		c.Writer.WriteString(line)
		flusher.Flush()
	}

	if fullContent.Len() > 0 {
		_, err := agentmodel.CreateMessage(c.Request.Context(), agentmodel.CreateMessageInput{
			SessionID: sessionID,
			Role:      "assistant",
			Content:   fullContent.String(),
		})
		if err != nil {
			log.Printf("[AgentChat] Failed to save assistant message: %v", err)
		}

		if err := agentmodel.UpdateSessionLastMessage(c.Request.Context(), sessionID); err != nil {
			log.Printf("[AgentChat] Failed to update session last message: %v", err)
		}

		if needTitle {
			submitTitleGenerationTask(sessionID, model, apiKey, userID)
		}
	}
}

func (h *ChatHandler) handleAgentNonStreamResponse(c *gin.Context, resp *http.Response, sessionID string, userMsg agentmodel.Message, model aimodel.ProviderModel, apiKey string, needTitle bool, userID string) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to read response")
		return
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to parse response")
		return
	}

	if len(chatResp.Choices) == 0 {
		response.Failed(c, response.ErrorUnknownCode, "No response from AI")
		return
	}

	assistantContent := chatResp.Choices[0].Message.Content

	assistantMsg, err := agentmodel.CreateMessage(c.Request.Context(), agentmodel.CreateMessageInput{
		SessionID:  sessionID,
		Role:       "assistant",
		Content:    assistantContent,
		TokenCount: chatResp.Usage.CompletionTokens,
		ModelUsed:  chatResp.Model,
	})
	if err != nil {
		log.Printf("[AgentChat] Failed to save assistant message: %v", err)
	}

	if err := agentmodel.UpdateSessionLastMessage(c.Request.Context(), sessionID); err != nil {
		log.Printf("[AgentChat] Failed to update session last message: %v", err)
	}

	if needTitle {
		submitTitleGenerationTask(sessionID, model, apiKey, userID)
	}

	response.Success(c, response.SuccessCode, gin.H{
		"user_message":      toMessageViewFromPtr(&userMsg),
		"assistant_message": toMessageViewFromPtr(&assistantMsg),
	})
}

func toMessageViewFromPtr(msg *agentmodel.Message) map[string]any {
	if msg == nil {
		return nil
	}
	return map[string]any{
		"id":          msg.ID,
		"session_id":  msg.SessionID,
		"role":        msg.Role,
		"content":     msg.Content,
		"token_count": msg.TokenCount,
		"model_used":  msg.ModelUsed,
		"created_at":  msg.CreatedAt,
	}
}

func buildChatMessages(systemPrompt string, history []agentmodel.Message, currentContent string) []AgentChatMessage {
	messages := make([]AgentChatMessage, 0, len(history)+2)

	if systemPrompt != "" {
		messages = append(messages, AgentChatMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	for _, msg := range history {
		messages = append(messages, AgentChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	messages = append(messages, AgentChatMessage{
		Role:    "user",
		Content: currentContent,
	})

	return messages
}
