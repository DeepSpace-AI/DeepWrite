package workspace

import "time"

type AnnotationType string

const (
	AnnotationTypeHighlight AnnotationType = "highlight"
	AnnotationTypeNote      AnnotationType = "note"
	AnnotationTypeDrawing   AnnotationType = "drawing"
)

type FileAnnotation struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	FileID      string         `json:"file_id" gorm:"type:uuid;not null;index"`
	UserID      string         `json:"user_id" gorm:"type:uuid;not null;index"`
	WorkspaceID string         `json:"workspace_id" gorm:"type:uuid;not null;index"`
	Type        AnnotationType `json:"type" gorm:"type:varchar(20);not null"`

	Page       int     `json:"page" gorm:"not null"`
	RectX      float64 `json:"rect_x" gorm:"type:decimal(10,6)"`
	RectY      float64 `json:"rect_y" gorm:"type:decimal(10,6)"`
	RectWidth  float64 `json:"rect_width" gorm:"type:decimal(10,6)"`
	RectHeight float64 `json:"rect_height" gorm:"type:decimal(10,6)"`

	Color   string `json:"color,omitempty" gorm:"type:varchar(20)"`
	Content string `json:"content,omitempty" gorm:"type:text"`
	Paths   string `json:"paths,omitempty" gorm:"type:text"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
