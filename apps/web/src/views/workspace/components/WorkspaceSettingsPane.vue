<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ApiError } from '@/api/http'
import {
  createWorkspaceInvitation,
  listWorkspaceInvitations,
  revokeWorkspaceInvitation,
  type WorkspaceInvitation,
  type WorkspaceInvitationRole,
} from '@/api/workspace'

const { t } = useI18n()
const route = useRoute()

const workspaceId = computed(() => String(route.params.id || '').trim())
const invitations = ref<WorkspaceInvitation[]>([])
const inviteEmail = ref('')
const inviteRole = ref<WorkspaceInvitationRole>('viewer')
const createdInviteToken = ref('')
const createMessage = ref('')
const errorMessage = ref('')
const isLoading = ref(false)
const isCreating = ref(false)
const revokingId = ref('')

const canSubmitInvite = computed(() => {
  return inviteEmail.value.trim().length > 0 && !isCreating.value
})

const roleOptions: WorkspaceInvitationRole[] = ['viewer', 'editor', 'admin']

function roleLabel(role: string) {
  const key = `workspace.detail.settingsInvitations.roleOptions.${role}`
  return t(key)
}

function statusLabel(status: string) {
  const key = `workspace.detail.settingsInvitations.status.${status}`
  return t(key)
}

function formatDateTime(value?: string) {
  if (!value) return '-'
  const ts = new Date(value)
  if (Number.isNaN(ts.getTime())) return '-'
  return ts.toLocaleString()
}

function getErrorText(error: unknown, fallbackKey: string) {
  if (error instanceof ApiError && error.message) {
    return error.message
  }
  if (error instanceof Error && error.message) {
    return error.message
  }
  return t(fallbackKey)
}

async function loadInvitations() {
  if (!workspaceId.value) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    invitations.value = await listWorkspaceInvitations(workspaceId.value)
  } catch (error) {
    errorMessage.value = getErrorText(error, 'workspace.detail.settingsInvitations.loadError')
  } finally {
    isLoading.value = false
  }
}

async function submitInvitation() {
  if (!workspaceId.value || !canSubmitInvite.value) return
  isCreating.value = true
  createMessage.value = ''
  errorMessage.value = ''
  createdInviteToken.value = ''

  try {
    const result = await createWorkspaceInvitation(workspaceId.value, {
      invitee_email: inviteEmail.value.trim(),
      role: inviteRole.value,
    })
    createdInviteToken.value = result.token || ''
    createMessage.value = result.reused
      ? t('workspace.detail.settingsInvitations.createReusedSuccess')
      : t('workspace.detail.settingsInvitations.createSuccess')
    inviteEmail.value = ''
    await loadInvitations()
  } catch (error) {
    errorMessage.value = getErrorText(error, 'workspace.detail.settingsInvitations.createError')
  } finally {
    isCreating.value = false
  }
}

async function revokeInvitation(invitationId: string) {
  if (!workspaceId.value || !invitationId || revokingId.value) return
  revokingId.value = invitationId
  errorMessage.value = ''
  try {
    await revokeWorkspaceInvitation(workspaceId.value, invitationId)
    await loadInvitations()
  } catch (error) {
    errorMessage.value = getErrorText(error, 'workspace.detail.settingsInvitations.revokeError')
  } finally {
    revokingId.value = ''
  }
}

async function copyToken() {
  if (!createdInviteToken.value) return
  try {
    await navigator.clipboard.writeText(createdInviteToken.value)
    createMessage.value = t('workspace.detail.settingsInvitations.copyTokenSuccess')
  } catch {
    errorMessage.value = t('workspace.detail.settingsInvitations.copyTokenError')
  }
}

onMounted(loadInvitations)

watch(
  () => workspaceId.value,
  () => {
    invitations.value = []
    createdInviteToken.value = ''
    createMessage.value = ''
    errorMessage.value = ''
    loadInvitations()
  },
)
</script>

<template>
  <section class="h-full min-h-0 rounded-sm border border-base-300 bg-base-100 shadow-sm flex flex-col">
    <div class="border-b border-base-300 px-4 py-3">
      <h3 class="text-base font-semibold text-base-content">{{ t('workspace.detail.navSettings') }}</h3>
    </div>

    <div class="min-h-0 flex-1 space-y-4 overflow-y-auto p-4">
      <section class="rounded-sm border border-base-300 bg-base-100 p-4">
        <h4 class="text-sm font-semibold text-base-content">{{ t('workspace.detail.settingsInvitations.title') }}</h4>
        <p class="mt-1 text-xs text-base-content/65">{{ t('workspace.detail.settingsInvitations.subtitle') }}</p>

        <form class="mt-4 grid gap-3 md:grid-cols-[minmax(0,1fr)_180px_auto]" @submit.prevent="submitInvitation">
          <label class="form-control w-full">
            <span class="label-text text-xs text-base-content/70">{{ t('workspace.detail.settingsInvitations.emailLabel') }}</span>
            <input
              v-model="inviteEmail"
              type="email"
              class="input input-sm input-bordered w-full rounded-sm"
              :placeholder="t('workspace.detail.settingsInvitations.emailPlaceholder')"
              :disabled="isCreating"
            />
          </label>

          <label class="form-control w-full">
            <span class="label-text text-xs text-base-content/70">{{ t('workspace.detail.settingsInvitations.roleLabel') }}</span>
            <select v-model="inviteRole" class="select select-sm select-bordered w-full rounded-sm" :disabled="isCreating">
              <option v-for="role in roleOptions" :key="role" :value="role">{{ roleLabel(role) }}</option>
            </select>
          </label>

          <button type="submit" class="btn btn-sm rounded-sm self-end" :disabled="!canSubmitInvite">
            {{ isCreating ? t('workspace.detail.settingsInvitations.creating') : t('workspace.detail.settingsInvitations.createAction') }}
          </button>
        </form>

        <div v-if="createMessage" class="mt-3 rounded-sm border border-success/30 bg-success/10 px-3 py-2 text-xs text-success">
          {{ createMessage }}
        </div>
        <div v-if="errorMessage" class="mt-3 rounded-sm border border-error/30 bg-error/10 px-3 py-2 text-xs text-error">
          {{ errorMessage }}
        </div>

        <div v-if="createdInviteToken" class="mt-3 rounded-sm border border-base-300 bg-base-200/40 p-3">
          <p class="text-xs font-medium text-base-content">{{ t('workspace.detail.settingsInvitations.tokenLabel') }}</p>
          <div class="mt-2 flex flex-col gap-2 md:flex-row md:items-center">
            <input type="text" class="input input-sm input-bordered w-full rounded-sm" :value="createdInviteToken" readonly />
            <button type="button" class="btn btn-sm btn-ghost rounded-sm" @click="copyToken">
              {{ t('workspace.detail.settingsInvitations.copyToken') }}
            </button>
          </div>
        </div>
      </section>

      <section class="rounded-sm border border-base-300 bg-base-100 p-4">
        <div class="flex items-center justify-between gap-3">
          <h4 class="text-sm font-semibold text-base-content">{{ t('workspace.detail.settingsInvitations.listTitle') }}</h4>
          <button type="button" class="btn btn-ghost btn-xs rounded-sm" :disabled="isLoading" @click="loadInvitations">
            {{ t('workspace.detail.settingsInvitations.refreshAction') }}
          </button>
        </div>

        <div v-if="isLoading" class="mt-3 text-xs text-base-content/60">
          {{ t('workspace.detail.settingsInvitations.loading') }}
        </div>

        <div v-else-if="!invitations.length" class="mt-3 rounded-sm border border-dashed border-base-300 bg-base-200/30 px-3 py-8 text-center text-sm text-base-content/60">
          {{ t('workspace.detail.settingsInvitations.empty') }}
        </div>

        <div v-else class="mt-3 space-y-2">
          <article
            v-for="invitation in invitations"
            :key="invitation.id"
            class="rounded-sm border border-base-300 bg-base-100 px-3 py-2"
          >
            <div class="flex flex-wrap items-center justify-between gap-2">
              <p class="text-sm font-medium text-base-content">{{ invitation.invitee_email }}</p>
              <span class="badge badge-outline badge-sm rounded-sm">{{ statusLabel(invitation.status) }}</span>
            </div>
            <div class="mt-1 text-xs text-base-content/65">
              {{ t('workspace.detail.settingsInvitations.metaRole', { role: roleLabel(invitation.role) }) }}
              ·
              {{ t('workspace.detail.settingsInvitations.metaExpiresAt', { time: formatDateTime(invitation.expires_at) }) }}
            </div>

            <div v-if="invitation.status === 'pending'" class="mt-2">
              <button
                type="button"
                class="btn btn-ghost btn-xs rounded-sm text-error"
                :disabled="!!revokingId"
                @click="revokeInvitation(invitation.id)"
              >
                {{ revokingId === invitation.id
                  ? t('workspace.detail.settingsInvitations.revoking')
                  : t('workspace.detail.settingsInvitations.revokeAction') }}
              </button>
            </div>
          </article>
        </div>
      </section>
    </div>
  </section>
</template>
