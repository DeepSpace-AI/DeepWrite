package handler

import (
	"strings"
	"time"

	agentmodel "github.com/deepwrite/serivces/gateway/models/agent"
	"github.com/deepwrite/serivces/gateway/pkg/request"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

type SessionShareHandler struct{}

func (h *SessionShareHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	shares, err := agentmodel.GetSessionSharesByUser(c.Request.Context(), userID)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get shares")
		return
	}

	items := make([]map[string]any, 0, len(shares))
	for _, share := range shares {
		items = append(items, toSessionShareView(&share))
	}

	response.Success(c, response.SuccessCode, gin.H{"items": items})
}

func (h *SessionShareHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	var req CreateSessionShareRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	isOwner, err := agentmodel.IsSessionOwner(c.Request.Context(), req.SessionID, userID)
	if err != nil {
		response.Failed(c, 404, "Session not found")
		return
	}
	if !isOwner {
		response.Failed(c, 403, "You don't have permission to share this session")
		return
	}

	session, err := agentmodel.GetSessionByID(c.Request.Context(), req.SessionID)
	if err != nil {
		response.Failed(c, 404, "Session not found")
		return
	}

	var expiresAt *time.Time
	if req.ExpiresInHours > 0 {
		t := time.Now().Add(time.Duration(req.ExpiresInHours) * time.Hour)
		expiresAt = &t
	}

	title := req.Title
	if title == "" {
		title = session.Title
	}

	share, err := agentmodel.CreateSessionShare(c.Request.Context(), agentmodel.CreateSessionShareInput{
		SessionID: req.SessionID,
		UserID:    userID,
		Title:     title,
		ExpiresAt: expiresAt,
		AllowCopy: req.AllowCopy,
		IsPublic:  req.IsPublic,
	})
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to create share")
		return
	}

	response.Success(c, response.SuccessCreatedCode, toSessionShareView(&share))
}

func (h *SessionShareHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	shareID := strings.TrimSpace(c.Param("id"))
	if shareID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "share_id is required")
		return
	}

	if err := agentmodel.DeleteSessionShare(c.Request.Context(), shareID); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to delete share")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{"id": shareID, "deleted": true})
}

func (h *SessionShareHandler) GetByToken(c *gin.Context) {
	token := strings.TrimSpace(c.Param("token"))
	if token == "" {
		response.Failed(c, response.ErrorBadRequestCode, "token is required")
		return
	}

	share, err := agentmodel.GetSessionShareByToken(c.Request.Context(), token)
	if err != nil {
		response.Failed(c, 404, "Share not found")
		return
	}

	if agentmodel.IsShareExpired(share) {
		response.Failed(c, 410, "Share has expired")
		return
	}

	session, err := agentmodel.GetSessionByID(c.Request.Context(), share.SessionID)
	if err != nil {
		response.Failed(c, 404, "Session not found")
		return
	}

	messages, err := agentmodel.GetMessagesBySession(c.Request.Context(), share.SessionID, 0, 0)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get messages")
		return
	}

	agentmodel.IncrementShareViewCount(c.Request.Context(), token)

	msgItems := make([]map[string]any, 0, len(messages))
	for _, msg := range messages {
		msgItems = append(msgItems, toMessageView(&msg))
	}

	response.Success(c, response.SuccessCode, gin.H{
		"share":    toSessionShareView(&share),
		"session":  toSessionView(&session),
		"messages": msgItems,
	})
}

func toSessionShareView(share *agentmodel.SessionShare) map[string]any {
	return map[string]any{
		"id":          share.ID,
		"session_id":  share.SessionID,
		"user_id":     share.UserID,
		"share_token": share.ShareToken,
		"title":       share.Title,
		"expires_at":  share.ExpiresAt,
		"view_count":  share.ViewCount,
		"allow_copy":  share.AllowCopy,
		"is_public":   share.IsPublic,
		"created_at":  share.CreatedAt,
		"updated_at":  share.UpdatedAt,
	}
}

type CreateSessionShareRequest struct {
	SessionID      string `json:"session_id" binding:"required"`
	Title          string `json:"title"`
	ExpiresInHours int    `json:"expires_in_hours"`
	AllowCopy      bool   `json:"allow_copy"`
	IsPublic       bool   `json:"is_public"`
}
