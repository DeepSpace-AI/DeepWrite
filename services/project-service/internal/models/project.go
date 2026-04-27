package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Project struct {
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	Description   sql.NullString `json:"-"`
	OwnerID       string         `json:"owner_id"`
	TeamID        sql.NullString `json:"-"`
	Status        string         `json:"status"`
	ResearchField sql.NullString `json:"-"`
	IsArchived    bool           `json:"is_archived"`
	ArchivedAt    sql.NullTime   `json:"-"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func (p Project) MarshalJSON() ([]byte, error) {
	type Alias Project
	return json.Marshal(&struct {
		Alias
		Description   *string `json:"description,omitempty"`
		TeamID        *string `json:"team_id,omitempty"`
		ResearchField *string `json:"research_field,omitempty"`
		ArchivedAt    *string `json:"archived_at,omitempty"`
	}{
		Alias:         (Alias)(p),
		Description:   nullStringToPtr(p.Description),
		TeamID:        nullStringToPtr(p.TeamID),
		ResearchField: nullStringToPtr(p.ResearchField),
		ArchivedAt:    nullTimeToPtr(p.ArchivedAt),
	})
}

func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func nullTimeToPtr(nt sql.NullTime) *string {
	if nt.Valid {
		s := nt.Time.Format(time.RFC3339)
		return &s
	}
	return nil
}

type ProjectMember struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
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
