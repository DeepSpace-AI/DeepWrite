package handler

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/deepwrite/serivces/gateway/models/document"
	"github.com/deepwrite/serivces/gateway/pkg/request"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DocumentHandler struct{}

// @Summary      创建文档
// @Description  创建一个新的文档并写入初始快照版本
// @Tags         Document
// @Accept       json
// @Produce      json
// @Param        request body request.CreateDocumentRequest true "文档创建参数"
// @Success      201 {object} response.Response "创建成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /documents [post]
func (h *DocumentHandler) Create(c *gin.Context) {
	var req request.CreateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "Invalid request: "+err.Error())
		return
	}

	newDoc := document.Document{
		WorkspaceID:     strings.TrimSpace(req.WorkspaceID),
		FolderID:        strings.TrimSpace(req.FolderID),
		Title:           strings.TrimSpace(req.Title),
		TiptapSchema:    strings.TrimSpace(req.TiptapSchema),
		TiptapSchemaVer: strings.TrimSpace(req.TiptapSchemaVer),
	}

	contentJSON, err := toDatatypesJSON(req.ContentJSON)
	if err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "content_json must be a valid JSON object")
		return
	}
	newDoc.ContentJSON = contentJSON

	if err := document.Create(c.Request.Context(), &newDoc, c.GetString("user_id")); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "创建文档失败")
		return
	}

	response.Success(c, response.SuccessCreatedCode, newDoc)
}

// @Summary      获取文档详情
// @Description  根据文档ID获取文档详情
// @Tags         Document
// @Accept       json
// @Produce      json
// @Param        id path string true "文档ID"
// @Success      200 {object} response.Response "获取成功"
// @Failure      404 {object} response.Response "文档不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /documents/{id} [get]
func (h *DocumentHandler) GetByID(c *gin.Context) {
	documentID := strings.TrimSpace(c.Param("id"))
	if documentID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "document id is required")
		return
	}

	doc, err := document.GetByID(c.Request.Context(), documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "文档不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取文档失败")
		return
	}

	response.Success(c, response.SuccessCode, doc)
}

// @Summary      文档列表
// @Description  按工作区获取文档列表
// @Tags         Document
// @Accept       json
// @Produce      json
// @Param        workspace_id query string true "工作区ID"
// @Param        limit query int false "分页大小"
// @Param        offset query int false "偏移量"
// @Success      200 {object} response.Response "获取成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /documents [get]
func (h *DocumentHandler) List(c *gin.Context) {
	workspaceID := strings.TrimSpace(c.Query("workspace_id"))
	if workspaceID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace_id is required")
		return
	}

	limit, offset := request.ParsePagination(c)
	docs, err := document.ListByWorkspace(c.Request.Context(), workspaceID, limit, offset)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "获取文档列表失败")
		return
	}

	response.Success(c, response.SuccessCode, docs)
}

// @Summary      保存文档版本
// @Description  保存当前文档内容为新版本，可选择是否标记为快照
// @Tags         Document
// @Accept       json
// @Produce      json
// @Param        id path string true "文档ID"
// @Param        request body request.SaveDocumentVersionRequest true "版本保存参数"
// @Success      200 {object} response.Response "保存成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      404 {object} response.Response "文档不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /documents/{id} [put]
func (h *DocumentHandler) SaveVersion(c *gin.Context) {
	documentID := strings.TrimSpace(c.Param("id"))
	if documentID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "document id is required")
		return
	}

	var req request.SaveDocumentVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "Invalid request: "+err.Error())
		return
	}

	contentJSON, err := toDatatypesJSON(req.ContentJSON)
	if err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "content_json must be a valid JSON object")
		return
	}

	updatedDoc, newVersion, err := document.SaveVersion(c.Request.Context(), document.SaveVersionInput{
		DocumentID:  documentID,
		Title:       strings.TrimSpace(req.Title),
		ContentJSON: contentJSON,
		Source:      strings.TrimSpace(req.Source),
		Snapshot:    req.Snapshot,
		Summary:     strings.TrimSpace(req.Summary),
		CreatedBy:   c.GetString("user_id"),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "文档不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "保存版本失败")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{
		"document": updatedDoc,
		"version":  newVersion,
	})
}

// @Summary      获取文档历史版本
// @Description  获取文档历史版本列表（按版本号倒序）
// @Tags         Document
// @Accept       json
// @Produce      json
// @Param        id path string true "文档ID"
// @Param        limit query int false "分页大小"
// @Param        offset query int false "偏移量"
// @Success      200 {object} response.Response "获取成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /documents/{id}/versions [get]
func (h *DocumentHandler) VersionHistory(c *gin.Context) {
	documentID := strings.TrimSpace(c.Param("id"))
	if documentID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "document id is required")
		return
	}

	limit, offset := request.ParsePagination(c)
	versions, err := document.GetVersionHistory(c.Request.Context(), documentID, limit, offset)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "获取历史版本失败")
		return
	}

	response.Success(c, response.SuccessCode, versions)
}

// @Summary      恢复历史版本
// @Description  从历史版本恢复文档，并生成新的快照版本
// @Tags         Document
// @Accept       json
// @Produce      json
// @Param        id path string true "文档ID"
// @Param        request body request.RestoreDocumentVersionRequest true "恢复参数"
// @Success      200 {object} response.Response "恢复成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      404 {object} response.Response "版本不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /documents/{id}/restore [post]
func (h *DocumentHandler) RestoreVersion(c *gin.Context) {
	documentID := strings.TrimSpace(c.Param("id"))
	if documentID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "document id is required")
		return
	}

	var req request.RestoreDocumentVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "Invalid request: "+err.Error())
		return
	}

	updatedDoc, restoredVersion, err := document.RestoreFromVersion(c.Request.Context(), documentID, req.Version, c.GetString("user_id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "历史版本不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "恢复历史版本失败")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{
		"document": updatedDoc,
		"version":  restoredVersion,
	})
}

func toDatatypesJSON(value map[string]any) (datatypes.JSON, error) {
	if value == nil {
		return nil, nil
	}

	b, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	return datatypes.JSON(b), nil
}
