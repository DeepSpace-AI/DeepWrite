package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/deepwrite/project-service/internal/models"
	"github.com/deepwrite/project-service/internal/repository"
)

type ProjectService struct {
	repo *repository.ProjectRepository
}

func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) Create(ctx context.Context, ownerID string, req *models.CreateProjectRequest) (*models.Project, error) {
	project := &models.Project{
		Title:         req.Title,
		Description:   sql.NullString{String: req.Description, Valid: req.Description != ""},
		OwnerID:       ownerID,
		Status:        "active",
	}
	if req.TeamID != "" {
		project.TeamID.String = req.TeamID
		project.TeamID.Valid = true
	}
	if req.ResearchField != "" {
		project.ResearchField.String = req.ResearchField
		project.ResearchField.Valid = true
	}

	if err := s.repo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	// Add owner as project member
	if err := s.repo.AddMember(ctx, project.ID, ownerID, "owner"); err != nil {
		return nil, fmt.Errorf("add owner: %w", err)
	}

	return project, nil
}

func (s *ProjectService) GetByID(ctx context.Context, id, userID string) (*models.Project, error) {
	project, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	isOwner, _ := s.repo.IsOwner(ctx, id, userID)
	isMember, _ := s.repo.IsMember(ctx, id, userID)
	if !isOwner && !isMember {
		return nil, models.ErrUnauthorized
	}

	return project, nil
}

func (s *ProjectService) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Project, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	projects, err := s.repo.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list projects: %w", err)
	}

	total, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("count projects: %w", err)
	}

	return projects, total, nil
}

func (s *ProjectService) Update(ctx context.Context, id, userID string, req *models.UpdateProjectRequest) (*models.Project, error) {
	isOwner, err := s.repo.IsOwner(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		role, _ := s.repo.GetMemberRole(ctx, id, userID)
		if role != "editor" {
			return nil, models.ErrUnauthorized
		}
	}

	project, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		project.Title = req.Title
	}
	if req.Description != "" {
		project.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.Status != "" {
		project.Status = req.Status
	}
	if req.ResearchField != "" {
		project.ResearchField.String = req.ResearchField
		project.ResearchField.Valid = true
	}

	if err := s.repo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("update project: %w", err)
	}

	return project, nil
}

func (s *ProjectService) Delete(ctx context.Context, id, userID string) error {
	isOwner, err := s.repo.IsOwner(ctx, id, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return models.ErrUnauthorized
	}

	return s.repo.Archive(ctx, id)
}

func (s *ProjectService) Archive(ctx context.Context, id, userID string) error {
	return s.Delete(ctx, id, userID)
}

func (s *ProjectService) Unarchive(ctx context.Context, id, userID string) error {
	isOwner, err := s.repo.IsOwner(ctx, id, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return models.ErrUnauthorized
	}

	return s.repo.Unarchive(ctx, id)
}

func (s *ProjectService) AddMember(ctx context.Context, projectID, actorID string, req *models.AddProjectMemberRequest) error {
	isOwner, _ := s.repo.IsOwner(ctx, projectID, actorID)
	role, _ := s.repo.GetMemberRole(ctx, projectID, actorID)
	if !isOwner && role != "editor" {
		return models.ErrUnauthorized
	}

	return s.repo.AddMember(ctx, projectID, req.UserID, req.Role)
}

func (s *ProjectService) RemoveMember(ctx context.Context, projectID, actorID, targetUserID string) error {
	isOwner, _ := s.repo.IsOwner(ctx, projectID, actorID)
	if !isOwner {
		actorRole, _ := s.repo.GetMemberRole(ctx, projectID, actorID)
		targetRole, _ := s.repo.GetMemberRole(ctx, projectID, targetUserID)
		if actorRole != "editor" || targetRole == "owner" || targetRole == "editor" {
			if actorID != targetUserID {
				return models.ErrUnauthorized
			}
		}
	}

	isTargetOwner, _ := s.repo.IsOwner(ctx, projectID, targetUserID)
	if isTargetOwner {
		return fmt.Errorf("cannot remove project owner")
	}

	return s.repo.RemoveMember(ctx, projectID, targetUserID)
}

func (s *ProjectService) GetMembers(ctx context.Context, projectID, userID string) ([]*models.ProjectMember, error) {
	isOwner, _ := s.repo.IsOwner(ctx, projectID, userID)
	isMember, _ := s.repo.IsMember(ctx, projectID, userID)
	if !isOwner && !isMember {
		return nil, models.ErrUnauthorized
	}

	return s.repo.GetMembers(ctx, projectID)
}

func (s *ProjectService) UpdateMemberRole(ctx context.Context, projectID, actorID, targetUserID string, req *models.UpdateProjectMemberRequest) error {
	isOwner, _ := s.repo.IsOwner(ctx, projectID, actorID)
	if !isOwner {
		return models.ErrUnauthorized
	}

	isTargetOwner, _ := s.repo.IsOwner(ctx, projectID, targetUserID)
	if isTargetOwner {
		return fmt.Errorf("cannot change owner role")
	}

	return s.repo.AddMember(ctx, projectID, targetUserID, req.Role)
}
