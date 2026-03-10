package document

import "time"

type CollabToken struct {
	ID          string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	DocumentID  string     `json:"document_id" gorm:"type:uuid;not null;index"`
	WorkspaceID string     `json:"workspace_id" gorm:"type:uuid;not null;index"`
	UserID      string     `json:"user_id" gorm:"type:uuid;not null;index"`
	Role        string     `json:"role" gorm:"type:varchar(50);not null"`
	TokenHash   string     `json:"-" gorm:"type:varchar(128);not null;uniqueIndex"`
	ExpiresAt   time.Time  `json:"expires_at" gorm:"not null;index"`
	UsedAt      *time.Time `json:"used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

type CollabUpdate struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	DocumentID  string    `json:"document_id" gorm:"type:uuid;not null;index:idx_collab_doc_seq"`
	WorkspaceID string    `json:"workspace_id" gorm:"type:uuid;not null;index"`
	SessionID   string    `json:"session_id" gorm:"type:varchar(100);not null;index"`
	UserID      string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Seq         int64     `json:"seq" gorm:"not null;index:idx_collab_doc_seq"`
	Payload     []byte    `json:"payload" gorm:"type:bytea;not null"`
	Kind        string    `json:"kind" gorm:"type:varchar(20);not null;default:'sync'"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type CollabState struct {
	DocumentID           string     `json:"document_id" gorm:"primaryKey;type:uuid"`
	WorkspaceID          string     `json:"workspace_id" gorm:"type:uuid;not null;index"`
	LastSeq              int64      `json:"last_seq" gorm:"not null;default:0"`
	FlushSeq             int64      `json:"flush_seq" gorm:"not null;default:0"`
	PendingUpdateCount   int64      `json:"pending_update_count" gorm:"not null;default:0"`
	SnapshotUpdateCount  int64      `json:"snapshot_update_count" gorm:"not null;default:0"`
	LastSnapshotAt       *time.Time `json:"last_snapshot_at,omitempty"`
	LastFlushedAt        *time.Time `json:"last_flushed_at,omitempty"`
	LastRecoveredUntilAt *time.Time `json:"last_recovered_until_at,omitempty"`
	UpdatedAt            time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt            time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

type CollabAudit struct {
	ID            string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	DocumentID    string    `json:"document_id" gorm:"type:uuid;not null;index"`
	WorkspaceID   string    `json:"workspace_id" gorm:"type:uuid;not null;index"`
	SessionID     string    `json:"session_id" gorm:"type:varchar(100);not null;index"`
	UserID        string    `json:"user_id" gorm:"type:uuid;not null;index"`
	JoinedAt      time.Time `json:"joined_at" gorm:"not null"`
	LeftAt        time.Time `json:"left_at" gorm:"not null"`
	FlushSeq      int64     `json:"flush_seq" gorm:"not null;default:0"`
	UpdateCount   int64     `json:"update_count" gorm:"not null;default:0"`
	BytesIn       int64     `json:"bytes_in" gorm:"not null;default:0"`
	AckLatencyP95 int64     `json:"ack_latency_p95" gorm:"not null;default:0"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}
