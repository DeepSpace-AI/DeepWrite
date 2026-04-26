package service

import (
	"context"
	"testing"

	"github.com/deepwrite/user-service/internal/models"
)

type mockTeamRepo struct {
	teams       map[string]*models.Team
	members     map[string][]*models.TeamMember
	ownerMap    map[string]string
}

func newMockTeamRepo() *mockTeamRepo {
	return &mockTeamRepo{
		teams:    make(map[string]*models.Team),
		members:  make(map[string][]*models.TeamMember),
		ownerMap: make(map[string]string),
	}
}

func (r *mockTeamRepo) Create(_ context.Context, team *models.Team) error {
	r.teams[team.ID] = team
	r.ownerMap[team.ID] = team.OwnerID
	return nil
}

func (r *mockTeamRepo) GetByID(_ context.Context, id string) (*models.Team, error) {
	if t, ok := r.teams[id]; ok {
		return t, nil
	}
	return nil, models.ErrTeamNotFound
}

func (r *mockTeamRepo) ListByUser(_ context.Context, userID string) ([]*models.Team, error) {
	var result []*models.Team
	for _, t := range r.teams {
		if t.OwnerID == userID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (r *mockTeamRepo) Update(_ context.Context, team *models.Team) error {
	r.teams[team.ID] = team
	return nil
}

func (r *mockTeamRepo) Delete(_ context.Context, id string) error {
	delete(r.teams, id)
	delete(r.members, id)
	delete(r.ownerMap, id)
	return nil
}

func (r *mockTeamRepo) AddMember(_ context.Context, teamID, userID, role string) error {
	r.members[teamID] = append(r.members[teamID], &models.TeamMember{
		TeamID: teamID,
		UserID: userID,
		Role:   role,
	})
	return nil
}

func (r *mockTeamRepo) RemoveMember(_ context.Context, teamID, userID string) error {
	if members, ok := r.members[teamID]; ok {
		var filtered []*models.TeamMember
		for _, m := range members {
			if m.UserID != userID {
				filtered = append(filtered, m)
			}
		}
		r.members[teamID] = filtered
	}
	return nil
}

func (r *mockTeamRepo) GetMembers(_ context.Context, teamID string) ([]*models.TeamMember, error) {
	return r.members[teamID], nil
}

func (r *mockTeamRepo) IsMember(_ context.Context, teamID, userID string) (bool, error) {
	for _, m := range r.members[teamID] {
		if m.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (r *mockTeamRepo) GetMemberRole(_ context.Context, teamID, userID string) (string, error) {
	for _, m := range r.members[teamID] {
		if m.UserID == userID {
			return m.Role, nil
		}
	}
	return "", nil
}

func (r *mockTeamRepo) IsOwner(_ context.Context, teamID, userID string) (bool, error) {
	return r.ownerMap[teamID] == userID, nil
}

func (r *mockTeamRepo) GetMemberCount(_ context.Context, teamID string) (int, error) {
	return len(r.members[teamID]), nil
}

func TestTeamService_Create(t *testing.T) {
	teamRepo := newMockTeamRepo()
	userRepo := newMockUserRepo()
	svc := NewTeamService(teamRepo, userRepo)

	req := &models.CreateTeamRequest{
		Name:       "Research Team",
		Description: "A team for research",
		MaxMembers: 10,
	}

	team, err := svc.Create(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("Create team failed: %v", err)
	}
	if team.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, team.Name)
	}
	if team.OwnerID != "user-1" {
		t.Errorf("Expected owner user-1, got %s", team.OwnerID)
	}
}

func TestTeamService_GetByID(t *testing.T) {
	teamRepo := newMockTeamRepo()
	userRepo := newMockUserRepo()
	svc := NewTeamService(teamRepo, userRepo)

	req := &models.CreateTeamRequest{Name: "Test Team"}
	team, _ := svc.Create(context.Background(), "user-1", req)

	result, err := svc.GetByID(context.Background(), team.ID, "user-1")
	if err != nil {
		t.Fatalf("Get team failed: %v", err)
	}
	if result.Name != "Test Team" {
		t.Errorf("Expected name Test Team, got %s", result.Name)
	}
}

func TestTeamService_GetByID_Unauthorized(t *testing.T) {
	teamRepo := newMockTeamRepo()
	userRepo := newMockUserRepo()
	svc := NewTeamService(teamRepo, userRepo)

	req := &models.CreateTeamRequest{Name: "Test Team"}
	team, _ := svc.Create(context.Background(), "user-1", req)

	_, err := svc.GetByID(context.Background(), team.ID, "user-2")
	if err != models.ErrUnauthorized {
		t.Errorf("Expected ErrUnauthorized, got %v", err)
	}
}

func TestTeamService_Delete(t *testing.T) {
	teamRepo := newMockTeamRepo()
	userRepo := newMockUserRepo()
	svc := NewTeamService(teamRepo, userRepo)

	req := &models.CreateTeamRequest{Name: "Test Team"}
	team, _ := svc.Create(context.Background(), "user-1", req)

	err := svc.Delete(context.Background(), team.ID, "user-1")
	if err != nil {
		t.Fatalf("Delete team failed: %v", err)
	}

	_, err = svc.GetByID(context.Background(), team.ID, "user-1")
	if err != models.ErrTeamNotFound {
		t.Errorf("Expected ErrTeamNotFound after delete, got %v", err)
	}
}

func TestTeamService_AddMember(t *testing.T) {
	teamRepo := newMockTeamRepo()
	userRepo := newMockUserRepo()
	svc := NewTeamService(teamRepo, userRepo)

	req := &models.CreateTeamRequest{Name: "Test Team", MaxMembers: 5}
	team, _ := svc.Create(context.Background(), "user-1", req)

	addReq := &models.AddTeamMemberRequest{
		UserID: "user-2",
		Role:   "member",
	}
	err := svc.AddMember(context.Background(), team.ID, "user-1", addReq)
	if err != nil {
		t.Fatalf("Add member failed: %v", err)
	}

	members, _ := svc.GetMembers(context.Background(), team.ID, "user-1")
	if len(members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(members))
	}
}

func TestTeamService_RemoveMember(t *testing.T) {
	teamRepo := newMockTeamRepo()
	userRepo := newMockUserRepo()
	svc := NewTeamService(teamRepo, userRepo)

	req := &models.CreateTeamRequest{Name: "Test Team", MaxMembers: 5}
	team, _ := svc.Create(context.Background(), "user-1", req)

	addReq := &models.AddTeamMemberRequest{UserID: "user-2", Role: "member"}
	_ = svc.AddMember(context.Background(), team.ID, "user-1", addReq)

	err := svc.RemoveMember(context.Background(), team.ID, "user-1", "user-2")
	if err != nil {
		t.Fatalf("Remove member failed: %v", err)
	}

	members, _ := svc.GetMembers(context.Background(), team.ID, "user-1")
	if len(members) != 1 {
		t.Errorf("Expected 1 member after removal, got %d", len(members))
	}
}
