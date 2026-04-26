package models

import (
	"time"
)

// DocumentStatus represents the status of a document
type DocumentStatus string

const (
	DocumentStatusDraft     DocumentStatus = "draft"
	DocumentStatusReviewing DocumentStatus = "reviewing"
	DocumentStatusPublished DocumentStatus = "published"
	DocumentStatusArchived  DocumentStatus = "archived"
)

// Document represents the metadata of a document stored in PostgreSQL
type Document struct {
	ID             string         `json:"id" db:"id"`
	ProjectID      string         `json:"project_id" db:"project_id"`
	Title          string         `json:"title" db:"title"`
	Abstract       string         `json:"abstract" db:"abstract"`
	AuthorID       string         `json:"author_id" db:"author_id"`
	Status         DocumentStatus `json:"status" db:"status"`
	WordCount      int            `json:"word_count" db:"word_count"`
	CitationCount  int            `json:"citation_count" db:"citation_count"`
	LastEditedBy   *string        `json:"last_edited_by,omitempty" db:"last_edited_by"`
	LastEditedAt   *time.Time     `json:"last_edited_at,omitempty" db:"last_edited_at"`
	MongoContentID *string        `json:"mongo_content_id,omitempty" db:"mongo_content_id"`
	Format         string         `json:"format" db:"format"`
	IsDeleted      bool           `json:"is_deleted" db:"is_deleted"`
	DeletedAt      *time.Time     `json:"deleted_at,omitempty" db:"deleted_at"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}

// DocumentContent represents the full content of a document stored in MongoDB
type DocumentContent struct {
	DocumentID string                 `json:"document_id" bson:"document_id"`
	ProjectID  string                 `json:"project_id" bson:"project_id"`
	Title      string                 `json:"title" bson:"title"`
	Content    map[string]interface{} `json:"content" bson:"content"`
	Metadata   ContentMetadata        `json:"metadata" bson:"metadata"`
	Citations  []Citation             `json:"citations,omitempty" bson:"citations,omitempty"`
	CreatedAt  time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" bson:"updated_at"`
}

// ContentMetadata contains metadata about document content
type ContentMetadata struct {
	WordCount    int    `json:"word_count" bson:"word_count"`
	CitationCount int   `json:"citation_count" bson:"citation_count"`
	LastEditedBy string `json:"last_edited_by" bson:"last_edited_by"`
	Format       string `json:"format" bson:"format"`
}

// Citation represents a citation within a document
type Citation struct {
	ReferenceID string `json:"reference_id" bson:"reference_id"`
	CitationKey string `json:"citation_key" bson:"citation_key"`
	Section     string `json:"section" bson:"section"`
	Paragraph   int    `json:"paragraph" bson:"paragraph"`
	Offset      int    `json:"offset" bson:"offset"`
}

// DocumentVersion represents a version snapshot of a document
type DocumentVersion struct {
	ID             string    `json:"id" db:"id"`
	DocumentID     string    `json:"document_id" db:"document_id"`
	VersionNumber  int       `json:"version_number" db:"version_number"`
	Title          string    `json:"title" db:"title"`
	WordCount      int       `json:"word_count" db:"word_count"`
	ChangeSummary  string    `json:"change_summary" db:"change_summary"`
	MongoSnapshotID *string  `json:"mongo_snapshot_id,omitempty" db:"mongo_snapshot_id"`
	CreatedBy      string    `json:"created_by" db:"created_by"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// DocumentCollaborator represents a collaborator on a document
type DocumentCollaborator struct {
	ID         string    `json:"id" db:"id"`
	DocumentID string    `json:"document_id" db:"document_id"`
	UserID     string    `json:"user_id" db:"user_id"`
	Role       string    `json:"role" db:"role"`
	JoinedAt   time.Time `json:"joined_at" db:"joined_at"`
}

// CreateDocumentRequest represents the request to create a document
type CreateDocumentRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	Title     string `json:"title" binding:"required,max=500"`
	Abstract  string `json:"abstract"`
	Format    string `json:"format" default:"markdown"`
}

// UpdateDocumentRequest represents the request to update a document
type UpdateDocumentRequest struct {
	Title    string `json:"title,omitempty"`
	Abstract string `json:"abstract,omitempty"`
	Status   string `json:"status,omitempty"`
}

// UpdateDocumentContentRequest represents the request to update document content
type UpdateDocumentContentRequest struct {
	Content   map[string]interface{} `json:"content" binding:"required"`
	Metadata  *ContentMetadata       `json:"metadata,omitempty"`
	Citations []Citation             `json:"citations,omitempty"`
}

// CreateVersionRequest represents the request to create a version
type CreateVersionRequest struct {
	ChangeSummary string `json:"change_summary" binding:"required"`
}

// ListDocumentsRequest represents query parameters for listing documents
type ListDocumentsRequest struct {
	ProjectID string `form:"project_id" binding:"required"`
	Status    string `form:"status"`
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=20"`
}

// ListVersionsRequest represents query parameters for listing versions
type ListVersionsRequest struct {
	Page  int `form:"page,default=1"`
	Limit int `form:"limit,default=20"`
}

// AddCollaboratorRequest represents the request to add a collaborator
type AddCollaboratorRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" default:"editor"`
}

// DocumentResponse represents a document with its content
type DocumentResponse struct {
	*Document
	Content *DocumentContent `json:"content,omitempty"`
}
