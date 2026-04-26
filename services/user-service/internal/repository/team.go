package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/deepwrite/user-service/internal/models"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(ctx context.Context, team *models.Team) error {
	query := `
		INSERT INTO teams (name, description, owner_id, max_members)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		team.Name,
		sql.NullString{String: team.Description, Valid: team.Description != ""},
		team.OwnerID, team.MaxMembers,
	).Scan(&team.ID, &team.CreatedAt, &team.UpdatedAt)
}

func (r *TeamRepository) GetByID(ctx context.Context, id string) (*models.Team, error) {
	team := &models.Team{}
	query := `
		SELECT id, name, description, owner_id, max_members, created_at, updated_at
		FROM teams WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&team.ID, &team.Name, &team.Description, &team.OwnerID,
		&team.MaxMembers, &team.CreatedAt, &team.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrTeamNotFound
	}
	return team, err
}

func (r *TeamRepository) ListByUser(ctx context.Context, userID string) ([]*models.Team, error) {
	query := `
		SELECT t.id, t.name, t.description, t.owner_id, t.max_members, t.created_at, t.updated_at
		FROM teams t
		LEFT JOIN team_members tm ON t.id = tm.team_id
		WHERE t.owner_id = $1 OR tm.user_id = $1
		GROUP BY t.id
		ORDER BY t.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*models.Team
	for rows.Next() {
		team := &models.Team{}
		if err := rows.Scan(
			&team.ID, &team.Name, &team.Description, &team.OwnerID,
			&team.MaxMembers, &team.CreatedAt, &team.UpdatedAt,
		); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	return teams, rows.Err()
}

func (r *TeamRepository) Update(ctx context.Context, team *models.Team) error {
	query := `
		UPDATE teams SET name = $2, description = $3, max_members = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		team.ID, team.Name,
		sql.NullString{String: team.Description, Valid: team.Description != ""},
		team.MaxMembers,
	).Scan(&team.UpdatedAt)
}

func (r *TeamRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM teams WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *TeamRepository) AddMember(ctx context.Context, teamID, userID, role string) error {
	query := `
		INSERT INTO team_members (team_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (team_id, user_id) DO UPDATE SET role = $3
	`
	_, err := r.db.ExecContext(ctx, query, teamID, userID, role)
	return err
}

func (r *TeamRepository) RemoveMember(ctx context.Context, teamID, userID string) error {
	query := `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, teamID, userID)
	return err
}

func (r *TeamRepository) GetMembers(ctx context.Context, teamID string) ([]*models.TeamMember, error) {
	query := `
		SELECT id, team_id, user_id, role, joined_at
		FROM team_members
		WHERE team_id = $1
		ORDER BY joined_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.TeamMember
	for rows.Next() {
		m := &models.TeamMember{}
		if err := rows.Scan(&m.ID, &m.TeamID, &m.UserID, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *TeamRepository) IsMember(ctx context.Context, teamID, userID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM team_members WHERE team_id = $1 AND user_id = $2`
	err := r.db.QueryRowContext(ctx, query, teamID, userID).Scan(&count)
	return count > 0, err
}

func (r *TeamRepository) GetMemberRole(ctx context.Context, teamID, userID string) (string, error) {
	var role string
	query := `SELECT role FROM team_members WHERE team_id = $1 AND user_id = $2`
	err := r.db.QueryRowContext(ctx, query, teamID, userID).Scan(&role)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("not a member")
	}
	return role, err
}

func (r *TeamRepository) IsOwner(ctx context.Context, teamID, userID string) (bool, error) {
	var ownerID string
	query := `SELECT owner_id FROM teams WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, teamID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		return false, models.ErrTeamNotFound
	}
	return ownerID == userID, err
}

func (r *TeamRepository) GetMemberCount(ctx context.Context, teamID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM team_members WHERE team_id = $1`
	err := r.db.QueryRowContext(ctx, query, teamID).Scan(&count)
	return count, err
}
