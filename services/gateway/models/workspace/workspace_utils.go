package workspace

import (
	"context"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/pkg/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UpsertWorkspaceFileInput struct {
	WorkspaceID string
	FolderID    *string
	ObjectKey   string
	FileName    string
	ContentType string
	Size        int64
	ETag        string
	UploadedBy  string
}

type UpdateWorkspaceInput struct {
	Name        string
	Description string
	Public      *bool
	Status      string
}

type CreateFolderInput struct {
	WorkspaceID string
	ParentID    *string
	Name        string
	Description string
}

type UpdateFolderInput struct {
	Name        string
	Description string
}

func (w *Workspace) IsOwner(userID string) bool {
	for _, member := range w.Members {
		if member.UserId == userID && member.Role == RoleOwner {
			return true
		}
	}
	return false
}

func (w *Workspace) IsAdmin(userID string) bool {
	for _, member := range w.Members {
		if member.UserId == userID && (member.Role == RoleAdmin || member.Role == RoleOwner) {
			return true
		}
	}
	return false
}

func (w *Workspace) IsEditor(userID string) bool {
	for _, member := range w.Members {
		if member.UserId == userID && (member.Role == RoleEditor || member.Role == RoleAdmin || member.Role == RoleOwner) {
			return true
		}
	}
	return false
}

func (w *Workspace) IsViewer(userID string) bool {
	for _, member := range w.Members {
		if member.UserId == userID && (member.Role == RoleViewer || member.Role == RoleEditor || member.Role == RoleAdmin || member.Role == RoleOwner) {
			return true
		}
	}
	return false
}

func (w *Workspace) GetUserRole(userID string) string {
	for _, member := range w.Members {
		if member.UserId == userID {
			return member.Role
		}
	}
	return ""
}

func GetWorkSpaceByID(ctx context.Context, id string) (Workspace, error) {
	var ws Workspace
	err := database.DB.WithContext(ctx).
		Preload("Members").
		Where("id = ?", strings.TrimSpace(id)).
		First(&ws).Error
	return ws, err
}

func ListByUser(ctx context.Context, userID string, limit, offset int) ([]Workspace, error) {
	query := database.DB.WithContext(ctx).
		Model(&Workspace{}).
		Distinct("workspaces.*").
		Joins("left join members on members.workspace_id = workspaces.id").
		Where("workspaces.owner_id = ? OR members.user_id = ?", strings.TrimSpace(userID), strings.TrimSpace(userID)).
		Order("workspaces.updated_at desc")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var workspaces []Workspace
	err := query.Preload("Members").Find(&workspaces).Error
	return workspaces, err
}

func Create(ctx context.Context, ws *Workspace, ownerID string) error {
	if ws == nil {
		return gorm.ErrInvalidData
	}

	now := time.Now()
	ws.OwnerID = strings.TrimSpace(ownerID)
	ws.Name = strings.TrimSpace(ws.Name)
	ws.Description = strings.TrimSpace(ws.Description)
	if strings.TrimSpace(ws.Status) == "" {
		ws.Status = "active"
	}
	ws.CreatedAt = now
	ws.UpdatedAt = now

	return database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ws).Error; err != nil {
			return err
		}

		ownerMember := Members{
			WorkspaceID: ws.ID,
			UserId:      ws.OwnerID,
			Role:        RoleOwner,
		}

		return tx.Create(&ownerMember).Error
	})
}

func Update(ctx context.Context, workspaceID string, input UpdateWorkspaceInput) (Workspace, error) {
	var ws Workspace
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", strings.TrimSpace(workspaceID)).
			First(&ws).Error; err != nil {
			return err
		}

		updates := map[string]any{}
		if strings.TrimSpace(input.Name) != "" {
			updates["name"] = strings.TrimSpace(input.Name)
		}
		updates["description"] = strings.TrimSpace(input.Description)
		if input.Public != nil {
			updates["public"] = *input.Public
		}
		if strings.TrimSpace(input.Status) != "" {
			updates["status"] = strings.TrimSpace(input.Status)
		}

		if len(updates) > 0 {
			if err := tx.Model(&Workspace{}).Where("id = ?", ws.ID).Updates(updates).Error; err != nil {
				return err
			}
		}

		return tx.Preload("Members").Where("id = ?", ws.ID).First(&ws).Error
	})

	return ws, err
}

func Delete(ctx context.Context, workspaceID string) error {
	return database.DB.WithContext(ctx).
		Delete(&Workspace{}, "id = ?", strings.TrimSpace(workspaceID)).Error
}

func CreateFolder(ctx context.Context, input CreateFolderInput) (Folder, error) {
	folder := Folder{
		WorkspaceID: strings.TrimSpace(input.WorkspaceID),
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
	}

	if input.ParentID != nil {
		parentID := strings.TrimSpace(*input.ParentID)
		if parentID != "" {
			folder.ParentID = &parentID
		}
	}

	err := database.DB.WithContext(ctx).Create(&folder).Error
	return folder, err
}

func GetFolderByID(ctx context.Context, folderID string) (Folder, error) {
	var folder Folder
	err := database.DB.WithContext(ctx).
		Where("id = ?", strings.TrimSpace(folderID)).
		First(&folder).Error
	return folder, err
}

func ListFoldersByWorkspace(ctx context.Context, workspaceID string, parentID string, limit, offset int) ([]Folder, error) {
	query := database.DB.WithContext(ctx).
		Where("workspace_id = ?", strings.TrimSpace(workspaceID)).
		Order("updated_at desc")

	trimmedParentID := strings.TrimSpace(parentID)
	if trimmedParentID != "" {
		query = query.Where("parent_id = ?", trimmedParentID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var folders []Folder
	err := query.Find(&folders).Error
	return folders, err
}

func UpdateFolder(ctx context.Context, folderID string, input UpdateFolderInput) (Folder, error) {
	var folder Folder
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", strings.TrimSpace(folderID)).
			First(&folder).Error; err != nil {
			return err
		}

		updates := map[string]any{
			"description": strings.TrimSpace(input.Description),
		}

		if strings.TrimSpace(input.Name) != "" {
			updates["name"] = strings.TrimSpace(input.Name)
		}

		if err := tx.Model(&Folder{}).Where("id = ?", folder.ID).Updates(updates).Error; err != nil {
			return err
		}

		return tx.Where("id = ?", folder.ID).First(&folder).Error
	})

	return folder, err
}

func DeleteFolder(ctx context.Context, folderID string) error {
	return database.DB.WithContext(ctx).
		Delete(&Folder{}, "id = ?", strings.TrimSpace(folderID)).Error
}

func UpsertWorkspaceFile(ctx context.Context, input UpsertWorkspaceFileInput) (WorkspaceFile, error) {
	file := WorkspaceFile{
		WorkspaceID: strings.TrimSpace(input.WorkspaceID),
		FolderID:    input.FolderID,
		ObjectKey:   strings.TrimSpace(input.ObjectKey),
		FileName:    strings.TrimSpace(input.FileName),
		ContentType: strings.TrimSpace(input.ContentType),
		Size:        input.Size,
		ETag:        strings.TrimSpace(input.ETag),
		UploadedBy:  strings.TrimSpace(input.UploadedBy),
	}

	if file.ContentType == "" {
		file.ContentType = "application/octet-stream"
	}

	err := database.DB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "object_key"}},
			DoUpdates: clause.AssignmentColumns([]string{"workspace_id", "folder_id", "file_name", "content_type", "size", "etag", "uploaded_by", "updated_at"}),
		}).
		Create(&file).Error
	if err != nil {
		return WorkspaceFile{}, err
	}

	var saved WorkspaceFile
	err = database.DB.WithContext(ctx).
		Where("object_key = ?", file.ObjectKey).
		First(&saved).Error
	return saved, err
}

func ListWorkspaceFiles(ctx context.Context, workspaceID string, folderID string, limit, offset int) ([]WorkspaceFile, error) {
	query := database.DB.WithContext(ctx).
		Where("workspace_id = ?", strings.TrimSpace(workspaceID)).
		Order("updated_at desc")

	trimmedFolderID := strings.TrimSpace(folderID)
	if trimmedFolderID != "" {
		query = query.Where("folder_id = ?", trimmedFolderID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var files []WorkspaceFile
	err := query.Find(&files).Error
	return files, err
}

func GetWorkspaceFileByID(ctx context.Context, fileID string) (WorkspaceFile, error) {
	var file WorkspaceFile
	err := database.DB.WithContext(ctx).
		Where("id = ?", strings.TrimSpace(fileID)).
		First(&file).Error
	return file, err
}

func DeleteWorkspaceFileByID(ctx context.Context, fileID string) error {
	return database.DB.WithContext(ctx).
		Delete(&WorkspaceFile{}, "id = ?", strings.TrimSpace(fileID)).Error
}

func ListWorkspaceFilesByIDs(ctx context.Context, fileIDs []string) ([]WorkspaceFile, error) {
	if len(fileIDs) == 0 {
		return []WorkspaceFile{}, nil
	}

	normalizedIDs := make([]string, 0, len(fileIDs))
	for _, fileID := range fileIDs {
		trimmed := strings.TrimSpace(fileID)
		if trimmed != "" {
			normalizedIDs = append(normalizedIDs, trimmed)
		}
	}

	if len(normalizedIDs) == 0 {
		return []WorkspaceFile{}, nil
	}

	var files []WorkspaceFile
	err := database.DB.WithContext(ctx).
		Where("id IN ?", normalizedIDs).
		Find(&files).Error
	return files, err
}

func DeleteWorkspaceFilesByIDs(ctx context.Context, fileIDs []string) error {
	if len(fileIDs) == 0 {
		return nil
	}

	normalizedIDs := make([]string, 0, len(fileIDs))
	for _, fileID := range fileIDs {
		trimmed := strings.TrimSpace(fileID)
		if trimmed != "" {
			normalizedIDs = append(normalizedIDs, trimmed)
		}
	}

	if len(normalizedIDs) == 0 {
		return nil
	}

	return database.DB.WithContext(ctx).
		Delete(&WorkspaceFile{}, "id IN ?", normalizedIDs).Error
}
