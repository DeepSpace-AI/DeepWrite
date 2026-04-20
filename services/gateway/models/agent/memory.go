package agent

import (
	"time"

	"gorm.io/datatypes"
)

type Memory struct {
	ID              string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AgentID         string         `json:"agent_id" gorm:"type:uuid;not null;index"`
	UserID          string         `json:"user_id" gorm:"type:uuid;not null;index"`
	SessionID       *string        `json:"session_id" gorm:"type:uuid;index"`
	WorkspaceID     *string        `json:"workspace_id" gorm:"type:uuid;index"`
	Type            string         `json:"type" gorm:"type:varchar(30);not null;index"`
	Content         string         `json:"content" gorm:"type:text;not null"`
	Summary         string         `json:"summary" gorm:"type:text"`
	Keywords        datatypes.JSON `json:"keywords" gorm:"type:jsonb;default:'[]'::jsonb"`
	EmbeddingVector datatypes.JSON `json:"embedding_vector" gorm:"type:jsonb"`
	SourceSessionID *string        `json:"source_session_id" gorm:"type:uuid"`
	Importance      float64        `json:"importance" gorm:"default:0.5"`
	AccessCount     int            `json:"access_count" gorm:"default:0"`
	ExpiresAt       *time.Time     `json:"expires_at"`
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Memory) TableName() string {
	return DefaultMemoryTableName
}
