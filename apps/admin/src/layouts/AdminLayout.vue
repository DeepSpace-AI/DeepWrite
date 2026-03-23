<script setup lang="ts">
import { computed, onBeforeMount, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import IconHomeOutline from '~icons/mdi/home-outline'
import IconRobotOutline from '~icons/mdi/robot-outline'
import IconAccountGroupOutline from '~icons/mdi/account-group-outline'
import IconFolderOutline from '~icons/mdi/folder-outline'
import IconCogOutline from '~icons/mdi/cog-outline'
import IconLogout from '~icons/mdi/logout'
import { logout } from '@/api/auth'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const isDark = ref(false)
const THEME_KEY = 'deepwrite_admin_theme'

const userInitial = computed(() => {
  const name = userStore.user?.displayName || userStore.user?.email || ''
  return name ? name.slice(0, 1).toUpperCase() : 'A'
})

function isNavActive(name: string) {
  if (name === 'dashboard') {
    return route.path === '/' || route.name === 'dashboard'
  }
  if (name === 'providers') {
    return route.path.startsWith('/providers')
  }
  if (name === 'users') {
    return route.name === 'users'
  }
  if (name === 'workspaces') {
    return route.name === 'workspaces'
  }
  if (name === 'system') {
    return route.name === 'system'
  }
  return route.name === name
}

function applyTheme(darkMode: boolean) {
  const theme = darkMode ? 'forest' : 'vellum-light'
  document.documentElement.setAttribute('data-theme', theme)
  localStorage.setItem(THEME_KEY, theme)
  isDark.value = darkMode
}

async function handleLogout() {
  logout()
  await router.push({ name: 'login' })
}

onBeforeMount(() => {
  const theme = localStorage.getItem(THEME_KEY)
  if (theme === 'forest') {
    applyTheme(true)
    return
  }
  if (theme === 'vellum-light') {
    applyTheme(false)
    return
  }
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  applyTheme(prefersDark)
})
</script>

<template>
  <section class="relative min-h-screen overflow-hidden bg-dot-grid" :data-theme="isDark ? 'forest' : 'vellum-light'">
    <div class="glow-blob -left-32 top-0 h-[28rem] w-[28rem] bg-[var(--glow-primary)] opacity-60" />
    <div class="glow-blob -right-32 bottom-0 h-[32rem] w-[32rem] bg-[var(--glow-secondary)] opacity-50" />
    <div class="glow-blob left-1/3 top-1/2 h-64 w-64 rounded-full bg-[var(--glow-primary)] opacity-30" />

    <div class="relative mx-auto px-4 py-4 lg:px-6">
      <aside class="glass-card fixed left-4 top-4 z-20 flex h-[calc(100vh-2rem)] w-72 flex-col rounded-2xl">
        <div class="border-b border-[var(--border-subtle)] px-5 py-5">
          <h1 class="text-xl font-bold text-pretty">DeepWrite</h1>
          <p class="mt-1 text-xs text-pretty-muted">系统治理台</p>
        </div>

        <nav class="flex-1 space-y-1 overflow-y-auto px-3 py-3">
          <RouterLink
            :to="{ name: 'dashboard' }"
            class="group flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm transition-all duration-200"
            :class="isNavActive('dashboard') 
              ? 'bg-[var(--glow-primary)] text-pretty border border-[var(--border-subtle)]' 
              : 'text-pretty-secondary hover:bg-[var(--glow-primary)] hover:text-pretty border border-transparent'"
          >
            <span class="inline-flex h-9 w-9 items-center justify-center rounded-xl transition-colors"
              :class="isNavActive('dashboard') 
                ? 'bg-[var(--glow-primary)] text-pretty' 
                : 'text-pretty-muted group-hover:text-pretty'">
              <IconHomeOutline class="h-4 w-4" />
            </span>
            <div>
              <p class="font-medium">仪表盘</p>
              <p class="text-xs text-pretty-muted">系统概览与统计</p>
            </div>
          </RouterLink>

          <RouterLink
            :to="{ name: 'users' }"
            class="group flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm transition-all duration-200"
            :class="isNavActive('users') 
              ? 'bg-[var(--glow-primary)] text-pretty border border-[var(--border-subtle)]' 
              : 'text-pretty-secondary hover:bg-[var(--glow-primary)] hover:text-pretty border border-transparent'"
          >
            <span class="inline-flex h-9 w-9 items-center justify-center rounded-xl transition-colors"
              :class="isNavActive('users') 
                ? 'bg-[var(--glow-primary)] text-pretty' 
                : 'text-pretty-muted group-hover:text-pretty'">
              <IconAccountGroupOutline class="h-4 w-4" />
            </span>
            <div>
              <p class="font-medium">用户管理</p>
              <p class="text-xs text-pretty-muted">成员检索与状态</p>
            </div>
          </RouterLink>

          <RouterLink
            :to="{ name: 'workspaces' }"
            class="group flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm transition-all duration-200"
            :class="isNavActive('workspaces') 
              ? 'bg-[var(--glow-primary)] text-pretty border border-[var(--border-subtle)]' 
              : 'text-pretty-secondary hover:bg-[var(--glow-primary)] hover:text-pretty border border-transparent'"
          >
            <span class="inline-flex h-9 w-9 items-center justify-center rounded-xl transition-colors"
              :class="isNavActive('workspaces') 
                ? 'bg-[var(--glow-primary)] text-pretty' 
                : 'text-pretty-muted group-hover:text-pretty'">
              <IconFolderOutline class="h-4 w-4" />
            </span>
            <div>
              <p class="font-medium">工作区</p>
              <p class="text-xs text-pretty-muted">平台工作区列表</p>
            </div>
          </RouterLink>

          <RouterLink
            :to="{ name: 'providers' }"
            class="group flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm transition-all duration-200"
            :class="isNavActive('providers') 
              ? 'bg-[var(--glow-primary)] text-pretty border border-[var(--border-subtle)]' 
              : 'text-pretty-secondary hover:bg-[var(--glow-primary)] hover:text-pretty border border-transparent'"
          >
            <span class="inline-flex h-9 w-9 items-center justify-center rounded-xl transition-colors"
              :class="isNavActive('providers') 
                ? 'bg-[var(--glow-primary)] text-pretty' 
                : 'text-pretty-muted group-hover:text-pretty'">
              <IconRobotOutline class="h-4 w-4" />
            </span>
            <div>
              <p class="font-medium">AI Provider</p>
              <p class="text-xs text-pretty-muted">模型配置与密钥</p>
            </div>
          </RouterLink>

          <RouterLink
            :to="{ name: 'system' }"
            class="group flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm transition-all duration-200"
            :class="isNavActive('system') 
              ? 'bg-[var(--glow-primary)] text-pretty border border-[var(--border-subtle)]' 
              : 'text-pretty-secondary hover:bg-[var(--glow-primary)] hover:text-pretty border border-transparent'"
          >
            <span class="inline-flex h-9 w-9 items-center justify-center rounded-xl transition-colors"
              :class="isNavActive('system') 
                ? 'bg-[var(--glow-primary)] text-pretty' 
                : 'text-pretty-muted group-hover:text-pretty'">
              <IconCogOutline class="h-4 w-4" />
            </span>
            <div>
              <p class="font-medium">系统配置</p>
              <p class="text-xs text-pretty-muted">平台设置与状态</p>
            </div>
          </RouterLink>
        </nav>

        <div class="mx-3 mb-3 mt-auto rounded-xl border border-[var(--border-subtle)] bg-[var(--bg-elevated)] p-3">
          <div class="flex items-center gap-3">
            <div class="inline-flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--glow-primary)] font-semibold text-pretty">
              {{ userInitial }}
            </div>
            <div class="min-w-0">
              <p class="truncate text-sm font-medium text-pretty">{{ userStore.user?.displayName || 'admin' }}</p>
              <p class="truncate text-xs text-pretty-muted">{{ userStore.user?.email || '' }}</p>
            </div>
            <button type="button" class="btn btn-ghost btn-sm rounded-xl" @click="handleLogout">
              <IconLogout class="h-4 w-4" />
            </button>
          </div>
          <label class="mt-3 flex items-center justify-between rounded-xl border border-[var(--border-subtle)] bg-[var(--bg-base)] px-3 py-2">
            <span class="text-xs text-pretty-muted">暗色模式</span>
            <input
              type="checkbox"
              class="toggle toggle-sm"
              :checked="isDark"
              @change="applyTheme(($event.target as HTMLInputElement).checked)"
            />
          </label>
        </div>
      </aside>

      <main class="lg:pl-[20rem]">
        <RouterView />
      </main>
    </div>
  </section>
</template>
