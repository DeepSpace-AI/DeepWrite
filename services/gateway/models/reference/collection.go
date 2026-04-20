package reference

import (
	"time"

	"gorm.io/gorm"
)

type Collection struct {
	ID          string  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OwnerID     string  `json:"owner_id" gorm:"type:uuid;not null;index"`
	WorkspaceID *string `json:"workspace_id,omitempty" gorm:"type:uuid;index"`
	ParentID    *string `json:"parent_id,omitempty" gorm:"type:uuid;index"`

	Name        string `json:"name" gorm:"type:varchar(255);not null"`
	Description string `json:"description,omitempty" gorm:"type:text"`
	Color       string `json:"color,omitempty" gorm:"type:varchar(20)"`
	Icon        string `json:"icon,omitempty" gorm:"type:varchar(50)"`
	SortOrder   int    `json:"sort_order" gorm:"not null;default:0"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	References []CollectionReference `json:"-" gorm:"foreignKey:CollectionID;constraint:OnDelete:CASCADE;"`
}

func (Collection) TableName() string {
	return "collections"
}

type CollectionReference struct {
	CollectionID string    `json:"collection_id" gorm:"primaryKey;type:uuid"`
	ReferenceID  string    `json:"reference_id" gorm:"primaryKey;type:uuid"`
	SortOrder    int       `json:"sort_order" gorm:"not null;default:0"`
	AddedAt      time.Time `json:"added_at" gorm:"autoCreateTime"`
}

func (CollectionReference) TableName() string {
	return "collection_references"
}

type CreateCollectionInput struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	ParentID    *string `json:"parent_id"`
	Color       string  `json:"color"`
	Icon        string  `json:"icon"`
	WorkspaceID *string `json:"workspace_id"`
}

type UpdateCollectionInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	ParentID    *string `json:"parent_id"`
	Color       *string `json:"color"`
	Icon        *string `json:"icon"`
	SortOrder   *int    `json:"sort_order"`
}

type CollectionTree struct {
	Collection
	Children       []*CollectionTree `json:"children,omitempty"`
	ReferenceCount int               `json:"reference_count,omitempty"`
}

type ListCollectionsParams struct {
	WorkspaceID string `form:"workspace_id"`
	ParentID    string `form:"parent_id"`
}
