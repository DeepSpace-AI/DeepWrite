<script setup lang="ts">
import { computed, onBeforeMount, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { ApiError } from '@/api/http'
import { fetchDashboardOverview, type DashboardOverview } from '@/api/dashboard'
import { useUserStore } from '@/stores/user'
import IconRefresh from '~icons/mdi/refresh'
import IconBookOpenPageVariant from '~icons/mdi/book-open-page-variant'
import IconFileDocumentOutline from '~icons/mdi/file-document-outline'
import IconAccountGroupOutline from '~icons/mdi/account-group-outline'
import IconBellOutline from '~icons/mdi/bell-outline'
import IconFolderOutline from '~icons/mdi/folder-outline'
import IconArrowRight from '~icons/mdi/arrow-right'
import IconClockOutline from '~icons/mdi/clock-outline'

const userStore = useUserStore()
const { user } = storeToRefs(userStore)
const { t, locale } = useI18n()

const overview = ref<DashboardOverview | null>(null)
const isLoading = ref(false)
const loadError = ref('')

const emptySummary = {
  workspace_count: 0,
  active_workspace_count: 0,
  document_count: 0,
  file_count: 0,
  pending_invitation_count: 0,
  collaborator_count: 0,
}

const userName = computed(() => user.value?.displayName || t('common.writer'))
const userMeta = computed(() => {
  if (!user.value) return t('dashboard.notLoggedIn')
  return `${user.value.role} · ${user.value.language || 'en'} · ${user.value.timezone || 'UTC'}`
})

const summary = computed(() => overview.value?.summary ?? emptySummary)
const workspaces = computed(() => overview.value?.workspaces ?? [])
const recentDocuments = computed(() => overview.value?.recent_documents ?? [])
const pendingInvitations = computed(() => overview.value?.pending_invitations ?? [])

const metricCards = computed(() => [
  { key: 'workspace', label: t('dashboard.summary.workspaces'), value: summary.value.workspace_count, icon: IconFolderOutline },
  { key: 'document', label: t('dashboard.summary.documents'), value: summary.value.document_count, icon: IconFileDocumentOutline },
  { key: 'file', label: t('dashboard.summary.files'), value: summary.value.file_count, icon: IconBookOpenPageVariant },
  { key: 'collaborator', label: t('dashboard.summary.collaborators'), value: summary.value.collaborator_count, icon: IconAccountGroupOutline },
  { key: 'invitation', label: t('dashboard.summary.invitations'), value: summary.value.pending_invitation_count, icon: IconBellOutline },
])

function formatNumber(value: number) {
  return new Intl.NumberFormat(locale.value).format(value)
}

function formatDateTime(value: string) {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat(locale.value, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(date)
}

function formatRelativeTime(value: string) {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value

  const diff = date.getTime() - Date.now()
  const minutes = Math.round(diff / 60000)
  const rtf = new Intl.RelativeTimeFormat(locale.value, { numeric: 'auto' })

  if (Math.abs(minutes) < 60) return rtf.format(minutes, 'minute')

  const hours = Math.round(minutes / 60)
  if (Math.abs(hours) < 24) return rtf.format(hours, 'hour')

  const days = Math.round(hours / 24)
  return rtf.format(days, 'day')
}

function roleLabel(role: string) {
  if (role === 'owner') return t('workspace.roleOwner')
  if (role === 'admin') return t('workspace.roleAdmin')
  if (role === 'editor') return t('workspace.roleEditor')
  if (role === 'viewer') return t('workspace.roleViewer')
  return role || '--'
}

function statusLabel(status: string) {
  if (status === 'archived') return t('workspace.statusArchived')
  return t('workspace.statusActive')
}

function visibilityLabel(isPublic: boolean) {
  return isPublic ? t('dashboard.workspaceVisibilityPublic') : t('dashboard.workspaceVisibilityPrivate')
}

function getGreeting() {
  const hour = new Date().getHours()
  if (hour < 12) return t('dashboard.greeting.morning')
  if (hour < 18) return t('dashboard.greeting.afternoon')
  return t('dashboard.greeting.evening')
}

async function loadOverview() {
  isLoading.value = true
  loadError.value = ''

  try {
    overview.value = await fetchDashboardOverview()
  } catch (error) {
    loadError.value = error instanceof ApiError ? error.message : t('dashboard.loadError')
  } finally {
    isLoading.value = false
  }
}

onBeforeMount(loadOverview)
</script>

<template>
  <section class="scholar-dashboard space-y-6 p-8">
    <!-- Hero Welcome Section -->
    <section class="scholar-welcome paper-card rounded-lg p-8">
      <div class="flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between">
        <div class="max-w-2xl">
          <p class="label-sm text-[var(--color-on-surface-variant)]">{{ t('dashboard.overviewLabel') }}</p>
          <h1 class="text-editorial mt-3 text-3xl font-semibold lg:text-4xl">
            {{ getGreeting() }}, {{ userName }}.
          </h1>
          <blockquote class="body-lg mt-4 text-[var(--color-on-surface-variant)] italic">
            "{{ t('dashboard.quote') }}"
          </blockquote>
        </div>

        <div class="flex flex-col items-start gap-4 lg:items-end">
          <button
            type="button"
            class="btn-tertiary flex items-center gap-2 rounded-lg px-4 py-2 text-sm"
            :disabled="isLoading"
            @click="loadOverview"
          >
            <IconRefresh class="h-4 w-4" :class="{ 'animate-spin': isLoading }" />
            {{ t('dashboard.refresh') }}
          </button>
          <div class="scholar-user-card rounded-xl p-4 text-sm">
            <p class="font-medium text-[var(--color-on-surface)]">{{ user?.email || 'guest@deepwrite.local' }}</p>
            <p class="mt-1 text-[var(--color-on-surface-variant)]">{{ userMeta }}</p>
          </div>
        </div>
      </div>

      <!-- Metrics Grid -->
      <div class="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
        <article
          v-for="item in metricCards"
          :key="item.key"
          class="scholar-metric-card rounded-xl p-5 transition-colors"
        >
          <div class="flex items-start justify-between">
            <p class="label-sm text-[var(--color-on-surface-variant)]">{{ item.label }}</p>
            <component :is="item.icon" class="h-5 w-5 text-[var(--color-tertiary)]" />
          </div>
          <p class="text-editorial mt-3 text-3xl font-semibold">{{ formatNumber(item.value) }}</p>
        </article>
      </div>
    </section>

    <!-- Error State -->
    <div v-if="loadError" class="paper-card rounded-xl p-4">
      <p class="text-[var(--color-error)]">{{ loadError }}</p>
    </div>

    <!-- Loading State -->
    <section v-if="isLoading && !overview" class="grid gap-6 lg:grid-cols-2">
      <div class="skeleton h-80 rounded-lg" />
      <div class="skeleton h-80 rounded-lg" />
      <div class="skeleton h-80 rounded-lg" />
      <div class="skeleton h-80 rounded-lg" />
    </section>

    <!-- Main Content Grid -->
    <section v-else class="grid gap-6 lg:grid-cols-2">
      <!-- Recent Workspaces -->
      <article class="scholar-section paper-card rounded-lg p-6">
        <div class="mb-5 flex items-start justify-between gap-4">
          <div>
            <h2 class="text-editorial text-xl font-semibold">{{ t('dashboard.workspacesTitle') }}</h2>
            <p class="body-md mt-1 text-[var(--color-on-surface-variant)]">
              {{ t('dashboard.workspacesCount', { count: formatNumber(summary.workspace_count) }) }}
            </p>
          </div>
          <RouterLink
            :to="{ name: 'workspace-list' }"
            class="btn-tertiary flex items-center gap-1 rounded-lg px-3 py-2 text-sm"
          >
            {{ t('dashboard.workspacesAction') }}
            <IconArrowRight class="h-4 w-4" />
          </RouterLink>
        </div>

        <div
          v-if="!workspaces.length"
          class="scholar-empty rounded-xl p-10 text-center"
        >
          <p class="text-[var(--color-on-surface-variant)]">{{ t('dashboard.emptyWorkspaces') }}</p>
          <p class="mt-2 text-sm text-[var(--color-on-surface-muted)]">{{ t('dashboard.emptySummary') }}</p>
        </div>

        <ul v-else class="space-y-3">
          <li v-for="item in workspaces" :key="item.id">
            <RouterLink
              :to="{ name: 'workspace-detail', params: { id: item.id } }"
              class="scholar-list-item block rounded-xl p-4 transition-colors"
            >
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div class="min-w-0 flex-1">
                  <p class="truncate font-medium text-[var(--color-on-surface)]">{{ item.name }}</p>
                  <p class="mt-1 text-sm text-[var(--color-on-surface-variant)]">
                    {{ item.description || t('dashboard.workspaceNoDescription') }}
                  </p>
                </div>
                <div class="flex flex-wrap items-center gap-2">
                  <span class="scholar-chip rounded-lg px-2 py-1 text-xs">{{ roleLabel(item.role) }}</span>
                  <span class="scholar-chip rounded-lg px-2 py-1 text-xs">{{ statusLabel(item.status) }}</span>
                  <span class="scholar-chip rounded-lg px-2 py-1 text-xs">{{ visibilityLabel(item.public) }}</span>
                </div>
              </div>

              <div class="mt-3 flex items-center justify-between text-sm text-[var(--color-on-surface-variant)]">
                <span>{{ t('dashboard.workspaceMeta', { members: formatNumber(item.member_count), documents: formatNumber(item.document_count), files: formatNumber(item.file_count) }) }}</span>
                <span class="flex items-center gap-1 text-[var(--color-on-surface-muted)]">
                  <IconClockOutline class="h-4 w-4" />
                  {{ t('dashboard.updatedAt', { time: formatDateTime(item.updated_at) }) }}
                </span>
              </div>
            </RouterLink>
          </li>
        </ul>
      </article>

      <!-- Recent Documents -->
      <article class="scholar-section paper-card rounded-lg p-6">
        <div class="mb-5 flex items-start justify-between gap-4">
          <div>
            <h2 class="text-editorial text-xl font-semibold">{{ t('dashboard.recentDocumentsTitle') }}</h2>
            <p class="body-md mt-1 text-[var(--color-on-surface-variant)]">
              {{ t('dashboard.recentDocumentsCount', { count: formatNumber(recentDocuments.length) }) }}
            </p>
          </div>
        </div>

        <div
          v-if="!recentDocuments.length"
          class="scholar-empty rounded-xl p-10 text-center"
        >
          <p class="text-[var(--color-on-surface-variant)]">{{ t('dashboard.emptyDocuments') }}</p>
        </div>

        <ul v-else class="space-y-3">
          <li v-for="item in recentDocuments" :key="item.id">
            <RouterLink
              :to="{ name: 'workspace-detail', params: { id: item.workspace_id }, query: { doc: item.id } }"
              class="scholar-list-item block rounded-xl p-4 transition-colors"
            >
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div class="min-w-0 flex-1">
                  <p class="truncate font-medium text-[var(--color-on-surface)]">{{ item.title }}</p>
                  <p class="mt-1 text-sm text-[var(--color-on-surface-variant)]">
                    {{ t('dashboard.documentInWorkspace', { name: item.workspace_name || '--' }) }}
                  </p>
                </div>
                <span class="scholar-chip rounded-lg px-2 py-1 text-xs">
                  {{ t('dashboard.documentVersion', { version: item.current_version || 1 }) }}
                </span>
              </div>

              <div class="mt-3 flex items-center justify-between text-sm">
                <span class="flex items-center gap-1 text-[var(--color-on-surface-muted)]">
                  <IconClockOutline class="h-4 w-4" />
                  {{ t('dashboard.updatedAt', { time: formatDateTime(item.updated_at) }) }}
                </span>
                <span class="text-[var(--color-tertiary)] hover:opacity-80">
                  {{ t('dashboard.recentDocumentsAction') }}
                </span>
              </div>
            </RouterLink>
          </li>
        </ul>
      </article>

      <!-- Pending Invitations -->
      <article class="scholar-section paper-card rounded-lg p-6">
        <div class="mb-5 flex items-start justify-between gap-4">
          <div>
            <h2 class="text-editorial text-xl font-semibold">{{ t('dashboard.pendingInvitationsTitle') }}</h2>
            <p class="body-md mt-1 text-[var(--color-on-surface-variant)]">
              {{ t('dashboard.pendingInvitationsCount', { count: formatNumber(summary.pending_invitation_count) }) }}
            </p>
          </div>
          <RouterLink
            :to="{ name: 'notifications' }"
            class="btn-tertiary flex items-center gap-1 rounded-lg px-3 py-2 text-sm"
          >
            {{ t('dashboard.pendingInvitationsAction') }}
            <IconArrowRight class="h-4 w-4" />
          </RouterLink>
        </div>

        <div
          v-if="!pendingInvitations.length"
          class="scholar-empty rounded-xl p-10 text-center"
        >
          <p class="text-[var(--color-on-surface-variant)]">{{ t('dashboard.emptyInvitations') }}</p>
        </div>

        <ul v-else class="space-y-3">
          <li
            v-for="item in pendingInvitations"
            :key="item.id"
            class="scholar-list-item rounded-xl p-4"
          >
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p class="font-medium text-[var(--color-on-surface)]">{{ item.workspace_name || '--' }}</p>
                <p class="mt-1 text-sm text-[var(--color-on-surface-variant)]">
                  {{ t('dashboard.invitationRole', { role: roleLabel(item.role) }) }}
                </p>
              </div>
              <span class="scholar-chip scholar-chip-warning rounded-lg px-2 py-1 text-xs">
                {{ formatRelativeTime(item.expires_at) }}
              </span>
            </div>

            <div class="mt-3 space-y-1 text-sm text-[var(--color-on-surface-muted)]">
              <p>{{ t('dashboard.invitationEmail', { email: item.invitee_email || '--' }) }}</p>
              <p>{{ t('dashboard.invitationExpires', { time: formatDateTime(item.expires_at) }) }}</p>
            </div>
          </li>
        </ul>
      </article>

      <!-- Resource Overview -->
      <article class="scholar-section paper-card rounded-lg p-6">
        <div class="mb-5">
          <h2 class="text-editorial text-xl font-semibold">{{ t('dashboard.resourceTitle') }}</h2>
          <p class="body-md mt-1 text-[var(--color-on-surface-variant)]">{{ t('dashboard.resourceSubtitle') }}</p>
        </div>

        <div class="grid gap-4 sm:grid-cols-3">
          <div class="scholar-metric-card rounded-xl p-5">
            <p class="label-sm text-[var(--color-on-surface-variant)]">{{ t('dashboard.resourceActive') }}</p>
            <p class="text-editorial mt-3 text-3xl font-semibold">{{ formatNumber(summary.active_workspace_count) }}</p>
          </div>
          <div class="scholar-metric-card rounded-xl p-5">
            <p class="label-sm text-[var(--color-on-surface-variant)]">{{ t('dashboard.resourceMembers') }}</p>
            <p class="text-editorial mt-3 text-3xl font-semibold">{{ formatNumber(summary.collaborator_count) }}</p>
          </div>
          <div class="scholar-metric-card rounded-xl p-5">
            <p class="label-sm text-[var(--color-on-surface-variant)]">{{ t('dashboard.resourcePending') }}</p>
            <p class="text-editorial mt-3 text-3xl font-semibold">{{ formatNumber(summary.pending_invitation_count) }}</p>
          </div>
        </div>

        <div class="mt-5 space-y-2 text-sm text-[var(--color-on-surface-variant)]">
          <p>{{ t('dashboard.workspaceMeta', { members: formatNumber(summary.collaborator_count), documents: formatNumber(summary.document_count), files: formatNumber(summary.file_count) }) }}</p>
          <p>{{ t('dashboard.updatedAt', { time: recentDocuments[0] ? formatDateTime(recentDocuments[0].updated_at) : '--' }) }}</p>
        </div>

        <div class="mt-5 flex flex-wrap gap-3">
          <RouterLink
            :to="{ name: 'workspace-list' }"
            class="btn-primary-vellum rounded-lg px-5 py-2.5 text-sm"
          >
            {{ t('workspace.createBtn') }}
          </RouterLink>
          <RouterLink
            :to="{ name: 'notifications' }"
            class="btn-tertiary rounded-lg px-5 py-2.5 text-sm"
          >
            {{ t('dashboard.openNotifications') }}
          </RouterLink>
        </div>
      </article>
    </section>
  </section>
</template>

<style scoped>
.scholar-welcome {
  background-color: var(--surface-raised);
}

.scholar-user-card {
  background-color: var(--surface-overlay);
}

.scholar-metric-card {
  background-color: var(--surface-focus);
}

.scholar-metric-card:hover {
  background-color: var(--surface-container-low);
}

.scholar-section {
  background-color: var(--surface-raised);
}

.scholar-empty {
  background-color: var(--surface-overlay);
}

.scholar-list-item {
  background-color: var(--surface-overlay);
}

.scholar-list-item:hover {
  background-color: var(--surface-sunken);
}

.scholar-chip {
  background-color: var(--surface-sunken);
  color: var(--color-on-surface-variant);
}

.scholar-chip-warning {
  background-color: color-mix(in oklch, var(--color-error) 15%, transparent);
  color: var(--color-error);
}

.paper-card {
  box-shadow: 0 8px 32px oklch(0.28 0.008 105 / 0.08);
}

[data-theme="vellum-dark"] .paper-card {
  box-shadow: 0 8px 32px oklch(0.15 0.02 75 / 0.20);
}

.btn-primary-vellum {
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-dim));
  color: var(--color-on-primary);
  font-family: var(--font-sans);
  font-weight: 500;
  transition: opacity 0.3s ease-out;
}

.btn-primary-vellum:hover {
  opacity: 0.88;
}

.btn-tertiary {
  background-color: transparent;
  color: var(--color-primary);
  font-family: var(--font-sans);
  font-weight: 500;
  transition: background-color 0.3s ease-out;
}

.btn-tertiary:hover {
  background-color: var(--surface-sunken);
}

.btn-tertiary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>