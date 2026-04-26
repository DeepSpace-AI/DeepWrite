package models

import "errors"

type CreateProjectRequest struct {
	Title         string   `json:"title" binding:"required,min=2,max=500"`
	Description   string   `json:"description,omitempty"`
	TeamID        string   `json:"team_id,omitempty"`
	ResearchField string   `json:"research_field,omitempty"`
}

type UpdateProjectRequest struct {
	Title         string `json:"title,omitempty"`
	Description   string `json:"description,omitempty"`
	Status        string `json:"status,omitempty"`
	ResearchField string `json:"research_field,omitempty"`
}

type AddProjectMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=owner editor viewer"`
}

type UpdateProjectMemberRequest struct {
	Role string `json:"role" binding:"required,oneof=editor viewer"`
}

var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrMemberExists         = errors.New("user is already a project member")
	ErrInvalidRole          = errors.New("invalid role")
)
