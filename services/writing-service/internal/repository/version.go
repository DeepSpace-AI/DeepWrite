package repository

import (
	"context"
	"database/sql"

	"github.com/deepwrite/writing-service/internal/models"
)

type VersionRepository interface {
	Create(ctx context.Context, version *models.DocumentVersion) error
	GetByDocumentID(ctx context.Context, documentID string, page, limit int) ([]*models.DocumentVersion, int, error)
	GetByID(ctx context.Context, id string) (*models.DocumentVersion, error)
	GetLatestVersionNumber(ctx context.Context, documentID string) (int, error)
}

type PostgresVersionRepository struct {
	db *sql.DB
}

func NewPostgresVersionRepository(db *sql.DB) *PostgresVersionRepository {
	return &PostgresVersionRepository{db: db}
}

func (r *PostgresVersionRepository) Create(ctx context.Context, version *models.DocumentVersion) error {
	query := `
		INSERT INTO document_versions (document_id, version_number, title, word_count, change_summary, mongo_snapshot_id, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query,
		version.DocumentID, version.VersionNumber, version.Title,
		version.WordCount, version.ChangeSummary, version.MongoSnapshotID,
		version.CreatedBy,
	).Scan(&version.ID, &version.CreatedAt)
}

func (r *PostgresVersionRepository) GetByDocumentID(ctx context.Context, documentID string, page, limit int) ([]*models.DocumentVersion, int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM document_versions WHERE document_id = $1`, documentID,
	).Scan(&count); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, document_id, version_number, title, word_count, change_summary, mongo_snapshot_id, created_by, created_at
		FROM document_versions WHERE document_id = $1
		ORDER BY version_number DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, documentID, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var versions []*models.DocumentVersion
	for rows.Next() {
		v := &models.DocumentVersion{}
		var mongoSnapshotID sql.NullString
		if err := rows.Scan(&v.ID, &v.DocumentID, &v.VersionNumber, &v.Title,
			&v.WordCount, &v.ChangeSummary, &mongoSnapshotID, &v.CreatedBy, &v.CreatedAt); err != nil {
			return nil, 0, err
		}
		if mongoSnapshotID.Valid {
			v.MongoSnapshotID = &mongoSnapshotID.String
		}
		versions = append(versions, v)
	}

	return versions, count, rows.Err()
}

func (r *PostgresVersionRepository) GetByID(ctx context.Context, id string) (*models.DocumentVersion, error) {
	query := `
		SELECT id, document_id, version_number, title, word_count, change_summary, mongo_snapshot_id, created_by, created_at
		FROM document_versions WHERE id = $1
	`
	v := &models.DocumentVersion{}
	var mongoSnapshotID sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&v.ID, &v.DocumentID, &v.VersionNumber, &v.Title,
		&v.WordCount, &v.ChangeSummary, &mongoSnapshotID, &v.CreatedBy, &v.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if mongoSnapshotID.Valid {
		v.MongoSnapshotID = &mongoSnapshotID.String
	}
	return v, err
}

func (r *PostgresVersionRepository) GetLatestVersionNumber(ctx context.Context, documentID string) (int, error) {
	var versionNumber int
	query := `SELECT COALESCE(MAX(version_number), 0) FROM document_versions WHERE document_id = $1`
	err := r.db.QueryRowContext(ctx, query, documentID).Scan(&versionNumber)
	return versionNumber, err
}

type CollaboratorRepository interface {
	Add(ctx context.Context, collaborator *models.DocumentCollaborator) error
	Remove(ctx context.Context, documentID, userID string) error
	GetByDocumentID(ctx context.Context, documentID string) ([]*models.DocumentCollaborator, error)
}

type PostgresCollaboratorRepository struct {
	db *sql.DB
}

func NewPostgresCollaboratorRepository(db *sql.DB) *PostgresCollaboratorRepository {
	return &PostgresCollaboratorRepository{db: db}
}

func (r *PostgresCollaboratorRepository) Add(ctx context.Context, collaborator *models.DocumentCollaborator) error {
	query := `
		INSERT INTO document_collaborators (document_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (document_id, user_id) DO UPDATE SET role = $3
		RETURNING id, joined_at
	`
	return r.db.QueryRowContext(ctx, query,
		collaborator.DocumentID, collaborator.UserID, collaborator.Role,
	).Scan(&collaborator.ID, &collaborator.JoinedAt)
}

func (r *PostgresCollaboratorRepository) Remove(ctx context.Context, documentID, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM document_collaborators WHERE document_id = $1 AND user_id = $2`,
		documentID, userID,
	)
	return err
}

func (r *PostgresCollaboratorRepository) GetByDocumentID(ctx context.Context, documentID string) ([]*models.DocumentCollaborator, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, document_id, user_id, role, joined_at FROM document_collaborators WHERE document_id = $1`,
		documentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collaborators []*models.DocumentCollaborator
	for rows.Next() {
		c := &models.DocumentCollaborator{}
		if err := rows.Scan(&c.ID, &c.DocumentID, &c.UserID, &c.Role, &c.JoinedAt); err != nil {
			return nil, err
		}
		collaborators = append(collaborators, c)
	}

	return collaborators, rows.Err()
}
