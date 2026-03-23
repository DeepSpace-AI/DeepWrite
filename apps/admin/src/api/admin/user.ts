import { http, unwrapResponse } from '@/api/http'

export interface UserProfile {
  id: string
  email: string
  display_name: string
  avatar_url?: string
  role?: string
  status?: string
  created_at?: string
  updated_at?: string
}

export interface UserListResponse {
  users: UserProfile[]
  total: number
  limit: number
  offset: number
}

export interface UpdateUserRequest {
  display_name?: string
  role?: string
  status?: string
}

export async function fetchUsers(params?: {
  limit?: number
  offset?: number
  keyword?: string
  status?: string
}): Promise<UserListResponse> {
  const res = await http.get('/admin/users', { params })
  return unwrapResponse<UserListResponse>(res)
}

export async function fetchUser(userId: string): Promise<UserProfile> {
  const res = await http.get(`/admin/users/${userId}`)
  return unwrapResponse<UserProfile>(res)
}

export async function updateUser(userId: string, data: UpdateUserRequest): Promise<UserProfile> {
  const res = await http.patch(`/admin/users/${userId}`, data)
  return unwrapResponse<UserProfile>(res)
}

export async function deleteUser(userId: string): Promise<void> {
  const res = await http.delete(`/admin/users/${userId}`)
  unwrapResponse<void>(res)
}

export interface ResetPasswordResponse {
  user_id: string
  email: string
  new_password: string
}

export async function resetUserPassword(userId: string): Promise<ResetPasswordResponse> {
  const res = await http.post(`/admin/users/${userId}/reset-password`)
  return unwrapResponse<ResetPasswordResponse>(res)
}
