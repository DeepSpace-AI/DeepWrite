import { http, unwrapResponse } from '@/api/http'

export interface Workspace {
  id: string
  name: string
  status: string
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

export type WorkspaceMemberRole = 'admin' | 'editor' | 'viewer'

export async function listWorkspaces(limit = 100, offset = 0): Promise<Workspace[]> {
  const res = await http.get('/workspaces', {
    params: { limit, offset },
  })
  return unwrapResponse<Workspace[]>(res)
}

export async function listWorkspaceMembers(workspaceId: string): Promise<WorkspaceMemberDetail[]> {
  const res = await http.get(`/workspaces/${workspaceId}/members`)
  return unwrapResponse<WorkspaceMemberDetail[]>(res)
}

export async function updateWorkspaceMemberRole(workspaceId: string, userId: string, role: WorkspaceMemberRole): Promise<WorkspaceMemberDetail> {
  const res = await http.put(`/workspaces/${workspaceId}/members/${userId}`, { role })
  return unwrapResponse<WorkspaceMemberDetail>(res)
}

export async function removeWorkspaceMember(workspaceId: string, userId: string): Promise<{ user_id: string }> {
  const res = await http.delete(`/workspaces/${workspaceId}/members/${userId}`)
  return unwrapResponse<{ user_id: string }>(res)
}
