package document

import (
	"time"

	"gorm.io/datatypes"
)

const (
	VersionSourceAutosave = "autosave"
	VersionSourceManual   = "manual"
	VersionSourceSnapshot = "snapshot"
)

type Document struct {
	ID          string  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	WorkspaceID string  `json:"workspace_id" gorm:"type:uuid;not null"`
	FolderID    *string `json:"folder_id,omitempty" gorm:"type:uuid;index"`

	Title             string         `json:"title" gorm:"type:varchar(255);not null"`
	ContentJSON       datatypes.JSON `json:"content_json" gorm:"type:jsonb;not null;default:'{}'::jsonb"`
	TiptapSchema      string         `json:"tiptap_schema" gorm:"type:varchar(50);not null;default:'doc'"`
	TiptapSchemaVer   string         `json:"tiptap_schema_ver" gorm:"type:varchar(50);not null;default:'v1'"`
	CurrentVersion    int64          `json:"current_version" gorm:"not null;default:1"`
	LatestSnapshotID  *string        `json:"latest_snapshot_id,omitempty" gorm:"type:uuid"`
	LatestSnapshotVer int64          `json:"latest_snapshot_version" gorm:"column:latest_snapshot_ver;not null;default:0"`
	LatestSnapshotAt  *time.Time     `json:"latest_snapshot_at,omitempty"`
	LastVersionedAt   *time.Time     `json:"last_versioned_at,omitempty"`
	Versions          []Version      `json:"-" gorm:"foreignKey:DocumentID;constraint:OnDelete:CASCADE;"`
	Public            bool           `json:"public" gorm:"not null;default:false"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type Version struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	DocumentID  string         `json:"document_id" gorm:"type:uuid;not null;index:idx_document_version,unique"`
	WorkspaceID string         `json:"workspace_id" gorm:"type:uuid;not null;index"`
	Version     int64          `json:"version" gorm:"not null;index:idx_document_version,unique"`
	Title       string         `json:"title" gorm:"type:varchar(255);not null"`
	ContentJSON datatypes.JSON `json:"content_json" gorm:"type:jsonb;not null;default:'{}'::jsonb"`
	Source      string         `json:"source" gorm:"type:varchar(20);not null;default:'autosave'"`
	Snapshot    bool           `json:"snapshot" gorm:"not null;default:false;index"`
	Summary     string         `json:"summary" gorm:"type:varchar(255)"`
	CreatedBy   *string        `json:"created_by,omitempty" gorm:"type:uuid"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
}

func (Version) TableName() string {
	return "document_versions"
}
