package handler

import (
	"strings"

	agentmodel "github.com/deepwrite/serivces/gateway/models/agent"
	"github.com/deepwrite/serivces/gateway/pkg/request"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

type SessionHandler struct{}

func (h *SessionHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	limit, offset := request.ParsePagination(c)

	filter := agentmodel.SessionListFilter{
		AgentID: c.Query("agent_id"),
		Search:  c.Query("search"),
		GroupID: c.Query("group_id"),
	}

	if pinnedStr := c.Query("pinned"); pinnedStr != "" {
		pinned := pinnedStr == "true"
		filter.Pinned = &pinned
	}

	if archivedStr := c.Query("archived"); archivedStr != "" {
		archived := archivedStr == "true"
		filter.Archived = &archived
	} else {
		showArchived := false
		filter.Archived = &showArchived
	}

	sessions, err := agentmodel.GetSessionsByUser(c.Request.Context(), userID, filter, limit, offset)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get sessions")
		return
	}

	items := make([]map[string]any, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, toSessionView(&session))
	}

	response.Success(c, response.SuccessCode, gin.H{
		"items":  items,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *SessionHandler) GetByID(c *gin.Context) {
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
		response.Failed(c, 404, "Session not found")
		return
	}

	if session.UserID != userID {
		response.Failed(c, 403, "You don't have permission to access this session")
		return
	}

	limit, offset := request.ParsePagination(c)
	messages, err := agentmodel.GetMessagesBySession(c.Request.Context(), sessionID, limit, offset)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get messages")
		return
	}

	msgItems := make([]map[string]any, 0, len(messages))
	for _, msg := range messages {
		msgItems = append(msgItems, toMessageView(&msg))
	}

	response.Success(c, response.SuccessCode, gin.H{
		"session":  toSessionView(&session),
		"messages": msgItems,
	})
}

func (h *SessionHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	var req CreateSessionRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	agent, err := agentmodel.GetAgentByID(c.Request.Context(), req.AgentID)
	if err != nil {
		response.Failed(c, 404, "Agent not found")
		return
	}

	if !agent.Enabled {
		response.Failed(c, 400, "Agent is disabled")
		return
	}

	var workspaceID *string
	if req.WorkspaceID != "" {
		workspaceID = &req.WorkspaceID
	}

	session, err := agentmodel.CreateSession(c.Request.Context(), agentmodel.CreateSessionInput{
		AgentID:     req.AgentID,
		UserID:      userID,
		WorkspaceID: workspaceID,
		Title:       req.Title,
		ContextDocs: req.ContextDocs,
		ContextRefs: req.ContextRefs,
	})
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to create session")
		return
	}

	if err := agentmodel.IncrementAgentUsage(c.Request.Context(), req.AgentID); err != nil {
	}

	response.Success(c, response.SuccessCreatedCode, toSessionView(&session))
}

func (h *SessionHandler) Update(c *gin.Context) {
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
		response.Failed(c, 403, "You don't have permission to update this session")
		return
	}

	var req UpdateSessionRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	input := agentmodel.UpdateSessionInput{
		Title:       req.Title,
		ContextDocs: req.ContextDocs,
		ContextRefs: req.ContextRefs,
		Summary:     req.Summary,
		Tags:        req.Tags,
	}

	if req.Pinned != nil {
		input.Pinned = req.Pinned
	}
	if req.Archived != nil {
		input.Archived = req.Archived
	}
	if req.GroupID != nil {
		input.GroupID = req.GroupID
	}

	session, err := agentmodel.UpdateSession(c.Request.Context(), sessionID, input)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to update session")
		return
	}

	response.Success(c, response.SuccessCode, toSessionView(&session))
}

func (h *SessionHandler) Delete(c *gin.Context) {
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
		response.Failed(c, 403, "You don't have permission to delete this session")
		return
	}

	if err := agentmodel.DeleteSession(c.Request.Context(), sessionID); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to delete session")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{"id": sessionID, "deleted": true})
}

func (h *SessionHandler) Archive(c *gin.Context) {
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
		response.Failed(c, 403, "You don't have permission to archive this session")
		return
	}

	session, err := agentmodel.ArchiveSession(c.Request.Context(), sessionID)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to archive session")
		return
	}

	response.Success(c, response.SuccessCode, toSessionView(&session))
}

func (h *SessionHandler) Unarchive(c *gin.Context) {
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
		response.Failed(c, 403, "You don't have permission to unarchive this session")
		return
	}

	session, err := agentmodel.UnarchiveSession(c.Request.Context(), sessionID)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to unarchive session")
		return
	}

	response.Success(c, response.SuccessCode, toSessionView(&session))
}

func (h *SessionHandler) Pin(c *gin.Context) {
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
		response.Failed(c, 403, "You don't have permission to pin this session")
		return
	}

	session, err := agentmodel.PinSession(c.Request.Context(), sessionID)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to pin session")
		return
	}

	response.Success(c, response.SuccessCode, toSessionView(&session))
}

func (h *SessionHandler) Unpin(c *gin.Context) {
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
		response.Failed(c, 403, "You don't have permission to unpin this session")
		return
	}

	session, err := agentmodel.UnpinSession(c.Request.Context(), sessionID)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to unpin session")
		return
	}

	response.Success(c, response.SuccessCode, toSessionView(&session))
}

func toSessionView(session *agentmodel.Session) map[string]any {
	return map[string]any{
		"id":              session.ID,
		"agent_id":        session.AgentID,
		"user_id":         session.UserID,
		"workspace_id":    session.WorkspaceID,
		"title":           session.Title,
		"context_docs":    session.ContextDocs,
		"context_refs":    session.ContextRefs,
		"summary":         session.Summary,
		"token_count":     session.TokenCount,
		"status":          session.Status,
		"pinned":          session.Pinned,
		"archived":        session.Archived,
		"group_id":        session.GroupID,
		"tags":            session.Tags,
		"last_message_at": session.LastMessageAt,
		"created_at":      session.CreatedAt,
		"updated_at":      session.UpdatedAt,
	}
}

type CreateSessionRequest struct {
	AgentID     string `json:"agent_id" binding:"required"`
	WorkspaceID string `json:"workspace_id"`
	Title       string `json:"title"`
	ContextDocs string `json:"context_docs"`
	ContextRefs string `json:"context_refs"`
}

type UpdateSessionRequest struct {
	Title       string  `json:"title"`
	ContextDocs string  `json:"context_docs"`
	ContextRefs string  `json:"context_refs"`
	Summary     string  `json:"summary"`
	Pinned      *bool   `json:"pinned"`
	Archived    *bool   `json:"archived"`
	GroupID     *string `json:"group_id"`
	Tags        string  `json:"tags"`
}
