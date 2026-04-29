package service

import (
	"github.com/deepwrite/submission-service/internal/models"
	"github.com/deepwrite/submission-service/internal/repository"
)

type ReviewService struct {
	repo            *repository.ReviewRepository
	submissionRepo  *repository.SubmissionRepository
}

func NewReviewService(repo *repository.ReviewRepository, submissionRepo *repository.SubmissionRepository) *ReviewService {
	return &ReviewService{repo: repo, submissionRepo: submissionRepo}
}

func (s *ReviewService) Create(submissionID, userID string, req *models.CreateReviewRequest) (*models.Review, error) {
	sub, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return nil, err
	}
	if sub.UserID != userID {
		return nil, models.ErrUnauthorized
	}

	review := &models.Review{
		SubmissionID:   submissionID,
		ReviewerName:   req.ReviewerName,
		ReviewType:     req.ReviewType,
		Content:        req.Content,
		Rating:         req.Rating,
		Recommendation: req.Recommendation,
		Status:         "pending",
	}
	if review.ReviewType == "" {
		review.ReviewType = "external"
	}
	if err := s.repo.Create(review); err != nil {
		return nil, err
	}
	return s.repo.GetByID(review.ID)
}

func (s *ReviewService) ListBySubmission(submissionID, userID string) ([]models.Review, error) {
	sub, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return nil, err
	}
	if sub.UserID != userID {
		return nil, models.ErrUnauthorized
	}
	return s.repo.ListBySubmission(submissionID)
}

func (s *ReviewService) Update(reviewID, userID string, req *models.UpdateReviewRequest) (*models.Review, error) {
	review, err := s.repo.GetByID(reviewID)
	if err != nil {
		return nil, err
	}

	sub, err := s.submissionRepo.GetByID(review.SubmissionID)
	if err != nil {
		return nil, err
	}
	if sub.UserID != userID {
		return nil, models.ErrUnauthorized
	}

	if err := s.repo.Update(reviewID, req); err != nil {
		return nil, err
	}
	return s.repo.GetByID(reviewID)
}

func (s *ReviewService) Delete(reviewID, userID string) error {
	review, err := s.repo.GetByID(reviewID)
	if err != nil {
		return err
	}

	sub, err := s.submissionRepo.GetByID(review.SubmissionID)
	if err != nil {
		return err
	}
	if sub.UserID != userID {
		return models.ErrUnauthorized
	}

	return s.repo.Delete(reviewID)
}
