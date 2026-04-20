package agent

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/pkg/database"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CreateMemoryInput struct {
	AgentID         string
	UserID          string
	SessionID       *string
	WorkspaceID     *string
	Type            string
	Content         string
	Summary         string
	Keywords        string
	EmbeddingVector string
	SourceSessionID *string
	Importance      float64
	ExpiresAt       *time.Time
}

type UpdateMemoryInput struct {
	Content    string
	Summary    string
	Keywords   string
	Importance float64
	ExpiresAt  *time.Time
}

func CreateMemory(ctx context.Context, input CreateMemoryInput) (Memory, error) {
	memory := Memory{
		AgentID:         strings.TrimSpace(input.AgentID),
		UserID:          strings.TrimSpace(input.UserID),
		SessionID:       input.SessionID,
		WorkspaceID:     input.WorkspaceID,
		Type:            normalizeMemoryType(input.Type),
		Content:         strings.TrimSpace(input.Content),
		Summary:         strings.TrimSpace(input.Summary),
		SourceSessionID: input.SourceSessionID,
		Importance:      input.Importance,
		ExpiresAt:       input.ExpiresAt,
	}

	if len(input.Keywords) > 0 {
		memory.Keywords = datatypes.JSON([]byte(input.Keywords))
	} else {
		memory.Keywords = datatypes.JSON([]byte("[]"))
	}

	if len(input.EmbeddingVector) > 0 {
		memory.EmbeddingVector = datatypes.JSON([]byte(input.EmbeddingVector))
	}

	if memory.Importance <= 0 {
		memory.Importance = 0.5
	}

	if err := database.DB.WithContext(ctx).Create(&memory).Error; err != nil {
		return Memory{}, err
	}

	return memory, nil
}

func GetMemoryByID(ctx context.Context, memoryID string) (Memory, error) {
	var memory Memory
	err := database.DB.WithContext(ctx).
		Where("id = ?", strings.TrimSpace(memoryID)).
		First(&memory).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Memory{}, errors.New("memory not found")
	}
	return memory, err
}

func GetMemoriesByAgentWorkspace(ctx context.Context, agentID, workspaceID string, limit, offset int) ([]Memory, error) {
	query := database.DB.WithContext(ctx).
		Where("agent_id = ? AND workspace_id = ?", agentID, workspaceID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("importance desc, updated_at desc")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var memories []Memory
	err := query.Find(&memories).Error
	return memories, err
}

func GetMemoriesBySession(ctx context.Context, sessionID string, limit, offset int) ([]Memory, error) {
	query := database.DB.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("importance desc, updated_at desc")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var memories []Memory
	err := query.Find(&memories).Error
	return memories, err
}

func GetMemoriesByAgent(ctx context.Context, agentID, userID string, limit, offset int) ([]Memory, error) {
	query := database.DB.WithContext(ctx).
		Where("agent_id = ? AND user_id = ?", agentID, userID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("importance desc, updated_at desc")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var memories []Memory
	err := query.Find(&memories).Error
	return memories, err
}

func GetMemoriesByWorkspace(ctx context.Context, workspaceID string, limit, offset int) ([]Memory, error) {
	query := database.DB.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("importance desc, updated_at desc")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var memories []Memory
	err := query.Find(&memories).Error
	return memories, err
}

func GetRecentMemories(ctx context.Context, agentID, workspaceID string, limit int) ([]Memory, error) {
	if limit <= 0 {
		limit = 10
	}

	var memories []Memory
	err := database.DB.WithContext(ctx).
		Where("agent_id = ? AND workspace_id = ?", agentID, workspaceID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("updated_at desc").
		Limit(limit).
		Find(&memories).Error
	return memories, err
}

func UpdateMemory(ctx context.Context, memoryID string, input UpdateMemoryInput) (Memory, error) {
	memory, err := GetMemoryByID(ctx, memoryID)
	if err != nil {
		return Memory{}, err
	}

	updates := map[string]any{}

	if content := strings.TrimSpace(input.Content); content != "" {
		updates["content"] = content
	}
	if summary := strings.TrimSpace(input.Summary); summary != "" {
		updates["summary"] = summary
	}
	if len(input.Keywords) > 0 {
		updates["keywords"] = datatypes.JSON([]byte(input.Keywords))
	}
	if input.Importance > 0 {
		updates["importance"] = input.Importance
	}
	if input.ExpiresAt != nil {
		updates["expires_at"] = input.ExpiresAt
	}

	if len(updates) > 0 {
		if err := database.DB.WithContext(ctx).
			Model(&Memory{}).
			Where("id = ?", memory.ID).
			Updates(updates).Error; err != nil {
			return Memory{}, err
		}
	}

	return GetMemoryByID(ctx, memoryID)
}

func IncrementMemoryAccess(ctx context.Context, memoryID string) error {
	return database.DB.WithContext(ctx).
		Model(&Memory{}).
		Where("id = ?", strings.TrimSpace(memoryID)).
		UpdateColumn("access_count", gorm.Expr("access_count + 1")).Error
}

func DeleteMemory(ctx context.Context, memoryID string) error {
	return database.DB.WithContext(ctx).
		Delete(&Memory{}, "id = ?", strings.TrimSpace(memoryID)).Error
}

func DeleteMemoriesByAgent(ctx context.Context, agentID string) error {
	return database.DB.WithContext(ctx).
		Delete(&Memory{}, "agent_id = ?", strings.TrimSpace(agentID)).Error
}

func IsMemoryOwner(ctx context.Context, memoryID, userID string) (bool, error) {
	memory, err := GetMemoryByID(ctx, memoryID)
	if err != nil {
		return false, err
	}
	return memory.UserID == userID, nil
}

func normalizeMemoryType(typ string) string {
	switch strings.TrimSpace(typ) {
	case MemoryTypePreference, MemoryTypeFact, MemoryTypeTask, MemoryTypeInsight:
		return typ
	default:
		return MemoryTypeFact
	}
}
