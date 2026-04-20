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

type CreateSessionInput struct {
	AgentID     string
	UserID      string
	WorkspaceID *string
	Title       string
	ContextDocs string
	ContextRefs string
}

type UpdateSessionInput struct {
	Title       string
	ContextDocs string
	ContextRefs string
	Summary     string
	Pinned      *bool
	Archived    *bool
	GroupID     *string
	Tags        string
}

type SessionListFilter struct {
	AgentID  string
	Search   string
	Pinned   *bool
	Archived *bool
	GroupID  string
}

func CreateSession(ctx context.Context, input CreateSessionInput) (Session, error) {
	session := Session{
		AgentID: strings.TrimSpace(input.AgentID),
		UserID:  strings.TrimSpace(input.UserID),
		Title:   strings.TrimSpace(input.Title),
		Status:  SessionStatusActive,
	}

	if input.WorkspaceID != nil && *input.WorkspaceID != "" {
		trimmed := strings.TrimSpace(*input.WorkspaceID)
		session.WorkspaceID = &trimmed
	}

	if len(input.ContextDocs) > 0 {
		session.ContextDocs = datatypes.JSON([]byte(input.ContextDocs))
	} else {
		session.ContextDocs = datatypes.JSON([]byte("[]"))
	}

	if len(input.ContextRefs) > 0 {
		session.ContextRefs = datatypes.JSON([]byte(input.ContextRefs))
	} else {
		session.ContextRefs = datatypes.JSON([]byte("[]"))
	}

	if err := database.DB.WithContext(ctx).Create(&session).Error; err != nil {
		return Session{}, err
	}

	return session, nil
}

func GetSessionByID(ctx context.Context, sessionID string) (Session, error) {
	var session Session
	err := database.DB.WithContext(ctx).
		Where("id = ?", strings.TrimSpace(sessionID)).
		First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Session{}, errors.New("session not found")
	}
	return session, err
}

func GetSessionsByUser(ctx context.Context, userID string, filter SessionListFilter, limit, offset int) ([]Session, error) {
	query := database.DB.WithContext(ctx).
		Where("user_id = ?", userID)

	if filter.Archived != nil {
		if *filter.Archived {
			query = query.Where("archived = ?", true)
		} else {
			query = query.Where("archived = ?", false)
		}
	} else {
		query = query.Where("archived = ?", false)
	}

	if filter.Pinned != nil {
		query = query.Where("pinned = ?", *filter.Pinned)
	}

	if filter.AgentID != "" {
		query = query.Where("agent_id = ?", filter.AgentID)
	}

	if filter.GroupID != "" {
		query = query.Where("group_id = ?", filter.GroupID)
	}

	if filter.Search != "" {
		search := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(summary) LIKE ?", search, search)
	}

	query = query.Order("pinned DESC, last_message_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var sessions []Session
	err := query.Find(&sessions).Error
	return sessions, err
}

func GetSessionsByWorkspace(ctx context.Context, workspaceID string, limit, offset int) ([]Session, error) {
	query := database.DB.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Where("status = ?", SessionStatusActive).
		Order("last_message_at desc")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var sessions []Session
	err := query.Find(&sessions).Error
	return sessions, err
}

func GetSessionsByAgent(ctx context.Context, agentID, userID string, limit, offset int) ([]Session, error) {
	query := database.DB.WithContext(ctx).
		Where("agent_id = ? AND user_id = ?", agentID, userID).
		Where("status = ?", SessionStatusActive).
		Order("last_message_at desc")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var sessions []Session
	err := query.Find(&sessions).Error
	return sessions, err
}

func UpdateSession(ctx context.Context, sessionID string, input UpdateSessionInput) (Session, error) {
	session, err := GetSessionByID(ctx, sessionID)
	if err != nil {
		return Session{}, err
	}

	updates := map[string]any{}

	if title := strings.TrimSpace(input.Title); title != "" {
		updates["title"] = title
	}
	if len(input.ContextDocs) > 0 {
		updates["context_docs"] = datatypes.JSON([]byte(input.ContextDocs))
	}
	if len(input.ContextRefs) > 0 {
		updates["context_refs"] = datatypes.JSON([]byte(input.ContextRefs))
	}
	if summary := strings.TrimSpace(input.Summary); summary != "" {
		updates["summary"] = summary
	}
	if input.Pinned != nil {
		updates["pinned"] = *input.Pinned
	}
	if input.Archived != nil {
		updates["archived"] = *input.Archived
	}
	if input.GroupID != nil {
		updates["group_id"] = input.GroupID
	}
	if len(input.Tags) > 0 {
		updates["tags"] = datatypes.JSON([]byte(input.Tags))
	}

	if len(updates) > 0 {
		if err := database.DB.WithContext(ctx).
			Model(&Session{}).
			Where("id = ?", session.ID).
			Updates(updates).Error; err != nil {
			return Session{}, err
		}
	}

	return GetSessionByID(ctx, sessionID)
}

func ArchiveSession(ctx context.Context, sessionID string) (Session, error) {
	session, err := GetSessionByID(ctx, sessionID)
	if err != nil {
		return Session{}, err
	}

	if err := database.DB.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", session.ID).
		Updates(map[string]any{
			"archived":   true,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return Session{}, err
	}

	return GetSessionByID(ctx, sessionID)
}

func UnarchiveSession(ctx context.Context, sessionID string) (Session, error) {
	session, err := GetSessionByID(ctx, sessionID)
	if err != nil {
		return Session{}, err
	}

	if err := database.DB.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", session.ID).
		Updates(map[string]any{
			"archived":   false,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return Session{}, err
	}

	return GetSessionByID(ctx, sessionID)
}

func PinSession(ctx context.Context, sessionID string) (Session, error) {
	session, err := GetSessionByID(ctx, sessionID)
	if err != nil {
		return Session{}, err
	}

	if err := database.DB.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", session.ID).
		Updates(map[string]any{
			"pinned":     true,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return Session{}, err
	}

	return GetSessionByID(ctx, sessionID)
}

func UnpinSession(ctx context.Context, sessionID string) (Session, error) {
	session, err := GetSessionByID(ctx, sessionID)
	if err != nil {
		return Session{}, err
	}

	if err := database.DB.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", session.ID).
		Updates(map[string]any{
			"pinned":     false,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return Session{}, err
	}

	return GetSessionByID(ctx, sessionID)
}

func UpdateSessionLastMessage(ctx context.Context, sessionID string) error {
	now := time.Now()
	return database.DB.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", strings.TrimSpace(sessionID)).
		Updates(map[string]any{
			"last_message_at": now,
			"updated_at":      now,
		}).Error
}

func UpdateSessionTitle(ctx context.Context, sessionID, title string) error {
	return database.DB.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", strings.TrimSpace(sessionID)).
		Updates(map[string]any{
			"title":      title,
			"updated_at": time.Now(),
		}).Error
}

func IncrementSessionTokenCount(ctx context.Context, sessionID string, tokens int) error {
	return database.DB.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", strings.TrimSpace(sessionID)).
		UpdateColumn("token_count", gorm.Expr("token_count + ?", tokens)).Error
}

func DeleteSession(ctx context.Context, sessionID string) error {
	return database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&Message{}, "session_id = ?", strings.TrimSpace(sessionID)).Error; err != nil {
			return err
		}
		return tx.Delete(&Session{}, "id = ?", strings.TrimSpace(sessionID)).Error
	})
}

func IsSessionOwner(ctx context.Context, sessionID, userID string) (bool, error) {
	session, err := GetSessionByID(ctx, sessionID)
	if err != nil {
		return false, err
	}
	return session.UserID == userID, nil
}
