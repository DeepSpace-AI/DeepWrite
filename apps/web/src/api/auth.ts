import { http, unwrapResponse } from '@/api/http'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'
import type { User } from '@/stores/user'

interface AuthUser {
  id: string
  email: string
  display_name?: string
  displayName?: string
  role: string
  status: string
}

interface UserProfile {
  id: string
  user_id: string
  display_name: string
  avatar_url: string
  bio: string
  language: string
  timezone: string
  created_at: string
  updated_at: string
}

interface ProfilePayload {
  id: string
  email: string
  user_profile: UserProfile
  role: string
  status: string
  last_login: string
  created_at: string
  updated_at: string
}

interface LoginPayload {
  access_token: string
  refresh_token: string
  token_type: string
  user: AuthUser
}

export interface LoginInput {
  email: string
  password: string
}

export interface RegisterInput {
  displayName: string
  email: string
  password: string
}

export interface ForgotPasswordInput {
  email: string
}

export async function login(input: LoginInput): Promise<LoginPayload> {
  const res = await http.post('/auth/login', {
    email: input.email,
    password: input.password,
  }, { skipAuthRefresh: true })
  return unwrapResponse<LoginPayload>(res)
}

export async function register(input: RegisterInput): Promise<void> {
  const res = await http.post('/auth/register', {
    display_name: input.displayName,
    email: input.email,
    password: input.password,
  }, { skipAuthRefresh: true })
  unwrapResponse(res)
}

export async function forgotPassword(input: ForgotPasswordInput): Promise<string> {
  const res = await http.post('/user/forgot-password', {
    email: input.email,
  }, { skipAuthRefresh: true })
  const data = unwrapResponse<{ message?: string }>(res)
  return data.message || '如果邮箱存在，我们已发送重置指引'
}

function mapLoginUserToStoreUser(input: AuthUser): User {
  const displayName = (input.display_name || input.displayName || input.email.split('@')[0] || input.email).toString()

  return {
    id: input.id,
    email: input.email,
    displayName,
    avatarUrl: '',
    bio: '',
    language: 'en',
    timezone: 'UTC',
    role: input.role,
    status: input.status,
    lastLogin: '',
  }
}

function mapProfileToStoreUser(input: ProfilePayload): User {
  const displayName = (input.user_profile?.display_name || input.email.split('@')[0] || input.email).toString()

  return {
    id: input.id,
    email: input.email,
    displayName,
    avatarUrl: input.user_profile?.avatar_url || '',
    bio: input.user_profile?.bio || '',
    language: input.user_profile?.language || 'en',
    timezone: input.user_profile?.timezone || 'UTC',
    role: input.role,
    status: input.status,
    lastLogin: input.last_login,
  }
}

export async function fetchCurrentUserProfile(): Promise<User> {
  const res = await http.get('/user/profile')
  const data = unwrapResponse<ProfilePayload>(res)

  const userStore = useUserStore()
  const normalizedUser = mapProfileToStoreUser(data)
  userStore.setUser(normalizedUser)

  // 兼容协作页中仍读取的 display_name 本地字段
  localStorage.setItem('deepwrite_display_name', normalizedUser.displayName)

  return normalizedUser
}

export async function saveSession(payload: LoginPayload) {
  const authStore = useAuthStore()
  const userStore = useUserStore()

  authStore.setTokens(payload.access_token, payload.refresh_token)

  // 先用登录返回值快速落地，再以 /user/profile 的真实资料覆盖
  const quickUser = mapLoginUserToStoreUser(payload.user)
  userStore.setUser(quickUser)
  localStorage.setItem('deepwrite_display_name', quickUser.displayName)

  try {
    await fetchCurrentUserProfile()
  } catch {
    // 网络波动时保留登录返回的兜底用户信息
  }
}

export function logout() {
  const authStore = useAuthStore()
  const userStore = useUserStore()

  authStore.clearTokens()
  userStore.clearUser()
}
