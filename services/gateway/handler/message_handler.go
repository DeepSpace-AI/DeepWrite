package handler

import (
	"strings"

	agentmodel "github.com/deepwrite/serivces/gateway/models/agent"
	"github.com/deepwrite/serivces/gateway/pkg/request"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

type MessageHandler struct{}

func (h *MessageHandler) List(c *gin.Context) {
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

	isOwner, err := agentmodel.IsSessionOwner(c.Request.Context(), sessionID, userID)
	if err != nil {
		response.Failed(c, 404, "Session not found")
		return
	}
	if !isOwner {
		response.Failed(c, 403, "You don't have permission to access this session")
		return
	}

	limit, offset := request.ParsePagination(c)
	messages, err := agentmodel.GetMessagesBySession(c.Request.Context(), sessionID, limit, offset)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get messages")
		return
	}

	items := make([]map[string]any, 0, len(messages))
	for _, msg := range messages {
		items = append(items, toMessageView(&msg))
	}

	response.Success(c, response.SuccessCode, gin.H{
		"items":  items,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *MessageHandler) Create(c *gin.Context) {
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

	isOwner, err := agentmodel.IsSessionOwner(c.Request.Context(), sessionID, userID)
	if err != nil {
		response.Failed(c, 404, "Session not found")
		return
	}
	if !isOwner {
		response.Failed(c, 403, "You don't have permission to access this session")
		return
	}

	var req CreateMessageRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	msg, err := agentmodel.CreateMessage(c.Request.Context(), agentmodel.CreateMessageInput{
		SessionID:  sessionID,
		Role:       req.Role,
		Content:    req.Content,
		TokenCount: req.TokenCount,
		ModelUsed:  req.ModelUsed,
		Sources:    req.Sources,
		Actions:    req.Actions,
	})
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to create message")
		return
	}

	response.Success(c, response.SuccessCreatedCode, toMessageView(&msg))
}

func toMessageView(msg *agentmodel.Message) map[string]any {
	return map[string]any{
		"id":          msg.ID,
		"session_id":  msg.SessionID,
		"role":        msg.Role,
		"content":     msg.Content,
		"token_count": msg.TokenCount,
		"model_used":  msg.ModelUsed,
		"sources":     msg.Sources,
		"actions":     msg.Actions,
		"created_at":  msg.CreatedAt,
	}
}

type CreateMessageRequest struct {
	Role       string `json:"role" binding:"required"`
	Content    string `json:"content" binding:"required"`
	TokenCount int    `json:"token_count"`
	ModelUsed  string `json:"model_used"`
	Sources    string `json:"sources"`
	Actions    string `json:"actions"`
}
