package agent

import (
	"context"
	"errors"
	"strings"

	"github.com/deepwrite/serivces/gateway/pkg/database"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CreateAgentInput struct {
	Name                 string
	Description          string
	Type                 string
	Category             string
	IconURL              string
	SystemPrompt         string
	IdentityPrompt       string
	CapabilityPrompt     string
	InstructionPrompt    string
	SafetyPrompt         string
	DefaultModel         string
	Skills               string
	Tools                string
	Temperature          float64
	MaxTokens            int
	OwnerID              *string
	Public               bool
	InjectUserContext    bool
	InjectMemory         bool
	InjectTime           bool
	InjectWorkspace      bool
	MemoryRetrievalCount int
	MemoryMinRelevance   float64
}

type UpdateAgentInput struct {
	Name                 string
	Description          string
	Category             string
	IconURL              string
	SystemPrompt         string
	IdentityPrompt       string
	CapabilityPrompt     string
	InstructionPrompt    string
	SafetyPrompt         string
	DefaultModel         string
	Skills               string
	Tools                string
	Temperature          float64
	MaxTokens            int
	Public               bool
	InjectUserContext    *bool
	InjectMemory         *bool
	InjectTime           *bool
	InjectWorkspace      *bool
	MemoryRetrievalCount *int
	MemoryMinRelevance   *float64
}

func CreateAgent(ctx context.Context, input CreateAgentInput) (Agent, error) {
	agent := Agent{
		Name:                 strings.TrimSpace(input.Name),
		Description:          strings.TrimSpace(input.Description),
		Type:                 normalizeAgentType(input.Type),
		Category:             normalizeAgentCategory(input.Category),
		IconURL:              strings.TrimSpace(input.IconURL),
		SystemPrompt:         strings.TrimSpace(input.SystemPrompt),
		IdentityPrompt:       strings.TrimSpace(input.IdentityPrompt),
		CapabilityPrompt:     strings.TrimSpace(input.CapabilityPrompt),
		InstructionPrompt:    strings.TrimSpace(input.InstructionPrompt),
		SafetyPrompt:         strings.TrimSpace(input.SafetyPrompt),
		DefaultModel:         strings.TrimSpace(input.DefaultModel),
		Temperature:          input.Temperature,
		MaxTokens:            input.MaxTokens,
		OwnerID:              input.OwnerID,
		Public:               input.Public,
		Enabled:              true,
		InjectUserContext:    input.InjectUserContext,
		InjectMemory:         input.InjectMemory,
		InjectTime:           input.InjectTime,
		InjectWorkspace:      input.InjectWorkspace,
		MemoryRetrievalCount: input.MemoryRetrievalCount,
		MemoryMinRelevance:   input.MemoryMinRelevance,
	}

	if agent.MemoryRetrievalCount == 0 {
		agent.MemoryRetrievalCount = 5
	}
	if agent.MemoryMinRelevance == 0 {
		agent.MemoryMinRelevance = 0.7
	}

	if len(input.Skills) > 0 {
		agent.Skills = datatypes.JSON([]byte(input.Skills))
	} else {
		agent.Skills = datatypes.JSON([]byte("[]"))
	}

	if len(input.Tools) > 0 {
		agent.Tools = datatypes.JSON([]byte(input.Tools))
	} else {
		agent.Tools = datatypes.JSON([]byte("[]"))
	}

	if err := database.DB.WithContext(ctx).Create(&agent).Error; err != nil {
		return Agent{}, err
	}

	return agent, nil
}

func GetAgentByID(ctx context.Context, agentID string) (Agent, error) {
	var agent Agent
	err := database.DB.WithContext(ctx).
		Where("id = ?", strings.TrimSpace(agentID)).
		First(&agent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Agent{}, errors.New("agent not found")
	}
	return agent, err
}

func GetOfficialAgents(ctx context.Context) ([]Agent, error) {
	var agents []Agent
	err := database.DB.WithContext(ctx).
		Where("type = ?", AgentTypeOfficial).
		Order("usage_count desc, created_at desc").
		Find(&agents).Error
	return agents, err
}

func GetPublicAgents(ctx context.Context) ([]Agent, error) {
	var agents []Agent
	err := database.DB.WithContext(ctx).
		Where("type = ? AND public = ?", AgentTypeUser, true).
		Order("rating desc, usage_count desc").
		Find(&agents).Error
	return agents, err
}

func GetAvailableAgents(ctx context.Context, userID string) ([]Agent, error) {
	var agents []Agent
	err := database.DB.WithContext(ctx).
		Where("(type = ? AND enabled = ?) OR (type = ? AND owner_id = ? AND enabled = ?) OR (type = ? AND public = ? AND enabled = ?)",
			AgentTypeOfficial, true,
			AgentTypeUser, userID, true,
			AgentTypeUser, true, true).
		Order("type desc, usage_count desc, created_at desc").
		Find(&agents).Error
	return agents, err
}

func GetUserAgents(ctx context.Context, userID string) ([]Agent, error) {
	var agents []Agent
	err := database.DB.WithContext(ctx).
		Where("owner_id = ?", userID).
		Order("created_at desc").
		Find(&agents).Error
	return agents, err
}

func UpdateAgent(ctx context.Context, agentID string, input UpdateAgentInput) (Agent, error) {
	agent, err := GetAgentByID(ctx, agentID)
	if err != nil {
		return Agent{}, err
	}

	updates := map[string]any{}

	if name := strings.TrimSpace(input.Name); name != "" {
		updates["name"] = name
	}
	if desc := strings.TrimSpace(input.Description); desc != "" {
		updates["description"] = desc
	}
	if cat := normalizeAgentCategory(input.Category); cat != "" {
		updates["category"] = cat
	}
	if icon := strings.TrimSpace(input.IconURL); icon != "" {
		updates["icon_url"] = icon
	}
	if prompt := strings.TrimSpace(input.SystemPrompt); prompt != "" {
		updates["system_prompt"] = prompt
	}
	if prompt := strings.TrimSpace(input.IdentityPrompt); prompt != "" {
		updates["identity_prompt"] = prompt
	}
	if prompt := strings.TrimSpace(input.CapabilityPrompt); prompt != "" {
		updates["capability_prompt"] = prompt
	}
	if prompt := strings.TrimSpace(input.InstructionPrompt); prompt != "" {
		updates["instruction_prompt"] = prompt
	}
	if prompt := strings.TrimSpace(input.SafetyPrompt); prompt != "" {
		updates["safety_prompt"] = prompt
	}
	if model := strings.TrimSpace(input.DefaultModel); model != "" {
		updates["default_model"] = model
	}
	if input.Temperature > 0 {
		updates["temperature"] = input.Temperature
	}
	if input.MaxTokens > 0 {
		updates["max_tokens"] = input.MaxTokens
	}
	updates["public"] = input.Public

	if input.InjectUserContext != nil {
		updates["inject_user_context"] = *input.InjectUserContext
	}
	if input.InjectMemory != nil {
		updates["inject_memory"] = *input.InjectMemory
	}
	if input.InjectTime != nil {
		updates["inject_time"] = *input.InjectTime
	}
	if input.InjectWorkspace != nil {
		updates["inject_workspace"] = *input.InjectWorkspace
	}
	if input.MemoryRetrievalCount != nil {
		updates["memory_retrieval_count"] = *input.MemoryRetrievalCount
	}
	if input.MemoryMinRelevance != nil {
		updates["memory_min_relevance"] = *input.MemoryMinRelevance
	}

	if len(input.Skills) > 0 {
		updates["skills"] = datatypes.JSON([]byte(input.Skills))
	}
	if len(input.Tools) > 0 {
		updates["tools"] = datatypes.JSON([]byte(input.Tools))
	}

	if len(updates) > 0 {
		if err := database.DB.WithContext(ctx).
			Model(&Agent{}).
			Where("id = ?", agent.ID).
			Updates(updates).Error; err != nil {
			return Agent{}, err
		}
	}

	return GetAgentByID(ctx, agentID)
}

func UpdateAgentStatus(ctx context.Context, agentID string, enabled bool) (Agent, error) {
	agent, err := GetAgentByID(ctx, agentID)
	if err != nil {
		return Agent{}, err
	}

	if err := database.DB.WithContext(ctx).
		Model(&Agent{}).
		Where("id = ?", agent.ID).
		Update("enabled", enabled).Error; err != nil {
		return Agent{}, err
	}

	return GetAgentByID(ctx, agentID)
}

func IncrementAgentUsage(ctx context.Context, agentID string) error {
	return database.DB.WithContext(ctx).
		Model(&Agent{}).
		Where("id = ?", strings.TrimSpace(agentID)).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}

func DeleteAgent(ctx context.Context, agentID string) error {
	return database.DB.WithContext(ctx).
		Delete(&Agent{}, "id = ?", strings.TrimSpace(agentID)).Error
}

func IsAgentOwner(ctx context.Context, agentID, userID string) (bool, error) {
	agent, err := GetAgentByID(ctx, agentID)
	if err != nil {
		return false, err
	}

	if agent.Type == AgentTypeOfficial {
		return false, nil
	}

	if agent.OwnerID == nil {
		return false, nil
	}

	return *agent.OwnerID == userID, nil
}

func normalizeAgentType(typ string) string {
	switch strings.TrimSpace(typ) {
	case AgentTypeOfficial:
		return AgentTypeOfficial
	default:
		return AgentTypeUser
	}
}

func normalizeAgentCategory(cat string) string {
	switch strings.TrimSpace(cat) {
	case AgentCategoryResearch, AgentCategoryWriting, AgentCategoryData, AgentCategoryPublishing:
		return cat
	default:
		return AgentCategoryGeneral
	}
}
