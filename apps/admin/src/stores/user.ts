import { ref } from 'vue'
import { defineStore } from 'pinia'
import { cacheManager, CACHE_TTL } from '@/utils/cache'

const CACHE_KEY_USER = 'deepwrite_admin_user'
const USER_TTL = CACHE_TTL.ONE_DAY

export interface User {
  id: string
  email: string
  displayName: string
  avatarUrl?: string
  role: string
  status: string
}

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)

  function setUser(userData: User) {
    user.value = userData
    cacheManager.set(CACHE_KEY_USER, userData, USER_TTL)
  }

  function clearUser() {
    user.value = null
    cacheManager.remove(CACHE_KEY_USER)
  }

  function initializeFromStorage() {
    const storedUser = cacheManager.get<User>(CACHE_KEY_USER)
    user.value = storedUser || null
  }

  return {
    user,
    setUser,
    clearUser,
    initializeFromStorage,
  }
})
