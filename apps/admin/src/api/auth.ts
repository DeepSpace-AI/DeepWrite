import { http, unwrapResponse, ApiError } from '@/api/http'
import { useAuthStore } from '@/stores/auth'
import { useUserStore, type User } from '@/stores/user'

interface AuthUser {
  id: string
  email: string
  display_name?: string
  displayName?: string
  role: string
  status: string
}

interface ProfilePayload {
  id: string
  email: string
  role: string
  status: string
  user_profile?: {
    display_name?: string
    avatar_url?: string
  }
}

interface LoginPayload {
  access_token: string
  refresh_token: string
  token_type: string
  scope?: string
  user: AuthUser
}

export interface LoginInput {
  email: string
  password: string
}

function mapLoginUserToStoreUser(input: AuthUser): User {
  const displayName = (input.display_name || input.displayName || input.email.split('@')[0] || input.email).toString()
  return {
    id: input.id,
    email: input.email,
    displayName,
    avatarUrl: '',
    role: input.role,
    status: input.status,
  }
}

function mapProfileToStoreUser(input: ProfilePayload): User {
  const displayName = (input.user_profile?.display_name || input.email.split('@')[0] || input.email).toString()
  return {
    id: input.id,
    email: input.email,
    displayName,
    avatarUrl: input.user_profile?.avatar_url || '',
    role: input.role,
    status: input.status,
  }
}

export async function login(input: LoginInput): Promise<LoginPayload> {
  const res = await http.post('/auth/admin/login', {
    email: input.email,
    password: input.password,
  }, { skipAuthRefresh: true })
  return unwrapResponse<LoginPayload>(res)
}

export async function fetchCurrentUserProfile(): Promise<User> {
  const res = await http.get('/user/profile')
  const data = unwrapResponse<ProfilePayload>(res)
  const userStore = useUserStore()
  const normalizedUser = mapProfileToStoreUser(data)
  userStore.setUser(normalizedUser)
  return normalizedUser
}

export async function saveSession(payload: LoginPayload) {
  if (payload.scope !== 'admin') {
    logout()
    throw new ApiError('当前登录未签发管理员专用令牌。', 403, 403)
  }

  const authStore = useAuthStore()
  const userStore = useUserStore()
  authStore.setTokens(payload.access_token, payload.refresh_token)

  const quickUser = mapLoginUserToStoreUser(payload.user)
  userStore.setUser(quickUser)

  const profile = await fetchCurrentUserProfile().catch(() => quickUser)
  if (profile.role.toLowerCase() !== 'admin') {
    logout()
    throw new ApiError('当前账号不是管理员，无法访问后台。', 403, 403)
  }
}

export function logout() {
  const authStore = useAuthStore()
  const userStore = useUserStore()
  authStore.clearTokens()
  userStore.clearUser()
}
