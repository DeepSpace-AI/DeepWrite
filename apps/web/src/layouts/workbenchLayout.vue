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
import IconCogOutline from '~icons/mdi/cog-outline'
import IconHelpCircleOutline from '~icons/mdi/help-circle-outline'
import IconChevronDown from '~icons/mdi/chevron-down'

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
  if (!target.closest('.scholar-user-menu')) {
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
  <section class="scholar-layout flex min-h-screen overflow-x-hidden" :data-theme="isDark ? 'vellum-dark' : 'vellum-light'">
    <!-- Left Sidebar -->
    <aside class="scholar-sidebar fixed left-0 top-0 z-20 flex h-screen w-64 flex-col overflow-y-auto">
      <!-- Logo Area -->
      <div class="px-6 py-8">
        <h1 class="text-editorial text-lg font-medium italic">DeepWrite</h1>
        <p class="label-sm mt-1 text-[var(--color-on-surface-variant)] opacity-60">{{ t('sidebar.tagline') }}</p>
      </div>

      <!-- Main Navigation -->
      <nav class="flex-1 space-y-1 px-3">
        <RouterLink
          :to="{ name: 'dashboard' }"
          class="scholar-nav-item group flex items-center gap-4 py-2.5 pl-4 transition-colors"
          :class="isNavActive('dashboard') ? 'nav-active' : ''"
        >
          <IconViewDashboardOutline class="h-5 w-5" />
          <span class="label-sm">{{ t('sidebar.workbench') }}</span>
        </RouterLink>

        <RouterLink
          :to="{ name: 'workspace-list' }"
          class="scholar-nav-item group flex items-center gap-4 py-2.5 pl-4 transition-colors"
          :class="isNavActive('workspace-list') ? 'nav-active' : ''"
        >
          <IconHomeOutline class="h-5 w-5" />
          <span class="label-sm">{{ t('sidebar.workspace') }}</span>
        </RouterLink>

        <RouterLink
          :to="{ name: 'library' }"
          class="scholar-nav-item group flex items-center gap-4 py-2.5 pl-4 transition-colors"
          :class="isNavActive('library') ? 'nav-active' : ''"
        >
          <IconBookmarkOutline class="h-5 w-5" />
          <span class="label-sm">{{ t('sidebar.library') }}</span>
        </RouterLink>

        <RouterLink
          :to="{ name: 'skills' }"
          class="scholar-nav-item group flex items-center gap-4 py-2.5 pl-4 transition-colors"
          :class="isNavActive('skills') ? 'nav-active' : ''"
        >
          <IconViewGridOutline class="h-5 w-5" />
          <span class="label-sm">{{ t('sidebar.skills') }}</span>
        </RouterLink>

        <RouterLink
          :to="{ name: 'agents' }"
          class="scholar-nav-item group flex items-center gap-4 py-2.5 pl-4 transition-colors"
          :class="isNavActive('agents') ? 'nav-active' : ''"
        >
          <IconRobotOutline class="h-5 w-5" />
          <span class="label-sm">{{ t('sidebar.agent') }}</span>
        </RouterLink>
      </nav>

      <!-- Bottom Section -->
      <div class="px-3 py-6">
        <RouterLink
          :to="{ name: 'profile' }"
          class="scholar-nav-item group flex items-center gap-4 py-2 pl-4 transition-colors"
        >
          <IconCogOutline class="h-5 w-5" />
          <span class="label-sm">{{ t('userMenu.settings') }}</span>
        </RouterLink>

        <RouterLink
          to="#"
          class="scholar-nav-item group flex items-center gap-4 py-2 pl-4 transition-colors"
        >
          <IconHelpCircleOutline class="h-5 w-5" />
          <span class="label-sm">{{ t('sidebar.help') }}</span>
        </RouterLink>

        <!-- User Profile Card -->
        <div class="scholar-user-card mt-6 flex items-center gap-3 rounded-lg px-3 py-3">
          <div class="scholar-avatar flex h-8 w-8 items-center justify-center rounded-full overflow-hidden">
            <img
              v-if="user?.avatarUrl"
              :src="user.avatarUrl"
              :alt="user.displayName"
              class="h-full w-full rounded-full object-cover"
            />
            <span v-else class="text-sm font-medium">{{ userInitial }}</span>
          </div>
          <div class="min-w-0 flex-1">
            <p class="label-sm truncate font-medium">{{ user?.displayName || t('sidebar.guestUser') }}</p>
            <p class="text-xs text-[var(--color-on-surface-variant)] opacity-70">{{ user?.role || 'Scholar' }}</p>
          </div>

          <!-- Notifications & Menu -->
          <div class="scholar-user-menu relative flex items-center">
            <button
              type="button"
              class="scholar-icon-btn relative rounded-md p-1.5"
              :aria-label="t('sidebar.notifications')"
              @click="openNotifications"
            >
              <IconBellOutline class="h-4 w-4" />
              <span
                v-if="pendingCount > 0"
                class="absolute -right-0.5 -top-0.5 min-w-4 rounded-full bg-[var(--color-error)] px-1 text-center text-[10px] font-semibold text-white"
              >
                {{ pendingCount > 99 ? '99+' : pendingCount }}
              </span>
            </button>

            <button
              type="button"
              class="scholar-icon-btn ml-1 rounded-md p-1.5"
              :aria-label="t('sidebar.openMenu')"
              @click="userMenuOpen = !userMenuOpen"
            >
              <IconChevronDown class="h-4 w-4 transition-transform" :class="{ 'rotate-180': userMenuOpen }" />
            </button>

            <!-- Dropdown Menu -->
            <Transition name="dropdown">
              <div
                v-if="userMenuOpen"
                class="scholar-dropdown absolute right-0 bottom-full z-50 mb-2 w-56 rounded-lg p-3"
              >
                <button
                  type="button"
                  class="scholar-dropdown-item w-full rounded-md px-3 py-2 text-left text-sm"
                  @click="openProfileCenter"
                >
                  {{ t('userMenu.profile') }}
                </button>

                <div class="my-2 h-px bg-[var(--surface-container-highest)]" />

                <label class="flex items-center justify-between rounded-md px-3 py-2 text-sm">
                  <span>{{ t('userMenu.darkMode') }}</span>
                  <input
                    type="checkbox"
                    class="toggle toggle-sm"
                    :checked="isDark"
                    @change="applyTheme(($event.target as HTMLInputElement).checked)"
                  />
                </label>

                <label class="mt-2 block px-3 py-1.5 text-xs text-[var(--color-on-surface-variant)]">{{ t('userMenu.language') }}</label>
                <select
                  v-model="selectedLanguage"
                  class="scholar-select mt-1 w-full rounded-md px-3 py-2 text-sm"
                  @change="updateUserPreferences"
                >
                  <option v-for="item in languageOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
                </select>

                <label class="mt-2 block px-3 py-1.5 text-xs text-[var(--color-on-surface-variant)]">{{ t('userMenu.timezone') }}</label>
                <select
                  v-model="selectedTimezone"
                  class="scholar-select mt-1 w-full rounded-md px-3 py-2 text-sm"
                  @change="updateUserPreferences"
                >
                  <option v-for="item in timezoneOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
                </select>

                <div class="my-2 h-px bg-[var(--surface-container-highest)]" />

                <button
                  type="button"
                  class="scholar-dropdown-item w-full rounded-md px-3 py-2 text-left text-sm text-[var(--color-error)]"
                  @click="handleLogout"
                >
                  {{ t('userMenu.logout') }}
                </button>
              </div>
            </Transition>
          </div>
        </div>
      </div>
    </aside>

    <!-- Main Content Area -->
    <main class="scholar-main ml-64 min-h-screen flex-1">
      <RouterView />
    </main>
  </section>
</template>

<style scoped>
.scholar-layout {
  background-color: var(--surface-base);
}

.scholar-sidebar {
  background-color: var(--surface-container-low);
  border: none;
}

.scholar-nav-item {
  color: var(--color-on-surface-variant);
  border-left: 2px solid transparent;
  border-radius: 0 0.375rem 0.375rem 0;
}

.scholar-nav-item:hover {
  background-color: var(--surface-container);
  color: var(--color-on-surface);
}

.scholar-nav-item.nav-active {
  color: var(--color-on-surface);
  border-left-color: var(--color-primary);
  font-weight: 500;
}

.scholar-user-card {
  background-color: var(--surface-container);
}

.scholar-avatar {
  background-color: var(--surface-container-high);
  color: var(--color-on-surface);
}

.scholar-icon-btn {
  color: var(--color-on-surface-variant);
  transition: background-color 0.2s, color 0.2s;
}

.scholar-icon-btn:hover {
  background-color: var(--surface-container-high);
  color: var(--color-on-surface);
}

.scholar-dropdown {
  background-color: var(--surface-raised);
  box-shadow: 0 8px 32px oklch(0.28 0.008 105 / 0.12);
}

[data-theme="vellum-dark"] .scholar-dropdown {
  box-shadow: 0 8px 32px oklch(0.15 0.02 75 / 0.25);
}

.scholar-dropdown-item {
  color: var(--color-on-surface);
  transition: background-color 0.2s;
}

.scholar-dropdown-item:hover {
  background-color: var(--surface-container);
}

.scholar-select {
  background-color: var(--surface-container);
  color: var(--color-on-surface);
  border: none;
  outline: none;
}

.scholar-select:focus {
  background-color: var(--surface-container-high);
}

.glow-blob {
  position: absolute;
  border-radius: 9999px;
  filter: blur(80px);
  pointer-events: none;
}

/* Dropdown Transition */
.dropdown-enter-active,
.dropdown-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(4px);
}
</style>