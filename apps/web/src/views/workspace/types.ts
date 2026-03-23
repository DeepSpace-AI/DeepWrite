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

export type AnnotationType = 'highlight' | 'note' | 'drawing'

export interface PDFAnnotation {
  id: string
  file_id: string
  user_id: string
  workspace_id: string
  type: AnnotationType
  page: number
  rect_x: number
  rect_y: number
  rect_width: number
  rect_height: number
  color?: string
  content?: string
  paths?: DrawingPath[]
  created_at: string
  updated_at: string
}

export interface DrawingPath {
  points: Array<{ x: number; y: number }>
  color: string
  width: number
}

export interface PDFPageAnnotation extends PDFAnnotation {
  absoluteRect: { x: number; y: number; width: number; height: number }
}
