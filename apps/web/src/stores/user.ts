import { ref } from 'vue'
import { defineStore } from 'pinia'
import { cacheManager, CACHE_TTL } from '@/utils/cache'

const CACHE_KEY_USER = 'deepwrite_user'
/** 用户信息缓存过期时间：24 小时 */
const USER_TTL = CACHE_TTL.ONE_DAY

export interface User {
  id: string
  email: string
  displayName: string
  avatarUrl?: string
  bio?: string
  language?: string
  timezone?: string
  role: string
  status: string
  lastLogin?: string
}

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)

  function setUser(userData: User) {
    user.value = userData
    // 同步用户信息到缓存（带过期时间和开发环境加密）
    cacheManager.set(CACHE_KEY_USER, userData, USER_TTL)
  }

  function clearUser() {
    user.value = null
    cacheManager.remove(CACHE_KEY_USER)
  }

  function initializeFromStorage() {
    const storedUser = cacheManager.get<User>(CACHE_KEY_USER)
    if (storedUser) {
      user.value = storedUser
      return
    }

    user.value = null
  }

  return {
    user,
    setUser,
    clearUser,
    initializeFromStorage,
  }
})
