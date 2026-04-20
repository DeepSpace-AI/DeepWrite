package agent

import (
	"time"

	"gorm.io/datatypes"
)

type Message struct {
	ID         string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	SessionID  string         `json:"session_id" gorm:"type:uuid;not null;index"`
	Role       string         `json:"role" gorm:"type:varchar(20);not null;index"`
	Content    string         `json:"content" gorm:"type:text;not null"`
	TokenCount int            `json:"token_count" gorm:"default:0"`
	ModelUsed  string         `json:"model_used" gorm:"type:varchar(100)"`
	Sources    datatypes.JSON `json:"sources" gorm:"type:jsonb;default:'[]'::jsonb"`
	Actions    datatypes.JSON `json:"actions" gorm:"type:jsonb;default:'[]'::jsonb"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
}

func (Message) TableName() string {
	return DefaultMessageTableName
}

type MessageSource struct {
	ID    string `json:"id"`
	Type  string `json:"type"` // document, reference
	Title string `json:"title"`
}

type MessageAction struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Type    string `json:"type"` // summarize, logic_check, translate, etc.
	Enabled bool   `json:"enabled"`
}
