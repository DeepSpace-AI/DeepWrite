package service

import (
	"context"
	"fmt"
	"time"

	"github.com/deepwrite/user-service/internal/config"
	"github.com/deepwrite/user-service/internal/models"
	"github.com/golang-jwt/jwt"
)

type UserService struct {
	userRepo UserRepository
	cfg      *config.JWTConfig
}

func NewUserService(userRepo UserRepository, cfg *config.JWTConfig) *UserService {
	return &UserService{userRepo: userRepo, cfg: cfg}
}

func (s *UserService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
	exists, err := s.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if exists {
		return nil, models.ErrEmailExists
	}

	user := &models.User{
		Email:            req.Email,
		Name:             req.Name,
		SubscriptionTier: "free",
	}

	if err := s.userRepo.Create(ctx, user, req.Password); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return s.generateAuthResponse(user)
}

func (s *UserService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == models.ErrUserNotFound {
			return nil, models.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	if !s.userRepo.VerifyPassword(user.PasswordHash, req.Password) {
		return nil, models.ErrInvalidCredentials
	}

	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	return s.generateAuthResponse(user)
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, id string, req *models.UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Institution != "" {
		user.Institution.String = req.Institution
		user.Institution.Valid = true
	}
	if req.ResearchFields != nil {
		user.ResearchFields = req.ResearchFields
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	return user, nil
}

func (s *UserService) generateAuthResponse(user *models.User) (*models.AuthResponse, error) {
	token, err := s.generateToken(user.ID, user.Email, user.SubscriptionTier)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	user.PasswordHash = ""

	return &models.AuthResponse{
		AccessToken:  token,
		RefreshToken: models.GenerateRefreshToken(),
		ExpiresIn:    s.cfg.AccessTokenTTL * 60,
		User:         *user,
	}, nil
}

func (s *UserService) generateToken(userID, email, tier string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"tier":  tier,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Duration(s.cfg.AccessTokenTTL) * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Secret))
}

func (s *UserService) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
