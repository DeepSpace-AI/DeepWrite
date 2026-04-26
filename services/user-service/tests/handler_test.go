package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deepwrite/user-service/internal/config"
	"github.com/deepwrite/user-service/internal/models"
	"github.com/deepwrite/user-service/internal/service"
	"github.com/gin-gonic/gin"
)

type testMockUserRepo struct {
	users       map[string]*models.User
	emailExists map[string]bool
}

func newTestMockUserRepo() *testMockUserRepo {
	return &testMockUserRepo{
		users:       make(map[string]*models.User),
		emailExists: make(map[string]bool),
	}
}

func (r *testMockUserRepo) Create(_ context.Context, user *models.User, password string) error {
	if r.emailExists[user.Email] {
		return models.ErrEmailExists
	}
	r.users[user.ID] = user
	r.emailExists[user.Email] = true
	return nil
}

func (r *testMockUserRepo) GetByEmail(_ context.Context, email string) (*models.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, models.ErrUserNotFound
}

func (r *testMockUserRepo) GetByID(_ context.Context, id string) (*models.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, models.ErrUserNotFound
}

func (r *testMockUserRepo) Update(_ context.Context, user *models.User) error {
	r.users[user.ID] = user
	return nil
}

func (r *testMockUserRepo) UpdateLastLogin(_ context.Context, _ string) error {
	return nil
}

func (r *testMockUserRepo) EmailExists(_ context.Context, email string) (bool, error) {
	return r.emailExists[email], nil
}

func (r *testMockUserRepo) VerifyPassword(hash, password string) bool {
	return hash == password
}

type testMockTeamRepo struct{}

func (r *testMockTeamRepo) Create(_ context.Context, _ *models.Team) error         { return nil }
func (r *testMockTeamRepo) GetByID(_ context.Context, _ string) (*models.Team, error) { return nil, models.ErrTeamNotFound }
func (r *testMockTeamRepo) ListByUser(_ context.Context, _ string) ([]*models.Team, error) { return nil, nil }
func (r *testMockTeamRepo) Update(_ context.Context, _ *models.Team) error         { return nil }
func (r *testMockTeamRepo) Delete(_ context.Context, _ string) error               { return nil }
func (r *testMockTeamRepo) AddMember(_ context.Context, _, _, _ string) error      { return nil }
func (r *testMockTeamRepo) RemoveMember(_ context.Context, _, _ string) error      { return nil }
func (r *testMockTeamRepo) GetMembers(_ context.Context, _ string) ([]*models.TeamMember, error) { return nil, nil }
func (r *testMockTeamRepo) IsMember(_ context.Context, _, _ string) (bool, error)  { return false, nil }
func (r *testMockTeamRepo) GetMemberRole(_ context.Context, _, _ string) (string, error) { return "", nil }
func (r *testMockTeamRepo) IsOwner(_ context.Context, _, _ string) (bool, error)   { return false, nil }
func (r *testMockTeamRepo) GetMemberCount(_ context.Context, _ string) (int, error) { return 0, nil }

func setupTestRouter() (*gin.Engine, *UserHandler, *TeamHandler) {
	gin.SetMode(gin.TestMode)

	userRepo := newTestMockUserRepo()
	teamRepo := &testMockTeamRepo{}
	cfg := &config.JWTConfig{Secret: "test-secret", AccessTokenTTL: 15}

	userSvc := service.NewUserService(userRepo, cfg)
	teamSvc := service.NewTeamService(teamRepo, userRepo)

	userHandler := NewUserHandler(userSvc)
	teamHandler := NewTeamHandler(teamSvc)

	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := router.Group("/api")
	{
		users := api.Group("/users")
		{
			users.POST("/register", userHandler.Register)
			users.POST("/login", userHandler.Login)
		}
	}

	return router, userHandler, teamHandler
}

func TestRegisterHandler(t *testing.T) {
	router, _, _ := setupTestRouter()

	reqBody := models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/users/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["success"] != true {
		t.Errorf("Expected success true, got %v", resp["success"])
	}
}

func TestRegisterHandler_InvalidInput(t *testing.T) {
	router, _, _ := setupTestRouter()

	reqBody := map[string]string{"email": "invalid"}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/users/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLoginHandler(t *testing.T) {
	router, _, _ := setupTestRouter()

	regReq := models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	jsonBody, _ := json.Marshal(regReq)
	req := httptest.NewRequest(http.MethodPost, "/api/users/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	loginReq := models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonBody, _ = json.Marshal(loginReq)
	req = httptest.NewRequest(http.MethodPost, "/api/users/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["success"] != true {
		t.Errorf("Expected success true, got %v", resp["success"])
	}
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	router, _, _ := setupTestRouter()

	loginReq := models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "wrongpassword",
	}
	jsonBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/users/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestHealthHandler(t *testing.T) {
	router, _, _ := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}
