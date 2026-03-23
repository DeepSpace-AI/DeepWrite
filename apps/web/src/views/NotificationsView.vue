<script setup lang="ts">
import { computed, onBeforeMount, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { useNotificationStore } from '@/stores/notification'
import type { WorkspaceInvitation } from '@/api/workspace'
import IconBellOutline from '~icons/mdi/bell-outline'
import IconRefresh from '~icons/mdi/refresh'

const notificationStore = useNotificationStore()
const { t } = useI18n()

const currentFilter = ref<'pending' | 'all'>('pending')
const actionLoadingMap = ref<Record<string, boolean>>({})
const pageError = ref('')

const { invitations, isLoading, pendingCount } = storeToRefs(notificationStore)

const isPendingMode = computed(() => currentFilter.value === 'pending')

function formatDate(value: string) {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function statusBadgeClass(status: string) {
  if (status === 'pending') return 'badge-warning'
  if (status === 'accepted') return 'badge-success'
  if (status === 'rejected') return 'badge-error'
  if (status === 'revoked') return 'badge-neutral'
  if (status === 'expired') return 'badge-ghost'
  return 'badge-ghost'
}

function workspaceShortId(item: WorkspaceInvitation) {
  return item.workspace_id?.slice(0, 8) || '--'
}

async function fetchInvitations() {
  pageError.value = ''
  try {
    const status = currentFilter.value === 'pending' ? 'pending' : undefined
    await notificationStore.loadInvitations(status)
  } catch (error) {
    pageError.value = error instanceof Error ? error.message : t('notifications.loadFailed')
  }
}

function setFilter(filter: 'pending' | 'all') {
  if (currentFilter.value === filter) return
  currentFilter.value = filter
  fetchInvitations()
}

async function resolveInvitation(item: WorkspaceInvitation, action: 'accept' | 'reject') {
  pageError.value = ''
  actionLoadingMap.value[item.id] = true
  try {
    if (action === 'accept') {
      await notificationStore.acceptInvitation(item)
    } else {
      await notificationStore.rejectInvitation(item)
    }

    await notificationStore.refreshPendingCount()
    if (isPendingMode.value) {
      await fetchInvitations()
    }
  } catch (error) {
    pageError.value = error instanceof Error ? error.message : t('notifications.actionFailed')
  } finally {
    actionLoadingMap.value[item.id] = false
  }
}

onBeforeMount(async () => {
  await fetchInvitations()
  await notificationStore.refreshPendingCount()
})
</script>

<template>
  <section class="space-y-5">
    <section class="paper-card rounded-lg p-5">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p class="text-[11px] font-mono uppercase tracking-[0.2em] text-pretty-muted">Workbench / Notifications</p>
          <h2 class="heading-serif mt-2 text-3xl font-bold text-pretty">{{ t('notifications.title') }}</h2>
          <p class="mt-2 max-w-3xl text-sm leading-7 text-pretty-secondary">{{ t('notifications.subtitle') }}</p>
        </div>
        <div class="flex items-center gap-2">
          <span class="badge rounded-md bg-amber-500/15 text-amber-600 text-xs">{{ t('notifications.pendingCount', { count: pendingCount }) }}</span>
          <button type="button" class="btn-tertiary rounded-md px-3 py-1.5 text-sm" :disabled="isLoading" @click="fetchInvitations">
            <IconRefresh class="h-4 w-4" />
            {{ t('notifications.refresh') }}
          </button>
        </div>
      </div>

      <div class="mt-4 flex items-center gap-2">
        <button
          type="button"
          class="rounded-md px-3 py-1.5 text-sm"
          :class="isPendingMode ? 'btn-primary-vellum' : 'btn-tertiary text-pretty-secondary hover:text-pretty'"
          @click="setFilter('pending')"
        >
          {{ t('notifications.pendingTab') }}
        </button>
        <button
          type="button"
          class="rounded-md px-3 py-1.5 text-sm"
          :class="!isPendingMode ? 'btn-primary-vellum' : 'btn-tertiary text-pretty-secondary hover:text-pretty'"
          @click="setFilter('all')"
        >
          {{ t('notifications.allTab') }}
        </button>
      </div>
    </section>

    <section class="paper-card rounded-lg p-5">
      <div v-if="pageError" class="mb-4 rounded-md bg-error/8 p-4 text-sm text-error">{{ pageError }}</div>

      <div v-if="isLoading" class="space-y-3">
        <div class="skeleton h-20 w-full rounded-lg" />
        <div class="skeleton h-20 w-full rounded-lg" />
      </div>

      <div v-else-if="!invitations.length" class="rounded-md bg-[var(--surface-overlay)] px-4 py-10 text-center">
        <IconBellOutline class="mx-auto h-6 w-6 text-pretty-muted" />
        <p class="mt-3 text-sm text-pretty-secondary">{{ t('notifications.empty') }}</p>
      </div>

      <ul v-else class="space-y-3">
        <li v-for="item in invitations" :key="item.id" class="surface-card rounded-md p-4">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <p class="text-sm font-semibold text-pretty">{{ t('notifications.workspaceInvite') }}</p>
              <p class="mt-1 text-xs text-pretty-secondary">
                {{ t('notifications.workspaceLabel', { id: workspaceShortId(item) }) }}
              </p>
            </div>
            <span class="badge rounded-md text-xs" :class="statusBadgeClass(item.status)">
              {{ t(`notifications.status.${item.status}`, item.status) }}
            </span>
          </div>

          <div class="mt-3 grid gap-2 text-xs text-pretty-secondary md:grid-cols-3">
            <p>{{ t('notifications.inviteeEmail', { email: item.invitee_email || '--' }) }}</p>
            <p>{{ t('notifications.roleLabel', { role: item.role }) }}</p>
            <p>{{ t('notifications.expireAt', { time: formatDate(item.expires_at) }) }}</p>
          </div>

          <div v-if="item.status === 'pending'" class="mt-4 rounded-md bg-[var(--surface-overlay)] p-3">
            <div class="flex items-center gap-2">
              <button
                type="button"
                class="rounded-md bg-emerald-500/85 px-3 py-1.5 text-sm text-white hover:bg-emerald-500"
                :disabled="actionLoadingMap[item.id]"
                @click="resolveInvitation(item, 'accept')"
              >
                {{ actionLoadingMap[item.id] ? t('notifications.processing') : t('notifications.accept') }}
              </button>
              <button
                type="button"
                class="rounded-md bg-red-500/85 px-3 py-1.5 text-sm text-white hover:bg-red-500"
                :disabled="actionLoadingMap[item.id]"
                @click="resolveInvitation(item, 'reject')"
              >
                {{ actionLoadingMap[item.id] ? t('notifications.processing') : t('notifications.reject') }}
              </button>
            </div>
          </div>
        </li>
      </ul>
    </section>
  </section>
</template>
