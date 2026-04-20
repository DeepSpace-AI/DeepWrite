package handler

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/models/reference"
	"github.com/deepwrite/serivces/gateway/models/workspace"
	"github.com/deepwrite/serivces/gateway/pkg/config"
	"github.com/deepwrite/serivces/gateway/pkg/database"
	"github.com/deepwrite/serivces/gateway/pkg/pdf"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/deepwrite/serivces/gateway/pkg/storages"
	"github.com/deepwrite/serivces/gateway/pkg/worker"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PDFHandler struct {
	unpaywallClient *pdf.UnpaywallClient
	storageClient   StorageClient
}

type StorageClient interface {
	UploadFile(ctx context.Context, key string, data []byte, contentType string) (string, error)
	GetFileURL(ctx context.Context, key string) (string, error)
}

func NewPDFHandler() *PDFHandler {
	email := os.Getenv("UNPAYWALL_EMAIL")
	if email == "" {
		email = "support@deepwrite.work"
	}

	return &PDFHandler{
		unpaywallClient: pdf.NewUnpaywallClient(email),
	}
}

type DownloadPDFRequest struct {
	DOI string `json:"doi" binding:"required"`
}

type DownloadPDFResponse struct {
	FileID    string               `json:"file_id"`
	URL       string               `json:"url"`
	Reference *reference.Reference `json:"reference,omitempty"`
	TaskID    string               `json:"task_id,omitempty"`
}

func (h *PDFHandler) DownloadByDOI(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	var req DownloadPDFRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "invalid request: "+err.Error())
		return
	}

	doi := strings.TrimSpace(req.DOI)
	if doi == "" {
		response.Failed(c, response.ErrorBadRequestCode, "DOI is required")
		return
	}

	pdfData, pdfURL, err := h.unpaywallClient.DownloadPDFByDOI(c.Request.Context(), doi)
	if err != nil {
		response.Failed(c, response.ErrorNotFoundCode, "failed to download PDF: "+err.Error())
		return
	}

	fileID := uuid.New().String()
	key := fmt.Sprintf("references/%s/%s.pdf", userID, fileID)

	fileURL, err := h.uploadToStorage(c.Request.Context(), key, pdfData)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to upload file: "+err.Error())
		return
	}

	workspaceFile := &workspace.WorkspaceFile{
		WorkspaceID: userID,
		ObjectKey:   key,
		FileName:    fmt.Sprintf("%s.pdf", doi),
		Size:        int64(len(pdfData)),
		ContentType: "application/pdf",
		UploadedBy:  userID,
	}

	if err := database.DB.Create(workspaceFile).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to create file record")
		return
	}

	ref := &reference.Reference{
		OwnerID: userID,
		Title:   fmt.Sprintf("DOI: %s", doi),
		DOI:     doi,
		URL:     pdfURL,
		Type:    reference.ReferenceTypeArticle,
		FileID:  &workspaceFile.ID,
	}

	if err := database.DB.Create(ref).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to create reference")
		return
	}

	taskID, _ := h.triggerPDFExtraction(c.Request.Context(), workspaceFile.ID, pdfData)

	response.Success(c, response.SuccessCode, DownloadPDFResponse{
		FileID:    workspaceFile.ID,
		URL:       fileURL,
		Reference: ref,
		TaskID:    taskID,
	})
}

type UploadPDFResponse struct {
	FileID    string               `json:"file_id"`
	URL       string               `json:"url"`
	TaskID    string               `json:"task_id,omitempty"`
	Reference *reference.Reference `json:"reference,omitempty"`
}

func (h *PDFHandler) Upload(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "file is required")
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		response.Failed(c, response.ErrorBadRequestCode, "only PDF files are supported")
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType != "application/pdf" && contentType != "application/octet-stream" {
		contentType = "application/pdf"
	}

	fileData, err := io.ReadAll(file)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to read file")
		return
	}

	if len(fileData) > 100*1024*1024 {
		response.Failed(c, response.ErrorBadRequestCode, "file size exceeds 100MB limit")
		return
	}

	fileID := uuid.New().String()
	key := fmt.Sprintf("references/%s/%s.pdf", userID, fileID)

	fileURL, err := h.uploadToStorage(c.Request.Context(), key, fileData)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to upload file: "+err.Error())
		return
	}

	filename := header.Filename
	if filename == "" {
		filename = fileID + ".pdf"
	}

	workspaceFile := &workspace.WorkspaceFile{
		WorkspaceID: userID,
		ObjectKey:   key,
		FileName:    filename,
		Size:        int64(len(fileData)),
		ContentType: "application/pdf",
		UploadedBy:  userID,
	}

	if err := database.DB.Create(workspaceFile).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to create file record")
		return
	}

	refTitle := filename
	if strings.HasSuffix(strings.ToLower(refTitle), ".pdf") {
		refTitle = strings.TrimSuffix(refTitle, ".pdf")
	}

	ref := &reference.Reference{
		OwnerID:       userID,
		Title:         refTitle,
		Type:          reference.ReferenceTypeUnknown,
		FileID:        &workspaceFile.ID,
		ExtractStatus: reference.ExtractStatusPending,
		Status:        reference.ReferenceStatusActive,
	}

	if err := database.DB.Create(ref).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to create reference")
		return
	}

	taskID, extractErr := h.triggerPDFExtractionWithRef(c.Request.Context(), workspaceFile.ID, ref.ID, fileData)
	if extractErr != nil {
		fmt.Printf("WARN: failed to trigger PDF extraction: %v\n", extractErr)
		database.DB.Model(ref).Updates(map[string]any{
			"extract_status": reference.ExtractStatusFailed,
			"extract_error":  extractErr.Error(),
		})
	} else {
		database.DB.Model(ref).Updates(map[string]any{
			"extract_task_id": taskID,
			"extract_status":  reference.ExtractStatusProcessing,
		})
		ref.ExtractTaskID = &taskID
		ref.ExtractStatus = reference.ExtractStatusProcessing
	}

	response.Success(c, response.SuccessCode, UploadPDFResponse{
		FileID:    workspaceFile.ID,
		URL:       fileURL,
		TaskID:    taskID,
		Reference: ref,
	})
}

type UploadForReferenceResponse struct {
	FileID    string               `json:"file_id"`
	URL       string               `json:"url"`
	TaskID    string               `json:"task_id,omitempty"`
	Reference *reference.Reference `json:"reference,omitempty"`
}

func (h *PDFHandler) UploadForReference(c *gin.Context) {
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
		response.Failed(c, response.ErrorNotFoundCode, "reference not found")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "file is required")
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		response.Failed(c, response.ErrorBadRequestCode, "only PDF files are supported")
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType != "application/pdf" && contentType != "application/octet-stream" {
		contentType = "application/pdf"
	}

	fileData, err := io.ReadAll(file)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to read file")
		return
	}

	if len(fileData) > 100*1024*1024 {
		response.Failed(c, response.ErrorBadRequestCode, "file size exceeds 100MB limit")
		return
	}

	fileID := uuid.New().String()
	key := fmt.Sprintf("references/%s/%s.pdf", userID, fileID)

	fileURL, err := h.uploadToStorage(c.Request.Context(), key, fileData)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to upload file: "+err.Error())
		return
	}

	filename := header.Filename
	if filename == "" {
		filename = fileID + ".pdf"
	}

	workspaceFile := &workspace.WorkspaceFile{
		WorkspaceID: userID,
		ObjectKey:   key,
		FileName:    filename,
		Size:        int64(len(fileData)),
		ContentType: "application/pdf",
		UploadedBy:  userID,
	}

	if err := database.DB.Create(workspaceFile).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to create file record")
		return
	}

	if err := database.DB.Model(&ref).Update("file_id", workspaceFile.ID).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to attach file to reference")
		return
	}

	ref.FileID = &workspaceFile.ID
	database.DB.Where("id = ?", refID).First(&ref)

	taskID, _ := h.triggerPDFExtraction(c.Request.Context(), workspaceFile.ID, fileData)

	response.Success(c, response.SuccessCode, UploadForReferenceResponse{
		FileID:    workspaceFile.ID,
		URL:       fileURL,
		TaskID:    taskID,
		Reference: &ref,
	})
}

type UploadWithURLResponse struct {
	FileID string `json:"file_id"`
	URL    string `json:"url"`
	TaskID string `json:"task_id,omitempty"`
}

func (h *PDFHandler) UploadWithURL(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "invalid request: "+err.Error())
		return
	}

	pdfURL := strings.TrimSpace(req.URL)
	if pdfURL == "" {
		response.Failed(c, response.ErrorBadRequestCode, "URL is required")
		return
	}

	httpClient := &http.Client{}
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, pdfURL, nil)
	if err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "invalid URL")
		return
	}

	httpReq.Header.Set("User-Agent", "DeepWrite/1.0")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to download from URL")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		response.Failed(c, response.ErrorUnknownCode, fmt.Sprintf("download failed with status %d", resp.StatusCode))
		return
	}

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to read file data")
		return
	}

	if len(fileData) > 100*1024*1024 {
		response.Failed(c, response.ErrorBadRequestCode, "file size exceeds 100MB limit")
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/pdf") && !strings.Contains(contentType, "application/octet-stream") {
		response.Failed(c, response.ErrorBadRequestCode, "URL does not point to a PDF file")
		return
	}

	fileID := uuid.New().String()
	key := fmt.Sprintf("references/%s/%s.pdf", userID, fileID)

	fileURL, err := h.uploadToStorage(c.Request.Context(), key, fileData)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to upload file: "+err.Error())
		return
	}

	filename := filepath.Base(pdfURL)
	if filename == "" || filename == "." {
		filename = fileID + ".pdf"
	}

	workspaceFile := &workspace.WorkspaceFile{
		WorkspaceID: userID,
		ObjectKey:   key,
		FileName:    filename,
		Size:        int64(len(fileData)),
		ContentType: "application/pdf",
		UploadedBy:  userID,
	}

	if err := database.DB.Create(workspaceFile).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to create file record")
		return
	}

	taskID, _ := h.triggerPDFExtraction(c.Request.Context(), workspaceFile.ID, fileData)

	response.Success(c, response.SuccessCode, UploadWithURLResponse{
		FileID: workspaceFile.ID,
		URL:    fileURL,
		TaskID: taskID,
	})
}

func (h *PDFHandler) uploadToStorage(ctx context.Context, key string, data []byte) (string, error) {
	storage := storages.GetDefault()
	if storage != nil {
		if err := storage.PutObject(ctx, key, bytes.NewReader(data), "application/pdf", nil); err != nil {
			return "", fmt.Errorf("failed to upload to storage: %w", err)
		}
		url, err := storage.PresignGetURL(ctx, key, 24*time.Hour)
		if err != nil {
			return "", fmt.Errorf("failed to get file URL: %w", err)
		}
		return url, nil
	}

	return h.uploadToLocal(ctx, key, data)
}

func (h *PDFHandler) uploadToLocal(ctx context.Context, key string, data []byte) (string, error) {
	uploadDir := os.Getenv("LOCAL_STORAGE_PATH")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	fullPath := filepath.Join(uploadDir, key)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return fmt.Sprintf("%s/files/%s", baseURL, key), nil
}

func (h *PDFHandler) triggerPDFExtraction(ctx context.Context, fileID string, fileData []byte) (string, error) {
	cfg := config.GetGlobalConfig()

	workerClient := worker.NewClient(
		worker.WithBaseURL(cfg.Worker.URL),
		worker.WithAuthToken(cfg.Worker.Token),
	)

	contentBase64 := base64.StdEncoding.EncodeToString(fileData)

	taskResp, err := workerClient.ExtractAndLookupPdf(ctx, fileID, contentBase64)
	if err != nil {
		return "", err
	}

	return taskResp.TaskID, nil
}

func (h *PDFHandler) triggerPDFExtractionWithRef(ctx context.Context, fileID string, refID string, fileData []byte) (string, error) {
	cfg := config.GetGlobalConfig()

	workerClient := worker.NewClient(
		worker.WithBaseURL(cfg.Worker.URL),
		worker.WithAuthToken(cfg.Worker.Token),
	)

	contentBase64 := base64.StdEncoding.EncodeToString(fileData)

	taskResp, err := workerClient.ExtractAndLookupPdfWithRef(ctx, fileID, refID, contentBase64)
	if err != nil {
		return "", err
	}

	return taskResp.TaskID, nil
}

type UpdateExtractResultRequest struct {
	ReferenceID string                   `json:"reference_id" binding:"required"`
	Status      string                   `json:"status" binding:"required"`
	TaskID      string                   `json:"task_id"`
	Error       string                   `json:"error,omitempty"`
	Reference   *worker.ReferenceFromPdf `json:"reference,omitempty"`
	Metadata    map[string]any           `json:"metadata,omitempty"`
}

func (h *PDFHandler) UpdateExtractResult(c *gin.Context) {
	token := c.GetHeader("X-Internal-Token")
	cfg := config.GetGlobalConfig()
	if cfg.Worker.Token != "" && token != cfg.Worker.Token {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	var req UpdateExtractResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "invalid request: "+err.Error())
		return
	}

	var ref reference.Reference
	if err := database.DB.Where("id = ?", req.ReferenceID).First(&ref).Error; err != nil {
		response.Failed(c, response.ErrorNotFoundCode, "reference not found")
		return
	}

	updates := map[string]any{
		"extract_status": req.Status,
	}

	if req.TaskID != "" {
		updates["extract_task_id"] = req.TaskID
	}

	if req.Error != "" {
		updates["extract_error"] = req.Error
	}

	if req.Reference != nil {
		if req.Reference.Title != "" {
			updates["title"] = req.Reference.Title
		}
		if req.Reference.DOI != "" {
			updates["doi"] = req.Reference.DOI
		}
		if req.Reference.Source != "" {
			updates["source"] = req.Reference.Source
		}
		if req.Reference.Abstract != "" {
			updates["abstract"] = req.Reference.Abstract
		}
		if req.Reference.Year > 0 {
			updates["year"] = req.Reference.Year
		}
		if req.Reference.Type != "" {
			updates["type"] = req.Reference.Type
		}
		if req.Reference.Volume != "" {
			updates["volume"] = req.Reference.Volume
		}
		if req.Reference.Issue != "" {
			updates["issue"] = req.Reference.Issue
		}
		if req.Reference.Pages != "" {
			updates["pages"] = req.Reference.Pages
		}
		if req.Reference.Publisher != "" {
			updates["publisher"] = req.Reference.Publisher
		}
		if req.Reference.URL != "" {
			updates["url"] = req.Reference.URL
		}
		if len(req.Reference.Authors) > 0 {
			authorsJSON, _ := json.Marshal(req.Reference.Authors)
			updates["authors"] = authorsJSON
		}
	}

	if err := database.DB.Model(&ref).Updates(updates).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to update reference")
		return
	}

	database.DB.Where("id = ?", req.ReferenceID).First(&ref)

	response.Success(c, response.SuccessCode, ref)
}

type BatchDownloadRequest struct {
	DOIs []string `json:"dois" binding:"required"`
}

type BatchDownloadResponse struct {
	Results []BatchDownloadResult `json:"results"`
}

type BatchDownloadResult struct {
	DOI     string `json:"doi"`
	Success bool   `json:"success"`
	FileID  string `json:"file_id,omitempty"`
	URL     string `json:"url,omitempty"`
	Error   string `json:"error,omitempty"`
	TaskID  string `json:"task_id,omitempty"`
}

func (h *PDFHandler) BatchDownload(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	var req BatchDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "invalid request: "+err.Error())
		return
	}

	if len(req.DOIs) == 0 {
		response.Failed(c, response.ErrorBadRequestCode, "DOIs are required")
		return
	}

	if len(req.DOIs) > 20 {
		response.Failed(c, response.ErrorBadRequestCode, "maximum 20 DOIs per batch")
		return
	}

	results := make([]BatchDownloadResult, len(req.DOIs))

	for i, doi := range req.DOIs {
		results[i] = h.downloadSingleDOI(c.Request.Context(), userID, strings.TrimSpace(doi))
	}

	response.Success(c, response.SuccessCode, BatchDownloadResponse{Results: results})
}

func (h *PDFHandler) downloadSingleDOI(ctx context.Context, userID, doi string) BatchDownloadResult {
	result := BatchDownloadResult{DOI: doi}

	pdfData, pdfURL, err := h.unpaywallClient.DownloadPDFByDOI(ctx, doi)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	fileID := uuid.New().String()
	key := fmt.Sprintf("references/%s/%s.pdf", userID, fileID)

	fileURL, err := h.uploadToStorage(ctx, key, pdfData)
	if err != nil {
		result.Error = "failed to upload file"
		return result
	}

	workspaceFile := &workspace.WorkspaceFile{
		WorkspaceID: userID,
		ObjectKey:   key,
		FileName:    fmt.Sprintf("%s.pdf", doi),
		Size:        int64(len(pdfData)),
		ContentType: "application/pdf",
		UploadedBy:  userID,
	}

	if err := database.DB.Create(workspaceFile).Error; err != nil {
		result.Error = "failed to create file record"
		return result
	}

	ref := &reference.Reference{
		OwnerID: userID,
		Title:   fmt.Sprintf("DOI: %s", doi),
		DOI:     doi,
		URL:     pdfURL,
		Type:    reference.ReferenceTypeArticle,
		FileID:  &workspaceFile.ID,
	}

	if err := database.DB.Create(ref).Error; err != nil {
		result.Error = "failed to create reference"
		return result
	}

	taskID, _ := h.triggerPDFExtraction(ctx, workspaceFile.ID, pdfData)

	result.Success = true
	result.FileID = workspaceFile.ID
	result.URL = fileURL
	result.TaskID = taskID

	return result
}

func SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func init() {
	uploadDir := os.Getenv("LOCAL_STORAGE_PATH")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	os.MkdirAll(uploadDir, 0755)
}
