package handler

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/models/reference"
	"github.com/deepwrite/serivces/gateway/pkg/config"
	"github.com/deepwrite/serivces/gateway/pkg/crossref"
	"github.com/deepwrite/serivces/gateway/pkg/database"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/deepwrite/serivces/gateway/pkg/worker"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReferenceHandler struct {
	crossrefClient *crossref.Client
}

func NewReferenceHandler() *ReferenceHandler {
	return &ReferenceHandler{
		crossrefClient: crossref.NewClient(
			crossref.WithMailto("support@deepwrite.work"),
		),
	}
}

func (h *ReferenceHandler) List(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	var params reference.ListReferencesParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "invalid query parameters")
		return
	}
	params.Normalize()

	query := database.DB.Model(&reference.Reference{}).Where("owner_id = ?", userID)

	if params.WorkspaceID != "" {
		query = query.Where("workspace_id = ?", params.WorkspaceID)
	}

	if params.Query != "" {
		searchTerm := "%" + strings.ToLower(params.Query) + "%"
		query = query.Where(
			"LOWER(title) LIKE ? OR LOWER(source) LIKE ? OR LOWER(abstract) LIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}

	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}

	if params.YearFrom != nil {
		query = query.Where("year >= ?", *params.YearFrom)
	}
	if params.YearTo != nil {
		query = query.Where("year <= ?", *params.YearTo)
	}

	if params.Starred != nil {
		query = query.Where("starred = ?", *params.Starred)
	}

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	if params.CollectionID != "" {
		query = query.Joins("JOIN collection_references cr ON cr.reference_id = references.id").
			Where("cr.collection_id = ?", params.CollectionID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to count references")
		return
	}

	orderClause := "created_at DESC"
	if params.SortBy != "" {
		order := "ASC"
		if params.SortOrder == "desc" {
			order = "DESC"
		}
		orderClause = params.SortBy + " " + order
	}

	offset := (params.Page - 1) * params.PageSize

	var items []reference.Reference
	if err := query.Order(orderClause).Offset(offset).Limit(params.PageSize).Find(&items).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to list references")
		return
	}

	response.Success(c, response.SuccessCode, reference.ListReferencesResponse{
		Items:    items,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	})
}

func (h *ReferenceHandler) GetByID(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Failed(c, response.ErrorBadRequestCode, "id is required")
		return
	}

	var ref reference.Reference
	if err := database.DB.Where("id = ? AND owner_id = ?", id, userID).First(&ref).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Failed(c, response.ErrorNotFoundCode, "reference not found")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "failed to get reference")
		return
	}

	response.Success(c, response.SuccessCode, ref)
}

func (h *ReferenceHandler) Create(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	var input reference.CreateReferenceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "invalid request: "+err.Error())
		return
	}

	ref := reference.Reference{
		OwnerID:     userID,
		Title:       strings.TrimSpace(input.Title),
		Source:      strings.TrimSpace(input.Source),
		DOI:         strings.TrimSpace(input.DOI),
		ISBN:        strings.TrimSpace(input.ISBN),
		URL:         strings.TrimSpace(input.URL),
		Abstract:    input.Abstract,
		Volume:      input.Volume,
		Issue:       input.Issue,
		Pages:       input.Pages,
		Publisher:   input.Publisher,
		Language:    input.Language,
		FileID:      input.FileID,
		WorkspaceID: input.WorkspaceID,
		CitationKey: strings.TrimSpace(input.CitationKey),
		BibtexRaw:   input.BibtexRaw,
		Starred:     false,
		Status:      reference.ReferenceStatusActive,
	}

	if input.Type != "" {
		ref.Type = input.Type
	} else {
		ref.Type = reference.ReferenceTypeUnknown
	}

	if input.Year != nil {
		ref.Year = input.Year
	}

	if len(input.Authors) > 0 {
		authorsJSON, _ := json.Marshal(input.Authors)
		ref.Authors = authorsJSON
	}

	if len(input.Keywords) > 0 {
		keywordsJSON, _ := json.Marshal(input.Keywords)
		ref.Keywords = keywordsJSON
	}

	if input.Metadata != nil {
		metadataJSON, _ := json.Marshal(input.Metadata)
		ref.Metadata = metadataJSON
	}

	if err := database.DB.Create(&ref).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to create reference")
		return
	}

	response.Success(c, response.SuccessCreatedCode, ref)
}

func (h *ReferenceHandler) Update(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Failed(c, response.ErrorBadRequestCode, "id is required")
		return
	}

	var ref reference.Reference
	if err := database.DB.Where("id = ? AND owner_id = ?", id, userID).First(&ref).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Failed(c, response.ErrorNotFoundCode, "reference not found")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "failed to get reference")
		return
	}

	var input reference.UpdateReferenceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "invalid request: "+err.Error())
		return
	}

	updates := make(map[string]any)

	if input.Title != nil {
		updates["title"] = strings.TrimSpace(*input.Title)
	}
	if input.Source != nil {
		updates["source"] = strings.TrimSpace(*input.Source)
	}
	if input.DOI != nil {
		updates["doi"] = strings.TrimSpace(*input.DOI)
	}
	if input.ISBN != nil {
		updates["isbn"] = strings.TrimSpace(*input.ISBN)
	}
	if input.URL != nil {
		updates["url"] = strings.TrimSpace(*input.URL)
	}
	if input.Abstract != nil {
		updates["abstract"] = *input.Abstract
	}
	if input.Volume != nil {
		updates["volume"] = *input.Volume
	}
	if input.Issue != nil {
		updates["issue"] = *input.Issue
	}
	if input.Pages != nil {
		updates["pages"] = *input.Pages
	}
	if input.Publisher != nil {
		updates["publisher"] = *input.Publisher
	}
	if input.Language != nil {
		updates["language"] = *input.Language
	}
	if input.FileID != nil {
		updates["file_id"] = input.FileID
	}
	if input.WorkspaceID != nil {
		updates["workspace_id"] = input.WorkspaceID
	}
	if input.CitationKey != nil {
		updates["citation_key"] = strings.TrimSpace(*input.CitationKey)
	}
	if input.BibtexRaw != nil {
		updates["bibtex_raw"] = *input.BibtexRaw
	}
	if input.Year != nil {
		updates["year"] = input.Year
	}
	if input.Type != nil {
		updates["type"] = *input.Type
	}
	if input.Starred != nil {
		updates["starred"] = *input.Starred
	}
	if input.Status != nil {
		updates["status"] = *input.Status
	}
	if input.Authors != nil {
		authorsJSON, _ := json.Marshal(input.Authors)
		updates["authors"] = authorsJSON
	}
	if input.Keywords != nil {
		keywordsJSON, _ := json.Marshal(input.Keywords)
		updates["keywords"] = keywordsJSON
	}
	if input.Metadata != nil {
		metadataJSON, _ := json.Marshal(input.Metadata)
		updates["metadata"] = metadataJSON
	}

	if len(updates) > 0 {
		if err := database.DB.Model(&ref).Updates(updates).Error; err != nil {
			response.Failed(c, response.ErrorUnknownCode, "failed to update reference")
			return
		}
	}

	if err := database.DB.Where("id = ?", id).First(&ref).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to get updated reference")
		return
	}

	response.Success(c, response.SuccessCode, ref)
}

func (h *ReferenceHandler) Delete(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Failed(c, response.ErrorBadRequestCode, "id is required")
		return
	}

	result := database.DB.Where("id = ? AND owner_id = ?", id, userID).Delete(&reference.Reference{})
	if result.Error != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to delete reference")
		return
	}

	if result.RowsAffected == 0 {
		response.Failed(c, response.ErrorNotFoundCode, "reference not found")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{"id": id})
}

func (h *ReferenceHandler) AttachFile(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Failed(c, response.ErrorBadRequestCode, "id is required")
		return
	}

	var input struct {
		FileID string `json:"file_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "file_id is required")
		return
	}

	result := database.DB.Model(&reference.Reference{}).
		Where("id = ? AND owner_id = ?", id, userID).
		Update("file_id", input.FileID)

	if result.Error != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to attach file")
		return
	}

	if result.RowsAffected == 0 {
		response.Failed(c, response.ErrorNotFoundCode, "reference not found")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{"file_id": input.FileID})
}

func (h *ReferenceHandler) DetachFile(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Failed(c, response.ErrorBadRequestCode, "id is required")
		return
	}

	result := database.DB.Model(&reference.Reference{}).
		Where("id = ? AND owner_id = ?", id, userID).
		Update("file_id", nil)

	if result.Error != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to detach file")
		return
	}

	if result.RowsAffected == 0 {
		response.Failed(c, response.ErrorNotFoundCode, "reference not found")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{"file_id": nil})
}

type DoiLookupResponse struct {
	Found     bool                 `json:"found"`
	Reference *reference.Reference `json:"reference,omitempty"`
	Error     string               `json:"error,omitempty"`
}

func (h *ReferenceHandler) LookupDOI(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	doi := strings.TrimSpace(c.Query("doi"))
	if doi == "" {
		response.Failed(c, response.ErrorBadRequestCode, "doi is required")
		return
	}

	ref, err := h.crossrefClient.LookupByDOI(c.Request.Context(), doi)
	if err != nil {
		response.Success(c, response.SuccessCode, DoiLookupResponse{
			Found: false,
			Error: err.Error(),
		})
		return
	}

	response.Success(c, response.SuccessCode, DoiLookupResponse{
		Found:     true,
		Reference: ref,
	})
}

type SearchResult struct {
	Items []reference.Reference `json:"items"`
	Total int                   `json:"total"`
}

func (h *ReferenceHandler) SearchCrossref(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		response.Failed(c, response.ErrorBadRequestCode, "query is required")
		return
	}

	limit := 20
	offset := 0

	items, total, err := h.crossrefClient.Search(c.Request.Context(), query, limit, offset)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to search: "+err.Error())
		return
	}

	response.Success(c, response.SuccessCode, SearchResult{
		Items: items,
		Total: total,
	})
}

type ExtractPdfFromReferenceInput struct {
	Wait bool `json:"wait"`
}

type ExtractPdfResponse struct {
	TaskID    string                   `json:"task_id,omitempty"`
	Status    string                   `json:"status,omitempty"`
	Extracted *worker.PdfExtractResult `json:"extracted,omitempty"`
	Error     string                   `json:"error,omitempty"`
}

func (h *ReferenceHandler) ExtractPdfFromReference(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	refID := strings.TrimSpace(c.Param("id"))
	if refID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "reference id is required")
		return
	}

	var ref reference.Reference
	if err := database.DB.Where("id = ? AND owner_id = ?", refID, userID).First(&ref).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Failed(c, response.ErrorNotFoundCode, "reference not found")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "failed to get reference")
		return
	}

	if ref.FileID == nil {
		response.Failed(c, response.ErrorBadRequestCode, "reference has no attached file")
		return
	}

	var input ExtractPdfFromReferenceInput
	c.ShouldBindJSON(&input)

	cfg := config.GetGlobalConfig()
	workerClient := worker.NewClient(
		worker.WithBaseURL(cfg.Worker.URL),
		worker.WithAuthToken(cfg.Worker.Token),
	)

	taskResp, err := workerClient.ExtractAndLookupPdf(c.Request.Context(), refID, "")
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to start extraction: "+err.Error())
		return
	}

	if !input.Wait {
		response.Success(c, response.SuccessCode, ExtractPdfResponse{
			TaskID: taskResp.TaskID,
		})
		return
	}

	status, err := workerClient.WaitForTask(c.Request.Context(), taskResp.TaskID, 2*time.Second, 60*time.Second)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to wait for task: "+err.Error())
		return
	}

	if status.Status != "SUCCESS" {
		response.Success(c, response.SuccessCode, ExtractPdfResponse{
			TaskID: taskResp.TaskID,
			Status: status.Status,
			Error:  status.Error,
		})
		return
	}

	var result worker.PdfExtractResult
	if err := json.Unmarshal(status.Result, &result); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to parse result")
		return
	}

	response.Success(c, response.SuccessCode, ExtractPdfResponse{
		TaskID:    taskResp.TaskID,
		Status:    status.Status,
		Extracted: &result,
	})
}

func (h *ReferenceHandler) GetExtractionTaskStatus(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	taskID := strings.TrimSpace(c.Param("task_id"))
	if taskID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "task_id is required")
		return
	}

	cfg := config.GetGlobalConfig()
	workerClient := worker.NewClient(
		worker.WithBaseURL(cfg.Worker.URL),
		worker.WithAuthToken(cfg.Worker.Token),
	)

	status, err := workerClient.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to get task status: "+err.Error())
		return
	}

	response.Success(c, response.SuccessCode, status)
}
