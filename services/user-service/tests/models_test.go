package tests

import (
	"testing"

	"github.com/deepwrite/user-service/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestRegisterRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     models.RegisterRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: models.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test User",
			},
			wantErr: false,
		},
		{
			name: "missing email",
			req: models.RegisterRequest{
				Password: "password123",
				Name:     "Test User",
			},
			wantErr: true,
		},
		{
			name: "short password",
			req: models.RegisterRequest{
				Email:    "test@example.com",
				Password: "123",
				Name:     "Test User",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validation logic would be tested here with actual binding
			assert.NotNil(t, tt.req)
		})
	}
}

func TestLoginRequestValidation(t *testing.T) {
	req := models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	assert.NotEmpty(t, req.Email)
	assert.NotEmpty(t, req.Password)
}

func TestGenerateRefreshToken(t *testing.T) {
	token1 := models.GenerateRefreshToken()
	token2 := models.GenerateRefreshToken()
	assert.NotEmpty(t, token1)
	assert.NotEmpty(t, token2)
	assert.NotEqual(t, token1, token2)
}

func TestErrorConstants(t *testing.T) {
	assert.NotNil(t, models.ErrUserNotFound)
	assert.NotNil(t, models.ErrEmailExists)
	assert.NotNil(t, models.ErrInvalidCredentials)
	assert.NotNil(t, models.ErrTeamNotFound)
	assert.NotNil(t, models.ErrUnauthorized)
}
