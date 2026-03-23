package handler

import (
	"github.com/deepwrite/serivces/gateway/models/document"
	"github.com/deepwrite/serivces/gateway/models/user"
	"github.com/deepwrite/serivces/gateway/models/workspace"
	"github.com/deepwrite/serivces/gateway/pkg/database"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

type AdminStats struct {
	TotalUsers       int64 `json:"total_users"`
	TotalWorkspaces  int64 `json:"total_workspaces"`
	TotalDocuments   int64 `json:"total_documents"`
	TotalFiles       int64 `json:"total_files"`
	ActiveWorkspaces int64 `json:"active_workspaces"`
	ActiveUsersToday int64 `json:"active_users_today"`
	StorageUsedBytes int64 `json:"storage_used_bytes"`
}

func (h *AdminUserHandler) GetStats(c *gin.Context) {
	var totalUsers int64
	if err := database.DB.WithContext(c.Request.Context()).Model(&user.User{}).Count(&totalUsers).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to count users")
		return
	}

	var totalWorkspaces int64
	if err := database.DB.WithContext(c.Request.Context()).Model(&workspace.Workspace{}).Count(&totalWorkspaces).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to count workspaces")
		return
	}

	var activeWorkspaces int64
	if err := database.DB.WithContext(c.Request.Context()).
		Model(&workspace.Workspace{}).
		Where("status != ?", "archived").
		Count(&activeWorkspaces).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to count active workspaces")
		return
	}

	var totalDocuments int64
	if err := database.DB.WithContext(c.Request.Context()).
		Model(&document.Document{}).
		Count(&totalDocuments).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to count documents")
		return
	}

	var totalFiles int64
	if err := database.DB.WithContext(c.Request.Context()).
		Model(&workspace.WorkspaceFile{}).
		Count(&totalFiles).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to count files")
		return
	}

	var activeUsersToday int64
	if err := database.DB.WithContext(c.Request.Context()).
		Model(&user.User{}).
		Where("last_login >= CURRENT_DATE").
		Count(&activeUsersToday).Error; err != nil {
		activeUsersToday = 0
	}

	var storageUsedBytes int64
	if err := database.DB.WithContext(c.Request.Context()).
		Model(&workspace.WorkspaceFile{}).
		Select("COALESCE(SUM(size), 0)").
		Scan(&storageUsedBytes).Error; err != nil {
		storageUsedBytes = 0
	}

	stats := AdminStats{
		TotalUsers:       totalUsers,
		TotalWorkspaces:  totalWorkspaces,
		TotalDocuments:   totalDocuments,
		TotalFiles:       totalFiles,
		ActiveWorkspaces: activeWorkspaces,
		ActiveUsersToday: activeUsersToday,
		StorageUsedBytes: storageUsedBytes,
	}

	response.Success(c, response.SuccessCode, stats)
}
