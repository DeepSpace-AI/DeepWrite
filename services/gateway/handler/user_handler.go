package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/models/user"
	"github.com/deepwrite/serivces/gateway/pkg/cache"
	"github.com/deepwrite/serivces/gateway/pkg/config"
	"github.com/deepwrite/serivces/gateway/pkg/mailer"
	"github.com/deepwrite/serivces/gateway/pkg/request"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/deepwrite/serivces/gateway/pkg/storages"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
}

const passwordResetTokenPrefix = "user:password_reset:"

// @Summary      获取用户信息
// @Description  获取当前登录用户的基本信息和个人资料
// @Tags         User
// @Accept       json
// @Produce      json
// @Success      200 {object} user.User "用户信息"
// @Failure      401 {object} response.Response "未授权"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	uid := c.GetString("user_id")
	if uid == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	user, err := user.GetUserByID(c.Request.Context(), uid)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "获取用户信息失败")
		return
	}

	response.Success(c, response.SuccessCode, user)
}

// @Summary      更新用户信息
// @Description  更新当前登录用户的个人资料信息
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        profile body request.UserProfileRequest true "用户个人资料"
// @Success      200 {object} user.User "更新后的用户信息"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /user/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req request.UserProfileRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	uid := c.GetString("user_id")
	if uid == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	u, err := user.GetUserByID(c.Request.Context(), uid)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "获取用户信息失败")
		return
	}

	u.UserProfile.AvatarURL = req.AvatarURL
	u.UserProfile.DisplayName = req.DisplayName
	u.UserProfile.Bio = req.Bio
	u.UserProfile.Language = req.Language
	u.UserProfile.Timezone = req.Timezone

	if err := user.UpdateUserProfile(c.Request.Context(), u); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "更新用户信息失败")
		return
	}

	response.Success(c, response.SuccessCode, u)
}

// @Summary      上传用户头像
// @Description  上传当前登录用户头像并自动更新用户资料中的 avatar_url
// @Tags         User
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "头像文件（支持 image/*）"
// @Success      200 {object} response.Response "上传成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /user/avatar [post]
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	uid := strings.TrimSpace(c.GetString("user_id"))
	if uid == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	storageClient := storages.GetDefault()
	if storageClient == nil {
		response.Failed(c, response.ErrorUnknownCode, "storage is not configured")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "file is required")
		return
	}

	if file.Size <= 0 {
		response.Failed(c, response.ErrorBadRequestCode, "file is empty")
		return
	}

	if file.Size > 5*1024*1024 {
		response.Failed(c, response.ErrorBadRequestCode, "file size must be <= 5MB")
		return
	}

	contentType := strings.TrimSpace(file.Header.Get("Content-Type"))
	if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		response.Failed(c, response.ErrorBadRequestCode, "only image files are allowed")
		return
	}

	src, err := file.Open()
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to open uploaded file")
		return
	}
	defer src.Close()

	objectKey := buildUserAvatarObjectKey(uid, file.Filename)
	if err := storageClient.PutObject(c.Request.Context(), objectKey, src, contentType, map[string]string{"user_id": uid, "kind": "avatar"}); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to upload avatar")
		return
	}

	previewURL, err := storageClient.PresignGetURL(c.Request.Context(), objectKey, 7*24*time.Hour)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to generate avatar url")
		return
	}

	u, err := user.GetUserByID(c.Request.Context(), uid)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "获取用户信息失败")
		return
	}
	u.UserProfile.AvatarURL = previewURL
	if err := user.UpdateUserProfile(c.Request.Context(), u); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "更新用户头像失败")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{
		"avatar_url": previewURL,
		"object_key": objectKey,
	})
}

// @Summary      修改密码
// @Description  修改当前登录用户的密码
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request body request.ChangePasswordRequest true "修改密码请求"
// @Success      200 {object} response.Response "密码修改成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /user/change-password [post]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req request.ChangePasswordRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	uid := c.GetString("user_id")
	if uid == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	u, err := user.GetUserByID(c.Request.Context(), uid)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "获取用户信息失败")
		return
	}

	if !u.VerifyPassword(req.OldPassword) {
		response.Failed(c, response.ErrorUnauthorizedCode, "旧密码不正确")
		return
	}
	if strings.TrimSpace(req.OldPassword) == strings.TrimSpace(req.NewPassword) {
		response.Failed(c, response.ErrorBadRequestCode, "新密码不能与旧密码一致")
		return
	}

	if err := u.ChangePassword(c.Request.Context(), req.NewPassword); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "修改密码失败")
		return
	}

	_ = h.sendPasswordChangedAlert(c.Request.Context(), u, c.ClientIP(), c.Request.UserAgent())

	response.Success(c, response.SuccessCode, "密码修改成功")
}

// @Summary      修改邮箱
// @Description  当前登录用户修改邮箱（需校验密码）
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body request.ChangeEmailRequest true "修改邮箱请求"
// @Success      200 {object} response.Response "邮箱修改成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      401 {object} response.Response "未授权或密码错误"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /user/change-email [post]
func (h *UserHandler) ChangeEmail(c *gin.Context) {
	var req request.ChangeEmailRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	uid := c.GetString("user_id")
	if uid == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	u, err := user.GetUserByID(c.Request.Context(), uid)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "获取用户信息失败")
		return
	}

	if !u.VerifyPassword(req.Password) {
		response.Failed(c, response.ErrorUnauthorizedCode, "密码错误")
		return
	}

	newEmail := normalizeEmail(req.NewEmail)
	if newEmail == "" {
		response.Failed(c, response.ErrorBadRequestCode, "邮箱格式不正确")
		return
	}
	if normalizeEmail(u.Email) == newEmail {
		response.Failed(c, response.ErrorBadRequestCode, "新邮箱不能与当前邮箱一致")
		return
	}

	existing, findErr := user.GetUserByEmail(c.Request.Context(), newEmail)
	if findErr == nil && existing.ID != u.ID {
		response.Failed(c, response.ErrorBadRequestCode, "邮箱已被占用")
		return
	}
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		response.Failed(c, response.ErrorUnknownCode, "校验邮箱失败")
		return
	}

	if err := u.ChangeEmail(c.Request.Context(), newEmail); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "修改邮箱失败")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{
		"id":    u.ID,
		"email": u.Email,
	})
}

// @Summary      忘记密码
// @Description  发送重置密码邮件（若邮箱存在）
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request body request.ForgotPasswordRequest true "忘记密码请求"
// @Success      200 {object} response.Response "请求已受理"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /user/forgot-password [post]
func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var req request.ForgotPasswordRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	cacheStore := cache.GetDefault()
	if cacheStore == nil {
		response.Failed(c, response.ErrorUnknownCode, "系统暂不可用")
		return
	}

	email := normalizeEmail(req.Email)
	u, err := user.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		response.Success(c, response.SuccessCode, gin.H{"message": "如果邮箱存在，我们已发送重置指引"})
		return
	}

	token, err := generateSecureToken(32)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "生成重置令牌失败")
		return
	}

	expireAt := time.Now().Add(30 * time.Minute)
	if err := cacheStore.Set(passwordResetTokenPrefix+token, u.ID, expireAt); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "保存重置令牌失败")
		return
	}

	go func(target user.User, resetToken string) {
		_ = h.sendResetPasswordMail(context.Background(), target, resetToken, 30)
	}(u, token)
	response.Success(c, response.SuccessCode, gin.H{"message": "如果邮箱存在，我们已发送重置指引"})
}

// @Summary      重置密码
// @Description  通过忘记密码令牌设置新密码
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request body request.ResetPasswordRequest true "重置密码请求"
// @Success      200 {object} response.Response "重置成功"
// @Failure      400 {object} response.Response "请求参数错误或令牌失效"
// @Failure      500 {object} response.Response "服务器错误"
// @Router       /user/reset-password [post]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req request.ResetPasswordRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	cacheStore := cache.GetDefault()
	if cacheStore == nil {
		response.Failed(c, response.ErrorUnknownCode, "系统暂不可用")
		return
	}

	key := passwordResetTokenPrefix + strings.TrimSpace(req.Token)
	rawUserID, err := cacheStore.Get(key)
	if err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "重置链接无效或已过期")
		return
	}

	uid := strings.TrimSpace(fmt.Sprint(rawUserID))
	if uid == "" {
		response.Failed(c, response.ErrorBadRequestCode, "重置链接无效或已过期")
		return
	}

	u, err := user.GetUserByID(c.Request.Context(), uid)
	if err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "用户不存在")
		return
	}

	if err := u.ChangePassword(c.Request.Context(), req.NewPassword); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "重置密码失败")
		return
	}

	_ = cacheStore.Delete(key)
	_ = h.sendPasswordChangedAlert(c.Request.Context(), u, c.ClientIP(), c.Request.UserAgent())

	response.Success(c, response.SuccessCode, "密码重置成功")
}

func (h *UserHandler) sendResetPasswordMail(ctx context.Context, u user.User, token string, expireMinutes int) error {
	sender := mailer.GetDefault()
	if sender == nil {
		return nil
	}

	cfg := config.GetGlobalConfig()
	templateDir := strings.TrimSpace(cfg.Mail.TemplateDir)
	if templateDir == "" {
		templateDir = "templates/mail"
	}

	resetURL := token
	htmlBody, err := mailer.RenderHTMLTemplateWithLayout(
		filepath.Join(templateDir, "base.html"),
		filepath.Join(templateDir, "reset_password.html"),
		gin.H{
			"Brand":         "DeepWrite",
			"Title":         "重置你的密码",
			"Preheader":     "我们收到了一次密码重置请求。",
			"SupportEmail":  cfg.Mail.FromMail,
			"Name":          h.displayName(u),
			"ResetURL":      resetURL,
			"ExpireMinutes": expireMinutes,
		},
	)
	if err != nil {
		return err
	}

	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return sender.Send(sendCtx, mailer.Message{
		To:       []string{u.Email},
		Subject:  "[DeepWrite] 重置密码",
		TextBody: "我们收到了你的密码重置请求，请使用该链接完成重置: " + resetURL,
		HTMLBody: htmlBody,
	})
}

func (h *UserHandler) sendPasswordChangedAlert(ctx context.Context, u user.User, ip, userAgent string) error {
	sender := mailer.GetDefault()
	if sender == nil {
		return nil
	}

	cfg := config.GetGlobalConfig()
	templateDir := strings.TrimSpace(cfg.Mail.TemplateDir)
	if templateDir == "" {
		templateDir = "templates/mail"
	}

	htmlBody, err := mailer.RenderHTMLTemplateWithLayout(
		filepath.Join(templateDir, "base.html"),
		filepath.Join(templateDir, "password_changed.html"),
		gin.H{
			"Brand":        "DeepWrite",
			"Title":        "密码已更新",
			"Preheader":    "你的账号密码刚刚发生变更。",
			"SupportEmail": cfg.Mail.FromMail,
			"Name":         h.displayName(u),
			"ChangedAt":    time.Now().Format("2006-01-02 15:04:05 MST"),
			"IP":           ip,
			"UserAgent":    userAgent,
		},
	)
	if err != nil {
		return err
	}

	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return sender.Send(sendCtx, mailer.Message{
		To:       []string{u.Email},
		Subject:  "[DeepWrite] 密码变更提醒",
		TextBody: "你的账号密码已修改。如果不是你本人操作，请立即重置密码。",
		HTMLBody: htmlBody,
	})
}

func (h *UserHandler) displayName(u user.User) string {
	name := strings.TrimSpace(u.UserProfile.DisplayName)
	if name != "" {
		return name
	}
	return strings.TrimSpace(u.Email)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func generateSecureToken(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func buildUserAvatarObjectKey(userID, filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".bin"
	}

	base := strings.TrimSpace(strings.TrimSuffix(filename, filepath.Ext(filename)))
	if base == "" {
		base = "avatar"
	}
	base = strings.ReplaceAll(base, " ", "_")
	base = strings.Trim(base, "/")

	return path.Join("users", userID, "avatar", fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), base, ext))
}
