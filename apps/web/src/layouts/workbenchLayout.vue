<script setup lang="ts">
import { computed, onBeforeMount, ref, watch, onMounted, onUnmounted } from 'vue'
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
const userMenuOpen = ref(false)

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
  const theme = darkMode ? 'vellum-dark' : 'vellum-light'
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
    console.error('Failed to update user preferences, rolling back', error)
  }
}

function openProfileCenter() {
  userMenuOpen.value = false
  router.push({ name: 'profile' })
}

function openSystemSettings() {
  userMenuOpen.value = false
  router.push({ name: 'dashboard', query: { tab: 'settings' } })
}

function openSubscription() {
  userMenuOpen.value = false
  router.push({ name: 'dashboard', query: { tab: 'subscription' } })
}

function openNotifications() {
  userMenuOpen.value = false
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
  if (theme === 'vellum-dark') {
    applyTheme(true)
  } else if (theme === 'vellum-light') {
    applyTheme(false)
  } else {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    applyTheme(prefersDark)
  }
})

function closeMenuOnClickOutside(event: MouseEvent) {
  const target = event.target as HTMLElement
  if (!target.closest('.relative')) {
    userMenuOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', closeMenuOnClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', closeMenuOnClickOutside)
})
</script>

<template>
  <section class="dot-grid relative min-h-screen overflow-hidden min-w-screen" :data-theme="isDark ? 'vellum-dark' : 'vellum-light'">
    <div class="glow-blob -left-28 top-0 h-[28rem] w-[28rem] bg-[var(--glow-primary)] opacity-50" />
    <div class="glow-blob -right-24 bottom-0 h-[32rem] w-[32rem] bg-[var(--glow-secondary)] opacity-40" />
    <div class="glow-blob left-1/3 top-1/2 h-64 w-64 -translate-x-1/2 -translate-y-1/2 rounded-full bg-[var(--glow-primary)] opacity-20" />

    <div class="relative mx-auto px-4 py-4 lg:px-6">
      <aside class="paper-card fixed left-4 top-4 z-20 flex h-[calc(100vh-2rem)] w-72 flex-col rounded-lg">
        <div class="px-5 py-5">
          <h1 class="text-editorial mt-2 text-xl font-semibold">DeepWrite</h1>
        </div>

        <nav class="flex-1 space-y-1 overflow-y-auto px-3 py-3">
          <RouterLink
            :to="{ name: 'dashboard' }"
            class="group flex items-center gap-3 rounded-md px-3 py-3 transition-all duration-300"
            :class="isNavActive('dashboard')
              ? 'bg-[var(--glow-primary)]'
              : 'hover:bg-[var(--glow-primary)]'"
          >
            <span
              class="inline-flex h-10 w-10 items-center justify-center rounded-md transition-colors"
              :class="isNavActive('dashboard')
                ? 'bg-[var(--glow-primary)]'
                : 'group-hover:bg-[var(--glow-primary)]'"
            >
              <IconViewDashboardOutline class="h-5 w-5" :class="isNavActive('dashboard') ? 'text-pretty' : 'text-pretty-muted'" />
            </span>
            <div>
              <p class="font-medium label-md" :class="isNavActive('dashboard') ? 'text-pretty' : 'text-pretty-secondary'">{{ t('sidebar.workbench') }}</p>
              <p class="text-xs text-pretty-muted mt-0.5">{{ t('sidebar.workbenchDesc') }}</p>
            </div>
          </RouterLink>

          <RouterLink
            :to="{ name: 'workspace-list' }"
            class="group flex items-center gap-3 rounded-md px-3 py-3 transition-all duration-300"
            :class="isNavActive('workspace-list')
              ? 'bg-[var(--glow-primary)]'
              : 'hover:bg-[var(--glow-primary)]'"
          >
            <span
              class="inline-flex h-10 w-10 items-center justify-center rounded-md transition-colors"
              :class="isNavActive('workspace-list')
                ? 'bg-[var(--glow-secondary)]'
                : 'group-hover:bg-[var(--glow-secondary)]'"
            >
              <IconHomeOutline class="h-5 w-5" :class="isNavActive('workspace-list') ? 'text-pretty' : 'text-pretty-muted'" />
            </span>
            <div>
              <p class="font-medium label-md" :class="isNavActive('workspace-list') ? 'text-pretty' : 'text-pretty-secondary'">{{ t('sidebar.workspace') }}</p>
              <p class="text-xs text-pretty-muted mt-0.5">{{ t('sidebar.workspaceDesc') }}</p>
            </div>
          </RouterLink>

          <RouterLink
            :to="{ name: 'library' }"
            class="group flex items-center gap-3 rounded-md px-3 py-3 transition-all duration-300"
            :class="isNavActive('library')
              ? 'bg-[var(--glow-primary)]'
              : 'hover:bg-[var(--glow-primary)]'"
          >
            <span
              class="inline-flex h-10 w-10 items-center justify-center rounded-md transition-colors"
              :class="isNavActive('library')
                ? 'bg-[var(--glow-secondary)]'
                : 'group-hover:bg-[var(--glow-secondary)]'"
            >
              <IconBookmarkOutline class="h-5 w-5" :class="isNavActive('library') ? 'text-pretty' : 'text-pretty-muted'" />
            </span>
            <div>
              <p class="font-medium label-md" :class="isNavActive('library') ? 'text-pretty' : 'text-pretty-secondary'">{{ t('sidebar.library') }}</p>
              <p class="text-xs text-pretty-muted mt-0.5">{{ t('sidebar.libraryDesc') }}</p>
            </div>
          </RouterLink>

          <RouterLink
            :to="{ name: 'skills' }"
            class="group flex items-center gap-3 rounded-md px-3 py-3 transition-all duration-300"
            :class="isNavActive('skills')
              ? 'bg-[var(--glow-primary)]'
              : 'hover:bg-[var(--glow-primary)]'"
          >
            <span
              class="inline-flex h-10 w-10 items-center justify-center rounded-md transition-colors"
              :class="isNavActive('skills')
                ? 'bg-[var(--glow-primary)]'
                : 'group-hover:bg-[var(--glow-primary)]'"
            >
              <IconViewGridOutline class="h-5 w-5" :class="isNavActive('skills') ? 'text-pretty' : 'text-pretty-muted'" />
            </span>
            <div>
              <p class="font-medium label-md" :class="isNavActive('skills') ? 'text-pretty' : 'text-pretty-secondary'">{{ t('sidebar.skills') }}</p>
              <p class="text-xs text-pretty-muted mt-0.5">{{ t('sidebar.skillsDesc') }}</p>
            </div>
          </RouterLink>

          <RouterLink
            :to="{ name: 'agents' }"
            class="group flex items-center gap-3 rounded-md px-3 py-3 transition-all duration-300"
            :class="isNavActive('agents')
              ? 'bg-[var(--glow-primary)]'
              : 'hover:bg-[var(--glow-primary)]'"
          >
            <span
              class="inline-flex h-10 w-10 items-center justify-center rounded-md transition-colors"
              :class="isNavActive('agents')
                ? 'bg-[var(--glow-primary)]'
                : 'group-hover:bg-[var(--glow-primary)]'"
            >
              <IconRobotOutline class="h-5 w-5" :class="isNavActive('agents') ? 'text-pretty' : 'text-pretty-muted'" />
            </span>
            <div>
              <p class="font-medium label-md" :class="isNavActive('agents') ? 'text-pretty' : 'text-pretty-secondary'">{{ t('sidebar.agent') }}</p>
              <p class="text-xs text-pretty-muted mt-0.5">{{ t('sidebar.agentDesc') }}</p>
            </div>
          </RouterLink>
        </nav>

        <div class="mx-3 mb-3 mt-auto rounded-md bg-[var(--surface-overlay)] p-4">
          <div class="flex items-start gap-3">
            <img
              v-if="user?.avatarUrl"
              :src="user.avatarUrl"
              :alt="`${user.displayName} avatar`"
              class="h-10 w-10 rounded-md object-cover"
            />
            <div
              v-else
              class="inline-flex h-10 w-10 items-center justify-center rounded-md bg-[var(--glow-primary)] font-semibold text-pretty"
            >
              {{ userInitial }}
            </div>
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-medium text-pretty">{{ user?.displayName || t('sidebar.guestUser') }}</p>
              <p class="truncate text-xs text-pretty-muted">{{ user?.email || 'guest@deepwrite.local' }}</p>
            </div>

            <div class="ml-auto flex items-center gap-1">
              <button
                type="button"
                class="btn-tertiary rounded-md p-2"
                :aria-label="t('sidebar.notifications')"
                @click="openNotifications"
              >
                <span class="relative inline-flex">
                  <IconBellOutline class="h-5 w-5" />
                  <span
                    v-if="pendingCount > 0"
                    class="absolute -right-1 -top-1 min-w-4 rounded-full bg-[var(--color-primary)] px-1 text-center text-[10px] font-semibold text-white"
                  >
                    {{ pendingCount > 99 ? '99+' : pendingCount }}
                  </span>
                </span>
              </button>

              <div class="relative">
                <button
                  type="button"
                  class="btn-tertiary rounded-md p-2"
                  :aria-label="t('sidebar.openMenu')"
                  @click="userMenuOpen = !userMenuOpen"
                >
                  <IconPlus class="h-5 w-5" />
                </button>
                <div
                  v-if="userMenuOpen"
                  class="absolute right-0 bottom-full z-50 mb-2 w-64 rounded-md p-4 shadow-lg paper-card bg-[var(--surface-raised)]"
                >
                  <button
                    type="button"
                    class="btn-tertiary mb-1 w-full justify-start rounded-md px-3 py-2 text-sm"
                    @click="openProfileCenter"
                  >
                    {{ t('userMenu.profile') }}
                  </button>
                  <button
                    type="button"
                    class="btn-tertiary mb-1 w-full justify-start rounded-md px-3 py-2 text-sm"
                    @click="openSystemSettings"
                  >
                    {{ t('userMenu.settings') }}
                  </button>
                  <button
                    type="button"
                    class="btn-tertiary mb-2 w-full justify-start rounded-md px-3 py-2 text-sm"
                    @click="openSubscription"
                  >
                    {{ t('userMenu.subscription') }}
                  </button>

                  <div class="my-3 rounded-sm bg-[var(--surface-overlay)] px-2 py-2" />

                  <label class="flex items-center justify-between gap-3 rounded-md px-2 py-2.5 text-sm">
                    <span class="text-pretty-secondary">{{ t('userMenu.darkMode') }}</span>
                    <input
                      type="checkbox"
                      class="toggle toggle-sm"
                      :checked="isDark"
                      @change="applyTheme(($event.target as HTMLInputElement).checked)"
                    />
                  </label>

                  <label class="mt-3 block px-2 py-1.5 text-xs text-pretty-muted">{{ t('userMenu.language') }}</label>
                  <select
                    v-model="selectedLanguage"
                    class="input-ghost mt-1 w-full rounded-md px-3 py-2 text-sm"
                    @change="updateUserPreferences"
                  >
                    <option v-for="item in languageOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
                  </select>

                  <label class="mt-3 block px-2 py-1.5 text-xs text-pretty-muted">{{ t('userMenu.timezone') }}</label>
                  <select
                    v-model="selectedTimezone"
                    class="input-ghost mt-1 w-full rounded-md px-3 py-2 text-sm"
                    @change="updateUserPreferences"
                  >
                    <option v-for="item in timezoneOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
                  </select>

                  <div class="my-3 rounded-sm bg-[var(--surface-overlay)] px-2 py-2" />

                  <button
                    type="button"
                    class="btn-tertiary w-full justify-start rounded-md px-3 py-2 text-sm text-red-500"
                    @click="handleLogout"
                  >
                    {{ t('userMenu.logout') }}
                  </button>
                </div>
              </div>
            </div>
          </div>
          <p v-if="user?.bio?.trim()" class="mt-3 line-clamp-2 text-xs leading-5 text-pretty-secondary">{{ user.bio }}</p>
          <div class="mt-3 flex items-center justify-between text-xs text-pretty-muted">
            <span>{{ user?.role || 'viewer' }}</span>
            <span class="inline-flex items-center gap-1.5">
              <span
                class="h-2 w-2 rounded-full"
                :class="user?.status === 'active' ? 'bg-[var(--color-secondary)]' : 'bg-amber-500'"
              />
              {{ t('userStatus.' + (user?.status || 'unknown')) }}
            </span>
          </div>
          <div class="mt-2 flex items-center justify-between text-[11px] text-pretty-muted/60">
            <span>{{ t('sidebar.langDisplay', { lang: user?.language || 'en' }) }}</span>
            <span>{{ user?.timezone || 'UTC' }}</span>
          </div>
        </div>
      </aside>

      <main class="space-y-5 lg:ml-[19rem] lg:min-h-[calc(100vh-2rem)]">
        <RouterView />
      </main>
    </div>
  </section>
</template>
