package request

type CreateDocumentRequest struct {
	WorkspaceID     string         `json:"workspace_id" binding:"required,uuid"`
	FolderID        string         `json:"folder_id" binding:"omitempty,uuid"`
	Title           string         `json:"title" binding:"required,max=255"`
	ContentJSON     map[string]any `json:"content_json"`
	TiptapSchema    string         `json:"tiptap_schema"`
	TiptapSchemaVer string         `json:"tiptap_schema_ver"`
}

type SaveDocumentVersionRequest struct {
	Title       string         `json:"title"`
	ContentJSON map[string]any `json:"content_json"`
	Source      string         `json:"source"`
	Snapshot    bool           `json:"snapshot"`
	Summary     string         `json:"summary"`
}

type UpdateDocumentMetaRequest struct {
	Title       string `json:"title" binding:"omitempty,max=255"`
	FolderID    string `json:"folder_id" binding:"omitempty,uuid"`
	ClearFolder bool   `json:"clear_folder"`
}

type RestoreDocumentVersionRequest struct {
	Version int64 `json:"version" binding:"required,min=1"`
}

type SyncCollabContentRequest struct {
	Title       string         `json:"title"`
	ContentJSON map[string]any `json:"content_json" binding:"required"`
}
