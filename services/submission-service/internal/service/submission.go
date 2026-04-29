package service

import (
	"github.com/deepwrite/submission-service/internal/models"
	"github.com/deepwrite/submission-service/internal/repository"
)

type SubmissionService struct {
	repo       *repository.SubmissionRepository
	journalRepo *repository.JournalRepository
}

func NewSubmissionService(repo *repository.SubmissionRepository, journalRepo *repository.JournalRepository) *SubmissionService {
	return &SubmissionService{repo: repo, journalRepo: journalRepo}
}

func (s *SubmissionService) Create(userID string, req *models.CreateSubmissionRequest) (*models.Submission, error) {
	submission := &models.Submission{
		ProjectID: req.ProjectID,
		UserID:    userID,
		Title:     req.Title,
		Keywords:  req.Keywords,
		Status:    "draft",
		Stage:     "preparation",
	}
	if req.Abstract != "" {
		submission.Abstract.String = req.Abstract
		submission.Abstract.Valid = true
	}
	if req.DocumentID != "" {
		submission.DocumentID.String = req.DocumentID
		submission.DocumentID.Valid = true
	}
	if err := s.repo.Create(submission); err != nil {
		return nil, err
	}
	return s.repo.GetByID(submission.ID)
}

func (s *SubmissionService) GetByID(id, userID string) (*models.Submission, error) {
	sub, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if sub.UserID != userID {
		return nil, models.ErrUnauthorized
	}
	if sub.JournalID.Valid {
		journal, _ := s.journalRepo.GetByID(int(sub.JournalID.Int32))
		sub.Journal = journal
	}
	history, _ := s.repo.GetHistory(id)
	sub.History = history
	return sub, nil
}

func (s *SubmissionService) ListByProject(projectID, userID string, page, limit int) ([]models.Submission, int, error) {
	return s.repo.ListByProject(projectID, userID, page, limit)
}

func (s *SubmissionService) ListByUser(userID string, page, limit int) ([]models.Submission, int, error) {
	return s.repo.ListByUser(userID, page, limit)
}

func (s *SubmissionService) Update(id, userID string, req *models.UpdateSubmissionRequest) (*models.Submission, error) {
	sub, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if sub.UserID != userID {
		return nil, models.ErrUnauthorized
	}
	if err := s.repo.Update(id, req); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

func (s *SubmissionService) Delete(id, userID string) error {
	sub, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if sub.UserID != userID {
		return models.ErrUnauthorized
	}
	return s.repo.Delete(id)
}

func (s *SubmissionService) SubmitToJournal(id, userID string, journalID int) (*models.Submission, error) {
	sub, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if sub.UserID != userID {
		return nil, models.ErrUnauthorized
	}

	_ = s.repo.Update(id, &models.UpdateSubmissionRequest{
		Status:    "submitted",
		Stage:     "under_review",
		JournalID: &journalID,
	})

	_ = s.repo.AddHistory(&models.History{
		SubmissionID: id,
		FromStatus:   sub.Status,
		ToStatus:     "submitted",
		Note:         "Submitted to journal",
		CreatedBy:    userID,
	})

	_ = s.repo.UpdateScore(id, 0)

	return s.repo.GetByID(id)
}

func (s *SubmissionService) UpdateStatus(id, userID, newStatus, note string) (*models.Submission, error) {
	sub, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if sub.UserID != userID {
		return nil, models.ErrUnauthorized
	}

	_ = s.repo.Update(id, &models.UpdateSubmissionRequest{Status: newStatus})

	_ = s.repo.AddHistory(&models.History{
		SubmissionID: id,
		FromStatus:   sub.Status,
		ToStatus:     newStatus,
		Note:         note,
		CreatedBy:    userID,
	})

	return s.repo.GetByID(id)
}

func (s *SubmissionService) AddHistory(id, userID string, req *models.AddHistoryRequest) error {
	sub, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if sub.UserID != userID {
		return models.ErrUnauthorized
	}
	return s.repo.AddHistory(&models.History{
		SubmissionID: id,
		FromStatus:   req.FromStatus,
		ToStatus:     req.ToStatus,
		Note:         req.Note,
		CreatedBy:    userID,
	})
}
