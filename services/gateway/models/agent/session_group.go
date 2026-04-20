package agent

import (
	"time"
)

const DefaultSessionGroupTableName = "dw_session_groups"

type SessionGroup struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID      string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Name        string    `json:"name" gorm:"type:varchar(100);not null"`
	Description string    `json:"description" gorm:"type:text"`
	Color       string    `json:"color" gorm:"type:varchar(20);default:'#6366f1'"`
	Icon        string    `json:"icon" gorm:"type:varchar(50)"`
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (SessionGroup) TableName() string {
	return DefaultSessionGroupTableName
}

type SessionShare struct {
	ID         string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	SessionID  string     `json:"session_id" gorm:"type:uuid;not null;index"`
	UserID     string     `json:"user_id" gorm:"type:uuid;not null;index"`
	ShareToken string     `json:"share_token" gorm:"type:varchar(32);uniqueIndex;not null"`
	Title      string     `json:"title" gorm:"type:varchar(255)"`
	ExpiresAt  *time.Time `json:"expires_at"`
	ViewCount  int        `json:"view_count" gorm:"default:0"`
	AllowCopy  bool       `json:"allow_copy" gorm:"default:true"`
	IsPublic   bool       `json:"is_public" gorm:"default:true"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (SessionShare) TableName() string {
	return "dw_session_shares"
}
