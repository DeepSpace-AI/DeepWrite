package reference

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ReferenceType string

const (
	ReferenceTypeArticle     ReferenceType = "article"
	ReferenceTypeBook        ReferenceType = "book"
	ReferenceTypeBookChapter ReferenceType = "book-chapter"
	ReferenceTypeConference  ReferenceType = "conference"
	ReferenceTypeThesis      ReferenceType = "thesis"
	ReferenceTypeReport      ReferenceType = "report"
	ReferenceTypeWeb         ReferenceType = "web"
	ReferenceTypePreprint    ReferenceType = "preprint"
	ReferenceTypeUnknown     ReferenceType = "unknown"
)

const (
	ReferenceStatusActive   = "active"
	ReferenceStatusReviewed = "reviewed"
	ReferenceStatusArchived = "archived"
)

const (
	ExtractStatusPending    = "pending"
	ExtractStatusProcessing = "processing"
	ExtractStatusComplete   = "complete"
	ExtractStatusFailed     = "failed"
)

type Author struct {
	Family  string `json:"family"`
	Given   string `json:"given,omitempty"`
	Suffix  string `json:"suffix,omitempty"`
	Literal string `json:"literal,omitempty"`
	ORCID   string `json:"orcid,omitempty"`
}

type Reference struct {
	ID          string  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OwnerID     string  `json:"owner_id" gorm:"type:uuid;not null;index"`
	WorkspaceID *string `json:"workspace_id,omitempty" gorm:"type:uuid;index"`

	Title    string         `json:"title" gorm:"type:varchar(1024);not null"`
	Authors  datatypes.JSON `json:"authors" gorm:"type:jsonb;default:'[]'::jsonb"`
	Year     *int           `json:"year,omitempty" gorm:"index"`
	Source   string         `json:"source,omitempty" gorm:"type:varchar(255)"`
	DOI      string         `json:"doi,omitempty" gorm:"type:varchar(255);uniqueIndex:idx_references_doi,where:deleted_at IS NULL"`
	ISBN     string         `json:"isbn,omitempty" gorm:"type:varchar(20)"`
	URL      string         `json:"url,omitempty" gorm:"type:varchar(1024)"`
	Abstract string         `json:"abstract,omitempty" gorm:"type:text"`
	Keywords datatypes.JSON `json:"keywords,omitempty" gorm:"type:jsonb"`
	Type     ReferenceType  `json:"type" gorm:"type:varchar(50);not null;default:'unknown';index"`

	Volume    string `json:"volume,omitempty" gorm:"type:varchar(50)"`
	Issue     string `json:"issue,omitempty" gorm:"type:varchar(50)"`
	Pages     string `json:"pages,omitempty" gorm:"type:varchar(50)"`
	Publisher string `json:"publisher,omitempty" gorm:"type:varchar(255)"`
	Language  string `json:"language,omitempty" gorm:"type:varchar(20)"`

	FileID      *string        `json:"file_id,omitempty" gorm:"type:uuid;index"`
	CitationKey string         `json:"citation_key,omitempty" gorm:"type:varchar(255);index"`
	BibtexRaw   string         `json:"bibtex_raw,omitempty" gorm:"type:text"`
	Metadata    datatypes.JSON `json:"metadata,omitempty" gorm:"type:jsonb"`

	ExtractStatus string  `json:"extract_status,omitempty" gorm:"type:varchar(50);default:'';index"`
	ExtractTaskID *string `json:"extract_task_id,omitempty" gorm:"type:varchar(255)"`
	ExtractError  string  `json:"extract_error,omitempty" gorm:"type:text"`

	Starred bool   `json:"starred" gorm:"not null;default:false;index"`
	Status  string `json:"status" gorm:"type:varchar(50);not null;default:'active';index"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (Reference) TableName() string {
	return "references"
}

type CreateReferenceInput struct {
	Title       string         `json:"title" binding:"required"`
	Authors     []Author       `json:"authors"`
	Year        *int           `json:"year"`
	Source      string         `json:"source"`
	DOI         string         `json:"doi"`
	ISBN        string         `json:"isbn"`
	URL         string         `json:"url"`
	Abstract    string         `json:"abstract"`
	Keywords    []string       `json:"keywords"`
	Type        ReferenceType  `json:"type"`
	Volume      string         `json:"volume"`
	Issue       string         `json:"issue"`
	Pages       string         `json:"pages"`
	Publisher   string         `json:"publisher"`
	Language    string         `json:"language"`
	FileID      *string        `json:"file_id"`
	WorkspaceID *string        `json:"workspace_id"`
	CitationKey string         `json:"citation_key"`
	BibtexRaw   string         `json:"bibtex_raw"`
	Metadata    map[string]any `json:"metadata"`
}

type UpdateReferenceInput struct {
	Title       *string        `json:"title"`
	Authors     []Author       `json:"authors"`
	Year        *int           `json:"year"`
	Source      *string        `json:"source"`
	DOI         *string        `json:"doi"`
	ISBN        *string        `json:"isbn"`
	URL         *string        `json:"url"`
	Abstract    *string        `json:"abstract"`
	Keywords    []string       `json:"keywords"`
	Type        *ReferenceType `json:"type"`
	Volume      *string        `json:"volume"`
	Issue       *string        `json:"issue"`
	Pages       *string        `json:"pages"`
	Publisher   *string        `json:"publisher"`
	Language    *string        `json:"language"`
	FileID      *string        `json:"file_id"`
	WorkspaceID *string        `json:"workspace_id"`
	CitationKey *string        `json:"citation_key"`
	BibtexRaw   *string        `json:"bibtex_raw"`
	Metadata    map[string]any `json:"metadata"`
	Starred     *bool          `json:"starred"`
	Status      *string        `json:"status"`
}

type ListReferencesParams struct {
	Query        string `form:"q"`
	Type         string `form:"type"`
	YearFrom     *int   `form:"year_from"`
	YearTo       *int   `form:"year_to"`
	CollectionID string `form:"collection_id"`
	Starred      *bool  `form:"starred"`
	Status       string `form:"status"`
	WorkspaceID  string `form:"workspace_id"`
	SortBy       string `form:"sort_by"`
	SortOrder    string `form:"sort_order"`
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
}

func (p *ListReferencesParams) Normalize() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	if p.SortBy == "" {
		p.SortBy = "created_at"
	}
	if p.SortOrder == "" {
		p.SortOrder = "desc"
	}
}

type ListReferencesResponse struct {
	Items    []Reference `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}
