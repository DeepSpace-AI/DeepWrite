package handler

import (
	"context"
	"crypto/rand"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/models/user"
	"github.com/deepwrite/serivces/gateway/pkg/database"
	"github.com/deepwrite/serivces/gateway/pkg/request"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminUserHandler struct{}

type UserResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpdateUserRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Role        *string `json:"role,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type UserListResponse struct {
	Users  []UserResponse `json:"users"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// @Summary      获取用户列表
// @Description  获取所有用户列表，支持分页和搜索
// @Tags         Admin/User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "每页数量" default(50)
// @Param        offset query int false "偏移量" default(0)
// @Param        keyword query string false "搜索关键词"
// @Param        status query string false "用户状态筛选"
// @Success      200 {object} response.Response{data=UserListResponse} "用户列表"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /admin/users [get]
func (h *AdminUserHandler) ListUsers(c *gin.Context) {
	limit := parseIntQuery(c, "limit", 50)
	offset := parseIntQuery(c, "offset", 0)
	keyword := strings.TrimSpace(c.Query("keyword"))
	status := strings.TrimSpace(c.Query("status"))

	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var total int64
	var users []user.User

	query := database.DB.WithContext(context.Background()).Model(&user.User{})

	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		query = query.Where("email ILIKE ? OR user_profiles.display_name ILIKE ?", likeKeyword, likeKeyword)
	}

	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Preload("UserProfile").Count(&total).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "查询用户列表失败")
		return
	}

	if err := query.Preload("UserProfile").Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "查询用户列表失败")
		return
	}

	userList := make([]UserResponse, 0, len(users))
	for _, u := range users {
		userList = append(userList, toUserResponse(u))
	}

	response.Success(c, response.SuccessCode, gin.H{
		"users":  userList,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// @Summary      获取用户详情
// @Description  根据ID获取用户详细信息
// @Tags         Admin/User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        userId path string true "用户ID"
// @Success      200 {object} response.Response{data=UserResponse} "用户信息"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限"
// @Failure      404 {object} response.Response "用户不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /admin/users/{userId} [get]
func (h *AdminUserHandler) GetUser(c *gin.Context) {
	userId := strings.TrimSpace(c.Param("userId"))
	if userId == "" {
		response.Failed(c, response.ErrorBadRequestCode, "user_id is required")
		return
	}

	u, err := user.GetUserByID(c.Request.Context(), userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "用户不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取用户信息失败")
		return
	}

	response.Success(c, response.SuccessCode, toUserResponse(u))
}

// @Summary      更新用户
// @Description  更新用户信息（角色、状态、显示名）
// @Tags         Admin/User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        userId path string true "用户ID"
// @Param        body body UpdateUserRequest true "更新信息"
// @Success      200 {object} response.Response{data=UserResponse} "更新后的用户信息"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限"
// @Failure      404 {object} response.Response "用户不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /admin/users/{userId} [patch]
func (h *AdminUserHandler) UpdateUser(c *gin.Context) {
	userId := strings.TrimSpace(c.Param("userId"))
	if userId == "" {
		response.Failed(c, response.ErrorBadRequestCode, "user_id is required")
		return
	}

	var req UpdateUserRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	u, err := user.GetUserByID(c.Request.Context(), userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "用户不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取用户信息失败")
		return
	}

	if req.DisplayName != nil {
		u.UserProfile.DisplayName = *req.DisplayName
	}
	if req.Role != nil {
		u.Role = *req.Role
	}
	if req.Status != nil {
		u.Status = *req.Status
	}

	if err := user.UpdateUserProfile(c.Request.Context(), u); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "更新用户信息失败")
		return
	}

	response.Success(c, response.SuccessCode, toUserResponse(u))
}

// @Summary      删除用户
// @Description  删除指定用户
// @Tags         Admin/User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        userId path string true "用户ID"
// @Success      200 {object} response.Response "删除成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限"
// @Failure      404 {object} response.Response "用户不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /admin/users/{userId} [delete]
func (h *AdminUserHandler) DeleteUser(c *gin.Context) {
	userId := strings.TrimSpace(c.Param("userId"))
	if userId == "" {
		response.Failed(c, response.ErrorBadRequestCode, "user_id is required")
		return
	}

	u, err := user.GetUserByID(c.Request.Context(), userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "用户不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取用户信息失败")
		return
	}

	if err := database.DB.WithContext(c.Request.Context()).Delete(&u).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "删除用户失败")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{"user_id": userId})
}

// @Summary      重置用户密码
// @Description  管理员重置用户密码，生成新的随机密码
// @Tags         Admin/User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        userId path string true "用户ID"
// @Success      200 {object} response.Response{data=ResetPasswordResponse} "新密码"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      403 {object} response.Response "无权限"
// @Failure      404 {object} response.Response "用户不存在"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /admin/users/{userId}/reset-password [post]
func (h *AdminUserHandler) ResetPassword(c *gin.Context) {
	userId := strings.TrimSpace(c.Param("userId"))
	if userId == "" {
		response.Failed(c, response.ErrorBadRequestCode, "user_id is required")
		return
	}

	u, err := user.GetUserByID(c.Request.Context(), userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "用户不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取用户信息失败")
		return
	}

	newPassword := generateRandomPassword(16)
	if err := u.ChangePassword(c.Request.Context(), newPassword); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "重置密码失败")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{
		"user_id":      userId,
		"email":        u.Email,
		"new_password": newPassword,
	})
}

func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

func toUserResponse(u user.User) UserResponse {
	resp := UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
	if u.UserProfile.DisplayName != "" {
		resp.DisplayName = u.UserProfile.DisplayName
	}
	if u.UserProfile.AvatarURL != "" {
		resp.AvatarURL = u.UserProfile.AvatarURL
	}
	return resp
}

func parseIntQuery(c *gin.Context, key string, defaultVal int) int {
	valStr := c.Query(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil || val < 0 {
		return defaultVal
	}
	return val
}
