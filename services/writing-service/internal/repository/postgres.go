package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/deepwrite/writing-service/internal/models"
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *models.Document) error
	GetByID(ctx context.Context, id string) (*models.Document, error)
	Update(ctx context.Context, doc *models.Document) error
	Delete(ctx context.Context, id string) error
	ListByProject(ctx context.Context, projectID, status string, page, limit int) ([]*models.Document, int, error)
	IncrementWordCount(ctx context.Context, id string, count int) error
	IncrementCitationCount(ctx context.Context, id string, count int) error
}

type PostgresDocumentRepository struct {
	db *sql.DB
}

func NewPostgresDocumentRepository(db *sql.DB) *PostgresDocumentRepository {
	return &PostgresDocumentRepository{db: db}
}

func (r *PostgresDocumentRepository) Create(ctx context.Context, doc *models.Document) error {
	query := `
		INSERT INTO documents (project_id, title, abstract, author_id, status, format)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		doc.ProjectID, doc.Title, doc.Abstract, doc.AuthorID,
		doc.Status, doc.Format,
	).Scan(&doc.ID, &doc.CreatedAt, &doc.UpdatedAt)
}

func (r *PostgresDocumentRepository) GetByID(ctx context.Context, id string) (*models.Document, error) {
	query := `
		SELECT id, project_id, title, abstract, author_id, status, word_count, citation_count,
			last_edited_by, last_edited_at, mongo_content_id, format, is_deleted, deleted_at, created_at, updated_at
		FROM documents WHERE id = $1 AND is_deleted = false
	`
	doc := &models.Document{}
	var lastEditedBy, mongoContentID sql.NullString
	var lastEditedAt, deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&doc.ID, &doc.ProjectID, &doc.Title, &doc.Abstract, &doc.AuthorID,
		&doc.Status, &doc.WordCount, &doc.CitationCount,
		&lastEditedBy, &lastEditedAt, &mongoContentID, &doc.Format,
		&doc.IsDeleted, &deletedAt, &doc.CreatedAt, &doc.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if lastEditedBy.Valid {
		doc.LastEditedBy = &lastEditedBy.String
	}
	if lastEditedAt.Valid {
		doc.LastEditedAt = &lastEditedAt.Time
	}
	if deletedAt.Valid {
		doc.DeletedAt = &deletedAt.Time
	}
	if mongoContentID.Valid {
		doc.MongoContentID = &mongoContentID.String
	}

	return doc, nil
}

func (r *PostgresDocumentRepository) Update(ctx context.Context, doc *models.Document) error {
	query := `
		UPDATE documents
		SET title = COALESCE($2, title),
			abstract = COALESCE($3, abstract),
			status = COALESCE($4, status),
			last_edited_by = $5,
			last_edited_at = $6,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND is_deleted = false
		RETURNING updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		doc.ID, doc.Title, doc.Abstract, doc.Status,
		doc.LastEditedBy, time.Now(),
	).Scan(&doc.UpdatedAt)
}

func (r *PostgresDocumentRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE documents SET is_deleted = true, deleted_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresDocumentRepository) ListByProject(ctx context.Context, projectID, status string, page, limit int) ([]*models.Document, int, error) {
	var count int
	countQuery := `SELECT COUNT(*) FROM documents WHERE project_id = $1 AND is_deleted = false`
	countArgs := []interface{}{projectID}
	if status != "" {
		countQuery += ` AND status = $2`
		countArgs = append(countArgs, status)
	}
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&count); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, project_id, title, abstract, author_id, status, word_count, citation_count,
			last_edited_by, last_edited_at, mongo_content_id, format, created_at, updated_at
		FROM documents WHERE project_id = $1 AND is_deleted = false
	`
	args := []interface{}{projectID}
	argIndex := 2
	if status != "" {
		query += fmt.Sprintf(` AND status = $%d`, argIndex)
		args = append(args, status)
		argIndex++
	}
	query += ` ORDER BY updated_at DESC`
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argIndex, argIndex+1)
	args = append(args, limit, (page-1)*limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var docs []*models.Document
	for rows.Next() {
		doc := &models.Document{}
		var lastEditedBy, mongoContentID sql.NullString
		var lastEditedAt sql.NullTime

		err := rows.Scan(
			&doc.ID, &doc.ProjectID, &doc.Title, &doc.Abstract, &doc.AuthorID,
			&doc.Status, &doc.WordCount, &doc.CitationCount,
			&lastEditedBy, &lastEditedAt, &mongoContentID, &doc.Format,
			&doc.CreatedAt, &doc.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		if lastEditedBy.Valid {
			doc.LastEditedBy = &lastEditedBy.String
		}
		if lastEditedAt.Valid {
			doc.LastEditedAt = &lastEditedAt.Time
		}
		if mongoContentID.Valid {
			doc.MongoContentID = &mongoContentID.String
		}
		docs = append(docs, doc)
	}

	return docs, count, rows.Err()
}

func (r *PostgresDocumentRepository) IncrementWordCount(ctx context.Context, id string, count int) error {
	query := `UPDATE documents SET word_count = word_count + $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, count)
	return err
}

func (r *PostgresDocumentRepository) IncrementCitationCount(ctx context.Context, id string, count int) error {
	query := `UPDATE documents SET citation_count = citation_count + $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, count)
	return err
}
