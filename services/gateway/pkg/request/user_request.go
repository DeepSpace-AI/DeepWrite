package request

type UserProfileRequest struct {
	DisplayName string `json:"display_name" binding:"required,max=30" example:"deepwrite-user"`
	AvatarURL   string `json:"avatar_url" binding:"required" example:"https://example.com/avatar.jpg"`
	Bio         string `json:"bio" binding:"max=255" example:"This is a user bio."`
	Language    string `json:"language" example:"en"`
	Timezone    string `json:"timezone" example:"UTC"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6" example:"123456"`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"654321"`
}

type ChangeEmailRequest struct {
	NewEmail string `json:"new_email" binding:"required,email" example:"newuser@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"123456"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required" example:"f2b412f7f05f4d199f8b7a6d577d4b2a"`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"654321"`
}
