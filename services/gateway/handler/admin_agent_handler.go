package handler

import (
	"strings"

	agentmodel "github.com/deepwrite/serivces/gateway/models/agent"
	"github.com/deepwrite/serivces/gateway/pkg/request"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

type AdminAgentHandler struct{}

func (h *AdminAgentHandler) ListOfficial(c *gin.Context) {
	agents, err := agentmodel.GetOfficialAgents(c.Request.Context())
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get official agents")
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

func (h *AdminAgentHandler) ListPublic(c *gin.Context) {
	agents, err := agentmodel.GetPublicAgents(c.Request.Context())
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get public agents")
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

func (h *AdminAgentHandler) GetByID(c *gin.Context) {
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

func (h *AdminAgentHandler) Create(c *gin.Context) {
	var req CreateAdminAgentRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	agent, err := agentmodel.CreateAgent(c.Request.Context(), agentmodel.CreateAgentInput{
		Name:         req.Name,
		Description:  req.Description,
		Type:         agentmodel.AgentTypeOfficial,
		Category:     req.Category,
		IconURL:      req.IconURL,
		SystemPrompt: req.SystemPrompt,
		DefaultModel: req.DefaultModel,
		Skills:       req.Skills,
		Tools:        req.Tools,
		Temperature:  req.Temperature,
		MaxTokens:    req.MaxTokens,
		OwnerID:      nil,
		Public:       true,
	})
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to create agent")
		return
	}

	response.Success(c, response.SuccessCreatedCode, toAgentView(&agent))
}

func (h *AdminAgentHandler) Update(c *gin.Context) {
	agentID := strings.TrimSpace(c.Param("id"))
	if agentID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "agent_id is required")
		return
	}

	var req UpdateAdminAgentRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	agent, err := agentmodel.UpdateAgent(c.Request.Context(), agentID, agentmodel.UpdateAgentInput{
		Name:         req.Name,
		Description:  req.Description,
		Category:     req.Category,
		IconURL:      req.IconURL,
		SystemPrompt: req.SystemPrompt,
		DefaultModel: req.DefaultModel,
		Skills:       req.Skills,
		Tools:        req.Tools,
		Temperature:  req.Temperature,
		MaxTokens:    req.MaxTokens,
		Public:       true,
	})
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to update agent")
		return
	}

	response.Success(c, response.SuccessCode, toAgentView(&agent))
}

func (h *AdminAgentHandler) Delete(c *gin.Context) {
	agentID := strings.TrimSpace(c.Param("id"))
	if agentID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "agent_id is required")
		return
	}

	if err := agentmodel.DeleteAgent(c.Request.Context(), agentID); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to delete agent")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{"id": agentID, "deleted": true})
}

func (h *AdminAgentHandler) UpdateStatus(c *gin.Context) {
	agentID := strings.TrimSpace(c.Param("id"))
	if agentID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "agent_id is required")
		return
	}

	var req UpdateStatusRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	agent, err := agentmodel.UpdateAgentStatus(c.Request.Context(), agentID, req.Enabled)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to update agent status")
		return
	}

	response.Success(c, response.SuccessCode, toAgentView(&agent))
}

func (h *AdminAgentHandler) Moderate(c *gin.Context) {
	agentID := strings.TrimSpace(c.Param("id"))
	if agentID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "agent_id is required")
		return
	}

	var req ModerateRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	agent, err := agentmodel.GetAgentByID(c.Request.Context(), agentID)
	if err != nil {
		response.Failed(c, 404, "Agent not found")
		return
	}

	if agent.Type != agentmodel.AgentTypeUser {
		response.Failed(c, 400, "Can only moderate user agents")
		return
	}

	agent, err = agentmodel.UpdateAgentStatus(c.Request.Context(), agentID, req.Enabled)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to moderate agent")
		return
	}

	response.Success(c, response.SuccessCode, toAgentView(&agent))
}

type CreateAdminAgentRequest struct {
	Name         string  `json:"name" binding:"required"`
	Description  string  `json:"description"`
	Category     string  `json:"category"`
	IconURL      string  `json:"icon_url"`
	SystemPrompt string  `json:"system_prompt" binding:"required"`
	DefaultModel string  `json:"default_model"`
	Skills       string  `json:"skills"`
	Tools        string  `json:"tools"`
	Temperature  float64 `json:"temperature"`
	MaxTokens    int     `json:"max_tokens"`
}

type UpdateAdminAgentRequest struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Category     string  `json:"category"`
	IconURL      string  `json:"icon_url"`
	SystemPrompt string  `json:"system_prompt"`
	DefaultModel string  `json:"default_model"`
	Skills       string  `json:"skills"`
	Tools        string  `json:"tools"`
	Temperature  float64 `json:"temperature"`
	MaxTokens    int     `json:"max_tokens"`
}

type UpdateStatusRequest struct {
	Enabled bool `json:"enabled"`
}

type ModerateRequest struct {
	Enabled bool `json:"enabled"`
}
