package service

import (
	"context"

	"github.com/deepwrite/user-service/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User, password string) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateLastLogin(ctx context.Context, id string) error
	EmailExists(ctx context.Context, email string) (bool, error)
	VerifyPassword(hash, password string) bool
}

type TeamRepository interface {
	Create(ctx context.Context, team *models.Team) error
	GetByID(ctx context.Context, id string) (*models.Team, error)
	ListByUser(ctx context.Context, userID string) ([]*models.Team, error)
	Update(ctx context.Context, team *models.Team) error
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, teamID, userID, role string) error
	RemoveMember(ctx context.Context, teamID, userID string) error
	GetMembers(ctx context.Context, teamID string) ([]*models.TeamMember, error)
	IsMember(ctx context.Context, teamID, userID string) (bool, error)
	GetMemberRole(ctx context.Context, teamID, userID string) (string, error)
	IsOwner(ctx context.Context, teamID, userID string) (bool, error)
	GetMemberCount(ctx context.Context, teamID string) (int, error)
}
