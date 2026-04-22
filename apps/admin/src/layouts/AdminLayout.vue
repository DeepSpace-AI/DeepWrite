<script setup lang="ts">
import { computed, onBeforeMount, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import IconHome from '~icons/mdi/home'
import IconAccountGroup from '~icons/mdi/account-group'
import IconFolder from '~icons/mdi/folder'
import IconRobot from '~icons/mdi/robot'
import IconCog from '~icons/mdi/cog'
import IconLogout from '~icons/mdi/logout'
import IconMoon from '~icons/mdi/moon-waning-crescent'
import IconSun from '~icons/mdi/white-balance-sunny'
import IconChevronLeft from '~icons/mdi/chevron-left'
import { logout } from '@/api/auth'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const isDark = ref(false)
const THEME_KEY = 'tech_theme'
const isSidebarCollapsed = ref(false)

const userInitial = computed(() => {
  const name = userStore.user?.displayName || userStore.user?.email || ''
  return name ? name.slice(0, 1).toUpperCase() : 'A'
})

const navItems = [
  { name: 'dashboard', label: '仪表盘', icon: IconHome, path: '/' },
  { name: 'users', label: '用户管理', icon: IconAccountGroup, path: '/users' },
  { name: 'workspaces', label: '工作区', icon: IconFolder, path: '/workspaces' },
  { name: 'providers', label: 'AI Provider', icon: IconRobot, path: '/providers' },
  { name: 'system', label: '系统配置', icon: IconCog, path: '/system' },
]

function isNavActive(name: string) {
  if (name === 'dashboard') {
    return route.path === '/' || route.name === 'dashboard'
  }
  if (name === 'providers') {
    return route.path.startsWith('/providers')
  }
  return route.name === name
}

function applyTheme(darkMode: boolean) {
  const theme = darkMode ? 'tech-dark' : 'tech-light'
  document.documentElement.setAttribute('data-theme', theme)
  localStorage.setItem(THEME_KEY, theme)
  isDark.value = darkMode
}

function toggleTheme() {
  applyTheme(!isDark.value)
}

async function handleLogout() {
  logout()
  await router.push({ name: 'login' })
}

onBeforeMount(() => {
  const theme = localStorage.getItem(THEME_KEY)
  if (theme === 'tech-dark') {
    applyTheme(true)
    return
  }
  if (theme === 'tech-light') {
    applyTheme(false)
    return
  }
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  applyTheme(prefersDark)
})
</script>

<template>
  <div class="min-h-screen bg-[var(--bg-base)]">
    <div class="flex min-h-screen">
      <aside 
        class="sidebar-tech fixed left-0 top-0 z-30 flex h-screen flex-col transition-all duration-200"
        :class="isSidebarCollapsed ? 'w-16' : 'w-60'"
      >
        <div class="flex h-16 items-center justify-between border-b border-[var(--border-subtle)] px-4">
          <div v-if="!isSidebarCollapsed" class="flex items-center gap-3">
            <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-[var(--accent-primary)]">
              <svg class="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div>
              <h1 class="text-display text-base leading-tight">DeepWrite</h1>
              <p class="text-muted text-xs">Admin Console</p>
            </div>
          </div>
          <button 
            class="flex h-8 w-8 items-center justify-center rounded transition-colors hover:bg-[var(--surface-hover)]"
            @click="isSidebarCollapsed = !isSidebarCollapsed"
          >
            <IconChevronLeft 
              class="h-5 w-5 text-[var(--text-muted)] transition-transform duration-200"
              :class="{ 'rotate-180': isSidebarCollapsed }"
            />
          </button>
        </div>

        <nav class="flex-1 overflow-y-auto p-2">
          <div class="space-y-0.5">
            <RouterLink
              v-for="item in navItems"
              :key="item.name"
              :to="item.path"
              class="sidebar-nav-item"
              :class="{ 'active': isNavActive(item.name), 'justify-center': isSidebarCollapsed }"
            >
              <span class="nav-icon flex h-8 w-8 shrink-0 items-center justify-center rounded">
                <component :is="item.icon" class="h-5 w-5" />
              </span>
              <span v-if="!isSidebarCollapsed" class="text-sm">{{ item.label }}</span>
            </RouterLink>
          </div>
        </nav>

        <div class="border-t border-[var(--border-subtle)] p-2">
          <div 
            class="flex items-center gap-3 rounded-lg p-2"
            :class="{ 'justify-center': isSidebarCollapsed }"
          >
            <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded bg-[var(--accent-primary)] text-sm font-medium text-white">
              {{ userInitial }}
            </div>
            <div v-if="!isSidebarCollapsed" class="min-w-0 flex-1">
              <p class="truncate text-sm font-medium">{{ userStore.user?.displayName || 'admin' }}</p>
              <p class="truncate text-xs text-[var(--text-muted)]">{{ userStore.user?.email || '' }}</p>
            </div>
            <div v-if="!isSidebarCollapsed" class="flex shrink-0 gap-0.5">
              <button 
                type="button" 
                class="flex h-7 w-7 items-center justify-center rounded transition-colors hover:bg-[var(--surface-hover)]"
                @click="toggleTheme"
                :title="isDark ? '切换到亮色模式' : '切换到暗色模式'"
              >
                <IconSun v-if="isDark" class="h-4 w-4 text-[var(--text-muted)]" />
                <IconMoon v-else class="h-4 w-4 text-[var(--text-muted)]" />
              </button>
              <button 
                type="button" 
                class="flex h-7 w-7 items-center justify-center rounded transition-colors hover:bg-[var(--surface-hover)]"
                @click="handleLogout"
                title="退出登录"
              >
                <IconLogout class="h-4 w-4 text-[var(--text-muted)]" />
              </button>
            </div>
          </div>
        </div>
      </aside>

      <main 
        class="flex-1 min-h-screen transition-all duration-200"
        :class="isSidebarCollapsed ? 'ml-16' : 'ml-60'"
      >
        <div class="p-6">
          <RouterView />
        </div>
      </main>
    </div>
  </div>
</template>