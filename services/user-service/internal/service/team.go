package service

import (
	"context"
	"fmt"

	"github.com/deepwrite/user-service/internal/models"
)

type TeamService struct {
	teamRepo TeamRepository
	userRepo UserRepository
}

func NewTeamService(teamRepo TeamRepository, userRepo UserRepository) *TeamService {
	return &TeamService{teamRepo: teamRepo, userRepo: userRepo}
}

func (s *TeamService) Create(ctx context.Context, ownerID string, req *models.CreateTeamRequest) (*models.Team, error) {
	maxMembers := req.MaxMembers
	if maxMembers <= 0 {
		maxMembers = 5
	}

	team := &models.Team{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
		MaxMembers:  maxMembers,
	}

	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, fmt.Errorf("create team: %w", err)
	}

	// Add owner as team member with admin role
	if err := s.teamRepo.AddMember(ctx, team.ID, ownerID, "owner"); err != nil {
		return nil, fmt.Errorf("add owner: %w", err)
	}

	return team, nil
}

func (s *TeamService) GetByID(ctx context.Context, id, userID string) (*models.Team, error) {
	team, err := s.teamRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check access
	isOwner, _ := s.teamRepo.IsOwner(ctx, id, userID)
	isMember, _ := s.teamRepo.IsMember(ctx, id, userID)
	if !isOwner && !isMember {
		return nil, models.ErrUnauthorized
	}

	return team, nil
}

func (s *TeamService) ListByUser(ctx context.Context, userID string) ([]*models.Team, error) {
	return s.teamRepo.ListByUser(ctx, userID)
}

func (s *TeamService) Update(ctx context.Context, id, userID string, req *models.UpdateTeamRequest) (*models.Team, error) {
	isOwner, err := s.teamRepo.IsOwner(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, models.ErrUnauthorized
	}

	team, err := s.teamRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		team.Name = req.Name
	}
	if req.Description != "" {
		team.Description = req.Description
	}
	if req.MaxMembers > 0 {
		team.MaxMembers = req.MaxMembers
	}

	if err := s.teamRepo.Update(ctx, team); err != nil {
		return nil, fmt.Errorf("update team: %w", err)
	}

	return team, nil
}

func (s *TeamService) Delete(ctx context.Context, id, userID string) error {
	isOwner, err := s.teamRepo.IsOwner(ctx, id, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return models.ErrUnauthorized
	}

	return s.teamRepo.Delete(ctx, id)
}

func (s *TeamService) AddMember(ctx context.Context, teamID, actorID string, req *models.AddTeamMemberRequest) error {
	// Check actor is owner or admin
	isOwner, _ := s.teamRepo.IsOwner(ctx, teamID, actorID)
	role, _ := s.teamRepo.GetMemberRole(ctx, teamID, actorID)
	if !isOwner && role != "admin" {
		return models.ErrUnauthorized
	}

	// Check team capacity
	count, err := s.teamRepo.GetMemberCount(ctx, teamID)
	if err != nil {
		return fmt.Errorf("get member count: %w", err)
	}

	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return err
	}

	if team.MaxMembers > 0 && count >= team.MaxMembers {
		return fmt.Errorf("team is full")
	}

	// Check user exists
	if _, err := s.userRepo.GetByID(ctx, req.UserID); err != nil {
		return fmt.Errorf("user not found")
	}

	return s.teamRepo.AddMember(ctx, teamID, req.UserID, req.Role)
}

func (s *TeamService) RemoveMember(ctx context.Context, teamID, actorID, targetUserID string) error {
	isOwner, _ := s.teamRepo.IsOwner(ctx, teamID, actorID)
	actorRole, _ := s.teamRepo.GetMemberRole(ctx, teamID, actorID)

	// Owner can remove anyone
	// Admin can remove members and viewers
	// Members can only remove themselves
	if !isOwner {
		if actorRole == "admin" {
			targetRole, _ := s.teamRepo.GetMemberRole(ctx, teamID, targetUserID)
			if targetRole == "owner" || targetRole == "admin" {
				return models.ErrUnauthorized
			}
		} else if actorID != targetUserID {
			return models.ErrUnauthorized
		}
	}

	// Cannot remove owner
	isTargetOwner, _ := s.teamRepo.IsOwner(ctx, teamID, targetUserID)
	if isTargetOwner {
		return fmt.Errorf("cannot remove team owner")
	}

	return s.teamRepo.RemoveMember(ctx, teamID, targetUserID)
}

func (s *TeamService) GetMembers(ctx context.Context, teamID, userID string) ([]*models.TeamMember, error) {
	// Check access
	isOwner, _ := s.teamRepo.IsOwner(ctx, teamID, userID)
	isMember, _ := s.teamRepo.IsMember(ctx, teamID, userID)
	if !isOwner && !isMember {
		return nil, models.ErrUnauthorized
	}

	return s.teamRepo.GetMembers(ctx, teamID)
}

func (s *TeamService) UpdateMemberRole(ctx context.Context, teamID, actorID, targetUserID string, req *models.UpdateTeamMemberRequest) error {
	isOwner, _ := s.teamRepo.IsOwner(ctx, teamID, actorID)
	if !isOwner {
		return models.ErrUnauthorized
	}

	// Cannot change owner role
	isTargetOwner, _ := s.teamRepo.IsOwner(ctx, teamID, targetUserID)
	if isTargetOwner {
		return fmt.Errorf("cannot change owner role")
	}

	return s.teamRepo.AddMember(ctx, teamID, targetUserID, req.Role)
}
