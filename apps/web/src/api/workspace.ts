import { http, unwrapResponse } from '@/api/http'
import type { WorkspaceFile, WorkspaceFolder } from '@/views/workspace/types'

export interface WorkspaceMember {
  workspace_id: string
  user_id: string
  role: 'owner' | 'admin' | 'editor' | 'viewer' | string
}

export interface WorkspaceMemberDetail {
  user_id: string
  role: 'owner' | 'admin' | 'editor' | 'viewer' | string
  email?: string
  display_name?: string
  avatar_url?: string
  status?: string
  is_workspace_owner?: boolean
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

export interface UpdateWorkspaceFileInput {
  folder_id?: string | null
  clear_folder?: boolean
  file_name?: string
}

export interface WorkspaceFileDetailResponse {
  file: WorkspaceFile
  preview_url?: string
}

export interface BatchDeleteWorkspaceFilesResult {
  deleted_ids: string[]
  failed: Array<{ file_id: string; reason: string }>
}

export type WorkspaceInvitationRole = 'admin' | 'editor' | 'viewer'
export type WorkspaceMemberRole = 'admin' | 'editor' | 'viewer'

export interface WorkspaceInvitation {
  id: string
  workspace_id: string
  inviter_id: string
  invitee_user_id?: string
  invitee_email: string
  role: WorkspaceInvitationRole | string
  status: 'pending' | 'accepted' | 'rejected' | 'expired' | 'revoked' | string
  expires_at: string
  accepted_at?: string
  rejected_at?: string
  revoked_at?: string
  created_at: string
  updated_at: string
  action_token?: string
}

export interface CreateWorkspaceInvitationInput {
  invitee_user_id?: string
  invitee_email?: string
  role: WorkspaceInvitationRole
}

export interface CreateWorkspaceInvitationResult {
  invitation: WorkspaceInvitation
  token: string
  reused: boolean
}

export interface ResolveWorkspaceInvitationInput {
  token?: string
  actionToken?: string
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

export async function updateWorkspaceFile(workspaceId: string, fileId: string, input: UpdateWorkspaceFileInput): Promise<WorkspaceFile> {
  const res = await http.put(`/workspaces/${workspaceId}/files/${fileId}`, {
    folder_id: input.folder_id || undefined,
    clear_folder: input.clear_folder ?? false,
    file_name: input.file_name,
  })
  return unwrapResponse<WorkspaceFile>(res)
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

export async function listWorkspaceInvitations(workspaceId: string, status?: string): Promise<WorkspaceInvitation[]> {
  const res = await http.get(`/workspaces/${workspaceId}/invitations`, {
    params: {
      status: status || undefined,
      limit: 100,
      offset: 0,
    },
  })
  return unwrapResponse<WorkspaceInvitation[]>(res)
}

export async function listWorkspaceMembers(workspaceId: string): Promise<WorkspaceMemberDetail[]> {
  const res = await http.get(`/workspaces/${workspaceId}/members`)
  return unwrapResponse<WorkspaceMemberDetail[]>(res)
}

export async function updateWorkspaceMemberRole(workspaceId: string, userId: string, role: WorkspaceMemberRole): Promise<WorkspaceMemberDetail> {
  const res = await http.put(`/workspaces/${workspaceId}/members/${userId}`, {
    role,
  })
  return unwrapResponse<WorkspaceMemberDetail>(res)
}

export async function removeWorkspaceMember(workspaceId: string, userId: string): Promise<{ user_id: string }> {
  const res = await http.delete(`/workspaces/${workspaceId}/members/${userId}`)
  return unwrapResponse<{ user_id: string }>(res)
}

export async function createWorkspaceInvitation(workspaceId: string, input: CreateWorkspaceInvitationInput): Promise<CreateWorkspaceInvitationResult> {
  const res = await http.post(`/workspaces/${workspaceId}/invitations`, {
    invitee_user_id: input.invitee_user_id || undefined,
    invitee_email: input.invitee_email || undefined,
    role: input.role,
  })
  return unwrapResponse<CreateWorkspaceInvitationResult>(res)
}

export async function revokeWorkspaceInvitation(workspaceId: string, invitationId: string): Promise<WorkspaceInvitation> {
  const res = await http.post(`/workspaces/${workspaceId}/invitations/${invitationId}/revoke`)
  return unwrapResponse<WorkspaceInvitation>(res)
}

export async function listMyWorkspaceInvitations(status?: string): Promise<WorkspaceInvitation[]> {
  const res = await http.get('/invitations/me', {
    params: {
      status: status || undefined,
      limit: 100,
      offset: 0,
    },
  })
  return unwrapResponse<WorkspaceInvitation[]>(res)
}

export async function acceptWorkspaceInvitation(invitationId: string, input: ResolveWorkspaceInvitationInput): Promise<WorkspaceInvitation> {
  const res = await http.post(`/invitations/${invitationId}/accept`, {
    token: input.token || undefined,
    action_token: input.actionToken || undefined,
  })
  return unwrapResponse<WorkspaceInvitation>(res)
}

export async function rejectWorkspaceInvitation(invitationId: string, input: ResolveWorkspaceInvitationInput): Promise<WorkspaceInvitation> {
  const res = await http.post(`/invitations/${invitationId}/reject`, {
    token: input.token || undefined,
    action_token: input.actionToken || undefined,
  })
  return unwrapResponse<WorkspaceInvitation>(res)
}
