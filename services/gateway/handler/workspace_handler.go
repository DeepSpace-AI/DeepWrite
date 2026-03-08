package handler

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/models/workspace"
	"github.com/deepwrite/serivces/gateway/pkg/config"
	"github.com/deepwrite/serivces/gateway/pkg/request"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/deepwrite/serivces/gateway/pkg/storages"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WorkspaceHandler struct{}

// @Summary      工作区列表
// @Description  获取当前登录用户可访问的工作区列表
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "分页大小"
// @Param        offset query int false "偏移量"
// @Success      200 {object} response.Response "获取成功"
// @Failure      401 {object} response.Response "未授权"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces [get]
func (h *WorkspaceHandler) List(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	limit, offset := request.ParsePagination(c)
	workspaces, err := workspace.ListByUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "获取工作区列表失败")
		return
	}

	response.Success(c, response.SuccessCode, workspaces)
}

// @Summary      创建工作区
// @Description  创建一个新的工作区，并将当前用户设置为 owner
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body request.CreateWorkspaceRequest true "工作区创建参数"
// @Success      201 {object} response.Response "创建成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces [post]
func (h *WorkspaceHandler) Create(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	var req request.CreateWorkspaceRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	ws := workspace.Workspace{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Public:      req.Public,
		Status:      "active",
	}

	if err := workspace.Create(c.Request.Context(), &ws, userID); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "创建工作区失败")
		return
	}

	created, err := workspace.GetWorkSpaceByID(c.Request.Context(), ws.ID)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "获取工作区详情失败")
		return
	}

	response.Success(c, response.SuccessCreatedCode, created)
}

// @Summary      获取工作区详情
// @Description  根据工作区ID获取详情；私有工作区需具备成员权限
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "工作区ID"
// @Success      200 {object} response.Response "获取成功"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限访问"
// @Failure      404 {object} response.Response "工作区不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces/{id} [get]
func (h *WorkspaceHandler) GetByID(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	workspaceID := strings.TrimSpace(c.Param("id"))
	if workspaceID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace id is required")
		return
	}

	ws, err := workspace.GetWorkSpaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "工作区不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取工作区失败")
		return
	}

	if !ws.Public && !ws.IsViewer(userID) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限访问该工作区")
		return
	}

	response.Success(c, response.SuccessCode, ws)
}

// @Summary      更新工作区
// @Description  更新工作区基础信息（需 admin 或 owner 权限）
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "工作区ID"
// @Param        request body request.UpdateWorkspaceRequest true "工作区更新参数"
// @Success      200 {object} response.Response "更新成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限修改"
// @Failure      404 {object} response.Response "工作区不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces/{id} [put]
func (h *WorkspaceHandler) Update(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	workspaceID := strings.TrimSpace(c.Param("id"))
	if workspaceID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace id is required")
		return
	}

	var req request.UpdateWorkspaceRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	current, err := workspace.GetWorkSpaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "工作区不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取工作区失败")
		return
	}

	if !current.IsAdmin(userID) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限修改该工作区")
		return
	}

	updated, err := workspace.Update(c.Request.Context(), workspaceID, workspace.UpdateWorkspaceInput{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Public:      req.Public,
		Status:      strings.TrimSpace(req.Status),
	})
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "更新工作区失败")
		return
	}

	response.Success(c, response.SuccessCode, updated)
}

// @Summary      删除工作区
// @Description  删除指定工作区（仅 owner 可执行）
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "工作区ID"
// @Success      200 {object} response.Response "删除成功"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限删除"
// @Failure      404 {object} response.Response "工作区不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces/{id} [delete]
func (h *WorkspaceHandler) Delete(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	workspaceID := strings.TrimSpace(c.Param("id"))
	if workspaceID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace id is required")
		return
	}

	current, err := workspace.GetWorkSpaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "工作区不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取工作区失败")
		return
	}

	if !current.IsOwner(userID) {
		response.Failed(c, response.ErrorForbiddenCode, "仅工作区所有者可删除")
		return
	}

	if err := workspace.Delete(c.Request.Context(), workspaceID); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "删除工作区失败")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{"id": workspaceID})
}

// @Summary      文件夹列表
// @Description  获取工作区下的文件夹列表，可按 parent_id 过滤
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "工作区ID"
// @Param        parent_id query string false "父文件夹ID"
// @Param        limit query int false "分页大小"
// @Param        offset query int false "偏移量"
// @Success      200 {object} response.Response "获取成功"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限访问"
// @Failure      404 {object} response.Response "工作区不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces/{id}/folders [get]
func (h *WorkspaceHandler) ListFolders(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	workspaceID := strings.TrimSpace(c.Param("id"))
	if workspaceID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace id is required")
		return
	}

	ws, err := workspace.GetWorkSpaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "工作区不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取工作区失败")
		return
	}

	if !ws.Public && !ws.IsViewer(userID) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限访问该工作区")
		return
	}

	limit, offset := request.ParsePagination(c)
	parentID := strings.TrimSpace(c.Query("parent_id"))

	folders, err := workspace.ListFoldersByWorkspace(c.Request.Context(), workspaceID, parentID, limit, offset)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "获取文件夹列表失败")
		return
	}

	response.Success(c, response.SuccessCode, folders)
}

// @Summary      创建文件夹
// @Description  在工作区下创建文件夹（需 editor/admin/owner 权限）
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "工作区ID"
// @Param        request body request.CreateFolderRequest true "文件夹创建参数"
// @Success      201 {object} response.Response "创建成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限创建"
// @Failure      404 {object} response.Response "工作区不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces/{id}/folders [post]
func (h *WorkspaceHandler) CreateFolder(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	workspaceID := strings.TrimSpace(c.Param("id"))
	if workspaceID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace id is required")
		return
	}

	var req request.CreateFolderRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	ws, err := workspace.GetWorkSpaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "工作区不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取工作区失败")
		return
	}

	if !ws.IsEditor(userID) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限创建文件夹")
		return
	}

	if req.ParentID != nil {
		parentFolder, err := workspace.GetFolderByID(c.Request.Context(), *req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Failed(c, 404, "父文件夹不存在")
				return
			}
			response.Failed(c, response.ErrorUnknownCode, "获取父文件夹失败")
			return
		}
		if parentFolder.WorkspaceID != workspaceID {
			response.Failed(c, response.ErrorBadRequestCode, "parent folder does not belong to workspace")
			return
		}
	}

	folder, err := workspace.CreateFolder(c.Request.Context(), workspace.CreateFolderInput{
		WorkspaceID: workspaceID,
		ParentID:    req.ParentID,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
	})
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "创建文件夹失败")
		return
	}

	response.Success(c, response.SuccessCreatedCode, folder)
}

// @Summary      获取文件夹详情
// @Description  获取工作区下指定文件夹详情
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "工作区ID"
// @Param        folder_id path string true "文件夹ID"
// @Success      200 {object} response.Response "获取成功"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限访问"
// @Failure      404 {object} response.Response "文件夹不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces/{id}/folders/{folder_id} [get]
func (h *WorkspaceHandler) GetFolderByID(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	workspaceID := strings.TrimSpace(c.Param("id"))
	folderID := strings.TrimSpace(c.Param("folder_id"))
	if workspaceID == "" || folderID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace id and folder id are required")
		return
	}

	ws, err := workspace.GetWorkSpaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "工作区不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取工作区失败")
		return
	}

	if !ws.Public && !ws.IsViewer(userID) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限访问该工作区")
		return
	}

	folder, err := workspace.GetFolderByID(c.Request.Context(), folderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "文件夹不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取文件夹失败")
		return
	}

	if folder.WorkspaceID != workspaceID {
		response.Failed(c, 404, "文件夹不存在")
		return
	}

	response.Success(c, response.SuccessCode, folder)
}

// @Summary      更新文件夹
// @Description  更新文件夹名称和描述（需 editor/admin/owner 权限）
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "工作区ID"
// @Param        folder_id path string true "文件夹ID"
// @Param        request body request.UpdateFolderRequest true "文件夹更新参数"
// @Success      200 {object} response.Response "更新成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限更新"
// @Failure      404 {object} response.Response "文件夹不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces/{id}/folders/{folder_id} [put]
func (h *WorkspaceHandler) UpdateFolder(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	workspaceID := strings.TrimSpace(c.Param("id"))
	folderID := strings.TrimSpace(c.Param("folder_id"))
	if workspaceID == "" || folderID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace id and folder id are required")
		return
	}

	var req request.UpdateFolderRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	ws, err := workspace.GetWorkSpaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "工作区不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取工作区失败")
		return
	}

	if !ws.IsEditor(userID) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限更新文件夹")
		return
	}

	existsFolder, err := workspace.GetFolderByID(c.Request.Context(), folderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "文件夹不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取文件夹失败")
		return
	}

	if existsFolder.WorkspaceID != workspaceID {
		response.Failed(c, 404, "文件夹不存在")
		return
	}

	updated, err := workspace.UpdateFolder(c.Request.Context(), folderID, workspace.UpdateFolderInput{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
	})
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "更新文件夹失败")
		return
	}

	response.Success(c, response.SuccessCode, updated)
}

// @Summary      删除文件夹
// @Description  删除指定文件夹（需 editor/admin/owner 权限）
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "工作区ID"
// @Param        folder_id path string true "文件夹ID"
// @Success      200 {object} response.Response "删除成功"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限删除"
// @Failure      404 {object} response.Response "文件夹不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces/{id}/folders/{folder_id} [delete]
func (h *WorkspaceHandler) DeleteFolder(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	workspaceID := strings.TrimSpace(c.Param("id"))
	folderID := strings.TrimSpace(c.Param("folder_id"))
	if workspaceID == "" || folderID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace id and folder id are required")
		return
	}

	ws, err := workspace.GetWorkSpaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "工作区不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取工作区失败")
		return
	}

	if !ws.IsEditor(userID) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限删除文件夹")
		return
	}

	folder, err := workspace.GetFolderByID(c.Request.Context(), folderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "文件夹不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取文件夹失败")
		return
	}

	if folder.WorkspaceID != workspaceID {
		response.Failed(c, 404, "文件夹不存在")
		return
	}

	if err := workspace.DeleteFolder(c.Request.Context(), folderID); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "删除文件夹失败")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{"id": folderID})
}

// @Summary      工作区文件列表
// @Description  获取工作区文件列表，可按 folder_id 过滤
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "工作区ID"
// @Param        folder_id query string false "文件夹ID"
// @Param        limit query int false "分页大小"
// @Param        offset query int false "偏移量"
// @Success      200 {object} response.Response "获取成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限访问"
// @Failure      404 {object} response.Response "工作区或文件夹不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces/{id}/files [get]
func (h *WorkspaceHandler) ListFiles(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	workspaceID := strings.TrimSpace(c.Param("id"))
	if workspaceID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace id is required")
		return
	}

	ws, err := workspace.GetWorkSpaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "工作区不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取工作区失败")
		return
	}

	if !ws.Public && !ws.IsViewer(userID) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限访问该工作区")
		return
	}

	folderID := strings.TrimSpace(c.Query("folder_id"))
	if folderID != "" {
		folder, err := workspace.GetFolderByID(c.Request.Context(), folderID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Failed(c, 404, "文件夹不存在")
				return
			}
			response.Failed(c, response.ErrorUnknownCode, "获取文件夹失败")
			return
		}
		if folder.WorkspaceID != workspaceID {
			response.Failed(c, response.ErrorBadRequestCode, "folder does not belong to workspace")
			return
		}
	}

	limit, offset := request.ParsePagination(c)
	files, err := workspace.ListWorkspaceFiles(c.Request.Context(), workspaceID, folderID, limit, offset)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "获取文件列表失败")
		return
	}

	response.Success(c, response.SuccessCode, files)
}

// @Summary      获取工作区文件详情
// @Description  根据 file_id 获取工作区文件详情，并返回预览URL
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "工作区ID"
// @Param        file_id path string true "文件ID"
// @Success      200 {object} response.Response "获取成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限访问"
// @Failure      404 {object} response.Response "工作区或文件不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces/{id}/files/{file_id} [get]
func (h *WorkspaceHandler) GetFileByID(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	workspaceID := strings.TrimSpace(c.Param("id"))
	fileID := strings.TrimSpace(c.Param("file_id"))
	if workspaceID == "" || fileID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace id and file id are required")
		return
	}

	ws, err := workspace.GetWorkSpaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "工作区不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取工作区失败")
		return
	}

	if !ws.Public && !ws.IsViewer(userID) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限访问该工作区")
		return
	}

	file, err := workspace.GetWorkspaceFileByID(c.Request.Context(), fileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "文件不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取文件失败")
		return
	}

	if file.WorkspaceID != workspaceID {
		response.Failed(c, 404, "文件不存在")
		return
	}

	storageClient := storages.GetDefault()
	if storageClient == nil {
		response.Failed(c, response.ErrorUnknownCode, "storage is not configured")
		return
	}

	previewURL, _ := storageClient.PresignGetURL(c.Request.Context(), file.ObjectKey, 15*time.Minute)

	response.Success(c, response.SuccessCode, gin.H{
		"file":        file,
		"preview_url": previewURL,
	})
}

// @Summary      获取工作区文件预上传URL
// @Description  生成 AWS S3 预签名 PUT URL，前端可直传文件到对象存储
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "工作区ID"
// @Param        request body request.PresignWorkspaceUploadRequest true "预上传参数"
// @Success      201 {object} response.Response "生成成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限操作"
// @Failure      404 {object} response.Response "工作区或文件夹不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces/{id}/files [post]
func (h *WorkspaceHandler) UploadFile(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	workspaceID := strings.TrimSpace(c.Param("id"))
	if workspaceID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace id is required")
		return
	}

	ws, err := workspace.GetWorkSpaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "工作区不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取工作区失败")
		return
	}

	if !ws.IsEditor(userID) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限上传文件")
		return
	}

	var req request.PresignWorkspaceUploadRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	folderID := strings.TrimSpace(req.FolderID)
	if folderID != "" {
		folder, err := workspace.GetFolderByID(c.Request.Context(), folderID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Failed(c, 404, "文件夹不存在")
				return
			}
			response.Failed(c, response.ErrorUnknownCode, "获取文件夹失败")
			return
		}
		if folder.WorkspaceID != workspaceID {
			response.Failed(c, response.ErrorBadRequestCode, "folder does not belong to workspace")
			return
		}
	}

	storageClient := storages.GetDefault()
	if storageClient == nil {
		response.Failed(c, response.ErrorUnknownCode, "storage is not configured")
		return
	}

	objectKey := buildWorkspaceFileKey(workspaceID, folderID, req.FileName)
	contentType := strings.TrimSpace(req.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	expiresIn := 15 * time.Minute
	if req.ExpiresIn > 0 {
		expiresIn = time.Duration(req.ExpiresIn) * time.Second
	}

	uploadURL, err := storageClient.PresignPutURL(c.Request.Context(), objectKey, expiresIn, contentType)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "生成上传地址失败")
		return
	}

	previewURL, _ := storageClient.PresignGetURL(c.Request.Context(), objectKey, expiresIn)

	response.Success(c, response.SuccessCreatedCode, gin.H{
		"workspace_id":  workspaceID,
		"folder_id":     folderID,
		"filename":      strings.TrimSpace(req.FileName),
		"content_type":  contentType,
		"object_key":    objectKey,
		"upload_method": "PUT",
		"upload_url":    uploadURL,
		"preview_url":   previewURL,
	})
}

// @Summary      确认工作区文件上传完成
// @Description  校验对象已上传并保存文件元数据
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "工作区ID"
// @Param        request body request.CompleteWorkspaceUploadRequest true "上传完成参数"
// @Success      200 {object} response.Response "确认成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限操作"
// @Failure      404 {object} response.Response "工作区或文件夹不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces/{id}/files/complete [post]
func (h *WorkspaceHandler) CompleteUpload(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	workspaceID := strings.TrimSpace(c.Param("id"))
	if workspaceID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace id is required")
		return
	}

	ws, err := workspace.GetWorkSpaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "工作区不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取工作区失败")
		return
	}

	if !ws.IsEditor(userID) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限上传文件")
		return
	}

	var req request.CompleteWorkspaceUploadRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	folderID := strings.TrimSpace(req.FolderID)
	var folderPtr *string
	if folderID != "" {
		folder, err := workspace.GetFolderByID(c.Request.Context(), folderID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Failed(c, 404, "文件夹不存在")
				return
			}
			response.Failed(c, response.ErrorUnknownCode, "获取文件夹失败")
			return
		}
		if folder.WorkspaceID != workspaceID {
			response.Failed(c, response.ErrorBadRequestCode, "folder does not belong to workspace")
			return
		}
		folderPtr = &folderID
	}

	storageClient := storages.GetDefault()
	if storageClient == nil {
		response.Failed(c, response.ErrorUnknownCode, "storage is not configured")
		return
	}

	objectKey := strings.TrimSpace(req.ObjectKey)
	if objectKey == "" {
		response.Failed(c, response.ErrorBadRequestCode, "object_key is required")
		return
	}

	info, err := storageClient.HeadObject(c.Request.Context(), objectKey)
	if err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "object not found or not uploaded")
		return
	}

	file, err := workspace.UpsertWorkspaceFile(c.Request.Context(), workspace.UpsertWorkspaceFileInput{
		WorkspaceID: workspaceID,
		FolderID:    folderPtr,
		ObjectKey:   objectKey,
		FileName:    strings.TrimSpace(req.FileName),
		ContentType: info.ContentType,
		Size:        info.Size,
		ETag:        info.ETag,
		UploadedBy:  userID,
	})
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "保存文件元数据失败")
		return
	}

	previewURL, _ := storageClient.PresignGetURL(c.Request.Context(), objectKey, 15*time.Minute)

	response.Success(c, response.SuccessCode, gin.H{
		"file":        file,
		"preview_url": previewURL,
	})
}

// @Summary      删除工作区文件
// @Description  删除对象存储文件并删除文件元数据（需 editor/admin/owner 权限）
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "工作区ID"
// @Param        file_id path string true "文件ID"
// @Success      200 {object} response.Response "删除成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限操作"
// @Failure      404 {object} response.Response "工作区或文件不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces/{id}/files/{file_id} [delete]
func (h *WorkspaceHandler) DeleteFile(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	workspaceID := strings.TrimSpace(c.Param("id"))
	fileID := strings.TrimSpace(c.Param("file_id"))
	if workspaceID == "" || fileID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace id and file id are required")
		return
	}

	ws, err := workspace.GetWorkSpaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "工作区不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取工作区失败")
		return
	}

	if !ws.IsEditor(userID) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限删除文件")
		return
	}

	file, err := workspace.GetWorkspaceFileByID(c.Request.Context(), fileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "文件不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取文件失败")
		return
	}

	if file.WorkspaceID != workspaceID {
		response.Failed(c, 404, "文件不存在")
		return
	}

	storageClient := storages.GetDefault()
	if storageClient == nil {
		response.Failed(c, response.ErrorUnknownCode, "storage is not configured")
		return
	}

	if err := storageClient.DeleteObject(c.Request.Context(), file.ObjectKey); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "删除对象存储文件失败")
		return
	}

	if err := workspace.DeleteWorkspaceFileByID(c.Request.Context(), fileID); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "删除文件元数据失败")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{"id": fileID, "object_key": file.ObjectKey})
}

// @Summary      批量删除工作区文件
// @Description  批量删除对象存储文件并删除文件元数据（需 editor/admin/owner 权限）
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "工作区ID"
// @Param        request body request.BatchDeleteWorkspaceFilesRequest true "批量删除参数"
// @Success      200 {object} response.Response "删除完成"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限操作"
// @Failure      404 {object} response.Response "工作区不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces/{id}/files/batch-delete [post]
func (h *WorkspaceHandler) BatchDeleteFiles(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	workspaceID := strings.TrimSpace(c.Param("id"))
	if workspaceID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace id is required")
		return
	}

	ws, err := workspace.GetWorkSpaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "工作区不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取工作区失败")
		return
	}

	if !ws.IsEditor(userID) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限删除文件")
		return
	}

	var req request.BatchDeleteWorkspaceFilesRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	files, err := workspace.ListWorkspaceFilesByIDs(c.Request.Context(), req.FileIDs)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "查询文件失败")
		return
	}

	storageClient := storages.GetDefault()
	if storageClient == nil {
		response.Failed(c, response.ErrorUnknownCode, "storage is not configured")
		return
	}

	requested := make(map[string]struct{}, len(req.FileIDs))
	for _, id := range req.FileIDs {
		requested[strings.TrimSpace(id)] = struct{}{}
	}

	deletableIDs := make([]string, 0, len(files))
	failed := make([]gin.H, 0)

	for _, file := range files {
		if file.WorkspaceID != workspaceID {
			failed = append(failed, gin.H{"file_id": file.ID, "reason": "file does not belong to workspace"})
			continue
		}

		if err := storageClient.DeleteObject(c.Request.Context(), file.ObjectKey); err != nil {
			failed = append(failed, gin.H{"file_id": file.ID, "reason": "delete object failed"})
			continue
		}

		deletableIDs = append(deletableIDs, file.ID)
		delete(requested, file.ID)
	}

	for id := range requested {
		if id != "" {
			failed = append(failed, gin.H{"file_id": id, "reason": "file not found"})
		}
	}

	if len(deletableIDs) > 0 {
		if err := workspace.DeleteWorkspaceFilesByIDs(c.Request.Context(), deletableIDs); err != nil {
			response.Failed(c, response.ErrorUnknownCode, "批量删除文件元数据失败")
			return
		}
	}

	response.Success(c, response.SuccessCode, gin.H{
		"deleted_ids": deletableIDs,
		"failed":      failed,
	})
}

// @Summary      工作区邀请列表
// @Description  获取指定工作区邀请记录（需 owner/admin 权限）
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "工作区ID"
// @Param        status query string false "邀请状态"
// @Param        limit query int false "分页大小"
// @Param        offset query int false "偏移量"
// @Success      200 {object} response.Response "获取成功"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限访问"
// @Failure      404 {object} response.Response "工作区不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces/{id}/invitations [get]
func (h *WorkspaceHandler) ListInvitations(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	workspaceID := strings.TrimSpace(c.Param("id"))
	if workspaceID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace id is required")
		return
	}

	ws, err := workspace.GetWorkSpaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "工作区不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取工作区失败")
		return
	}

	if !ws.IsAdmin(userID) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限查看邀请")
		return
	}

	status := strings.TrimSpace(c.Query("status"))
	limit, offset := request.ParsePagination(c)
	invitations, err := workspace.ListInvitationsByWorkspace(c.Request.Context(), workspaceID, status, limit, offset)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "获取邀请列表失败")
		return
	}

	response.Success(c, response.SuccessCode, invitations)
}

// @Summary      创建工作区邀请
// @Description  通过邮箱或用户ID创建邀请（需 owner/admin 权限）
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "工作区ID"
// @Param        request body request.CreateWorkspaceInvitationRequest true "邀请参数"
// @Success      201 {object} response.Response "创建成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限创建"
// @Failure      404 {object} response.Response "工作区不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces/{id}/invitations [post]
func (h *WorkspaceHandler) CreateInvitation(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	userEmail := strings.TrimSpace(c.GetString("user_email"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	workspaceID := strings.TrimSpace(c.Param("id"))
	if workspaceID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace id is required")
		return
	}

	var req request.CreateWorkspaceInvitationRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	if req.InviteeUserID == nil && strings.TrimSpace(req.InviteeEmail) == "" {
		response.Failed(c, response.ErrorBadRequestCode, "invitee_user_id or invitee_email is required")
		return
	}

	if !workspace.IsValidInviteRole(req.Role) {
		response.Failed(c, response.ErrorBadRequestCode, "invalid invitation role")
		return
	}

	ws, err := workspace.GetWorkSpaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "工作区不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取工作区失败")
		return
	}

	if !ws.IsAdmin(userID) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限发起邀请")
		return
	}

	resolvedUserID, resolvedEmail, err := workspace.ResolveInvitee(c.Request.Context(), req.InviteeUserID, req.InviteeEmail)
	if err != nil {
		if errors.Is(err, workspace.ErrInvitationTargetInvalid) || errors.Is(err, gorm.ErrInvalidData) {
			response.Failed(c, response.ErrorBadRequestCode, "invalid invitee target")
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, response.ErrorBadRequestCode, "invitee user not found")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "解析邀请对象失败")
		return
	}

	if resolvedUserID != nil && strings.TrimSpace(*resolvedUserID) == userID {
		response.Failed(c, response.ErrorBadRequestCode, "cannot invite yourself")
		return
	}
	if resolvedEmail != "" && strings.EqualFold(resolvedEmail, userEmail) {
		response.Failed(c, response.ErrorBadRequestCode, "cannot invite yourself")
		return
	}

	created, token, err := workspace.CreateInvitation(c.Request.Context(), workspace.CreateInvitationInput{
		WorkspaceID:   workspaceID,
		InviterID:     userID,
		InviteeUserID: resolvedUserID,
		InviteeEmail:  resolvedEmail,
		Role:          req.Role,
		ExpiresAt:     getInvitationExpireAt(),
	})
	if err != nil {
		if errors.Is(err, workspace.ErrInvitationPendingExists) {
			response.Success(c, response.SuccessCode, gin.H{
				"invitation": created,
				"token":      "",
				"reused":     true,
			})
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "创建邀请失败")
		return
	}

	response.Success(c, response.SuccessCreatedCode, gin.H{
		"invitation": created,
		"token":      token,
		"reused":     false,
	})
}

// @Summary      撤销工作区邀请
// @Description  撤销待处理邀请（需 owner/admin 权限）
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "工作区ID"
// @Param        invite_id path string true "邀请ID"
// @Success      200 {object} response.Response "撤销成功"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限撤销"
// @Failure      404 {object} response.Response "邀请不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /workspaces/{id}/invitations/{invite_id}/revoke [post]
func (h *WorkspaceHandler) RevokeInvitation(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	workspaceID := strings.TrimSpace(c.Param("id"))
	inviteID := strings.TrimSpace(c.Param("invite_id"))
	if workspaceID == "" || inviteID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "workspace id and invite id are required")
		return
	}

	ws, err := workspace.GetWorkSpaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "工作区不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取工作区失败")
		return
	}

	if !ws.IsAdmin(userID) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限撤销邀请")
		return
	}

	invitation, err := workspace.GetInvitationByID(c.Request.Context(), inviteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "邀请不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取邀请失败")
		return
	}

	if invitation.WorkspaceID != workspaceID {
		response.Failed(c, 404, "邀请不存在")
		return
	}

	revoked, err := workspace.RevokeInvitation(c.Request.Context(), inviteID)
	if err != nil {
		h.handleInvitationStateError(c, err, "撤销邀请失败")
		return
	}

	response.Success(c, response.SuccessCode, revoked)
}

// @Summary      我的邀请列表
// @Description  获取当前用户收到的邀请列表
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        status query string false "邀请状态"
// @Param        limit query int false "分页大小"
// @Param        offset query int false "偏移量"
// @Success      200 {object} response.Response "获取成功"
// @Failure      401 {object} response.Response "未授权"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /invitations/me [get]
func (h *WorkspaceHandler) ListMyInvitations(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	userEmail := strings.TrimSpace(c.GetString("user_email"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	status := strings.TrimSpace(c.Query("status"))
	limit, offset := request.ParsePagination(c)
	invitations, err := workspace.ListInvitationsForUser(c.Request.Context(), userID, userEmail, status, limit, offset)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "获取邀请列表失败")
		return
	}

	response.Success(c, response.SuccessCode, invitations)
}

// @Summary      接受邀请
// @Description  当前用户通过 token 接受邀请并加入工作区
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        invite_id path string true "邀请ID"
// @Param        request body request.ResolveWorkspaceInvitationRequest true "接受参数"
// @Success      200 {object} response.Response "接受成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限操作"
// @Failure      404 {object} response.Response "邀请不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /invitations/{invite_id}/accept [post]
func (h *WorkspaceHandler) AcceptInvitation(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	userEmail := strings.TrimSpace(c.GetString("user_email"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	inviteID := strings.TrimSpace(c.Param("invite_id"))
	if inviteID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "invite id is required")
		return
	}

	var req request.ResolveWorkspaceInvitationRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	accepted, err := workspace.AcceptInvitation(c.Request.Context(), inviteID, req.Token, userID, userEmail)
	if err != nil {
		h.handleInvitationStateError(c, err, "接受邀请失败")
		return
	}

	response.Success(c, response.SuccessCode, accepted)
}

// @Summary      拒绝邀请
// @Description  当前用户通过 token 拒绝邀请
// @Tags         Workspace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        invite_id path string true "邀请ID"
// @Param        request body request.ResolveWorkspaceInvitationRequest true "拒绝参数"
// @Success      200 {object} response.Response "拒绝成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限操作"
// @Failure      404 {object} response.Response "邀请不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /invitations/{invite_id}/reject [post]
func (h *WorkspaceHandler) RejectInvitation(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	userEmail := strings.TrimSpace(c.GetString("user_email"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	inviteID := strings.TrimSpace(c.Param("invite_id"))
	if inviteID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "invite id is required")
		return
	}

	var req request.ResolveWorkspaceInvitationRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	rejected, err := workspace.RejectInvitation(c.Request.Context(), inviteID, req.Token, userID, userEmail)
	if err != nil {
		h.handleInvitationStateError(c, err, "拒绝邀请失败")
		return
	}

	response.Success(c, response.SuccessCode, rejected)
}

func getInvitationExpireAt() time.Time {
	expireHours := config.GetInt("WORKSPACE.INVITATION_EXPIRE_HOURS")
	if expireHours <= 0 {
		expireHours = 24 * 7
	}
	return time.Now().Add(time.Duration(expireHours) * time.Hour)
}

func (h *WorkspaceHandler) handleInvitationStateError(c *gin.Context, err error, fallbackMsg string) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.Failed(c, 404, "邀请不存在")
	case errors.Is(err, workspace.ErrInvitationNotPending):
		response.Failed(c, response.ErrorBadRequestCode, "邀请状态不是待处理")
	case errors.Is(err, workspace.ErrInvitationExpired):
		response.Failed(c, response.ErrorBadRequestCode, "邀请已过期")
	case errors.Is(err, workspace.ErrInvitationTokenInvalid):
		response.Failed(c, response.ErrorBadRequestCode, "邀请token无效")
	case errors.Is(err, workspace.ErrInvitationTargetInvalid):
		response.Failed(c, response.ErrorForbiddenCode, "当前用户无权限处理该邀请")
	default:
		response.Failed(c, response.ErrorUnknownCode, fallbackMsg)
	}
}

func buildWorkspaceFileKey(workspaceID, folderID, filename string) string {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	name = strings.TrimSpace(name)
	if name == "" {
		name = "file"
	}
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.Trim(name, "/")

	timestamp := time.Now().UnixNano()
	if strings.TrimSpace(folderID) == "" {
		return path.Join("workspaces", workspaceID, "root", fmt.Sprintf("%d_%s%s", timestamp, name, ext))
	}

	return path.Join("workspaces", workspaceID, "folders", folderID, fmt.Sprintf("%d_%s%s", timestamp, name, ext))
}
