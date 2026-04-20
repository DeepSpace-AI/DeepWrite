<script setup lang="ts">
import { computed, onBeforeMount, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { useNotificationStore } from '@/stores/notification'
import type { WorkspaceInvitation } from '@/api/workspace'
import IconBellOutline from '~icons/mdi/bell-outline'
import IconRefresh from '~icons/mdi/refresh'
import IconCheck from '~icons/mdi/check'
import IconClose from '~icons/mdi/close'
import IconClockOutline from '~icons/mdi/clock-outline'

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

function statusClass(status: string) {
  switch (status) {
    case 'pending':
      return 'status-pending'
    case 'accepted':
      return 'status-accepted'
    case 'rejected':
      return 'status-rejected'
    case 'expired':
      return 'status-expired'
    default:
      return 'status-default'
  }
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
  <section class="scholar-notifications p-8">
    <div class="mx-auto max-w-4xl">
      <!-- Header -->
      <header class="mb-8">
        <p class="label-sm text-[var(--color-on-surface-variant)] uppercase tracking-widest">
          Workbench / Notifications
        </p>
        <h1 class="text-editorial mt-3 text-4xl font-light text-[var(--color-on-background)]">
          {{ t('notifications.title') }}
        </h1>
        <p class="body-md mt-3 text-[var(--color-on-surface-variant)]">
          {{ t('notifications.subtitle') }}
        </p>
      </header>

      <!-- Stats & Actions -->
      <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
        <div class="flex items-center gap-3">
          <span class="scholar-badge rounded-full px-4 py-2 text-sm">
            {{ t('notifications.pendingCount', { count: pendingCount }) }}
          </span>
          <button
            type="button"
            class="scholar-refresh-btn flex items-center gap-2 rounded-lg px-4 py-2"
            :disabled="isLoading"
            @click="fetchInvitations"
          >
            <IconRefresh class="h-4 w-4" :class="{ 'animate-spin': isLoading }" />
            {{ t('notifications.refresh') }}
          </button>
        </div>

        <!-- Filter Tabs -->
        <div class="scholar-tabs flex rounded-lg p-1">
          <button
            type="button"
            class="scholar-tab rounded-lg px-4 py-2"
            :class="isPendingMode ? 'active' : ''"
            @click="setFilter('pending')"
          >
            {{ t('notifications.pendingTab') }}
          </button>
          <button
            type="button"
            class="scholar-tab rounded-lg px-4 py-2"
            :class="!isPendingMode ? 'active' : ''"
            @click="setFilter('all')"
          >
            {{ t('notifications.allTab') }}
          </button>
        </div>
      </div>

      <!-- Error -->
      <div v-if="pageError" class="scholar-error mb-6 rounded-lg p-4">
        {{ pageError }}
      </div>

      <!-- Loading -->
      <div v-if="isLoading" class="space-y-4">
        <div class="skeleton h-24 rounded-xl" />
        <div class="skeleton h-24 rounded-xl" />
      </div>

      <!-- Empty State -->
      <div v-else-if="!invitations.length" class="scholar-empty rounded-xl p-12 text-center">
        <IconBellOutline class="mx-auto h-12 w-12 text-[var(--color-on-surface-muted)]" />
        <p class="mt-4 text-[var(--color-on-surface-variant)]">{{ t('notifications.empty') }}</p>
      </div>

      <!-- Notifications List -->
      <ul v-else class="space-y-4">
        <li
          v-for="item in invitations"
          :key="item.id"
          class="scholar-notification rounded-xl p-6"
        >
          <!-- Header -->
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h3 class="text-editorial text-lg font-medium text-[var(--color-on-background)]">
                {{ t('notifications.workspaceInvite') }}
              </h3>
              <p class="label-sm mt-1 text-[var(--color-on-surface-variant)]">
                {{ t('notifications.workspaceLabel', { id: workspaceShortId(item) }) }}
              </p>
            </div>
            <span class="scholar-status rounded-full px-3 py-1 text-xs" :class="statusClass(item.status)">
              {{ t(`notifications.status.${item.status}`, item.status) }}
            </span>
          </div>

          <!-- Meta -->
          <div class="mt-4 flex flex-wrap items-center gap-6 text-sm text-[var(--color-on-surface-variant)]">
            <span>{{ t('notifications.inviteeEmail', { email: item.invitee_email || '--' }) }}</span>
            <span>{{ t('notifications.roleLabel', { role: item.role }) }}</span>
            <span class="flex items-center gap-1">
              <IconClockOutline class="h-4 w-4" />
              {{ t('notifications.expireAt', { time: formatDate(item.expires_at) }) }}
            </span>
          </div>

          <!-- Actions -->
          <div v-if="item.status === 'pending'" class="mt-6 flex items-center gap-3 rounded-lg p-4">
            <button
              type="button"
              class="scholar-accept-btn flex items-center gap-2 rounded-lg px-5 py-2.5"
              :disabled="actionLoadingMap[item.id]"
              @click="resolveInvitation(item, 'accept')"
            >
              <IconCheck class="h-4 w-4" />
              {{ actionLoadingMap[item.id] ? t('notifications.processing') : t('notifications.accept') }}
            </button>
            <button
              type="button"
              class="scholar-reject-btn flex items-center gap-2 rounded-lg px-5 py-2.5"
              :disabled="actionLoadingMap[item.id]"
              @click="resolveInvitation(item, 'reject')"
            >
              <IconClose class="h-4 w-4" />
              {{ actionLoadingMap[item.id] ? t('notifications.processing') : t('notifications.reject') }}
            </button>
          </div>
        </li>
      </ul>
    </div>
  </section>
</template>

<style scoped>
.scholar-badge {
  background-color: color-mix(in oklch, var(--color-tertiary) 15%, transparent);
  color: var(--color-tertiary);
}

.scholar-refresh-btn {
  background-color: transparent;
  color: var(--color-primary);
  font-weight: 500;
  transition: background-color 0.2s;
}

.scholar-refresh-btn:hover {
  background-color: var(--surface-container-high);
}

.scholar-refresh-btn:disabled {
  opacity: 0.5;
}

.scholar-tabs {
  background-color: var(--surface-container-low);
}

.scholar-tab {
  color: var(--color-on-surface-variant);
  font-size: 0.875rem;
  font-weight: 500;
  transition: background-color 0.2s, color 0.2s;
}

.scholar-tab:hover {
  background-color: var(--surface-container);
}

.scholar-tab.active {
  background-color: var(--surface-container-lowest);
  color: var(--color-on-surface);
}

.scholar-error {
  background-color: var(--color-error-container);
  color: var(--color-on-error-container);
}

.scholar-empty {
  background-color: var(--surface-container-low);
}

.scholar-notification {
  background-color: var(--surface-container-lowest);
  box-shadow: 0 4px 24px oklch(0.28 0.008 105 / 0.06);
}

[data-theme="vellum-dark"] .scholar-notification {
  box-shadow: 0 4px 24px oklch(0.15 0.02 75 / 0.15);
}

.scholar-status {
  font-weight: 500;
  text-transform: capitalize;
}

.scholar-status.status-pending {
  background-color: color-mix(in oklch, oklch(0.75 0.15 85) 20%, transparent);
  color: oklch(0.55 0.12 85);
}

.scholar-status.status-accepted {
  background-color: var(--color-tertiary-container);
  color: var(--color-on-tertiary-container);
}

.scholar-status.status-rejected {
  background-color: var(--color-error-container);
  color: var(--color-on-error-container);
}

.scholar-status.status-expired {
  background-color: var(--surface-container-high);
  color: var(--color-on-surface-variant);
}

.scholar-status.status-default {
  background-color: var(--surface-container);
  color: var(--color-on-surface-variant);
}

.scholar-accept-btn {
  background-color: var(--color-tertiary);
  color: var(--color-on-tertiary);
  font-weight: 500;
  transition: opacity 0.2s;
}

.scholar-accept-btn:hover {
  opacity: 0.88;
}

.scholar-reject-btn {
  background-color: var(--color-error);
  color: var(--color-on-error);
  font-weight: 500;
  transition: opacity 0.2s;
}

.scholar-reject-btn:hover {
  opacity: 0.88;
}
</style>