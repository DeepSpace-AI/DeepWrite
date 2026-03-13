export type MainPaneType = 'writing' | 'literature' | 'analysis' | 'settings'

export interface WorkspaceDocument {
  id: string
  workspace_id: string
  folder_id?: string | null
  title: string
  content_json?: Record<string, unknown> | null
  tiptap_schema?: string
  tiptap_schema_ver?: string
  current_version?: number
  latest_snapshot_version?: number
  latest_snapshot_at?: string | null
  public?: boolean
  created_at?: string
  updated_at: string
}

export interface WorkspaceFolder {
  id: string
  workspace_id: string
  parent_id?: string | null
  name: string
  description?: string
  created_at?: string
  updated_at: string
}

export interface WorkspaceFile {
  id: string
  workspace_id: string
  folder_id?: string | null
  object_key: string
  file_name: string
  content_type: string
  size: number
  etag?: string
  uploaded_by?: string
  created_at?: string
  updated_at: string
  preview_url?: string
}

export interface VisibleFolderNode {
  folder: WorkspaceFolder
  depth: number
  expanded: boolean
}

export interface Collaborator {
  id: string
  name: string
  avatarUrl?: string
  color?: string
}
