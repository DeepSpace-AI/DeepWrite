package agent

import (
	"time"

	"gorm.io/datatypes"
)

type Session struct {
	ID            string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AgentID       string         `json:"agent_id" gorm:"type:uuid;not null;index"`
	UserID        string         `json:"user_id" gorm:"type:uuid;not null;index"`
	WorkspaceID   *string        `json:"workspace_id" gorm:"type:uuid;index"`
	Title         string         `json:"title" gorm:"type:varchar(255)"`
	ContextDocs   datatypes.JSON `json:"context_docs" gorm:"type:jsonb;default:'[]'::jsonb"`
	ContextRefs   datatypes.JSON `json:"context_refs" gorm:"type:jsonb;default:'[]'::jsonb"`
	Summary       string         `json:"summary" gorm:"type:text"`
	TokenCount    int            `json:"token_count" gorm:"default:0"`
	Status        string         `json:"status" gorm:"type:varchar(20);not null;default:'active';index"`
	Pinned        bool           `json:"pinned" gorm:"not null;default:false;index"`
	Archived      bool           `json:"archived" gorm:"not null;default:false;index"`
	GroupID       *string        `json:"group_id" gorm:"type:uuid;index"`
	Tags          datatypes.JSON `json:"tags" gorm:"type:jsonb;default:'[]'::jsonb"`
	LastMessageAt time.Time      `json:"last_message_at" gorm:"autoCreateTime"`
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Session) TableName() string {
	return DefaultSessionTableName
}

type ContextDoc struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ContextRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
