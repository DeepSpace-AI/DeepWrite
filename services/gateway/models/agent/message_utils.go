package agent

import (
	"context"
	"strings"

	"github.com/deepwrite/serivces/gateway/pkg/database"
	"gorm.io/datatypes"
)

type CreateMessageInput struct {
	SessionID  string
	Role       string
	Content    string
	TokenCount int
	ModelUsed  string
	Sources    string
	Actions    string
}

func CreateMessage(ctx context.Context, input CreateMessageInput) (Message, error) {
	msg := Message{
		SessionID:  strings.TrimSpace(input.SessionID),
		Role:       normalizeMessageRole(input.Role),
		Content:    strings.TrimSpace(input.Content),
		TokenCount: input.TokenCount,
		ModelUsed:  strings.TrimSpace(input.ModelUsed),
	}

	if len(input.Sources) > 0 {
		msg.Sources = datatypes.JSON([]byte(input.Sources))
	} else {
		msg.Sources = datatypes.JSON([]byte("[]"))
	}

	if len(input.Actions) > 0 {
		msg.Actions = datatypes.JSON([]byte(input.Actions))
	} else {
		msg.Actions = datatypes.JSON([]byte("[]"))
	}

	if err := database.DB.WithContext(ctx).Create(&msg).Error; err != nil {
		return Message{}, err
	}

	if err := UpdateSessionLastMessage(ctx, msg.SessionID); err != nil {
		return Message{}, err
	}

	if msg.TokenCount > 0 {
		if err := IncrementSessionTokenCount(ctx, msg.SessionID, msg.TokenCount); err != nil {
			return Message{}, err
		}
	}

	return msg, nil
}

func GetMessagesBySession(ctx context.Context, sessionID string, limit, offset int) ([]Message, error) {
	query := database.DB.WithContext(ctx).
		Where("session_id = ?", strings.TrimSpace(sessionID)).
		Order("created_at asc")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var messages []Message
	err := query.Find(&messages).Error
	return messages, err
}

func GetRecentMessages(ctx context.Context, sessionID string, limit int) ([]Message, error) {
	if limit <= 0 {
		limit = 20
	}

	var messages []Message
	err := database.DB.WithContext(ctx).
		Where("session_id = ?", strings.TrimSpace(sessionID)).
		Order("created_at desc").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

func GetMessageByID(ctx context.Context, messageID string) (Message, error) {
	var msg Message
	err := database.DB.WithContext(ctx).
		Where("id = ?", strings.TrimSpace(messageID)).
		First(&msg).Error
	return msg, err
}

func DeleteMessage(ctx context.Context, messageID string) error {
	return database.DB.WithContext(ctx).
		Delete(&Message{}, "id = ?", strings.TrimSpace(messageID)).Error
}

func DeleteMessagesBySession(ctx context.Context, sessionID string) error {
	return database.DB.WithContext(ctx).
		Delete(&Message{}, "session_id = ?", strings.TrimSpace(sessionID)).Error
}

func CountMessagesBySession(ctx context.Context, sessionID string) (int64, error) {
	var count int64
	err := database.DB.WithContext(ctx).
		Model(&Message{}).
		Where("session_id = ?", strings.TrimSpace(sessionID)).
		Count(&count).Error
	return count, err
}

func normalizeMessageRole(role string) string {
	switch strings.TrimSpace(role) {
	case MessageRoleUser, MessageRoleAssistant, MessageRoleSystem:
		return role
	default:
		return MessageRoleUser
	}
}
