<script setup lang="ts">
import { computed, onBeforeMount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { logout, updateCurrentUserPreferences } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'
import { useNotificationStore } from '@/stores/notification'
import { useI18n } from 'vue-i18n'
import { getCachedLocale, setAppLocale } from '../../locales/i18n'
import IconViewDashboardOutline from '~icons/mdi/view-dashboard-outline'
import IconHomeOutline from '~icons/mdi/home-outline'
import IconBookmarkOutline from '~icons/mdi/bookmark-outline'
import IconViewGridOutline from '~icons/mdi/view-grid-outline'
import IconRobotOutline from '~icons/mdi/robot-outline'
import IconBellOutline from '~icons/mdi/bell-outline'
import IconPlus from '~icons/mdi/plus'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const userStore = useUserStore()
const notificationStore = useNotificationStore()
const { t } = useI18n()

const { user } = storeToRefs(userStore)
const { pendingCount } = storeToRefs(notificationStore)
const selectedLanguage = ref(getCachedLocale() ?? 'zh-CN')
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

function isNavActive(name: 'dashboard' | 'workspace-list' | 'library' | 'skills' | 'agents') {
  return route.name === name
}

function applyTheme(darkMode: boolean) {
  const theme = darkMode ? 'forest' : 'bumblebee'
  document.documentElement.setAttribute('data-theme', theme)
  localStorage.setItem(THEME_KEY, theme)
  isDark.value = darkMode
}

async function updateUserPreferences() {
  const nextLanguage = selectedLanguage.value
  const nextTimezone = selectedTimezone.value

  setAppLocale(nextLanguage)
  if (!user.value) return

  const previousUser = { ...user.value }
  userStore.setUser({
    ...user.value,
    language: nextLanguage,
    timezone: nextTimezone,
  })

  try {
    await updateCurrentUserPreferences({
      language: nextLanguage,
      timezone: nextTimezone,
    })
  } catch (error) {
    userStore.setUser(previousUser)
    setAppLocale(previousUser.language || getCachedLocale() || 'zh-CN')
    console.error('更新用户偏好失败，将回滚本地设置', error)
  }
}

function openProfileCenter() {
  router.push({ name: 'profile' })
}

function openSystemSettings() {
  router.push({ name: 'dashboard', query: { tab: 'settings' } })
}

function openSubscription() {
  router.push({ name: 'dashboard', query: { tab: 'subscription' } })
}

function openNotifications() {
  router.push({ name: 'notifications' })
}

async function handleLogout() {
  notificationStore.clear()
  logout()
  await router.push({ name: 'login' })
}

watch(
  user,
  (nextUser) => {
    selectedTimezone.value = nextUser?.timezone || 'UTC'

    if (nextUser?.language === 'en' || nextUser?.language === 'zh-CN') {
      selectedLanguage.value = nextUser.language
      setAppLocale(nextUser.language)
      return
    }

    const cachedLocale = getCachedLocale() ?? 'zh-CN'
    selectedLanguage.value = cachedLocale
    setAppLocale(cachedLocale)
  },
  { immediate: true },
)

onBeforeMount(() => {
  authStore.initializeFromStorage()
  userStore.initializeFromStorage()
  notificationStore.refreshPendingCount()

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
      <aside class="rounded-sm border border-base-300 bg-base-100/95 shadow-sm backdrop-blur lg:fixed lg:left-5 lg:top-4 lg:z-20 lg:flex lg:h-[calc(100vh-2rem)] lg:w-70 lg:flex-col">
        <div class="border-b border-base-300 px-5 py-4">
           <h1 class="heading-serif mt-2 text-xl font-bold text-base-content">{{ t('sidebar.title') }}</h1>
        </div>

        <nav class="px-3 py-3 lg:flex-1 lg:overflow-y-auto">
          <RouterLink
            :to="{ name: 'dashboard' }"
            class="group mb-1 flex items-center gap-3 rounded-sm px-3 py-2 text-sm transition"
            :class="isNavActive('dashboard')
              ? 'border border-base-300/70 bg-base-200 text-base-content'
              : 'border border-transparent text-base-content/78 hover:bg-base-200'"
          >
            <span class="inline-flex h-7 w-7 items-center justify-center rounded-sm" :class="isNavActive('dashboard') ? 'bg-primary/10 text-primary' : 'bg-base-200/70 text-base-content/65'">
              <IconViewDashboardOutline class="h-4 w-4" />
            </span>
            <div>
                <p class="font-medium" :class="isNavActive('dashboard') ? 'text-base-content' : 'text-base-content/78'">{{ t('sidebar.workbench') }}</p>
                <p class="text-xs" :class="isNavActive('dashboard') ? 'text-base-content/50' : 'text-base-content/45'">{{ t('sidebar.workbenchDesc') }}</p>
            </div>
          </RouterLink>

          <RouterLink
            :to="{ name: 'workspace-list' }"
            class="group mb-1 flex items-center gap-3 rounded-sm px-3 py-2 text-sm transition"
            :class="isNavActive('workspace-list')
              ? 'border border-base-300/70 bg-base-200 text-base-content'
              : 'border border-transparent text-base-content/78 hover:bg-base-200'"
          >
            <span class="inline-flex h-7 w-7 items-center justify-center rounded-sm" :class="isNavActive('workspace-list') ? 'bg-primary/10 text-primary' : 'bg-base-200/70 text-base-content/65'">
              <IconHomeOutline class="h-4 w-4" />
            </span>
            <div>
                <p class="font-medium" :class="isNavActive('workspace-list') ? 'text-base-content' : 'text-base-content/78'">{{ t('sidebar.workspace') }}</p>
                <p class="text-xs" :class="isNavActive('workspace-list') ? 'text-base-content/50' : 'text-base-content/45'">{{ t('sidebar.workspaceDesc') }}</p>
            </div>
          </RouterLink>

          <RouterLink
            :to="{ name: 'library' }"
            class="group mb-1 flex items-center gap-3 rounded-sm px-3 py-2 text-sm transition"
            :class="isNavActive('library')
              ? 'border border-base-300/70 bg-base-200 text-base-content'
              : 'border border-transparent text-base-content/78 hover:bg-base-200'"
          >
            <span class="inline-flex h-7 w-7 items-center justify-center rounded-sm" :class="isNavActive('library') ? 'bg-secondary/10 text-secondary' : 'bg-base-200/70 text-base-content/65'">
              <IconBookmarkOutline class="h-4 w-4" />
            </span>
            <div>
                <p class="font-medium" :class="isNavActive('library') ? 'text-base-content' : 'text-base-content/78'">{{ t('sidebar.library') }}</p>
                <p class="text-xs" :class="isNavActive('library') ? 'text-base-content/50' : 'text-base-content/45'">{{ t('sidebar.libraryDesc') }}</p>
            </div>
          </RouterLink>

          <RouterLink
            :to="{ name: 'skills' }"
            class="group mb-1 flex items-center gap-3 rounded-sm px-3 py-2 text-sm transition"
            :class="isNavActive('skills')
              ? 'border border-base-300/70 bg-base-200 text-base-content'
              : 'border border-transparent text-base-content/78 hover:bg-base-200'"
          >
            <span class="inline-flex h-7 w-7 items-center justify-center rounded-sm" :class="isNavActive('skills') ? 'bg-accent/10 text-accent' : 'bg-base-200/70 text-base-content/65'">
              <IconViewGridOutline class="h-4 w-4" />
            </span>
            <div>
                <p class="font-medium" :class="isNavActive('skills') ? 'text-base-content' : 'text-base-content/78'">{{ t('sidebar.skills') }}</p>
                <p class="text-xs" :class="isNavActive('skills') ? 'text-base-content/50' : 'text-base-content/45'">{{ t('sidebar.skillsDesc') }}</p>
            </div>
          </RouterLink>

          <RouterLink
            :to="{ name: 'agents' }"
            class="group flex items-center gap-3 rounded-sm px-3 py-2 text-sm transition"
            :class="isNavActive('agents')
              ? 'border border-base-300/70 bg-base-200 text-base-content'
              : 'border border-transparent text-base-content/78 hover:bg-base-200'"
          >
            <span class="inline-flex h-7 w-7 items-center justify-center rounded-sm" :class="isNavActive('agents') ? 'bg-primary/10 text-primary' : 'bg-base-200/70 text-base-content/65'">
              <IconRobotOutline class="h-4 w-4" />
            </span>
            <div>
                <p class="font-medium" :class="isNavActive('agents') ? 'text-base-content' : 'text-base-content/78'">{{ t('sidebar.agent') }}</p>
                <p class="text-xs" :class="isNavActive('agents') ? 'text-base-content/50' : 'text-base-content/45'">{{ t('sidebar.agentDesc') }}</p>
            </div>
          </RouterLink>
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
              <p class="truncate text-sm font-medium text-base-content">{{ user?.displayName || t('sidebar.guestUser') }}</p>
              <p class="truncate text-xs text-base-content/55">{{ user?.email || 'guest@deepwrite.local' }}</p>
            </div>

            <div class="ml-auto flex items-center gap-1">
              <button
                type="button"
                class="btn btn-ghost btn-xs rounded-sm"
                :aria-label="t('sidebar.notifications')"
                @click="openNotifications"
              >
                <span class="relative inline-flex">
                  <IconBellOutline class="h-4 w-4" />
                  <span
                    v-if="pendingCount > 0"
                    class="absolute -right-2 -top-1 min-w-4 rounded-full bg-error px-1 text-center text-[10px] font-semibold leading-4 text-error-content"
                  >
                    {{ pendingCount > 99 ? '99+' : pendingCount }}
                  </span>
                </span>
              </button>

              <div class="dropdown dropdown-end dropdown-top">
                <button tabindex="0" class="btn btn-ghost btn-xs rounded-sm" :aria-label="t('sidebar.openMenu')">
                  <IconPlus class="h-4 w-4" />
                </button>
                <div tabindex="0" class="dropdown-content z-30 mt-2 w-64 rounded-sm border border-base-300 bg-base-100 p-2 shadow-lg">
                  <button type="button" class="btn btn-ghost btn-sm justify-start rounded-sm" @click="openProfileCenter">
                    {{ t('userMenu.profile') }}
                  </button>
                  <button type="button" class="btn btn-ghost btn-sm justify-start rounded-sm" @click="openSystemSettings">
                    {{ t('userMenu.settings') }}
                  </button>
                  <button type="button" class="btn btn-ghost btn-sm justify-start rounded-sm" @click="openSubscription">
                    {{ t('userMenu.subscription') }}
                  </button>

                  <div class="my-2 border-t border-base-300" />

                  <label class="flex items-center justify-between gap-3 rounded-sm px-2 py-2 text-sm">
                    <span class="text-base-content/80">{{ t('userMenu.darkMode') }}</span>
                    <input
                      type="checkbox"
                      class="toggle toggle-sm"
                      :checked="isDark"
                      @change="applyTheme(($event.target as HTMLInputElement).checked)"
                    />
                  </label>

                  <label class="block px-2 py-1.5 text-xs text-base-content/60">{{ t('userMenu.language') }}</label>
                  <select v-model="selectedLanguage" class="select select-bordered select-sm w-full rounded-sm" @change="updateUserPreferences">
                    <option v-for="item in languageOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
                  </select>

                  <label class="mt-2 block px-2 py-1.5 text-xs text-base-content/60">{{ t('userMenu.timezone') }}</label>
                  <select v-model="selectedTimezone" class="select select-bordered select-sm w-full rounded-sm" @change="updateUserPreferences">
                    <option v-for="item in timezoneOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
                  </select>

                  <div class="my-2 border-t border-base-300" />

                  <button type="button" class="btn btn-ghost btn-sm w-full justify-start rounded-sm text-error" @click="handleLogout">
                    {{ t('userMenu.logout') }}
                  </button>
                </div>
              </div>
            </div>
          </div>
          <p v-if="user?.bio?.trim()" class="mt-3 line-clamp-2 text-xs leading-5 text-base-content/60">{{ user.bio }}</p>
          <div class="mt-3 flex items-center justify-between text-xs text-base-content/55">
            <span>{{ user?.role || 'viewer' }}</span>
            <span class="inline-flex items-center gap-1">
              <span :class="user?.status === 'active' ? 'status status-success' : 'status status-warning'" />
              {{ t('userStatus.' + (user?.status || 'unknown')) }}
            </span>
          </div>
          <div class="mt-2 flex items-center justify-between text-[11px] text-base-content/50">
            <span>{{ t('sidebar.langDisplay', { lang: user?.language || 'en' }) }}</span>
            <span>{{ user?.timezone || 'UTC' }}</span>
          </div>
        </div>
      </aside>

      <main class="space-y-5 lg:ml-75 lg:min-h-[calc(100vh-2rem)]">
        <RouterView />
      </main>
    </div>
  </section>
</template>
