package repository

import (
	"database/sql"

	"github.com/deepwrite/submission-service/internal/models"
)

type ReviewRepository struct {
	db *sql.DB
}

func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Create(review *models.Review) error {
	return r.db.QueryRow(
		`INSERT INTO submission_reviews (submission_id, reviewer_name, review_type, content, rating, recommendation, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`,
		review.SubmissionID, review.ReviewerName, review.ReviewType, review.Content, review.Rating, review.Recommendation, review.Status,
	).Scan(&review.ID, &review.CreatedAt)
}

func (r *ReviewRepository) GetByID(id string) (*models.Review, error) {
	var review models.Review
	err := r.db.QueryRow(
		"SELECT id, submission_id, reviewer_name, review_type, status, content, rating, recommendation, received_at, responded_at, created_at FROM submission_reviews WHERE id = $1",
		id,
	).Scan(&review.ID, &review.SubmissionID, &review.ReviewerName, &review.ReviewType, &review.Status, &review.Content, &review.Rating, &review.Recommendation, &review.ReceivedAt, &review.RespondedAt, &review.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *ReviewRepository) ListBySubmission(submissionID string) ([]models.Review, error) {
	rows, err := r.db.Query(
		"SELECT id, submission_id, reviewer_name, review_type, status, content, rating, recommendation, received_at, responded_at, created_at FROM submission_reviews WHERE submission_id = $1 ORDER BY created_at DESC",
		submissionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var review models.Review
		if err := rows.Scan(&review.ID, &review.SubmissionID, &review.ReviewerName, &review.ReviewType, &review.Status, &review.Content, &review.Rating, &review.Recommendation, &review.ReceivedAt, &review.RespondedAt, &review.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func (r *ReviewRepository) Update(id string, req *models.UpdateReviewRequest) error {
	_, err := r.db.Exec(
		`UPDATE submission_reviews SET content = COALESCE(NULLIF($1, ''), content), rating = COALESCE(NULLIF($2, 0), rating),
		recommendation = COALESCE(NULLIF($3, ''), recommendation), status = COALESCE(NULLIF($4, ''), status)
		WHERE id = $5`,
		req.Content, req.Rating, req.Recommendation, req.Status, id,
	)
	return err
}

func (r *ReviewRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM submission_reviews WHERE id = $1", id)
	return err
}
