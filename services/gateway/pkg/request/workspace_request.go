package request

type CreateWorkspaceRequest struct {
	Name        string `json:"name" binding:"required,max=255"`
	Description string `json:"description" binding:"max=1000"`
	Public      bool   `json:"public"`
}

type UpdateWorkspaceRequest struct {
	Name        string `json:"name" binding:"omitempty,max=255"`
	Description string `json:"description" binding:"max=1000"`
	Public      *bool  `json:"public"`
	Status      string `json:"status" binding:"omitempty,oneof=active archived"`
}

type CreateFolderRequest struct {
	ParentID    *string `json:"parent_id" binding:"omitempty,uuid"`
	Name        string  `json:"name" binding:"required,max=255"`
	Description string  `json:"description" binding:"max=1000"`
}

type UpdateFolderRequest struct {
	Name        string `json:"name" binding:"omitempty,max=255"`
	Description string `json:"description" binding:"max=1000"`
}

type PresignWorkspaceUploadRequest struct {
	FolderID    string `json:"folder_id" binding:"omitempty,uuid"`
	FileName    string `json:"file_name" binding:"required,max=255"`
	ContentType string `json:"content_type" binding:"omitempty,max=255"`
	ExpiresIn   int64  `json:"expires_in" binding:"omitempty,min=60,max=3600"`
}

type CompleteWorkspaceUploadRequest struct {
	FolderID  string `json:"folder_id" binding:"omitempty,uuid"`
	ObjectKey string `json:"object_key" binding:"required,max=1024"`
	FileName  string `json:"file_name" binding:"required,max=255"`
}

type BatchDeleteWorkspaceFilesRequest struct {
	FileIDs []string `json:"file_ids" binding:"required,min=1,dive,required,uuid"`
}

type CreateWorkspaceInvitationRequest struct {
	InviteeUserID *string `json:"invitee_user_id" binding:"omitempty,uuid"`
	InviteeEmail  string  `json:"invitee_email" binding:"omitempty,email"`
	Role          string  `json:"role" binding:"required,oneof=admin editor viewer"`
}

type ResolveWorkspaceInvitationRequest struct {
	Token       string `json:"token" binding:"omitempty"`
	ActionToken string `json:"action_token" binding:"omitempty"`
}

type UpdateWorkspaceMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin editor viewer"`
}
