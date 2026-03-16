<script setup lang="ts">
import { computed, onBeforeMount, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import IconRobotOutline from '~icons/mdi/robot-outline'
import IconAccountGroupOutline from '~icons/mdi/account-group-outline'
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
  if (name === 'providers') {
    return route.path.startsWith('/providers')
  }
  return route.name === name
}

function applyTheme(darkMode: boolean) {
  const theme = darkMode ? 'forest' : 'bumblebee'
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
  if (theme === 'bumblebee') {
    applyTheme(false)
    return
  }
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  applyTheme(prefersDark)
})
</script>

<template>
  <section class="dot-grid relative min-h-screen overflow-hidden bg-base-100">
    <div class="pointer-events-none absolute -left-28 top-0 h-72 w-72 rounded-full bg-secondary/10 blur-3xl" />
    <div class="pointer-events-none absolute -right-24 bottom-0 h-80 w-80 rounded-full bg-primary/10 blur-3xl" />
    <div class="relative mx-auto px-4 py-4 lg:px-5">
      <aside class="rounded-sm border border-base-300 bg-base-100/95 shadow-sm backdrop-blur lg:fixed lg:left-5 lg:top-4 lg:z-20 lg:flex lg:h-[calc(100vh-2rem)] lg:w-72 lg:flex-col">
        <div class="border-b border-base-300 px-5 py-4">
          <h1 class="text-xl font-bold text-base-content">DeepWrite Admin</h1>
          <p class="mt-1 text-xs text-base-content/60">系统治理台</p>
        </div>

        <nav class="px-3 py-3 lg:flex-1 lg:overflow-y-auto">
          <RouterLink
            :to="{ name: 'providers' }"
            class="group mb-1 flex items-center gap-3 rounded-sm px-3 py-2 text-sm transition"
            :class="isNavActive('providers') ? 'border border-base-300/70 bg-base-200 text-base-content' : 'border border-transparent text-base-content/78 hover:bg-base-200'"
          >
            <span class="inline-flex h-8 w-8 items-center justify-center rounded-sm" :class="isNavActive('providers') ? 'bg-primary/10 text-primary' : 'bg-base-200/70 text-base-content/65'">
              <IconRobotOutline class="h-4 w-4" />
            </span>
            <div>
              <p class="font-medium">AI Provider 管理</p>
              <p class="text-xs text-base-content/55">模型配置、密钥与启停控制</p>
            </div>
          </RouterLink>

          <RouterLink
            :to="{ name: 'users' }"
            class="group mb-1 flex items-center gap-3 rounded-sm px-3 py-2 text-sm transition"
            :class="isNavActive('users') ? 'border border-base-300/70 bg-base-200 text-base-content' : 'border border-transparent text-base-content/78 hover:bg-base-200'"
          >
            <span class="inline-flex h-8 w-8 items-center justify-center rounded-sm" :class="isNavActive('users') ? 'bg-primary/10 text-primary' : 'bg-base-200/70 text-base-content/65'">
              <IconAccountGroupOutline class="h-4 w-4" />
            </span>
            <div>
              <p class="font-medium">用户管理</p>
              <p class="text-xs text-base-content/55">成员检索、角色与状态治理</p>
            </div>
          </RouterLink>

          <RouterLink
            :to="{ name: 'system' }"
            class="group flex items-center gap-3 rounded-sm px-3 py-2 text-sm transition"
            :class="isNavActive('system') ? 'border border-base-300/70 bg-base-200 text-base-content' : 'border border-transparent text-base-content/78 hover:bg-base-200'"
          >
            <span class="inline-flex h-8 w-8 items-center justify-center rounded-sm" :class="isNavActive('system') ? 'bg-primary/10 text-primary' : 'bg-base-200/70 text-base-content/65'">
              <IconCogOutline class="h-4 w-4" />
            </span>
            <div>
              <p class="font-medium">系统配置</p>
              <p class="text-xs text-base-content/55">平台设置与运行状态</p>
            </div>
          </RouterLink>
        </nav>

        <div class="mx-3 mb-3 mt-auto rounded-sm border border-base-300 bg-base-200/70 p-3">
          <div class="flex items-center gap-3">
            <div class="inline-flex h-10 w-10 items-center justify-center rounded-sm bg-primary text-primary-content font-semibold">
              {{ userInitial }}
            </div>
            <div class="min-w-0">
              <p class="truncate text-sm font-medium text-base-content">{{ userStore.user?.displayName || 'admin' }}</p>
              <p class="truncate text-xs text-base-content/55">{{ userStore.user?.email || '' }}</p>
            </div>
            <button type="button" class="btn btn-ghost btn-sm rounded-sm" @click="handleLogout">
              <IconLogout class="h-4 w-4" />
            </button>
          </div>
          <label class="mt-3 flex items-center justify-between rounded-sm border border-base-300 bg-base-100 px-2 py-1.5">
            <span class="text-xs text-base-content/70">暗色模式</span>
            <input
              type="checkbox"
              class="toggle toggle-sm"
              :checked="isDark"
              @change="applyTheme(($event.target as HTMLInputElement).checked)"
            />
          </label>
        </div>
      </aside>

      <main class="lg:pl-[19rem]">
        <RouterView />
      </main>
    </div>
  </section>
</template>
