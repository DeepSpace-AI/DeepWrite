import { ref } from 'vue'
import { defineStore } from 'pinia'
import {
  acceptWorkspaceInvitation,
  listMyWorkspaceInvitations,
  rejectWorkspaceInvitation,
  type WorkspaceInvitation,
} from '@/api/workspace'

export const useNotificationStore = defineStore('notification', () => {
  const invitations = ref<WorkspaceInvitation[]>([])
  const pendingCount = ref(0)
  const isLoading = ref(false)
  const errorMessage = ref('')

  function clear() {
    invitations.value = []
    pendingCount.value = 0
    errorMessage.value = ''
  }

  async function loadInvitations(status?: string) {
    isLoading.value = true
    errorMessage.value = ''
    try {
      const rows = await listMyWorkspaceInvitations(status)
      invitations.value = rows
      pendingCount.value = rows.filter((item) => item.status === 'pending').length
      return rows
    } catch (error) {
      errorMessage.value = error instanceof Error ? error.message : '加载通知失败'
      throw error
    } finally {
      isLoading.value = false
    }
  }

  async function refreshPendingCount() {
    try {
      const rows = await listMyWorkspaceInvitations('pending')
      pendingCount.value = rows.filter((item) => item.status === 'pending').length
      return pendingCount.value
    } catch {
      pendingCount.value = 0
      return 0
    }
  }

  async function acceptInvitation(invitation: WorkspaceInvitation) {
    const next = await acceptWorkspaceInvitation(invitation.id, {
      actionToken: invitation.action_token,
    })
    invitations.value = invitations.value.map((item) => (item.id === invitation.id ? next : item))
    if (pendingCount.value > 0) {
      pendingCount.value -= 1
    }
    return next
  }

  async function rejectInvitation(invitation: WorkspaceInvitation) {
    const next = await rejectWorkspaceInvitation(invitation.id, {
      actionToken: invitation.action_token,
    })
    invitations.value = invitations.value.map((item) => (item.id === invitation.id ? next : item))
    if (pendingCount.value > 0) {
      pendingCount.value -= 1
    }
    return next
  }

  return {
    invitations,
    pendingCount,
    isLoading,
    errorMessage,
    clear,
    loadInvitations,
    refreshPendingCount,
    acceptInvitation,
    rejectInvitation,
  }
})
