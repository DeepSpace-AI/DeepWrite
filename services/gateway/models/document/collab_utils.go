package document

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/pkg/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func generateRawToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func IssueCollabToken(ctx context.Context, doc Document, userID, role string, ttl time.Duration) (string, CollabToken, error) {
	raw, err := generateRawToken()
	if err != nil {
		return "", CollabToken{}, err
	}

	now := time.Now()
	token := CollabToken{
		DocumentID:  doc.ID,
		WorkspaceID: doc.WorkspaceID,
		UserID:      strings.TrimSpace(userID),
		Role:        strings.TrimSpace(role),
		TokenHash:   hashToken(raw),
		ExpiresAt:   now.Add(ttl),
	}

	if err := database.DB.WithContext(ctx).Create(&token).Error; err != nil {
		return "", CollabToken{}, err
	}

	return raw, token, nil
}

func ConsumeCollabToken(ctx context.Context, rawToken string, documentID string, singleUse bool) (CollabToken, error) {
	if strings.TrimSpace(rawToken) == "" {
		return CollabToken{}, gorm.ErrRecordNotFound
	}

	var token CollabToken
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ? AND document_id = ?", hashToken(rawToken), strings.TrimSpace(documentID)).
			First(&token).Error; err != nil {
			return err
		}

		now := time.Now()
		if now.After(token.ExpiresAt) {
			return errors.New("collab token expired")
		}

		if singleUse && token.UsedAt != nil {
			return errors.New("collab token already used")
		}

		if singleUse {
			token.UsedAt = &now
			if err := tx.Model(&CollabToken{}).Where("id = ?", token.ID).Update("used_at", now).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return token, err
}

func ListCollabUpdates(ctx context.Context, documentID string, limit int) ([]CollabUpdate, error) {
	if limit <= 0 {
		limit = 5000
	}

	var updates []CollabUpdate
	err := database.DB.WithContext(ctx).
		Where("document_id = ?", strings.TrimSpace(documentID)).
		Order("seq asc").
		Limit(limit).
		Find(&updates).Error
	return updates, err
}

func GetOrCreateCollabState(ctx context.Context, documentID, workspaceID string) (CollabState, error) {
	var state CollabState
	err := database.DB.WithContext(ctx).Where("document_id = ?", strings.TrimSpace(documentID)).First(&state).Error
	if err == nil {
		return state, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return CollabState{}, err
	}

	state = CollabState{
		DocumentID:  strings.TrimSpace(documentID),
		WorkspaceID: strings.TrimSpace(workspaceID),
	}
	if err := database.DB.WithContext(ctx).Create(&state).Error; err != nil {
		return CollabState{}, err
	}
	return state, nil
}

func UpsertCollabState(ctx context.Context, state CollabState) error {
	return database.DB.WithContext(ctx).Save(&state).Error
}

func SaveCollabUpdates(ctx context.Context, updates []CollabUpdate) error {
	if len(updates) == 0 {
		return nil
	}
	return database.DB.WithContext(ctx).Create(&updates).Error
}

func CreateCollabAudit(ctx context.Context, audit CollabAudit) error {
	return database.DB.WithContext(ctx).Create(&audit).Error
}
