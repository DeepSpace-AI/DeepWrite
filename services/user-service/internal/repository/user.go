package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/deepwrite/user-service/internal/models"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	user.PasswordHash = string(hash)

	query := `
		INSERT INTO users (email, password_hash, name, institution, research_fields, subscription_tier)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		user.Email, user.PasswordHash, user.Name,
		sql.NullString{String: user.Institution.String, Valid: user.Institution.String != ""},
		pq.Array(user.ResearchFields), user.SubscriptionTier,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, password_hash, name, avatar_url, institution, research_fields,
		       subscription_tier, subscription_expires_at, email_verified_at, last_login_at, created_at, updated_at
		FROM users WHERE email = $1
	`
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.AvatarURL, &user.Institution, pq.Array(&user.ResearchFields),
		&user.SubscriptionTier, &user.SubscriptionExpiresAt, &user.EmailVerifiedAt,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrUserNotFound
	}
	return user, err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, name, avatar_url, institution, research_fields,
		       subscription_tier, subscription_expires_at, email_verified_at, last_login_at, created_at, updated_at
		FROM users WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Name,
		&user.AvatarURL, &user.Institution, pq.Array(&user.ResearchFields),
		&user.SubscriptionTier, &user.SubscriptionExpiresAt, &user.EmailVerifiedAt,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrUserNotFound
	}
	return user, err
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET name = $2, institution = $3, research_fields = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		user.ID, user.Name,
		sql.NullString{String: user.Institution.String, Valid: user.Institution.String != ""},
		pq.Array(user.ResearchFields),
	).Scan(&user.UpdatedAt)
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	query := `UPDATE users SET last_login_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE email = $1`
	err := r.db.QueryRowContext(ctx, query, email).Scan(&count)
	return count > 0, err
}

func (r *UserRepository) VerifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
