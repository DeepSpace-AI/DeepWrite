package document

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/pkg/database"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SaveVersionInput struct {
	DocumentID  string
	Title       string
	ContentJSON datatypes.JSON
	Source      string
	Snapshot    bool
	Summary     string
	CreatedBy   string
}

func Create(ctx context.Context, doc *Document, createdBy string) error {
	if doc == nil {
		return errors.New("document is nil")
	}

	now := time.Now()
	doc.normalizeForCreate(now)

	return database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(doc).Error; err != nil {
			return err
		}

		initialVersion := Version{
			DocumentID:  doc.ID,
			WorkspaceID: doc.WorkspaceID,
			Version:     doc.CurrentVersion,
			Title:       doc.Title,
			ContentJSON: doc.ContentJSON,
			Source:      VersionSourceManual,
			Snapshot:    true,
			CreatedBy:   strings.TrimSpace(createdBy),
			CreatedAt:   now,
		}

		if err := tx.Create(&initialVersion).Error; err != nil {
			return err
		}

		doc.LatestSnapshotID = &initialVersion.ID
		doc.LatestSnapshotVer = initialVersion.Version
		doc.LatestSnapshotAt = &initialVersion.CreatedAt
		doc.LastVersionedAt = &initialVersion.CreatedAt

		updates := map[string]any{
			"latest_snapshot_id":      doc.LatestSnapshotID,
			"latest_snapshot_version": doc.LatestSnapshotVer,
			"latest_snapshot_at":      doc.LatestSnapshotAt,
			"last_versioned_at":       doc.LastVersionedAt,
		}

		return tx.Model(&Document{}).Where("id = ?", doc.ID).Updates(updates).Error
	})
}

func GetByID(ctx context.Context, documentID string) (Document, error) {
	var doc Document
	err := database.DB.WithContext(ctx).
		Where("id = ?", documentID).
		First(&doc).Error
	return doc, err
}

func ListByWorkspace(ctx context.Context, workspaceID string, limit, offset int) ([]Document, error) {
	query := database.DB.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("updated_at desc")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var docs []Document
	err := query.Find(&docs).Error
	return docs, err
}

func Delete(ctx context.Context, documentID string) error {
	return database.DB.WithContext(ctx).
		Delete(&Document{}, "id = ?", documentID).Error
}

func SaveVersion(ctx context.Context, input SaveVersionInput) (Document, Version, error) {
	var (
		doc            Document
		createdVersion Version
	)

	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", input.DocumentID).
			First(&doc).Error; err != nil {
			return err
		}

		now := time.Now()
		nextVersion := doc.CurrentVersion + 1
		source := normalizeSource(input.Source)

		if len(input.ContentJSON) == 0 {
			input.ContentJSON = doc.ContentJSON
		}

		title := strings.TrimSpace(input.Title)
		if title == "" {
			title = doc.Title
		}

		createdVersion = Version{
			DocumentID:  doc.ID,
			WorkspaceID: doc.WorkspaceID,
			Version:     nextVersion,
			Title:       title,
			ContentJSON: input.ContentJSON,
			Source:      source,
			Snapshot:    input.Snapshot,
			Summary:     strings.TrimSpace(input.Summary),
			CreatedBy:   strings.TrimSpace(input.CreatedBy),
			CreatedAt:   now,
		}

		if err := tx.Create(&createdVersion).Error; err != nil {
			return err
		}

		updates := map[string]any{
			"title":             createdVersion.Title,
			"content_json":      createdVersion.ContentJSON,
			"current_version":   createdVersion.Version,
			"last_versioned_at": createdVersion.CreatedAt,
		}

		if createdVersion.Snapshot {
			updates["latest_snapshot_id"] = createdVersion.ID
			updates["latest_snapshot_version"] = createdVersion.Version
			updates["latest_snapshot_at"] = createdVersion.CreatedAt
		}

		if err := tx.Model(&Document{}).Where("id = ?", doc.ID).Updates(updates).Error; err != nil {
			return err
		}

		return tx.Where("id = ?", doc.ID).First(&doc).Error
	})

	return doc, createdVersion, err
}

func GetVersionHistory(ctx context.Context, documentID string, limit, offset int) ([]Version, error) {
	query := database.DB.WithContext(ctx).
		Where("document_id = ?", documentID).
		Order("version desc")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var versions []Version
	err := query.Find(&versions).Error
	return versions, err
}

func GetVersionByNumber(ctx context.Context, documentID string, version int64) (Version, error) {
	var v Version
	err := database.DB.WithContext(ctx).
		Where("document_id = ? AND version = ?", documentID, version).
		First(&v).Error
	return v, err
}

func RestoreFromVersion(ctx context.Context, documentID string, version int64, createdBy string) (Document, Version, error) {
	historyVersion, err := GetVersionByNumber(ctx, documentID, version)
	if err != nil {
		return Document{}, Version{}, err
	}

	return SaveVersion(ctx, SaveVersionInput{
		DocumentID:  documentID,
		Title:       historyVersion.Title,
		ContentJSON: historyVersion.ContentJSON,
		Source:      VersionSourceSnapshot,
		Snapshot:    true,
		Summary:     "restore from history version",
		CreatedBy:   createdBy,
	})
}

func normalizeSource(source string) string {
	switch strings.TrimSpace(source) {
	case VersionSourceAutosave:
		return VersionSourceAutosave
	case VersionSourceSnapshot:
		return VersionSourceSnapshot
	default:
		return VersionSourceManual
	}
}

func (d *Document) normalizeForCreate(now time.Time) {
	if len(d.ContentJSON) == 0 {
		d.ContentJSON = datatypes.JSON([]byte(`{"type":"doc","content":[]}`))
	}

	d.TiptapSchema = strings.TrimSpace(d.TiptapSchema)
	if d.TiptapSchema == "" {
		d.TiptapSchema = "doc"
	}

	d.TiptapSchemaVer = strings.TrimSpace(d.TiptapSchemaVer)
	if d.TiptapSchemaVer == "" {
		d.TiptapSchemaVer = "v1"
	}

	if d.CurrentVersion <= 0 {
		d.CurrentVersion = 1
	}

	d.LatestSnapshotVer = 0
	d.LatestSnapshotID = nil
	d.LatestSnapshotAt = nil
	d.LastVersionedAt = &now
}
