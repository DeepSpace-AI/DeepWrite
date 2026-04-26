package models

import (
	"database/sql"
	"time"
)

type Project struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Description    string         `json:"description,omitempty"`
	OwnerID        string         `json:"owner_id"`
	TeamID         sql.NullString `json:"team_id,omitempty"`
	Status         string         `json:"status"`
	ResearchField  sql.NullString `json:"research_field,omitempty"`
	IsArchived     bool           `json:"is_archived"`
	ArchivedAt     sql.NullTime   `json:"archived_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type ProjectMember struct {
	ID       string    `json:"id"`
	ProjectID string   `json:"project_id"`
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type ProjectSettings struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Settings  []byte    `json:"settings"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProjectTag struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}
