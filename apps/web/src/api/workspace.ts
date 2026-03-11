import { http, unwrapResponse } from '@/api/http'
import type { WorkspaceFile, WorkspaceFolder } from '@/views/workspace/types'

export interface WorkspaceMember {
  workspace_id: string
  user_id: string
  role: 'owner' | 'admin' | 'editor' | 'viewer' | string
}

export interface Workspace {
  id: string
  name: string
  description: string
  owner_id: string
  members?: WorkspaceMember[]
  public: boolean
  status: 'active' | 'archived' | string
  created_at: string
  updated_at: string
}

export interface CreateWorkspaceInput {
  name: string
  description?: string
  public?: boolean
}

export interface UpdateWorkspaceInput {
  name?: string
  description?: string
  public?: boolean
  status?: 'active' | 'archived'
}

export interface ListFoldersParams {
  parent_id?: string
  limit?: number
  offset?: number
}

export interface CreateFolderInput {
  parent_id?: string | null
  name: string
  description?: string
}

export interface UpdateFolderInput {
  name?: string
  description?: string
}

export interface ListFilesParams {
  folder_id?: string
  limit?: number
  offset?: number
}

export interface PresignWorkspaceUploadInput {
  folder_id?: string | null
  file_name: string
  content_type?: string
  expires_in?: number
}

export interface WorkspaceUploadTicket {
  workspace_id: string
  folder_id?: string | null
  filename: string
  content_type: string
  object_key: string
  upload_method: 'PUT' | string
  upload_url: string
  preview_url?: string
}

export interface CompleteWorkspaceUploadInput {
  folder_id?: string | null
  object_key: string
  file_name: string
}

export interface WorkspaceFileDetailResponse {
  file: WorkspaceFile
  preview_url?: string
}

export interface BatchDeleteWorkspaceFilesResult {
  deleted_ids: string[]
  failed: Array<{ file_id: string; reason: string }>
}

export async function listWorkspaces(limit = 100, offset = 0): Promise<Workspace[]> {
  const res = await http.get('/workspaces', {
    params: { limit, offset },
  })
  return unwrapResponse<Workspace[]>(res)
}

export async function createWorkspace(input: CreateWorkspaceInput): Promise<Workspace> {
  const res = await http.post('/workspaces', {
    name: input.name,
    description: input.description || '',
    public: !!input.public,
  })
  return unwrapResponse<Workspace>(res)
}

export async function updateWorkspace(id: string, input: UpdateWorkspaceInput): Promise<Workspace> {
  const res = await http.put(`/workspaces/${id}`, input)
  return unwrapResponse<Workspace>(res)
}

export async function deleteWorkspace(id: string): Promise<void> {
  const res = await http.delete(`/workspaces/${id}`)
  unwrapResponse<unknown>(res)
}

export async function listWorkspaceFolders(workspaceId: string, params: ListFoldersParams = {}): Promise<WorkspaceFolder[]> {
  const res = await http.get(`/workspaces/${workspaceId}/folders`, {
    params: {
      parent_id: params.parent_id,
      limit: params.limit ?? 100,
      offset: params.offset ?? 0,
    },
  })
  return unwrapResponse<WorkspaceFolder[]>(res)
}

export async function createWorkspaceFolder(workspaceId: string, input: CreateFolderInput): Promise<WorkspaceFolder> {
  const res = await http.post(`/workspaces/${workspaceId}/folders`, {
    parent_id: input.parent_id || undefined,
    name: input.name,
    description: input.description || '',
  })
  return unwrapResponse<WorkspaceFolder>(res)
}

export async function updateWorkspaceFolder(workspaceId: string, folderId: string, input: UpdateFolderInput): Promise<WorkspaceFolder> {
  const res = await http.put(`/workspaces/${workspaceId}/folders/${folderId}`, {
    name: input.name,
    description: input.description || '',
  })
  return unwrapResponse<WorkspaceFolder>(res)
}

export async function deleteWorkspaceFolder(workspaceId: string, folderId: string): Promise<void> {
  const res = await http.delete(`/workspaces/${workspaceId}/folders/${folderId}`)
  unwrapResponse<unknown>(res)
}

export async function listWorkspaceFiles(workspaceId: string, params: ListFilesParams = {}): Promise<WorkspaceFile[]> {
  const res = await http.get(`/workspaces/${workspaceId}/files`, {
    params: {
      folder_id: params.folder_id,
      limit: params.limit ?? 100,
      offset: params.offset ?? 0,
    },
  })
  return unwrapResponse<WorkspaceFile[]>(res)
}

export async function getWorkspaceFileDetail(workspaceId: string, fileId: string): Promise<WorkspaceFileDetailResponse> {
  const res = await http.get(`/workspaces/${workspaceId}/files/${fileId}`)
  return unwrapResponse<WorkspaceFileDetailResponse>(res)
}

export async function presignWorkspaceUpload(workspaceId: string, input: PresignWorkspaceUploadInput): Promise<WorkspaceUploadTicket> {
  const res = await http.post(`/workspaces/${workspaceId}/files`, {
    folder_id: input.folder_id || undefined,
    file_name: input.file_name,
    content_type: input.content_type || undefined,
    expires_in: input.expires_in,
  })
  return unwrapResponse<WorkspaceUploadTicket>(res)
}

export async function completeWorkspaceUpload(workspaceId: string, input: CompleteWorkspaceUploadInput): Promise<WorkspaceFileDetailResponse> {
  const res = await http.post(`/workspaces/${workspaceId}/files/complete`, {
    folder_id: input.folder_id || undefined,
    object_key: input.object_key,
    file_name: input.file_name,
  })
  return unwrapResponse<WorkspaceFileDetailResponse>(res)
}

export async function deleteWorkspaceFile(workspaceId: string, fileId: string): Promise<void> {
  const res = await http.delete(`/workspaces/${workspaceId}/files/${fileId}`)
  unwrapResponse<unknown>(res)
}

export async function batchDeleteWorkspaceFiles(workspaceId: string, fileIds: string[]): Promise<BatchDeleteWorkspaceFilesResult> {
  const res = await http.post(`/workspaces/${workspaceId}/files/batch-delete`, {
    file_ids: fileIds,
  })
  return unwrapResponse<BatchDeleteWorkspaceFilesResult>(res)
}
