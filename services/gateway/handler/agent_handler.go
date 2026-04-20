package handler

import (
	"context"
	"strings"

	agentmodel "github.com/deepwrite/serivces/gateway/models/agent"
	"github.com/deepwrite/serivces/gateway/pkg/database"
	"github.com/deepwrite/serivces/gateway/pkg/request"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

type AgentHandler struct{}

func (h *AgentHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	agents, err := agentmodel.GetAvailableAgents(c.Request.Context(), userID)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get agents")
		return
	}

	items := make([]map[string]any, 0, len(agents))
	for _, agent := range agents {
		items = append(items, toAgentView(&agent))
	}

	response.Success(c, response.SuccessCode, gin.H{
		"items": items,
	})
}

func (h *AgentHandler) GetByID(c *gin.Context) {
	agentID := strings.TrimSpace(c.Param("id"))
	if agentID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "agent_id is required")
		return
	}

	agent, err := agentmodel.GetAgentByID(c.Request.Context(), agentID)
	if err != nil {
		response.Failed(c, 404, "Agent not found")
		return
	}

	response.Success(c, response.SuccessCode, toAgentView(&agent))
}

func (h *AgentHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	var req CreateAgentRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	ownerID := userID
	agent, err := agentmodel.CreateAgent(c.Request.Context(), agentmodel.CreateAgentInput{
		Name:                 req.Name,
		Description:          req.Description,
		Type:                 agentmodel.AgentTypeUser,
		Category:             req.Category,
		IconURL:              req.IconURL,
		SystemPrompt:         req.SystemPrompt,
		IdentityPrompt:       req.IdentityPrompt,
		CapabilityPrompt:     req.CapabilityPrompt,
		InstructionPrompt:    req.InstructionPrompt,
		SafetyPrompt:         req.SafetyPrompt,
		DefaultModel:         req.DefaultModel,
		Skills:               req.Skills,
		Tools:                req.Tools,
		Temperature:          req.Temperature,
		MaxTokens:            req.MaxTokens,
		OwnerID:              &ownerID,
		Public:               req.Public,
		InjectUserContext:    req.InjectUserContext,
		InjectMemory:         req.InjectMemory,
		InjectTime:           req.InjectTime,
		InjectWorkspace:      req.InjectWorkspace,
		MemoryRetrievalCount: req.MemoryRetrievalCount,
		MemoryMinRelevance:   req.MemoryMinRelevance,
	})
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to create agent")
		return
	}

	response.Success(c, response.SuccessCreatedCode, toAgentView(&agent))
}

func (h *AgentHandler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	agentID := strings.TrimSpace(c.Param("id"))
	if agentID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "agent_id is required")
		return
	}

	isOwner, err := agentmodel.IsAgentOwner(c.Request.Context(), agentID, userID)
	if err != nil {
		response.Failed(c, 404, "Agent not found")
		return
	}
	if !isOwner {
		response.Failed(c, 403, "You don't have permission to update this agent")
		return
	}

	var req UpdateAgentRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	agent, err := agentmodel.UpdateAgent(c.Request.Context(), agentID, agentmodel.UpdateAgentInput{
		Name:                 req.Name,
		Description:          req.Description,
		Category:             req.Category,
		IconURL:              req.IconURL,
		SystemPrompt:         req.SystemPrompt,
		IdentityPrompt:       req.IdentityPrompt,
		CapabilityPrompt:     req.CapabilityPrompt,
		InstructionPrompt:    req.InstructionPrompt,
		SafetyPrompt:         req.SafetyPrompt,
		DefaultModel:         req.DefaultModel,
		Skills:               req.Skills,
		Tools:                req.Tools,
		Temperature:          req.Temperature,
		MaxTokens:            req.MaxTokens,
		Public:               req.Public,
		InjectUserContext:    req.InjectUserContext,
		InjectMemory:         req.InjectMemory,
		InjectTime:           req.InjectTime,
		InjectWorkspace:      req.InjectWorkspace,
		MemoryRetrievalCount: req.MemoryRetrievalCount,
		MemoryMinRelevance:   req.MemoryMinRelevance,
	})
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to update agent")
		return
	}

	response.Success(c, response.SuccessCode, toAgentView(&agent))
}

func (h *AgentHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	agentID := strings.TrimSpace(c.Param("id"))
	if agentID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "agent_id is required")
		return
	}

	isOwner, err := agentmodel.IsAgentOwner(c.Request.Context(), agentID, userID)
	if err != nil {
		response.Failed(c, 404, "Agent not found")
		return
	}
	if !isOwner {
		response.Failed(c, 403, "You don't have permission to delete this agent")
		return
	}

	if err := agentmodel.DeleteAgent(c.Request.Context(), agentID); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to delete agent")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{"id": agentID, "deleted": true})
}

func (h *AgentHandler) SetPublic(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	agentID := strings.TrimSpace(c.Param("id"))
	if agentID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "agent_id is required")
		return
	}

	isOwner, err := agentmodel.IsAgentOwner(c.Request.Context(), agentID, userID)
	if err != nil {
		response.Failed(c, 404, "Agent not found")
		return
	}
	if !isOwner {
		response.Failed(c, 403, "You don't have permission to update this agent")
		return
	}

	var req SetPublicRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	agent, err := agentmodel.GetAgentByID(c.Request.Context(), agentID)
	if err != nil {
		response.Failed(c, 404, "Agent not found")
		return
	}

	updates := map[string]any{"public": req.Public}
	if err := updateAgentFields(c.Request.Context(), agentID, updates); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to update agent")
		return
	}

	agent, _ = agentmodel.GetAgentByID(c.Request.Context(), agentID)
	response.Success(c, response.SuccessCode, toAgentView(&agent))
}

func (h *AgentHandler) Activate(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	agentID := strings.TrimSpace(c.Param("id"))
	if agentID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "agent_id is required")
		return
	}

	agent, err := agentmodel.GetAgentByID(c.Request.Context(), agentID)
	if err != nil {
		response.Failed(c, 404, "Agent not found")
		return
	}

	if !agent.Enabled {
		response.Failed(c, 400, "Agent is disabled")
		return
	}

	if err := agentmodel.IncrementAgentUsage(c.Request.Context(), agentID); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to activate agent")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{"activated": true})
}

func (h *AgentHandler) ListMyAgents(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	agents, err := agentmodel.GetUserAgents(c.Request.Context(), userID)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get agents")
		return
	}

	items := make([]map[string]any, 0, len(agents))
	for _, agent := range agents {
		items = append(items, toAgentView(&agent))
	}

	response.Success(c, response.SuccessCode, gin.H{
		"items": items,
	})
}

func toAgentView(agent *agentmodel.Agent) map[string]any {
	return map[string]any{
		"id":                     agent.ID,
		"name":                   agent.Name,
		"description":            agent.Description,
		"type":                   agent.Type,
		"category":               agent.Category,
		"icon_url":               agent.IconURL,
		"system_prompt":          agent.SystemPrompt,
		"identity_prompt":        agent.IdentityPrompt,
		"capability_prompt":      agent.CapabilityPrompt,
		"instruction_prompt":     agent.InstructionPrompt,
		"safety_prompt":          agent.SafetyPrompt,
		"default_model":          agent.DefaultModel,
		"skills":                 agent.Skills,
		"tools":                  agent.Tools,
		"temperature":            agent.Temperature,
		"max_tokens":             agent.MaxTokens,
		"owner_id":               agent.OwnerID,
		"public":                 agent.Public,
		"rating":                 agent.Rating,
		"usage_count":            agent.UsageCount,
		"enabled":                agent.Enabled,
		"inject_user_context":    agent.InjectUserContext,
		"inject_memory":          agent.InjectMemory,
		"inject_time":            agent.InjectTime,
		"inject_workspace":       agent.InjectWorkspace,
		"memory_retrieval_count": agent.MemoryRetrievalCount,
		"memory_min_relevance":   agent.MemoryMinRelevance,
		"created_at":             agent.CreatedAt,
		"updated_at":             agent.UpdatedAt,
	}
}

type CreateAgentRequest struct {
	Name                 string  `json:"name" binding:"required"`
	Description          string  `json:"description"`
	Category             string  `json:"category"`
	IconURL              string  `json:"icon_url"`
	SystemPrompt         string  `json:"system_prompt"`
	IdentityPrompt       string  `json:"identity_prompt"`
	CapabilityPrompt     string  `json:"capability_prompt"`
	InstructionPrompt    string  `json:"instruction_prompt"`
	SafetyPrompt         string  `json:"safety_prompt"`
	DefaultModel         string  `json:"default_model"`
	Skills               string  `json:"skills"`
	Tools                string  `json:"tools"`
	Temperature          float64 `json:"temperature"`
	MaxTokens            int     `json:"max_tokens"`
	Public               bool    `json:"public"`
	InjectUserContext    bool    `json:"inject_user_context"`
	InjectMemory         bool    `json:"inject_memory"`
	InjectTime           bool    `json:"inject_time"`
	InjectWorkspace      bool    `json:"inject_workspace"`
	MemoryRetrievalCount int     `json:"memory_retrieval_count"`
	MemoryMinRelevance   float64 `json:"memory_min_relevance"`
}

type UpdateAgentRequest struct {
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	Category             string   `json:"category"`
	IconURL              string   `json:"icon_url"`
	SystemPrompt         string   `json:"system_prompt"`
	IdentityPrompt       string   `json:"identity_prompt"`
	CapabilityPrompt     string   `json:"capability_prompt"`
	InstructionPrompt    string   `json:"instruction_prompt"`
	SafetyPrompt         string   `json:"safety_prompt"`
	DefaultModel         string   `json:"default_model"`
	Skills               string   `json:"skills"`
	Tools                string   `json:"tools"`
	Temperature          float64  `json:"temperature"`
	MaxTokens            int      `json:"max_tokens"`
	Public               bool     `json:"public"`
	InjectUserContext    *bool    `json:"inject_user_context"`
	InjectMemory         *bool    `json:"inject_memory"`
	InjectTime           *bool    `json:"inject_time"`
	InjectWorkspace      *bool    `json:"inject_workspace"`
	MemoryRetrievalCount *int     `json:"memory_retrieval_count"`
	MemoryMinRelevance   *float64 `json:"memory_min_relevance"`
}

type SetPublicRequest struct {
	Public bool `json:"public"`
}

func updateAgentFields(ctx context.Context, agentID string, updates map[string]any) error {
	return database.DB.WithContext(ctx).
		Model(&agentmodel.Agent{}).
		Where("id = ?", agentID).
		Updates(updates).Error
}
