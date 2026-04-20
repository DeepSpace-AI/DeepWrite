package handler

import (
	"strings"
	"time"

	agentmodel "github.com/deepwrite/serivces/gateway/models/agent"
	"github.com/deepwrite/serivces/gateway/pkg/request"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

type MemoryHandler struct{}

func (h *MemoryHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	agentID := strings.TrimSpace(c.Query("agent_id"))
	workspaceID := strings.TrimSpace(c.Query("workspace_id"))
	sessionID := strings.TrimSpace(c.Query("session_id"))

	limit, offset := request.ParsePagination(c)

	var memories []agentmodel.Memory
	var err error

	switch {
	case sessionID != "":
		memories, err = agentmodel.GetMemoriesBySession(c.Request.Context(), sessionID, limit, offset)
	case agentID != "" && workspaceID != "":
		memories, err = agentmodel.GetMemoriesByAgentWorkspace(c.Request.Context(), agentID, workspaceID, limit, offset)
	case agentID != "":
		memories, err = agentmodel.GetMemoriesByAgent(c.Request.Context(), agentID, userID, limit, offset)
	case workspaceID != "":
		memories, err = agentmodel.GetMemoriesByWorkspace(c.Request.Context(), workspaceID, limit, offset)
	default:
		response.Failed(c, response.ErrorBadRequestCode, "session_id, agent_id or workspace_id is required")
		return
	}

	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get memories")
		return
	}

	items := make([]map[string]any, 0, len(memories))
	for _, memory := range memories {
		items = append(items, toMemoryView(&memory))
	}

	response.Success(c, response.SuccessCode, gin.H{
		"items":  items,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *MemoryHandler) GetByID(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	memoryID := strings.TrimSpace(c.Param("id"))
	if memoryID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "memory_id is required")
		return
	}

	memory, err := agentmodel.GetMemoryByID(c.Request.Context(), memoryID)
	if err != nil {
		response.Failed(c, 404, "Memory not found")
		return
	}

	if memory.UserID != userID {
		response.Failed(c, 403, "You don't have permission to access this memory")
		return
	}

	if err := agentmodel.IncrementMemoryAccess(c.Request.Context(), memoryID); err != nil {
	}

	response.Success(c, response.SuccessCode, toMemoryView(&memory))
}

func (h *MemoryHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	var req CreateMemoryRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			response.Failed(c, response.ErrorBadRequestCode, "Invalid expires_at format, use RFC3339")
			return
		}
		expiresAt = &t
	}

	memory, err := agentmodel.CreateMemory(c.Request.Context(), agentmodel.CreateMemoryInput{
		AgentID:         req.AgentID,
		UserID:          userID,
		SessionID:       req.SessionID,
		WorkspaceID:     req.WorkspaceID,
		Type:            req.Type,
		Content:         req.Content,
		Summary:         req.Summary,
		Keywords:        req.Keywords,
		EmbeddingVector: req.EmbeddingVector,
		SourceSessionID: req.SourceSessionID,
		Importance:      req.Importance,
		ExpiresAt:       expiresAt,
	})
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to create memory")
		return
	}

	response.Success(c, response.SuccessCreatedCode, toMemoryView(&memory))
}

func (h *MemoryHandler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	memoryID := strings.TrimSpace(c.Param("id"))
	if memoryID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "memory_id is required")
		return
	}

	isOwner, err := agentmodel.IsMemoryOwner(c.Request.Context(), memoryID, userID)
	if err != nil {
		response.Failed(c, 404, "Memory not found")
		return
	}
	if !isOwner {
		response.Failed(c, 403, "You don't have permission to update this memory")
		return
	}

	var req UpdateMemoryRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			response.Failed(c, response.ErrorBadRequestCode, "Invalid expires_at format, use RFC3339")
			return
		}
		expiresAt = &t
	}

	memory, err := agentmodel.UpdateMemory(c.Request.Context(), memoryID, agentmodel.UpdateMemoryInput{
		Content:    req.Content,
		Summary:    req.Summary,
		Keywords:   req.Keywords,
		Importance: req.Importance,
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to update memory")
		return
	}

	response.Success(c, response.SuccessCode, toMemoryView(&memory))
}

func (h *MemoryHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	memoryID := strings.TrimSpace(c.Param("id"))
	if memoryID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "memory_id is required")
		return
	}

	isOwner, err := agentmodel.IsMemoryOwner(c.Request.Context(), memoryID, userID)
	if err != nil {
		response.Failed(c, 404, "Memory not found")
		return
	}
	if !isOwner {
		response.Failed(c, 403, "You don't have permission to delete this memory")
		return
	}

	if err := agentmodel.DeleteMemory(c.Request.Context(), memoryID); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to delete memory")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{"id": memoryID, "deleted": true})
}

func (h *MemoryHandler) GetRecent(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	agentID := strings.TrimSpace(c.Query("agent_id"))
	workspaceID := strings.TrimSpace(c.Query("workspace_id"))
	sessionID := strings.TrimSpace(c.Query("session_id"))

	if agentID == "" && sessionID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "agent_id or session_id is required")
		return
	}

	limit := 10
	var memories []agentmodel.Memory
	var err error

	if sessionID != "" {
		memories, err = agentmodel.GetMemoriesBySession(c.Request.Context(), sessionID, limit, 0)
	} else if workspaceID != "" {
		memories, err = agentmodel.GetRecentMemories(c.Request.Context(), agentID, workspaceID, limit)
	} else {
		memories, err = agentmodel.GetMemoriesByAgent(c.Request.Context(), agentID, userID, limit, 0)
	}

	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get recent memories")
		return
	}

	items := make([]map[string]any, 0, len(memories))
	for _, memory := range memories {
		items = append(items, toMemoryView(&memory))
	}

	response.Success(c, response.SuccessCode, gin.H{
		"items": items,
	})
}

func toMemoryView(memory *agentmodel.Memory) map[string]any {
	return map[string]any{
		"id":                memory.ID,
		"agent_id":          memory.AgentID,
		"user_id":           memory.UserID,
		"session_id":        memory.SessionID,
		"workspace_id":      memory.WorkspaceID,
		"type":              memory.Type,
		"content":           memory.Content,
		"summary":           memory.Summary,
		"keywords":          memory.Keywords,
		"source_session_id": memory.SourceSessionID,
		"importance":        memory.Importance,
		"access_count":      memory.AccessCount,
		"expires_at":        memory.ExpiresAt,
		"created_at":        memory.CreatedAt,
		"updated_at":        memory.UpdatedAt,
	}
}

type CreateMemoryRequest struct {
	AgentID         string  `json:"agent_id" binding:"required"`
	SessionID       *string `json:"session_id"`
	WorkspaceID     *string `json:"workspace_id"`
	Type            string  `json:"type" binding:"required"`
	Content         string  `json:"content" binding:"required"`
	Summary         string  `json:"summary"`
	Keywords        string  `json:"keywords"`
	EmbeddingVector string  `json:"embedding_vector"`
	SourceSessionID *string `json:"source_session_id"`
	Importance      float64 `json:"importance"`
	ExpiresAt       *string `json:"expires_at"`
}

type UpdateMemoryRequest struct {
	Content    string  `json:"content"`
	Summary    string  `json:"summary"`
	Keywords   string  `json:"keywords"`
	Importance float64 `json:"importance"`
	ExpiresAt  *string `json:"expires_at"`
}
