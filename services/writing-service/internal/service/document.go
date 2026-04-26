package service

import (
	"context"
	"fmt"
	"time"

	"github.com/deepwrite/writing-service/internal/models"
	"github.com/deepwrite/writing-service/internal/repository"
)

type DocumentService struct {
	docRepo    repository.DocumentRepository
	contentRepo repository.ContentRepository
	versionRepo repository.VersionRepository
	snapshotRepo repository.SnapshotRepository
	collabRepo  repository.CollaboratorRepository
}

func NewDocumentService(
	docRepo repository.DocumentRepository,
	contentRepo repository.ContentRepository,
	versionRepo repository.VersionRepository,
	snapshotRepo repository.SnapshotRepository,
	collabRepo repository.CollaboratorRepository,
) *DocumentService {
	return &DocumentService{
		docRepo:      docRepo,
		contentRepo:  contentRepo,
		versionRepo:  versionRepo,
		snapshotRepo: snapshotRepo,
		collabRepo:   collabRepo,
	}
}

func (s *DocumentService) CreateDocument(ctx context.Context, req *models.CreateDocumentRequest, authorID string) (*models.DocumentResponse, error) {
	doc := &models.Document{
		ProjectID: req.ProjectID,
		Title:     req.Title,
		Abstract:  req.Abstract,
		AuthorID:  authorID,
		Status:    models.DocumentStatusDraft,
		Format:    req.Format,
	}
	if doc.Format == "" {
		doc.Format = "markdown"
	}

	if err := s.docRepo.Create(ctx, doc); err != nil {
		return nil, err
	}

	content := &models.DocumentContent{
		DocumentID: doc.ID,
		ProjectID:  doc.ProjectID,
		Title:      doc.Title,
		Content: map[string]interface{}{
			"abstract":     "",
			"introduction": "",
			"methods":      "",
			"results":      "",
			"discussion":   "",
			"conclusion":   "",
		},
		Metadata: models.ContentMetadata{
			WordCount:    0,
			CitationCount: 0,
			LastEditedBy: authorID,
			Format:       doc.Format,
		},
	}
	if err := s.contentRepo.Create(ctx, content); err != nil {
		return nil, err
	}

	if err := s.collabRepo.Add(ctx, &models.DocumentCollaborator{
		DocumentID: doc.ID,
		UserID:     authorID,
		Role:       "owner",
	}); err != nil {
		return nil, err
	}

	return &models.DocumentResponse{Document: doc}, nil
}

func (s *DocumentService) GetDocument(ctx context.Context, id string, includeContent bool) (*models.DocumentResponse, error) {
	doc, err := s.docRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("document not found")
	}

	resp := &models.DocumentResponse{Document: doc}
	if includeContent {
		content, err := s.contentRepo.GetByDocumentID(ctx, id)
		if err != nil {
			return nil, err
		}
		resp.Content = content
	}

	return resp, nil
}

func (s *DocumentService) UpdateDocument(ctx context.Context, id string, req *models.UpdateDocumentRequest, userID string) (*models.Document, error) {
	doc, err := s.docRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("document not found")
	}

	if req.Title != "" {
		doc.Title = req.Title
	}
	if req.Abstract != "" {
		doc.Abstract = req.Abstract
	}
	if req.Status != "" {
		doc.Status = models.DocumentStatus(req.Status)
	}

	now := time.Now()
	doc.LastEditedBy = &userID
	doc.LastEditedAt = &now

	if err := s.docRepo.Update(ctx, doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *DocumentService) UpdateContent(ctx context.Context, id string, req *models.UpdateDocumentContentRequest, userID string) (*models.DocumentContent, error) {
	doc, err := s.docRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("document not found")
	}

	content := &models.DocumentContent{
		Title:     doc.Title,
		Content:   req.Content,
		Citations: req.Citations,
	}
	if req.Metadata != nil {
		content.Metadata = *req.Metadata
	} else {
		content.Metadata = models.ContentMetadata{
			LastEditedBy: userID,
			Format:       doc.Format,
		}
	}
	content.Metadata.LastEditedBy = userID

	if err := s.contentRepo.Update(ctx, id, content); err != nil {
		return nil, err
	}

	if req.Metadata != nil {
		_ = s.docRepo.IncrementWordCount(ctx, id, req.Metadata.WordCount-doc.WordCount)
	}

	now := time.Now()
	doc.LastEditedBy = &userID
	doc.LastEditedAt = &now
	_ = s.docRepo.Update(ctx, doc)

	return content, nil
}

func (s *DocumentService) DeleteDocument(ctx context.Context, id string) error {
	if err := s.docRepo.Delete(ctx, id); err != nil {
		return err
	}
	return s.contentRepo.Delete(ctx, id)
}

func (s *DocumentService) ListDocuments(ctx context.Context, req *models.ListDocumentsRequest) ([]*models.Document, int, error) {
	return s.docRepo.ListByProject(ctx, req.ProjectID, req.Status, req.Page, req.Limit)
}

func (s *DocumentService) CreateVersion(ctx context.Context, documentID string, req *models.CreateVersionRequest, userID string) (*models.DocumentVersion, error) {
	doc, err := s.docRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("document not found")
	}

	content, err := s.contentRepo.GetByDocumentID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	snapshotID, err := s.snapshotRepo.Create(ctx, documentID, content)
	if err != nil {
		return nil, err
	}

	latestVersion, err := s.versionRepo.GetLatestVersionNumber(ctx, documentID)
	if err != nil {
		return nil, err
	}

	version := &models.DocumentVersion{
		DocumentID:      documentID,
		VersionNumber:   latestVersion + 1,
		Title:           doc.Title,
		WordCount:       doc.WordCount,
		ChangeSummary:   req.ChangeSummary,
		MongoSnapshotID: &snapshotID,
		CreatedBy:       userID,
	}

	if err := s.versionRepo.Create(ctx, version); err != nil {
		return nil, err
	}

	return version, nil
}

func (s *DocumentService) ListVersions(ctx context.Context, documentID string, req *models.ListVersionsRequest) ([]*models.DocumentVersion, int, error) {
	return s.versionRepo.GetByDocumentID(ctx, documentID, req.Page, req.Limit)
}

func (s *DocumentService) AddCollaborator(ctx context.Context, documentID string, req *models.AddCollaboratorRequest) error {
	collaborator := &models.DocumentCollaborator{
		DocumentID: documentID,
		UserID:     req.UserID,
		Role:       req.Role,
	}
	if collaborator.Role == "" {
		collaborator.Role = "editor"
	}
	return s.collabRepo.Add(ctx, collaborator)
}

func (s *DocumentService) RemoveCollaborator(ctx context.Context, documentID, userID string) error {
	return s.collabRepo.Remove(ctx, documentID, userID)
}

func (s *DocumentService) GetCollaborators(ctx context.Context, documentID string) ([]*models.DocumentCollaborator, error) {
	return s.collabRepo.GetByDocumentID(ctx, documentID)
}
