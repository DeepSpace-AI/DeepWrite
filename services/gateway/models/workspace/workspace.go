package workspace

import "time"

const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleViewer = "viewer"

	InvitationStatusPending  = "pending"
	InvitationStatusAccepted = "accepted"
	InvitationStatusRejected = "rejected"
	InvitationStatusExpired  = "expired"
	InvitationStatusRevoked  = "revoked"
)

type Workspace struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:text"`
	OwnerID     string    `json:"owner_id" gorm:"type:uuid;not null"`
	Members     []Members `json:"members" gorm:"foreignKey:WorkspaceID;constraint:OnDelete:CASCADE;"`
	Public      bool      `json:"public" gorm:"not null;default:false"`
	Status      string    `json:"status" gorm:"type:varchar(50);not null;default:'active'"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type Members struct {
	WorkspaceID string `json:"workspace_id" gorm:"type:uuid;not null"`
	UserId      string `json:"user_id" gorm:"type:uuid;not null"`
	Role        string `json:"role" gorm:"type:varchar(50);not null"` // e.g., "admin", "member"
}

type Folder struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	WorkspaceID string    `json:"workspace_id" gorm:"type:uuid;not null"`
	ParentID    *string   `json:"parent_id,omitempty" gorm:"type:uuid;"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type WorkspaceInvitation struct {
	ID            string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	WorkspaceID   string     `json:"workspace_id" gorm:"type:uuid;not null;index"`
	InviterID     string     `json:"inviter_id" gorm:"type:uuid;not null;index"`
	InviteeUserID *string    `json:"invitee_user_id,omitempty" gorm:"type:uuid;index"`
	InviteeEmail  string     `json:"invitee_email" gorm:"type:varchar(255);not null;index"`
	Role          string     `json:"role" gorm:"type:varchar(50);not null"`
	Status        string     `json:"status" gorm:"type:varchar(50);not null;default:'pending';index"`
	TokenHash     string     `json:"-" gorm:"type:varchar(128);not null"`
	ExpiresAt     time.Time  `json:"expires_at" gorm:"not null;index"`
	AcceptedAt    *time.Time `json:"accepted_at,omitempty"`
	RejectedAt    *time.Time `json:"rejected_at,omitempty"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

type WorkspaceFile struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	WorkspaceID string    `json:"workspace_id" gorm:"type:uuid;not null;index"`
	FolderID    *string   `json:"folder_id,omitempty" gorm:"type:uuid;index"`
	ObjectKey   string    `json:"object_key" gorm:"type:varchar(1024);not null;uniqueIndex:idx_workspace_object"`
	FileName    string    `json:"file_name" gorm:"type:varchar(255);not null"`
	ContentType string    `json:"content_type" gorm:"type:varchar(255);not null;default:'application/octet-stream'"`
	Size        int64     `json:"size" gorm:"not null;default:0"`
	ETag        string    `json:"etag" gorm:"type:varchar(255)"`
	UploadedBy  string    `json:"uploaded_by" gorm:"type:uuid;not null;index"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
