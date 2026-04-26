package service

import (
	"context"
	"testing"

	"github.com/deepwrite/user-service/internal/config"
	"github.com/deepwrite/user-service/internal/models"
)

type mockUserRepo struct {
	users       map[string]*models.User
	emailExists map[string]bool
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:       make(map[string]*models.User),
		emailExists: make(map[string]bool),
	}
}

func (r *mockUserRepo) Create(_ context.Context, user *models.User, password string) error {
	if r.emailExists[user.Email] {
		return models.ErrEmailExists
	}
	r.users[user.ID] = user
	r.emailExists[user.Email] = true
	return nil
}

func (r *mockUserRepo) GetByEmail(_ context.Context, email string) (*models.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, models.ErrUserNotFound
}

func (r *mockUserRepo) GetByID(_ context.Context, id string) (*models.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, models.ErrUserNotFound
}

func (r *mockUserRepo) Update(_ context.Context, user *models.User) error {
	r.users[user.ID] = user
	return nil
}

func (r *mockUserRepo) UpdateLastLogin(_ context.Context, _ string) error {
	return nil
}

func (r *mockUserRepo) EmailExists(_ context.Context, email string) (bool, error) {
	return r.emailExists[email], nil
}

func (r *mockUserRepo) VerifyPassword(hash, password string) bool {
	return hash == password
}

func TestUserService_Register(t *testing.T) {
	repo := newMockUserRepo()
	cfg := &config.JWTConfig{Secret: "test-secret", AccessTokenTTL: 15}
	svc := NewUserService(repo, cfg)

	req := &models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	resp, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("Expected access token to be generated")
	}
	if resp.User.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, resp.User.Email)
	}
}

func TestUserService_Register_EmailExists(t *testing.T) {
	repo := newMockUserRepo()
	cfg := &config.JWTConfig{Secret: "test-secret", AccessTokenTTL: 15}
	svc := NewUserService(repo, cfg)

	req := &models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	_, _ = svc.Register(context.Background(), req)
	_, err := svc.Register(context.Background(), req)
	if err != models.ErrEmailExists {
		t.Errorf("Expected ErrEmailExists, got %v", err)
	}
}

func TestUserService_Login(t *testing.T) {
	repo := newMockUserRepo()
	cfg := &config.JWTConfig{Secret: "test-secret", AccessTokenTTL: 15}
	svc := NewUserService(repo, cfg)

	regReq := &models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	_, _ = svc.Register(context.Background(), regReq)

	loginReq := &models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	resp, err := svc.Login(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("Expected access token")
	}
}

func TestUserService_Login_InvalidCredentials(t *testing.T) {
	repo := newMockUserRepo()
	cfg := &config.JWTConfig{Secret: "test-secret", AccessTokenTTL: 15}
	svc := NewUserService(repo, cfg)

	loginReq := &models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "wrongpassword",
	}
	_, err := svc.Login(context.Background(), loginReq)
	if err != models.ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserService_ValidateToken(t *testing.T) {
	repo := newMockUserRepo()
	cfg := &config.JWTConfig{Secret: "test-secret", AccessTokenTTL: 15}
	svc := NewUserService(repo, cfg)

	regReq := &models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	resp, _ := svc.Register(context.Background(), regReq)

	claims, err := svc.ValidateToken(resp.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if (*claims)["email"] != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %v", (*claims)["email"])
	}
}

func TestUserService_ValidateToken_Invalid(t *testing.T) {
	repo := newMockUserRepo()
	cfg := &config.JWTConfig{Secret: "test-secret", AccessTokenTTL: 15}
	svc := NewUserService(repo, cfg)

	_, err := svc.ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}
