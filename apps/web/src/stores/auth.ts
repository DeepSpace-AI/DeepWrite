import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { cacheManager, CACHE_TTL } from '@/utils/cache'

const CACHE_KEY_ACCESS = 'deepwrite_access_token'
const CACHE_KEY_REFRESH = 'deepwrite_refresh_token'
/** Token 的缓存过期时间：24 小时 */
const TOKEN_TTL = CACHE_TTL.ONE_DAY

function decodeBase64Url(input: string): string {
  const normalized = input.replace(/-/g, '+').replace(/_/g, '/')
  const paddingLength = (4 - (normalized.length % 4)) % 4
  return atob(`${normalized}${'='.repeat(paddingLength)}`)
}

function isJwtExpired(token: string, skewSeconds = 30): boolean {
  try {
    const parts = token.split('.')
    const payloadPart = parts[1]
    if (!payloadPart) {
      return true
    }

    const payload = JSON.parse(decodeBase64Url(payloadPart)) as { exp?: number }
    if (typeof payload.exp !== 'number') {
      return true
    }

    const nowInSeconds = Math.floor(Date.now() / 1000)
    return payload.exp <= nowInSeconds + skewSeconds
  } catch {
    return true
  }
}

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string>('')
  const refreshToken = ref<string>('')

  const isAuthenticated = computed(() => !!accessToken.value)

  function setTokens(access: string, refresh: string) {
    accessToken.value = access
    refreshToken.value = refresh
    // 同步到缓存（带过期时间和开发环境加密）
    cacheManager.set(CACHE_KEY_ACCESS, access, TOKEN_TTL)
    cacheManager.set(CACHE_KEY_REFRESH, refresh, TOKEN_TTL)
  }

  function setAccessToken(token: string) {
    accessToken.value = token
    cacheManager.set(CACHE_KEY_ACCESS, token, TOKEN_TTL)
  }

  function clearTokens() {
    accessToken.value = ''
    refreshToken.value = ''
    cacheManager.removeMultiple([CACHE_KEY_ACCESS, CACHE_KEY_REFRESH])
  }

  function getEffectiveAccessToken() {
    return accessToken.value || cacheManager.get<string>(CACHE_KEY_ACCESS) || ''
  }

  function hasValidAccessToken() {
    const token = getEffectiveAccessToken().trim()
    if (!token) {
      return false
    }

    return !isJwtExpired(token)
  }

  function initializeFromStorage() {
    const storedAccess = cacheManager.get<string>(CACHE_KEY_ACCESS) || ''
    const storedRefresh = cacheManager.get<string>(CACHE_KEY_REFRESH) || ''
    accessToken.value = storedAccess
    refreshToken.value = storedRefresh
  }

  return {
    accessToken,
    refreshToken,
    isAuthenticated,
    setTokens,
    setAccessToken,
    clearTokens,
    getEffectiveAccessToken,
    hasValidAccessToken,
    initializeFromStorage,
  }
})
