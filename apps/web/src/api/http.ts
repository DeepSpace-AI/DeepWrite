import axios, { type AxiosResponse } from 'axios'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'
import { cacheManager } from '@/utils/cache'

declare module 'axios' {
  interface AxiosRequestConfig {
    skipAuthRefresh?: boolean
    _retry?: boolean
  }
}

export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
  time: number
}

export class ApiError extends Error {
  code: number
  status: number

  constructor(message: string, code = -1, status = 0) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
  }
}

export const gatewayBase = (import.meta.env.VITE_GATEWAY_BASE_URL as string | undefined)?.trim() || 'http://localhost:8080'

export const http = axios.create({
  baseURL: `${gatewayBase}/api/v1`,
  timeout: 15000,
  validateStatus: () => true,
})

const refreshHttp = axios.create({
  baseURL: `${gatewayBase}/api/v1`,
  timeout: 15000,
  validateStatus: () => true,
})

let refreshPromise: Promise<string> | null = null
const CACHE_KEY_ACCESS = 'deepwrite_access_token'
const CACHE_KEY_REFRESH = 'deepwrite_refresh_token'

function getAccessToken() {
  const authStore = useAuthStore()
  return authStore.accessToken || cacheManager.get<string>(CACHE_KEY_ACCESS) || ''
}

function getRefreshToken() {
  const authStore = useAuthStore()
  return authStore.refreshToken || cacheManager.get<string>(CACHE_KEY_REFRESH) || ''
}

function persistTokens(accessToken: string, refreshToken: string) {
  const authStore = useAuthStore()
  authStore.setTokens(accessToken, refreshToken)
}

function clearTokens() {
  const authStore = useAuthStore()
  const userStore = useUserStore()
  authStore.clearTokens()
  userStore.clearUser()
}

async function refreshAccessToken(): Promise<string> {
  if (refreshPromise) {
    return refreshPromise
  }

  const refreshToken = getRefreshToken().trim()
  if (!refreshToken) {
    throw new ApiError('登录已过期，请重新登录。', 401, 401)
  }

  refreshPromise = (async () => {
    const res = await refreshHttp.post('/auth/refresh', { refresh_token: refreshToken }, { skipAuthRefresh: true })
    const payload = res.data as ApiEnvelope<{ access_token: string; refresh_token: string }> | undefined

    if (!payload || payload.code !== 200 || !payload.data?.access_token || !payload.data?.refresh_token) {
      throw new ApiError(payload?.message || '会话续期失败，请重新登录。', payload?.code ?? 401, res.status)
    }

    persistTokens(payload.data.access_token, payload.data.refresh_token)
    return payload.data.access_token
  })().finally(() => {
    refreshPromise = null
  })

  return refreshPromise
}

http.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(async (res) => {
  const payload = res.data as Partial<ApiEnvelope<unknown>> | undefined
  const code = typeof payload?.code === 'number' ? payload.code : undefined
  const config = res.config

  const isUnauthorized = res.status === 401 || code === 401
  const canRetry = !config.skipAuthRefresh && !config._retry

  if (isUnauthorized && canRetry) {
    try {
      const newAccessToken = await refreshAccessToken()
      config._retry = true
      config.headers.Authorization = `Bearer ${newAccessToken}`
      return http(config)
    } catch (error) {
      clearTokens()
      if (error instanceof ApiError) {
        throw error
      }
      throw new ApiError('登录已过期，请重新登录。', 401, 401)
    }
  }

  return res
})

export function unwrapResponse<T>(res: AxiosResponse<ApiEnvelope<T>>): T {
  const payload = res.data
  if (!payload || typeof payload.code !== 'number') {
    throw new ApiError('响应格式错误', -1, res.status)
  }

  if (payload.code === 200 || payload.code === 201) {
    return payload.data
  }

  throw new ApiError(payload.message || '请求失败', payload.code, res.status)
}
