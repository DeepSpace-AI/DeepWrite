import { http, unwrapResponse } from '@/api/http'

export interface DashboardSummary {
  workspace_count: number
  active_workspace_count: number
  document_count: number
  file_count: number
  pending_invitation_count: number
  collaborator_count: number
}

export interface DashboardWorkspaceItem {
  id: string
  name: string
  description: string
  status: string
  role: string
  member_count: number
  document_count: number
  file_count: number
  updated_at: string
  created_at: string
  public: boolean
  owner_id: string
}

export interface DashboardDocumentItem {
  id: string
  workspace_id: string
  workspace_name: string
  title: string
  current_version: number
  updated_at: string
  created_at: string
}

export interface DashboardInvitationItem {
  id: string
  workspace_id: string
  workspace_name: string
  invitee_email: string
  role: string
  status: string
  expires_at: string
  created_at: string
}

export interface DashboardOverview {
  summary: DashboardSummary
  workspaces: DashboardWorkspaceItem[]
  recent_documents: DashboardDocumentItem[]
  pending_invitations: DashboardInvitationItem[]
}

export async function fetchDashboardOverview(): Promise<DashboardOverview> {
  const res = await http.get('/dashboard/overview')
  return unwrapResponse<DashboardOverview>(res)
}
