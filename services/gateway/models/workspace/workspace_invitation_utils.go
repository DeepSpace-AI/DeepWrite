package workspace

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/models/user"
	"github.com/deepwrite/serivces/gateway/pkg/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInvitationPendingExists = errors.New("pending invitation already exists")
	ErrInvitationExpired       = errors.New("invitation is expired")
	ErrInvitationNotPending    = errors.New("invitation is not pending")
	ErrInvitationTokenInvalid  = errors.New("invitation token is invalid")
	ErrInvitationTargetInvalid = errors.New("invitation target is invalid")
)

type CreateInvitationInput struct {
	WorkspaceID   string
	InviterID     string
	InviteeUserID *string
	InviteeEmail  string
	Role          string
	ExpiresAt     time.Time
}

func CreateInvitation(ctx context.Context, input CreateInvitationInput) (WorkspaceInvitation, string, error) {
	invitation, err := findPendingInvitation(ctx, strings.TrimSpace(input.WorkspaceID), normalizeEmail(input.InviteeEmail), normalizeOptionalID(input.InviteeUserID))
	if err == nil {
		if invitation.ExpiresAt.After(time.Now()) {
			return invitation, "", ErrInvitationPendingExists
		}
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return WorkspaceInvitation{}, "", err
	}

	token, err := generateInvitationToken()
	if err != nil {
		return WorkspaceInvitation{}, "", err
	}

	inviteeUserID := normalizeOptionalID(input.InviteeUserID)
	inviteeEmail := normalizeEmail(input.InviteeEmail)
	role := strings.TrimSpace(input.Role)
	if role == "" {
		role = RoleViewer
	}

	created := WorkspaceInvitation{
		WorkspaceID:   strings.TrimSpace(input.WorkspaceID),
		InviterID:     strings.TrimSpace(input.InviterID),
		InviteeUserID: inviteeUserID,
		InviteeEmail:  inviteeEmail,
		Role:          role,
		Status:        InvitationStatusPending,
		TokenHash:     hashInvitationToken(token),
		ExpiresAt:     input.ExpiresAt,
	}

	if err := database.DB.WithContext(ctx).Create(&created).Error; err != nil {
		return WorkspaceInvitation{}, "", err
	}

	return created, token, nil
}

func GetInvitationByID(ctx context.Context, invitationID string) (WorkspaceInvitation, error) {
	var invitation WorkspaceInvitation
	err := database.DB.WithContext(ctx).
		Where("id = ?", strings.TrimSpace(invitationID)).
		First(&invitation).Error
	return invitation, err
}

func ListInvitationsByWorkspace(ctx context.Context, workspaceID, status string, limit, offset int) ([]WorkspaceInvitation, error) {
	query := database.DB.WithContext(ctx).
		Where("workspace_id = ?", strings.TrimSpace(workspaceID)).
		Order("created_at desc")

	if strings.TrimSpace(status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(status))
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var invitations []WorkspaceInvitation
	err := query.Find(&invitations).Error
	return invitations, err
}

func ListInvitationsForUser(ctx context.Context, userID, email, status string, limit, offset int) ([]WorkspaceInvitation, error) {
	query := database.DB.WithContext(ctx).
		Where("invitee_user_id = ? OR invitee_email = ?", strings.TrimSpace(userID), normalizeEmail(email)).
		Order("created_at desc")

	if strings.TrimSpace(status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(status))
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var invitations []WorkspaceInvitation
	err := query.Find(&invitations).Error
	return invitations, err
}

func RevokeInvitation(ctx context.Context, invitationID string) (WorkspaceInvitation, error) {
	var invitation WorkspaceInvitation
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", strings.TrimSpace(invitationID)).
			First(&invitation).Error; err != nil {
			return err
		}

		if invitation.Status != InvitationStatusPending {
			return ErrInvitationNotPending
		}

		now := time.Now()
		invitation.Status = InvitationStatusRevoked
		invitation.RevokedAt = &now
		invitation.TokenHash = ""

		return tx.Model(&WorkspaceInvitation{}).Where("id = ?", invitation.ID).Updates(map[string]any{
			"status":     invitation.Status,
			"revoked_at": invitation.RevokedAt,
			"token_hash": invitation.TokenHash,
		}).Error
	})

	return invitation, err
}

func AcceptInvitation(ctx context.Context, invitationID, token, userID, userEmail string) (WorkspaceInvitation, error) {
	var invitation WorkspaceInvitation
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", strings.TrimSpace(invitationID)).
			First(&invitation).Error; err != nil {
			return err
		}

		if invitation.Status != InvitationStatusPending {
			return ErrInvitationNotPending
		}

		if invitation.ExpiresAt.Before(time.Now()) {
			now := time.Now()
			_ = tx.Model(&WorkspaceInvitation{}).Where("id = ?", invitation.ID).Updates(map[string]any{
				"status": InvitationStatusExpired,
			}).Error
			invitation.Status = InvitationStatusExpired
			invitation.UpdatedAt = now
			return ErrInvitationExpired
		}

		trimmedToken := strings.TrimSpace(token)
		if trimmedToken != "" {
			if invitation.TokenHash == "" || invitation.TokenHash != hashInvitationToken(trimmedToken) {
				return ErrInvitationTokenInvalid
			}
		}

		if !isInviteeMatched(invitation, userID, userEmail) {
			return ErrInvitationTargetInvalid
		}

		now := time.Now()
		invitation.Status = InvitationStatusAccepted
		invitation.AcceptedAt = &now
		invitation.TokenHash = ""

		if err := upsertMember(tx, invitation.WorkspaceID, strings.TrimSpace(userID), invitation.Role); err != nil {
			return err
		}

		return tx.Model(&WorkspaceInvitation{}).Where("id = ?", invitation.ID).Updates(map[string]any{
			"status":      invitation.Status,
			"accepted_at": invitation.AcceptedAt,
			"token_hash":  invitation.TokenHash,
		}).Error
	})

	return invitation, err
}

func RejectInvitation(ctx context.Context, invitationID, token, userID, userEmail string) (WorkspaceInvitation, error) {
	var invitation WorkspaceInvitation
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", strings.TrimSpace(invitationID)).
			First(&invitation).Error; err != nil {
			return err
		}

		if invitation.Status != InvitationStatusPending {
			return ErrInvitationNotPending
		}

		if invitation.ExpiresAt.Before(time.Now()) {
			_ = tx.Model(&WorkspaceInvitation{}).Where("id = ?", invitation.ID).Updates(map[string]any{
				"status": InvitationStatusExpired,
			}).Error
			invitation.Status = InvitationStatusExpired
			return ErrInvitationExpired
		}

		trimmedToken := strings.TrimSpace(token)
		if trimmedToken != "" {
			if invitation.TokenHash == "" || invitation.TokenHash != hashInvitationToken(trimmedToken) {
				return ErrInvitationTokenInvalid
			}
		}

		if !isInviteeMatched(invitation, userID, userEmail) {
			return ErrInvitationTargetInvalid
		}

		now := time.Now()
		invitation.Status = InvitationStatusRejected
		invitation.RejectedAt = &now
		invitation.TokenHash = ""

		return tx.Model(&WorkspaceInvitation{}).Where("id = ?", invitation.ID).Updates(map[string]any{
			"status":      invitation.Status,
			"rejected_at": invitation.RejectedAt,
			"token_hash":  invitation.TokenHash,
		}).Error
	})

	return invitation, err
}

func ResolveInvitee(ctx context.Context, inviteeUserID *string, inviteeEmail string) (*string, string, error) {
	normalizedEmail := normalizeEmail(inviteeEmail)
	normalizedUserID := normalizeOptionalID(inviteeUserID)

	if normalizedUserID == nil && normalizedEmail == "" {
		return nil, "", gorm.ErrInvalidData
	}

	if normalizedUserID != nil {
		u, err := user.GetUserByID(ctx, *normalizedUserID)
		if err != nil {
			return nil, "", err
		}
		if normalizedEmail != "" && normalizeEmail(u.Email) != normalizedEmail {
			return nil, "", ErrInvitationTargetInvalid
		}
		email := normalizeEmail(u.Email)
		return normalizedUserID, email, nil
	}

	u, err := user.GetUserByEmail(ctx, normalizedEmail)
	if err == nil {
		id := strings.TrimSpace(u.ID)
		return &id, normalizedEmail, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}

	return nil, normalizedEmail, nil
}

func IsValidInviteRole(role string) bool {
	switch strings.TrimSpace(role) {
	case RoleAdmin, RoleEditor, RoleViewer:
		return true
	default:
		return false
	}
}

func findPendingInvitation(ctx context.Context, workspaceID, inviteeEmail string, inviteeUserID *string) (WorkspaceInvitation, error) {
	query := database.DB.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Where("status = ?", InvitationStatusPending)

	if inviteeUserID != nil {
		query = query.Where("invitee_user_id = ?", *inviteeUserID)
	} else {
		query = query.Where("invitee_email = ?", inviteeEmail)
	}

	var invitation WorkspaceInvitation
	err := query.First(&invitation).Error
	return invitation, err
}

func upsertMember(tx *gorm.DB, workspaceID, userID, role string) error {
	var member Members
	err := tx.Where("workspace_id = ? AND user_id = ?", workspaceID, userID).First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		member = Members{
			WorkspaceID: workspaceID,
			UserId:      userID,
			Role:        role,
		}
		return tx.Create(&member).Error
	}
	if err != nil {
		return err
	}

	return tx.Model(&Members{}).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Update("role", role).Error
}

func isInviteeMatched(invitation WorkspaceInvitation, userID, userEmail string) bool {
	trimmedUserID := strings.TrimSpace(userID)
	normalizedEmail := normalizeEmail(userEmail)

	if invitation.InviteeUserID != nil {
		return strings.TrimSpace(*invitation.InviteeUserID) == trimmedUserID
	}

	return normalizeEmail(invitation.InviteeEmail) == normalizedEmail
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeOptionalID(v *string) *string {
	if v == nil {
		return nil
	}
	id := strings.TrimSpace(*v)
	if id == "" {
		return nil
	}
	return &id
}

func generateInvitationToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashInvitationToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}
