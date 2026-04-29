package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/deepwrite/submission-service/internal/models"
)

type SubmissionRepository struct {
	db *sql.DB
}

func NewSubmissionRepository(db *sql.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

func (r *SubmissionRepository) Create(s *models.Submission) error {
	return r.db.QueryRow(
		`INSERT INTO submissions (project_id, document_id, user_id, title, abstract, keywords, status, stage, notes)
		VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`,
		s.ProjectID, s.DocumentID.String, s.UserID, s.Title, s.Abstract.String, s.Keywords, s.Status, s.Stage, s.Notes.String,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *SubmissionRepository) GetByID(id string) (*models.Submission, error) {
	var s models.Submission
	err := r.db.QueryRow(
		`SELECT id, project_id, document_id, user_id, journal_id, title, abstract, keywords, status, stage,
		manuscript_url, supplementary_files, submission_date, last_status_change, notes, recommendation_score, created_at, updated_at
		FROM submissions WHERE id = $1`, id,
	).Scan(&s.ID, &s.ProjectID, &s.DocumentID, &s.UserID, &s.JournalID, &s.Title, &s.Abstract, &s.Keywords, &s.Status, &s.Stage,
		&s.ManuscriptURL, &s.SupplementaryFiles, &s.SubmissionDate, &s.LastStatusChange, &s.Notes, &s.RecommendationScore, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, models.ErrSubmissionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SubmissionRepository) ListByProject(projectID string, userID string, page, limit int) ([]models.Submission, int, error) {
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM submissions WHERE project_id = $1 AND user_id = $2", projectID, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(
		`SELECT id, project_id, document_id, user_id, journal_id, title, abstract, keywords, status, stage,
		manuscript_url, supplementary_files, submission_date, last_status_change, notes, recommendation_score, created_at, updated_at
		FROM submissions WHERE project_id = $1 AND user_id = $2
		ORDER BY updated_at DESC LIMIT $3 OFFSET $4`,
		projectID, userID, limit, (page-1)*limit,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var submissions []models.Submission
	for rows.Next() {
		var s models.Submission
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.DocumentID, &s.UserID, &s.JournalID, &s.Title, &s.Abstract, &s.Keywords, &s.Status, &s.Stage,
			&s.ManuscriptURL, &s.SupplementaryFiles, &s.SubmissionDate, &s.LastStatusChange, &s.Notes, &s.RecommendationScore, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		submissions = append(submissions, s)
	}
	return submissions, total, nil
}

func (r *SubmissionRepository) ListByUser(userID string, page, limit int) ([]models.Submission, int, error) {
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM submissions WHERE user_id = $1", userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(
		`SELECT id, project_id, document_id, user_id, journal_id, title, abstract, keywords, status, stage,
		manuscript_url, supplementary_files, submission_date, last_status_change, notes, recommendation_score, created_at, updated_at
		FROM submissions WHERE user_id = $1
		ORDER BY updated_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, (page-1)*limit,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var submissions []models.Submission
	for rows.Next() {
		var s models.Submission
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.DocumentID, &s.UserID, &s.JournalID, &s.Title, &s.Abstract, &s.Keywords, &s.Status, &s.Stage,
			&s.ManuscriptURL, &s.SupplementaryFiles, &s.SubmissionDate, &s.LastStatusChange, &s.Notes, &s.RecommendationScore, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		submissions = append(submissions, s)
	}
	return submissions, total, nil
}

func (r *SubmissionRepository) Update(id string, req *models.UpdateSubmissionRequest) error {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Title != "" {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, req.Title)
		argIdx++
	}
	if req.Abstract != "" {
		setClauses = append(setClauses, fmt.Sprintf("abstract = $%d", argIdx))
		args = append(args, req.Abstract)
		argIdx++
	}
	if req.Keywords != nil {
		setClauses = append(setClauses, fmt.Sprintf("keywords = $%d", argIdx))
		args = append(args, req.Keywords)
		argIdx++
	}
	if req.Status != "" {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, req.Status)
		argIdx++
	}
	if req.Stage != "" {
		setClauses = append(setClauses, fmt.Sprintf("stage = $%d", argIdx))
		args = append(args, req.Stage)
		argIdx++
	}
	if req.JournalID != nil {
		setClauses = append(setClauses, fmt.Sprintf("journal_id = $%d", argIdx))
		args = append(args, *req.JournalID)
		argIdx++
	}
	if req.Notes != "" {
		setClauses = append(setClauses, fmt.Sprintf("notes = $%d", argIdx))
		args = append(args, req.Notes)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP", "last_status_change = CURRENT_TIMESTAMP")
	query := fmt.Sprintf("UPDATE submissions SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
	args = append(args, id)

	_, err := r.db.Exec(query, args...)
	return err
}

func (r *SubmissionRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM submissions WHERE id = $1", id)
	return err
}

func (r *SubmissionRepository) AddHistory(h *models.History) error {
	return r.db.QueryRow(
		`INSERT INTO submission_history (submission_id, from_status, to_status, note, created_by)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		h.SubmissionID, h.FromStatus, h.ToStatus, h.Note, h.CreatedBy,
	).Scan(&h.ID, &h.CreatedAt)
}

func (r *SubmissionRepository) GetHistory(submissionID string) ([]models.History, error) {
	rows, err := r.db.Query(
		"SELECT id, submission_id, from_status, to_status, note, created_by, created_at FROM submission_history WHERE submission_id = $1 ORDER BY created_at DESC",
		submissionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.History
	for rows.Next() {
		var h models.History
		if err := rows.Scan(&h.ID, &h.SubmissionID, &h.FromStatus, &h.ToStatus, &h.Note, &h.CreatedBy, &h.CreatedAt); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}

func (r *SubmissionRepository) UpdateScore(id string, score float64) error {
	_, err := r.db.Exec("UPDATE submissions SET recommendation_score = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", score, id)
	return err
}
