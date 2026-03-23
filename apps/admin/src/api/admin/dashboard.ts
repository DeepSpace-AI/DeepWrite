import { http, unwrapResponse } from '@/api/http'

export interface AdminStats {
  total_users: number
  total_workspaces: number
  total_documents: number
  total_files: number
  active_workspaces: number
  active_users_today: number
  storage_used_bytes: number
}

export async function fetchAdminStats(): Promise<AdminStats> {
  const res = await http.get('/admin/stats')
  return unwrapResponse<AdminStats>(res)
}
