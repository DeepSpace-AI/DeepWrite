import { http, unwrapResponse } from '@/api/http'

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
