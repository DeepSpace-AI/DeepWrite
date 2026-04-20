package agent

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/pkg/database"
	"gorm.io/gorm"
)

type CreateSessionGroupInput struct {
	UserID      string
	Name        string
	Description string
	Color       string
	Icon        string
}

type UpdateSessionGroupInput struct {
	Name        string
	Description string
	Color       string
	Icon        string
	SortOrder   *int
}

func CreateSessionGroup(ctx context.Context, input CreateSessionGroupInput) (SessionGroup, error) {
	group := SessionGroup{
		UserID:      strings.TrimSpace(input.UserID),
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Color:       input.Color,
		Icon:        input.Icon,
	}

	if group.Color == "" {
		group.Color = "#6366f1"
	}

	if err := database.DB.WithContext(ctx).Create(&group).Error; err != nil {
		return SessionGroup{}, err
	}

	return group, nil
}

func GetSessionGroupByID(ctx context.Context, groupID string) (SessionGroup, error) {
	var group SessionGroup
	err := database.DB.WithContext(ctx).
		Where("id = ?", strings.TrimSpace(groupID)).
		First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return SessionGroup{}, errors.New("session group not found")
	}
	return group, err
}

func GetSessionGroupsByUser(ctx context.Context, userID string) ([]SessionGroup, error) {
	var groups []SessionGroup
	err := database.DB.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Order("sort_order ASC, created_at ASC").
		Find(&groups).Error
	return groups, err
}

func UpdateSessionGroup(ctx context.Context, groupID string, input UpdateSessionGroupInput) (SessionGroup, error) {
	group, err := GetSessionGroupByID(ctx, groupID)
	if err != nil {
		return SessionGroup{}, err
	}

	updates := map[string]any{}

	if name := strings.TrimSpace(input.Name); name != "" {
		updates["name"] = name
	}
	if input.Description != "" {
		updates["description"] = input.Description
	}
	if input.Color != "" {
		updates["color"] = input.Color
	}
	if input.Icon != "" {
		updates["icon"] = input.Icon
	}
	if input.SortOrder != nil {
		updates["sort_order"] = *input.SortOrder
	}

	if len(updates) > 0 {
		if err := database.DB.WithContext(ctx).
			Model(&SessionGroup{}).
			Where("id = ?", group.ID).
			Updates(updates).Error; err != nil {
			return SessionGroup{}, err
		}
	}

	return GetSessionGroupByID(ctx, groupID)
}

func DeleteSessionGroup(ctx context.Context, groupID string) error {
	return database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Session{}).
			Where("group_id = ?", groupID).
			Update("group_id", nil).Error; err != nil {
			return err
		}
		return tx.Delete(&SessionGroup{}, "id = ?", groupID).Error
	})
}

func IsSessionGroupOwner(ctx context.Context, groupID, userID string) (bool, error) {
	group, err := GetSessionGroupByID(ctx, groupID)
	if err != nil {
		return false, err
	}
	return group.UserID == userID, nil
}

type CreateSessionShareInput struct {
	SessionID string
	UserID    string
	Title     string
	ExpiresAt *time.Time
	AllowCopy bool
	IsPublic  bool
}

func CreateSessionShare(ctx context.Context, input CreateSessionShareInput) (SessionShare, error) {
	token := generateShareToken()

	share := SessionShare{
		SessionID:  strings.TrimSpace(input.SessionID),
		UserID:     strings.TrimSpace(input.UserID),
		ShareToken: token,
		Title:      strings.TrimSpace(input.Title),
		ExpiresAt:  input.ExpiresAt,
		AllowCopy:  input.AllowCopy,
		IsPublic:   input.IsPublic,
	}

	if err := database.DB.WithContext(ctx).Create(&share).Error; err != nil {
		return SessionShare{}, err
	}

	return share, nil
}

func GetSessionShareByToken(ctx context.Context, token string) (SessionShare, error) {
	var share SessionShare
	err := database.DB.WithContext(ctx).
		Where("share_token = ?", strings.TrimSpace(token)).
		First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return SessionShare{}, errors.New("share not found")
	}
	return share, err
}

func GetSessionSharesByUser(ctx context.Context, userID string) ([]SessionShare, error) {
	var shares []SessionShare
	err := database.DB.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Order("created_at DESC").
		Find(&shares).Error
	return shares, err
}

func DeleteSessionShare(ctx context.Context, shareID string) error {
	return database.DB.WithContext(ctx).
		Delete(&SessionShare{}, "id = ?", shareID).Error
}

func IncrementShareViewCount(ctx context.Context, token string) error {
	return database.DB.WithContext(ctx).
		Model(&SessionShare{}).
		Where("share_token = ?", token).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func IsShareExpired(share SessionShare) bool {
	if share.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*share.ExpiresAt)
}

func generateShareToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
