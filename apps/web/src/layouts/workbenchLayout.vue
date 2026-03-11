<script setup lang="ts">
import { computed, onBeforeMount, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { logout } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const authStore = useAuthStore()
const userStore = useUserStore()

const { user } = storeToRefs(userStore)
const selectedLanguage = ref('en')
const selectedTimezone = ref('UTC')
const isDark = ref(false)

const THEME_KEY = 'deepwrite_theme'

const languageOptions = [
  { label: 'English', value: 'en' },
  { label: '简体中文', value: 'zh-CN' },
]

const timezoneOptions = [
  { label: 'UTC', value: 'UTC' },
  { label: 'Asia/Shanghai', value: 'Asia/Shanghai' },
  { label: 'Asia/Tokyo', value: 'Asia/Tokyo' },
  { label: 'Europe/Berlin', value: 'Europe/Berlin' },
  { label: 'America/New_York', value: 'America/New_York' },
]

const userInitial = computed(() => {
  const name = user.value?.displayName || user.value?.email || ''
  return name ? name.slice(0, 1).toUpperCase() : 'D'
})

function applyTheme(darkMode: boolean) {
  const theme = darkMode ? 'forest' : 'bumblebee'
  document.documentElement.setAttribute('data-theme', theme)
  localStorage.setItem(THEME_KEY, theme)
  isDark.value = darkMode
}

function updateUserPreferences() {
  if (!user.value) return

  userStore.setUser({
    ...user.value,
    language: selectedLanguage.value,
    timezone: selectedTimezone.value,
  })
}

function openProfileCenter() {
  router.push({ name: 'dashboard', query: { tab: 'profile' } })
}

function openSystemSettings() {
  router.push({ name: 'dashboard', query: { tab: 'settings' } })
}

function openSubscription() {
  router.push({ name: 'dashboard', query: { tab: 'subscription' } })
}

async function handleLogout() {
  logout()
  await router.push({ name: 'login' })
}

watch(
  user,
  (nextUser) => {
    selectedLanguage.value = nextUser?.language || 'en'
    selectedTimezone.value = nextUser?.timezone || 'UTC'
  },
  { immediate: true },
)

onBeforeMount(() => {
  authStore.initializeFromStorage()
  userStore.initializeFromStorage()

  const theme = localStorage.getItem(THEME_KEY)
  if (theme === 'forest') {
    applyTheme(true)
  } else if (theme === 'bumblebee') {
    applyTheme(false)
  } else {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    applyTheme(prefersDark)
  }
})
</script>

<template>
  <section class="dot-grid relative min-h-screen overflow-hidden bg-base-100 min-w-screen">
    <div class="pointer-events-none absolute -left-28 top-0 h-72 w-72 rounded-full bg-secondary/10 blur-3xl" />
    <div class="pointer-events-none absolute -right-24 bottom-0 h-80 w-80 rounded-full bg-primary/10 blur-3xl" />

    <div class="relative mx-auto px-4 py-4 lg:px-5">
      <aside class="rounded-sm border border-base-300 bg-base-100/95 shadow-sm backdrop-blur lg:fixed lg:left-5 lg:top-4 lg:z-20 lg:flex lg:h-[calc(100vh-2rem)] lg:w-[280px] lg:flex-col">
        <div class="border-b border-base-300 px-5 py-4">
          <h1 class="heading-serif mt-2 text-xl font-bold text-base-content">DeepWrite 工作台</h1>
        </div>

        <nav class="px-3 py-3 lg:flex-1 lg:overflow-y-auto">
          <RouterLink :to="{ name: 'dashboard' }" class="group mb-1 flex items-center gap-3 rounded-sm border border-base-300/70 bg-base-200 px-3 py-2 text-sm text-base-content">
            <span class="inline-flex h-7 w-7 items-center justify-center rounded-sm bg-primary/10 text-primary">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" class="h-4 w-4">
                <path d="M4 5h16v4H4z" />
                <path d="M4 11h7v8H4z" />
                <path d="M13 11h7v3h-7z" />
                <path d="M13 16h7v3h-7z" />
              </svg>
            </span>
            <div>
              <p class="font-medium">工作台</p>
              <p class="text-xs text-base-content/50">总览与任务视图</p>
            </div>
          </RouterLink>

          <RouterLink :to="{ name: 'workspace-list' }" class="group mb-1 flex items-center gap-3 rounded-sm px-3 py-2 text-sm text-base-content/78 transition hover:bg-base-200">
            <span class="inline-flex h-7 w-7 items-center justify-center rounded-sm bg-primary/10 text-primary">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" class="h-4 w-4">
                <path d="M3 10.5 12 3l9 7.5" />
                <path d="M5.5 9.5V20h13V9.5" />
              </svg>
            </span>
            <div>
              <p class="font-medium">工作空间</p>
              <p class="text-xs text-base-content/50">项目与协作房间</p>
            </div>
          </RouterLink>

          <button class="group mb-1 flex w-full items-center gap-3 rounded-sm px-3 py-2 text-left text-sm text-base-content/78 transition hover:bg-base-200">
            <span class="inline-flex h-7 w-7 items-center justify-center rounded-sm bg-secondary/10 text-secondary">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" class="h-4 w-4">
                <path d="M6.5 4.5h11a1 1 0 0 1 1 1v13l-3-2-3 2-3-2-3 2v-13a1 1 0 0 1 1-1Z" />
              </svg>
            </span>
            <div>
              <p class="font-medium">文献库</p>
              <p class="text-xs text-base-content/50">资料、摘要与引用</p>
            </div>
          </button>

          <button class="group mb-1 flex w-full items-center gap-3 rounded-sm px-3 py-2 text-left text-sm text-base-content/78 transition hover:bg-base-200">
            <span class="inline-flex h-7 w-7 items-center justify-center rounded-sm bg-accent/10 text-accent">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" class="h-4 w-4">
                <path d="M4 6.5h16M4 12h16M4 17.5h16" />
                <path d="M8.5 4v16M15.5 4v16" />
              </svg>
            </span>
            <div>
              <p class="font-medium">技能市场</p>
              <p class="text-xs text-base-content/50">模板、提示词与插件</p>
            </div>
          </button>

          <button class="group flex w-full items-center gap-3 rounded-sm px-3 py-2 text-left text-sm text-base-content/78 transition hover:bg-base-200">
            <span class="inline-flex h-7 w-7 items-center justify-center rounded-sm bg-primary/10 text-primary">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" class="h-4 w-4">
                <path d="M8 10.5a4 4 0 1 1 8 0V13a4 4 0 0 1-8 0v-2.5Z" />
                <path d="M12 17v3" />
                <path d="M7 20h10" />
                <path d="M4.5 8.5 7 10M19.5 8.5 17 10" />
              </svg>
            </span>
            <div>
              <p class="font-medium">Agent 助手</p>
              <p class="text-xs text-base-content/50">智能协作与编写建议</p>
            </div>
          </button>
        </nav>

        <div class="mx-3 mt-auto border-t border-base-300" />

        <div class="mx-3 mb-3 mt-3 rounded-sm border border-base-300 bg-base-200/70 p-3">
          <div class="flex items-start gap-3">
            <img
              v-if="user?.avatarUrl"
              :src="user.avatarUrl"
              :alt="`${user.displayName} avatar`"
              class="h-10 w-10 rounded-sm border border-base-300 object-cover"
            />
            <div v-else class="inline-flex h-10 w-10 items-center justify-center rounded-sm bg-primary text-primary-content font-semibold">
              {{ userInitial }}
            </div>
            <div class="min-w-0">
              <p class="truncate text-sm font-medium text-base-content">{{ user?.displayName || '访客用户' }}</p>
              <p class="truncate text-xs text-base-content/55">{{ user?.email || 'guest@deepwrite.local' }}</p>
            </div>

            <div class="dropdown dropdown-end dropdown-top ml-auto">
              <button tabindex="0" class="btn btn-ghost btn-xs rounded-sm" aria-label="打开用户菜单">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" class="h-4 w-4">
                  <path d="M5 12h14M12 5v14" />
                </svg>
              </button>
              <div tabindex="0" class="dropdown-content z-[30] mt-2 w-64 rounded-sm border border-base-300 bg-base-100 p-2 shadow-lg">
                <button type="button" class="btn btn-ghost btn-sm justify-start rounded-sm" @click="openProfileCenter">
                  个人中心
                </button>
                <button type="button" class="btn btn-ghost btn-sm justify-start rounded-sm" @click="openSystemSettings">
                  系统设置
                </button>
                <button type="button" class="btn btn-ghost btn-sm justify-start rounded-sm" @click="openSubscription">
                  套餐订阅
                </button>

                <div class="my-2 border-t border-base-300" />

                <label class="flex items-center justify-between gap-3 rounded-sm px-2 py-2 text-sm">
                  <span class="text-base-content/80">暗色模式</span>
                  <input
                    type="checkbox"
                    class="toggle toggle-sm"
                    :checked="isDark"
                    @change="applyTheme(($event.target as HTMLInputElement).checked)"
                  />
                </label>

                <label class="block px-2 py-1.5 text-xs text-base-content/60">语言</label>
                <select v-model="selectedLanguage" class="select select-bordered select-sm w-full rounded-sm" @change="updateUserPreferences">
                  <option v-for="item in languageOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
                </select>

                <label class="mt-2 block px-2 py-1.5 text-xs text-base-content/60">时区</label>
                <select v-model="selectedTimezone" class="select select-bordered select-sm w-full rounded-sm" @change="updateUserPreferences">
                  <option v-for="item in timezoneOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
                </select>

                <div class="my-2 border-t border-base-300" />

                <button type="button" class="btn btn-ghost btn-sm w-full justify-start rounded-sm text-error" @click="handleLogout">
                  登出
                </button>
              </div>
            </div>
          </div>
          <p v-if="user?.bio" class="mt-3 line-clamp-2 text-xs leading-5 text-base-content/60">{{ user.bio }}</p>
          <div class="mt-3 flex items-center justify-between text-xs text-base-content/55">
            <span>{{ user?.role || 'viewer' }}</span>
            <span class="inline-flex items-center gap-1">
              <span :class="user?.status === 'active' ? 'status status-success' : 'status status-warning'" />
              {{ user?.status || 'unknown' }}
            </span>
          </div>
          <div class="mt-2 flex items-center justify-between text-[11px] text-base-content/50">
            <span>语言 {{ user?.language || 'en' }}</span>
            <span>{{ user?.timezone || 'UTC' }}</span>
          </div>
        </div>
      </aside>

      <main class="space-y-5 lg:ml-[300px] lg:min-h-[calc(100vh-2rem)]">
        <RouterView />
      </main>
    </div>
  </section>
</template>
