package reference

import (
	"time"

	"gorm.io/gorm"
)

type ReferenceNote struct {
	ID          string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ReferenceID string `json:"reference_id" gorm:"type:uuid;not null;index"`
	UserID      string `json:"user_id" gorm:"type:uuid;not null;index"`

	Title    string `json:"title,omitempty" gorm:"type:varchar(255)"`
	Content  string `json:"content" gorm:"type:text;not null"`
	PageFrom *int   `json:"page_from,omitempty"`
	PageTo   *int   `json:"page_to,omitempty"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (ReferenceNote) TableName() string {
	return "reference_notes"
}

type CreateNoteInput struct {
	Title    string `json:"title"`
	Content  string `json:"content" binding:"required"`
	PageFrom *int   `json:"page_from"`
	PageTo   *int   `json:"page_to"`
}

type UpdateNoteInput struct {
	Title    *string `json:"title"`
	Content  *string `json:"content"`
	PageFrom *int    `json:"page_from"`
	PageTo   *int    `json:"page_to"`
}

type ListNotesParams struct {
	ReferenceID string `form:"reference_id" binding:"required"`
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
}

func (p *ListNotesParams) Normalize() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 50
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}
