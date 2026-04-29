package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrSubmissionNotFound = errors.New("submission not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrJournalNotFound    = errors.New("journal not found")
)

type Journal struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Publisher      string    `json:"publisher"`
	ISSN           string    `json:"issn"`
	Category       string    `json:"category"`
	Subcategory    string    `json:"subcategory"`
	ImpactFactor   float64   `json:"impact_factor"`
	Quartile       string    `json:"quartile"`
	OpenAccess     bool      `json:"open_access"`
	WebsiteURL     string    `json:"website_url"`
	ReviewTimeDays int       `json:"review_time_days"`
	AcceptanceRate float64   `json:"acceptance_rate"`
	Keywords       []string  `json:"keywords"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Submission struct {
	ID                  string          `json:"id"`
	ProjectID           string          `json:"project_id"`
	DocumentID          sql.NullString  `json:"-"`
	UserID              string          `json:"user_id"`
	JournalID           sql.NullInt32   `json:"-"`
	Title               string          `json:"title"`
	Abstract            sql.NullString  `json:"-"`
	Keywords            []string        `json:"keywords"`
	Status              string          `json:"status"`
	Stage               string          `json:"stage"`
	ManuscriptURL       sql.NullString  `json:"-"`
	SupplementaryFiles  []byte          `json:"-"`
	SubmissionDate      sql.NullTime    `json:"-"`
	LastStatusChange    time.Time       `json:"last_status_change"`
	Notes               sql.NullString  `json:"-"`
	RecommendationScore sql.NullFloat64 `json:"-"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`

	Journal *Journal `json:"journal,omitempty"`
	Reviews []Review `json:"reviews,omitempty"`
	History []History `json:"history,omitempty"`
}

func (s Submission) MarshalJSON() ([]byte, error) {
	type Alias Submission
	supFiles := []map[string]interface{}{}
	if len(s.SupplementaryFiles) > 0 {
		json.Unmarshal(s.SupplementaryFiles, &supFiles)
	}
	return json.Marshal(&struct {
		Alias
		DocumentID          *string                  `json:"document_id,omitempty"`
		Abstract            *string                  `json:"abstract,omitempty"`
		ManuscriptURL       *string                  `json:"manuscript_url,omitempty"`
		SupplementaryFiles  []map[string]interface{} `json:"supplementary_files,omitempty"`
		SubmissionDate      *string                  `json:"submission_date,omitempty"`
		Notes               *string                  `json:"notes,omitempty"`
		RecommendationScore *float64                 `json:"recommendation_score,omitempty"`
	}{
		Alias:               (Alias)(s),
		DocumentID:          nullStringToPtr(s.DocumentID),
		Abstract:            nullStringToPtr(s.Abstract),
		ManuscriptURL:       nullStringToPtr(s.ManuscriptURL),
		SupplementaryFiles:  supFiles,
		SubmissionDate:      nullTimeToPtr(s.SubmissionDate),
		Notes:               nullStringToPtr(s.Notes),
		RecommendationScore: nullFloat64ToPtr(s.RecommendationScore),
	})
}

type Review struct {
	ID             string     `json:"id"`
	SubmissionID   string     `json:"submission_id"`
	ReviewerName   string     `json:"reviewer_name"`
	ReviewType     string     `json:"review_type"`
	Status         string     `json:"status"`
	Content        string     `json:"content"`
	Rating         int        `json:"rating"`
	Recommendation string     `json:"recommendation"`
	ReceivedAt     time.Time  `json:"received_at"`
	RespondedAt    *time.Time `json:"responded_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type History struct {
	ID           string    `json:"id"`
	SubmissionID string    `json:"submission_id"`
	FromStatus   string    `json:"from_status"`
	ToStatus     string    `json:"to_status"`
	Note         string    `json:"note"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
}

func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func nullTimeToPtr(nt sql.NullTime) *string {
	if nt.Valid {
		s := nt.Time.Format(time.RFC3339)
		return &s
	}
	return nil
}

func nullFloat64ToPtr(nf sql.NullFloat64) *float64 {
	if nf.Valid {
		return &nf.Float64
	}
	return nil
}
