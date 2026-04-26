package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/deepwrite/project-service/internal/models"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, project *models.Project) error {
	query := `
		INSERT INTO projects (title, description, owner_id, team_id, status, research_field)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		project.Title,
		sql.NullString{String: project.Description, Valid: project.Description != ""},
		project.OwnerID,
		sql.NullString{String: project.TeamID.String, Valid: project.TeamID.String != ""},
		project.Status,
		sql.NullString{String: project.ResearchField.String, Valid: project.ResearchField.String != ""},
	).Scan(&project.ID, &project.CreatedAt, &project.UpdatedAt)
}

func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*models.Project, error) {
	project := &models.Project{}
	query := `
		SELECT id, title, description, owner_id, team_id, status, research_field,
		       is_archived, archived_at, created_at, updated_at
		FROM projects WHERE id = $1 AND is_archived = false
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID, &project.Title, &project.Description, &project.OwnerID,
		&project.TeamID, &project.Status, &project.ResearchField,
		&project.IsArchived, &project.ArchivedAt, &project.CreatedAt, &project.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrProjectNotFound
	}
	return project, err
}

func (r *ProjectRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Project, error) {
	query := `
		SELECT DISTINCT p.id, p.title, p.description, p.owner_id, p.team_id, p.status, p.research_field,
		       p.is_archived, p.archived_at, p.created_at, p.updated_at
		FROM projects p
		LEFT JOIN project_members pm ON p.id = pm.project_id
		WHERE p.is_archived = false AND (p.owner_id = $1 OR pm.user_id = $1)
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		p := &models.Project{}
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.OwnerID,
			&p.TeamID, &p.Status, &p.ResearchField,
			&p.IsArchived, &p.ArchivedAt, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *ProjectRepository) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	query := `
		SELECT COUNT(DISTINCT p.id)
		FROM projects p
		LEFT JOIN project_members pm ON p.id = pm.project_id
		WHERE p.is_archived = false AND (p.owner_id = $1 OR pm.user_id = $1)
	`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *ProjectRepository) Update(ctx context.Context, project *models.Project) error {
	query := `
		UPDATE projects SET title = $2, description = $3, status = $4,
		    research_field = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		project.ID, project.Title,
		sql.NullString{String: project.Description, Valid: project.Description != ""},
		project.Status,
		sql.NullString{String: project.ResearchField.String, Valid: project.ResearchField.String != ""},
	).Scan(&project.UpdatedAt)
}

func (r *ProjectRepository) Archive(ctx context.Context, id string) error {
	query := `
		UPDATE projects SET is_archived = true, archived_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *ProjectRepository) Unarchive(ctx context.Context, id string) error {
	query := `
		UPDATE projects SET is_archived = false, archived_at = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *ProjectRepository) AddMember(ctx context.Context, projectID, userID, role string) error {
	query := `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (project_id, user_id) DO UPDATE SET role = $3
	`
	_, err := r.db.ExecContext(ctx, query, projectID, userID, role)
	return err
}

func (r *ProjectRepository) RemoveMember(ctx context.Context, projectID, userID string) error {
	query := `DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, projectID, userID)
	return err
}

func (r *ProjectRepository) GetMembers(ctx context.Context, projectID string) ([]*models.ProjectMember, error) {
	query := `
		SELECT id, project_id, user_id, role, joined_at
		FROM project_members
		WHERE project_id = $1
		ORDER BY joined_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.ProjectMember
	for rows.Next() {
		m := &models.ProjectMember{}
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.UserID, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *ProjectRepository) IsMember(ctx context.Context, projectID, userID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM project_members WHERE project_id = $1 AND user_id = $2`
	err := r.db.QueryRowContext(ctx, query, projectID, userID).Scan(&count)
	return count > 0, err
}

func (r *ProjectRepository) GetMemberRole(ctx context.Context, projectID, userID string) (string, error) {
	var role string
	query := `SELECT role FROM project_members WHERE project_id = $1 AND user_id = $2`
	err := r.db.QueryRowContext(ctx, query, projectID, userID).Scan(&role)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("not a member")
	}
	return role, err
}

func (r *ProjectRepository) IsOwner(ctx context.Context, projectID, userID string) (bool, error) {
	var ownerID string
	query := `SELECT owner_id FROM projects WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, projectID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		return false, models.ErrProjectNotFound
	}
	return ownerID == userID, err
}
