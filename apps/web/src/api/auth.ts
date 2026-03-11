import { http, unwrapResponse } from '@/api/http'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'
import type { User } from '@/stores/user'
import { cacheManager, CACHE_TTL } from '@/utils/cache'

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

type SupportedLocale = 'zh-CN' | 'en'

const CACHE_KEY_LOCALE = 'deepwrite_locale'
const CACHE_KEY_DISPLAY_NAME = 'deepwrite_display_name'
const DISPLAY_NAME_TTL = CACHE_TTL.ONE_DAY

function isSupportedLocale(value: unknown): value is SupportedLocale {
  return value === 'zh-CN' || value === 'en'
}

function resolvePreferredLocale(...candidates: Array<string | undefined>): SupportedLocale {
  for (const candidate of candidates) {
    if (isSupportedLocale(candidate)) {
      return candidate
    }
  }

  const cachedLocale = cacheManager.get<string>(CACHE_KEY_LOCALE)
  if (isSupportedLocale(cachedLocale)) {
    return cachedLocale
  }

  return 'zh-CN'
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

export interface UpdateUserPreferencesInput {
  language: string
  timezone: string
}

export interface UpdateCurrentUserProfileInput {
  displayName: string
  avatarUrl: string
  bio: string
  language: string
  timezone: string
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

function mapLoginUserToStoreUser(input: AuthUser, fallbackLocale?: string): User {
  const displayName = (input.display_name || input.displayName || input.email.split('@')[0] || input.email).toString()

  return {
    id: input.id,
    email: input.email,
    displayName,
    avatarUrl: '',
    bio: '',
    language: resolvePreferredLocale(fallbackLocale),
    timezone: 'UTC',
    role: input.role,
    status: input.status,
    lastLogin: '',
  }
}

function mapProfileToStoreUser(input: ProfilePayload, fallbackLocale?: string): User {
  const displayName = (input.user_profile?.display_name || input.email.split('@')[0] || input.email).toString()

  return {
    id: input.id,
    email: input.email,
    displayName,
    avatarUrl: input.user_profile?.avatar_url || '',
    bio: input.user_profile?.bio || '',
    language: resolvePreferredLocale(input.user_profile?.language, fallbackLocale),
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
  const normalizedUser = mapProfileToStoreUser(data, userStore.user?.language)
  userStore.setUser(normalizedUser)

  // 兼容协作页中仍读取的 display_name 本地字段
  cacheManager.set(CACHE_KEY_DISPLAY_NAME, normalizedUser.displayName, DISPLAY_NAME_TTL)

  return normalizedUser
}

export async function updateCurrentUserPreferences(input: UpdateUserPreferencesInput): Promise<User> {
  const userStore = useUserStore()
  const currentUser = userStore.user

  if (!currentUser) {
    throw new Error('用户未登录，无法更新偏好设置')
  }

  return updateCurrentUserProfile({
    displayName: currentUser.displayName,
    avatarUrl: currentUser.avatarUrl || '',
    bio: currentUser.bio || '',
    language: input.language,
    timezone: input.timezone,
  })
}

export async function updateCurrentUserProfile(input: UpdateCurrentUserProfileInput): Promise<User> {
  const userStore = useUserStore()
  const currentUser = userStore.user

  if (!currentUser) {
    throw new Error('用户未登录，无法更新资料')
  }

  // 后端 /user/profile 要求 display_name/avatar_url/bio 为必填，
  // 这里确保所有字段都有值。
  const payload = {
    display_name: (input.displayName || currentUser.email.split('@')[0] || 'deepwrite-user').slice(0, 30),
    avatar_url: input.avatarUrl || 'about:blank',
    bio: input.bio.trim().slice(0, 255),
    language: input.language,
    timezone: input.timezone,
  }

  const res = await http.put('/user/profile', payload)
  unwrapResponse<unknown>(res)

  return fetchCurrentUserProfile()
}

export async function uploadUserAvatar(file: File): Promise<string> {
  const userStore = useUserStore()

  const formData = new FormData()
  formData.append('file', file)

  const res = await http.post('/user/avatar', formData)
  const data = unwrapResponse<{ avatar_url: string }>(res)

  if (userStore.user) {
    userStore.setUser({ ...userStore.user, avatarUrl: data.avatar_url })
  }

  return data.avatar_url
}

export async function saveSession(payload: LoginPayload) {
  const authStore = useAuthStore()
  const userStore = useUserStore()

  authStore.setTokens(payload.access_token, payload.refresh_token)

  // 先用登录返回值快速落地，再以 /user/profile 的真实资料覆盖
  const quickUser = mapLoginUserToStoreUser(payload.user, userStore.user?.language)
  userStore.setUser(quickUser)
  cacheManager.set(CACHE_KEY_DISPLAY_NAME, quickUser.displayName, DISPLAY_NAME_TTL)

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
