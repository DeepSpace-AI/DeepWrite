package models

type CreateSubmissionRequest struct {
	ProjectID  string   `json:"project_id" binding:"required"`
	DocumentID string   `json:"document_id,omitempty"`
	Title      string   `json:"title" binding:"required,max=500"`
	Abstract   string   `json:"abstract,omitempty"`
	Keywords   []string `json:"keywords,omitempty"`
}

type UpdateSubmissionRequest struct {
	Title      string   `json:"title,omitempty"`
	Abstract   string   `json:"abstract,omitempty"`
	Keywords   []string `json:"keywords,omitempty"`
	Status     string   `json:"status,omitempty"`
	Stage      string   `json:"stage,omitempty"`
	JournalID  *int     `json:"journal_id,omitempty"`
	Notes      string   `json:"notes,omitempty"`
}

type SubmitToJournalRequest struct {
	JournalID int `json:"journal_id" binding:"required"`
}

type CreateReviewRequest struct {
	ReviewerName   string `json:"reviewer_name,omitempty"`
	ReviewType     string `json:"review_type,omitempty"`
	Content        string `json:"content" binding:"required"`
	Rating         int    `json:"rating,omitempty"`
	Recommendation string `json:"recommendation,omitempty"`
}

type UpdateReviewRequest struct {
	Content        string `json:"content,omitempty"`
	Rating         int    `json:"rating,omitempty"`
	Recommendation string `json:"recommendation,omitempty"`
	Status         string `json:"status,omitempty"`
}

type JournalSearchRequest struct {
	Query    string `form:"query"`
	Category string `form:"category"`
	Quartile string `form:"quartile"`
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
}

type RecommendJournalsRequest struct {
	Title    string   `json:"title" binding:"required"`
	Abstract string   `json:"abstract,omitempty"`
	Keywords []string `json:"keywords,omitempty"`
}

type AddHistoryRequest struct {
	FromStatus string `json:"from_status,omitempty"`
	ToStatus   string `json:"to_status" binding:"required"`
	Note       string `json:"note,omitempty"`
}
