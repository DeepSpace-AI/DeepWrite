<script setup lang="ts">
import { computed, onBeforeMount, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { ApiError } from '@/api/http'
import { fetchDashboardOverview, type DashboardOverview } from '@/api/dashboard'
import { useUserStore } from '@/stores/user'
import IconRefresh from '~icons/mdi/refresh'

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
  { key: 'workspace', label: t('dashboard.summary.workspaces'), value: summary.value.workspace_count },
  { key: 'document', label: t('dashboard.summary.documents'), value: summary.value.document_count },
  { key: 'file', label: t('dashboard.summary.files'), value: summary.value.file_count },
  { key: 'invitation', label: t('dashboard.summary.invitations'), value: summary.value.pending_invitation_count },
  { key: 'collaborator', label: t('dashboard.summary.collaborators'), value: summary.value.collaborator_count },
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
  <section class="space-y-5">
    <section class="paper-card rounded-lg p-6">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div class="max-w-3xl">
          <p class="label-sm">{{ t('dashboard.overviewLabel') }}</p>
          <h2 class="text-editorial mt-2 text-3xl font-bold">{{ t('dashboard.welcomeBack', { name: userName }) }}</h2>
          <p class="body-lg text-pretty-secondary mt-2">{{ t('dashboard.subtitle') }}</p>
        </div>

        <div class="flex flex-col items-end gap-3">
          <button type="button" class="btn-tertiary px-4 py-2 rounded-lg text-sm" :disabled="isLoading" @click="loadOverview">
            <IconRefresh class="h-4 w-4 mr-2" />
            {{ t('dashboard.refresh') }}
          </button>
          <div class="rounded-xl bg-[var(--surface-overlay)] px-4 py-3 text-right text-sm">
            <p class="font-medium text-pretty-secondary">{{ user?.email || 'guest@deepwrite.local' }}</p>
            <p class="mt-1 text-pretty-muted">{{ userMeta }}</p>
          </div>
        </div>
      </div>

      <div class="mt-6 grid gap-4 sm:grid-cols-2 xl:grid-cols-5">
        <article v-for="item in metricCards" :key="item.key" class="surface-card rounded-xl px-4 py-4">
          <p class="label-sm">{{ item.label }}</p>
          <p class="mt-2 text-2xl font-semibold text-editorial">{{ formatNumber(item.value) }}</p>
        </article>
      </div>
    </section>

    <div v-if="loadError" class="paper-card rounded-xl p-4 text-sm">
      <p class="text-red-500">{{ loadError }}</p>
    </div>

    <section v-if="isLoading && !overview" class="grid gap-5 xl:grid-cols-2">
      <div class="skeleton h-72 rounded-lg" />
      <div class="skeleton h-72 rounded-lg" />
      <div class="skeleton h-72 rounded-lg" />
      <div class="skeleton h-72 rounded-lg" />
    </section>

    <section v-else class="grid gap-5 xl:grid-cols-2">
      <article class="surface-card rounded-lg p-5">
        <div class="mb-4 flex items-center justify-between gap-3">
          <div>
            <h3 class="text-lg font-semibold text-pretty">{{ t('dashboard.workspacesTitle') }}</h3>
            <p class="mt-1 text-sm text-pretty-muted">{{ t('dashboard.workspacesCount', { count: formatNumber(summary.workspace_count) }) }}</p>
          </div>
          <RouterLink :to="{ name: 'workspace-list' }" class="btn-tertiary px-3 py-1.5 rounded-lg text-sm">{{ t('dashboard.workspacesAction') }}</RouterLink>
        </div>

        <div v-if="!workspaces.length" class="rounded-xl bg-[var(--surface-overlay)] px-4 py-10 text-center text-sm text-pretty-secondary">
          <p>{{ t('dashboard.emptyWorkspaces') }}</p>
          <p class="mt-2 text-xs text-pretty-muted">{{ t('dashboard.emptySummary') }}</p>
        </div>

        <ul v-else class="space-y-3">
          <li v-for="item in workspaces" :key="item.id">
            <RouterLink :to="{ name: 'workspace-detail', params: { id: item.id } }" class="block rounded-xl bg-[var(--surface-overlay)] px-4 py-4 transition hover:bg-[var(--surface-sunken)]">
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p class="text-sm font-semibold text-pretty">{{ item.name }}</p>
                  <p class="mt-1 text-sm text-pretty-secondary">{{ item.description || t('dashboard.workspaceNoDescription') }}</p>
                </div>
                <div class="flex flex-wrap items-center gap-2 text-sm">
                  <span class="label-sm px-2 py-1 rounded-lg bg-[var(--surface-sunken)]">{{ roleLabel(item.role) }}</span>
                  <span class="label-sm px-2 py-1 rounded-lg bg-[var(--surface-sunken)]">{{ statusLabel(item.status) }}</span>
                  <span class="label-sm px-2 py-1 rounded-lg bg-[var(--surface-sunken)]">{{ visibilityLabel(item.public) }}</span>
                </div>
              </div>

              <p class="mt-3 text-sm text-pretty-secondary">
                {{ t('dashboard.workspaceMeta', { members: formatNumber(item.member_count), documents: formatNumber(item.document_count), files: formatNumber(item.file_count) }) }}
              </p>
              <p class="mt-2 text-sm text-pretty-muted">{{ t('dashboard.updatedAt', { time: formatDateTime(item.updated_at) }) }}</p>
            </RouterLink>
          </li>
        </ul>
      </article>

      <article class="surface-card rounded-lg p-5">
        <div class="mb-4 flex items-center justify-between gap-3">
          <div>
            <h3 class="text-lg font-semibold text-pretty">{{ t('dashboard.recentDocumentsTitle') }}</h3>
            <p class="mt-1 text-sm text-pretty-muted">{{ t('dashboard.recentDocumentsCount', { count: formatNumber(recentDocuments.length) }) }}</p>
          </div>
        </div>

        <div v-if="!recentDocuments.length" class="rounded-xl bg-[var(--surface-overlay)] px-4 py-10 text-center text-sm text-pretty-secondary">
          {{ t('dashboard.emptyDocuments') }}
        </div>

        <ul v-else class="space-y-3">
          <li v-for="item in recentDocuments" :key="item.id">
            <RouterLink
              :to="{ name: 'workspace-detail', params: { id: item.workspace_id }, query: { doc: item.id } }"
              class="block rounded-xl bg-[var(--surface-overlay)] px-4 py-4 transition hover:bg-[var(--surface-sunken)]"
            >
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p class="text-sm font-semibold text-pretty">{{ item.title }}</p>
                  <p class="mt-1 text-sm text-pretty-secondary">{{ t('dashboard.documentInWorkspace', { name: item.workspace_name || '--' }) }}</p>
                </div>
                <span class="label-sm px-2 py-1 rounded-lg bg-[var(--surface-sunken)]">{{ t('dashboard.documentVersion', { version: item.current_version || 1 }) }}</span>
              </div>

              <div class="mt-3 flex flex-wrap items-center justify-between gap-3 text-sm text-pretty-secondary">
                <span>{{ t('dashboard.updatedAt', { time: formatDateTime(item.updated_at) }) }}</span>
                <span class="text-[var(--color-primary)] hover:opacity-80">
                  {{ t('dashboard.recentDocumentsAction') }}
                </span>
              </div>
            </RouterLink>
          </li>
        </ul>
      </article>

      <article class="surface-card rounded-lg p-5">
        <div class="mb-4 flex items-center justify-between gap-3">
          <div>
            <h3 class="text-lg font-semibold text-pretty">{{ t('dashboard.pendingInvitationsTitle') }}</h3>
            <p class="mt-1 text-sm text-pretty-muted">{{ t('dashboard.pendingInvitationsCount', { count: formatNumber(summary.pending_invitation_count) }) }}</p>
          </div>
          <RouterLink :to="{ name: 'notifications' }" class="btn-tertiary px-3 py-1.5 rounded-lg text-sm">{{ t('dashboard.pendingInvitationsAction') }}</RouterLink>
        </div>

        <div v-if="!pendingInvitations.length" class="rounded-xl bg-[var(--surface-overlay)] px-4 py-10 text-center text-sm text-pretty-secondary">
          {{ t('dashboard.emptyInvitations') }}
        </div>

        <ul v-else class="space-y-3">
          <li v-for="item in pendingInvitations" :key="item.id" class="rounded-xl bg-[var(--surface-overlay)] px-4 py-4">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p class="text-sm font-semibold text-pretty">{{ item.workspace_name || '--' }}</p>
                <p class="mt-1 text-sm text-pretty-secondary">{{ t('dashboard.invitationRole', { role: roleLabel(item.role) }) }}</p>
              </div>
              <span class="label-sm px-2 py-1 rounded-lg bg-amber-500/15 text-amber-600">{{ formatRelativeTime(item.expires_at) }}</span>
            </div>

            <div class="mt-3 space-y-1 text-sm text-pretty-muted">
              <p>{{ t('dashboard.invitationEmail', { email: item.invitee_email || '--' }) }}</p>
              <p>{{ t('dashboard.invitationExpires', { time: formatDateTime(item.expires_at) }) }}</p>
            </div>
          </li>
        </ul>
      </article>

      <article class="surface-card rounded-lg p-5">
        <div>
          <h3 class="text-lg font-semibold text-pretty">{{ t('dashboard.resourceTitle') }}</h3>
          <p class="mt-1 body-lg text-pretty-secondary">{{ t('dashboard.resourceSubtitle') }}</p>
        </div>

        <div class="mt-5 grid gap-3 sm:grid-cols-3">
          <div class="surface-card rounded-xl px-4 py-4">
            <p class="label-sm">{{ t('dashboard.resourceActive') }}</p>
            <p class="mt-2 text-2xl font-semibold text-editorial">{{ formatNumber(summary.active_workspace_count) }}</p>
          </div>
          <div class="surface-card rounded-xl px-4 py-4">
            <p class="label-sm">{{ t('dashboard.resourceMembers') }}</p>
            <p class="mt-2 text-2xl font-semibold text-editorial">{{ formatNumber(summary.collaborator_count) }}</p>
          </div>
          <div class="surface-card rounded-xl px-4 py-4">
            <p class="label-sm">{{ t('dashboard.resourcePending') }}</p>
            <p class="mt-2 text-2xl font-semibold text-editorial">{{ formatNumber(summary.pending_invitation_count) }}</p>
          </div>
        </div>

        <div class="mt-5 space-y-2 body-md text-pretty-secondary">
          <p>{{ t('dashboard.workspaceMeta', { members: formatNumber(summary.collaborator_count), documents: formatNumber(summary.document_count), files: formatNumber(summary.file_count) }) }}</p>
          <p>{{ t('dashboard.updatedAt', { time: recentDocuments[0] ? formatDateTime(recentDocuments[0].updated_at) : '--' }) }}</p>
        </div>

        <div class="mt-5 flex flex-wrap gap-2">
          <RouterLink :to="{ name: 'workspace-list' }" class="btn-primary-vellum px-4 py-2 rounded-lg text-sm">{{ t('workspace.createBtn') }}</RouterLink>
          <RouterLink :to="{ name: 'notifications' }" class="btn-tertiary px-4 py-2 rounded-lg text-sm">{{ t('dashboard.openNotifications') }}</RouterLink>
        </div>
      </article>
    </section>
  </section>
</template>
