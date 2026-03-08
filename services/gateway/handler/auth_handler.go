package handler

import (
	"context"
	"strings"

	"github.com/deepwrite/serivces/gateway/models/user"
	"github.com/deepwrite/serivces/gateway/pkg/hash"
	"github.com/deepwrite/serivces/gateway/pkg/jwt"
	"github.com/deepwrite/serivces/gateway/pkg/request"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
}

// @Summary      用户注册
// @Description  注册一个新用户
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body request.RegisterRequest true "注册信息"
// @Success      201 {object} user.User "注册成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	ctx := context.Background()
	if _, err := user.GetUserByEmail(ctx, req.Email); err == nil {
		response.Failed(c, response.ErrorBadRequestCode, "user already exists")
		return
	}

	newUser := user.User{
		Email:    req.Email,
		Password: req.Password,
		Status:   "active",
		Role:     "user",
		UserProfile: user.Profile{
			DisplayName: req.DisplayName,
		},
	}

	if err := user.Create(ctx, &newUser); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "创建用户失败")
		return
	}

	response.Success(c, response.SuccessCreatedCode, newUser)
}

// @Summary      用户登录
// @Description  使用邮箱和密码登录，获取访问令牌和刷新令牌
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body request.LoginRequest true "登录信息"
// @Success      200 {object} response.Response "登录成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "认证失败"
// @Failure      403 {object} response.Response "用户被禁用"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	ctx := context.Background()
	currentUser, err := user.GetUserByEmail(ctx, req.Email)
	if err != nil {
		response.Failed(c, response.ErrorUnauthorizedCode, "email or password is incorrect")
		return
	}

	if !hash.VerifyPassword(currentUser.Password, req.Password) {
		response.Failed(c, response.ErrorUnauthorizedCode, "email or password is incorrect")
		return
	}

	if !currentUser.IsActive() {
		response.Failed(c, response.ErrorForbiddenCode, "user is inactive")
		return
	}

	token, err := jwt.GenerateAccessToken(currentUser.ID, currentUser.Email, currentUser.Role)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to generate token")
		return
	}

	refreshToken, err := jwt.GenerateRefreshToken(currentUser.ID, currentUser.Email, currentUser.Role)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to generate refresh token")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{
		"access_token":  token,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"user": gin.H{
			"id":     currentUser.ID,
			"email":  currentUser.Email,
			"role":   currentUser.Role,
			"status": currentUser.Status,
		},
	})
}

// @Summary      刷新令牌
// @Description  使用刷新令牌获取新的访问令牌和刷新令牌
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body request.RefreshTokenRequest true "刷新令牌信息"
// @Success      200 {object} response.Response "刷新成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "认证失败"
// @Failure      403 {object} response.Response "用户被禁用"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req request.RefreshTokenRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	claims, err := jwt.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		response.Failed(c, response.ErrorUnauthorizedCode, "invalid or expired refresh token")
		return
	}

	ctx := context.Background()
	currentUser, err := user.GetUserByEmail(ctx, claims.Email)
	if err != nil {
		response.Failed(c, response.ErrorUnauthorizedCode, "user not found")
		return
	}

	if !currentUser.IsActive() {
		response.Failed(c, response.ErrorForbiddenCode, "user is inactive")
		return
	}

	accessToken, err := jwt.GenerateAccessToken(currentUser.ID, currentUser.Email, currentUser.Role)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to generate access token")
		return
	}

	newRefreshToken, err := jwt.GenerateRefreshToken(currentUser.ID, currentUser.Email, currentUser.Role)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to generate refresh token")
		return
	}

	if err := jwt.RevokeToken(req.RefreshToken, claims); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to revoke old refresh token")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
	})
}

// @Summary      用户登出
// @Description  主动使当前访问令牌失效；可选传入刷新令牌并一并失效
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body request.LogoutRequest false "登出信息（可选刷新令牌）"
// @Success      200 {object} response.Response "登出成功"
// @Failure      401 {object} response.Response "认证失败"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req request.LogoutRequest
	_ = c.ShouldBindJSON(&req)

	if rawTokenValue, ok := c.Get("access_token_raw"); ok {
		if rawToken, ok := rawTokenValue.(string); ok {
			claims, _ := jwt.GetCurrentClaims(c)
			if err := jwt.RevokeToken(rawToken, claims); err != nil {
				response.Failed(c, response.ErrorUnknownCode, "failed to revoke access token")
				return
			}
		}
	}

	if strings.TrimSpace(req.RefreshToken) != "" {
		refreshClaims, err := jwt.ParseRefreshToken(req.RefreshToken)
		if err != nil {
			response.Failed(c, response.ErrorUnauthorizedCode, "invalid or expired refresh token")
			return
		}
		if err := jwt.RevokeToken(req.RefreshToken, refreshClaims); err != nil {
			response.Failed(c, response.ErrorUnknownCode, "failed to revoke refresh token")
			return
		}
	}

	response.Success(c, response.SuccessCode, gin.H{"message": "logout success"})
}

// @Summary      获取当前用户信息
// @Description  根据访问令牌返回当前登录用户的基础信息
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Response "获取成功"
// @Failure      401 {object} response.Response "未授权"
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	claims, ok := jwt.GetCurrentClaims(c)
	if !ok {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{
		"id":    claims.UserID,
		"email": claims.Email,
		"role":  claims.Role,
	})
}
