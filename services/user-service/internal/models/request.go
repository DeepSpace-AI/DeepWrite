package models

import (
	"errors"
	"time"
)

type RegisterRequest struct {
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=6"`
	Name     string   `json:"name" binding:"required,min=2"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	User         User      `json:"user"`
}

type UpdateUserRequest struct {
	Name           string   `json:"name,omitempty"`
	Institution    string   `json:"institution,omitempty"`
	ResearchFields []string `json:"research_fields,omitempty"`
}

type CreateTeamRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description,omitempty"`
	MaxMembers  int    `json:"max_members,omitempty"`
}

type UpdateTeamRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	MaxMembers  int    `json:"max_members,omitempty"`
}

type AddTeamMemberRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	Role   string `json:"role" binding:"required,oneof=admin member viewer"`
}

type UpdateTeamMemberRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member viewer"`
}

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailExists       = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTeamNotFound      = errors.New("team not found")
	ErrTeamMemberExists  = errors.New("user is already a team member")
	ErrUnauthorized      = errors.New("unauthorized")
)

func GenerateRefreshToken() string {
	return "refresh_" + time.Now().Format("20060102150405") + "_random"
}
