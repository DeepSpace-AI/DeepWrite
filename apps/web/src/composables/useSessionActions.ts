import { ref } from 'vue'
import { gatewayBase } from '@/api/http'
import {
  createSessionShare,
  deleteSessionShare,
  listSessionShares,
  getSessionExportUrl,
  type SessionShare,
} from '@/api/agent'
import { cacheManager } from '@/utils/cache'

export function useSessionActions() {
  const isExporting = ref(false)
  const isCreatingShare = ref(false)
  const shares = ref<SessionShare[]>([])
  const shareError = ref<string | null>(null)

  async function exportSession(sessionId: string, format: 'markdown' | 'json') {
    isExporting.value = true
    
    try {
      const token = cacheManager.get<string>('deepwrite_access_token') || ''
      const url = `${gatewayBase}/api/v1/sessions/${encodeURIComponent(sessionId)}/export/${format}`
      
      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      })
      
      if (!response.ok) {
        throw new Error(`Export failed: ${response.status}`)
      }
      
      const blob = await response.blob()
      const disposition = response.headers.get('Content-Disposition')
      let filename = `conversation.${format === 'markdown' ? 'md' : 'json'}`
      
      if (disposition) {
        const match = disposition.match(/filename="?(.+)"?/)
        if (match) {
          filename = match[1]!
        }
      }
      
      const downloadUrl = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = downloadUrl
      a.download = filename
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(downloadUrl)
    } finally {
      isExporting.value = false
    }
  }

  async function shareSession(
    sessionId: string,
    options?: {
      title?: string
      expiresInHours?: number
      allowCopy?: boolean
      isPublic?: boolean
    }
  ): Promise<SessionShare | null> {
    isCreatingShare.value = true
    shareError.value = null
    
    try {
      const share = await createSessionShare({
        session_id: sessionId,
        title: options?.title,
        expires_in_hours: options?.expiresInHours,
        allow_copy: options?.allowCopy ?? true,
        is_public: options?.isPublic ?? true,
      })
      
      shares.value.unshift(share)
      return share
    } catch (e) {
      shareError.value = e instanceof Error ? e.message : 'Failed to create share'
      return null
    } finally {
      isCreatingShare.value = false
    }
  }

  async function loadShares() {
    try {
      const res = await listSessionShares()
      shares.value = res.items || []
    } catch (e) {
      console.error('Failed to load shares:', e)
    }
  }

  async function removeShare(shareId: string) {
    await deleteSessionShare(shareId)
    shares.value = shares.value.filter(s => s.id !== shareId)
  }

  function getShareUrl(shareToken: string): string {
    return `${window.location.origin}/share/${shareToken}`
  }

  async function copyShareUrl(shareToken: string): Promise<boolean> {
    const url = getShareUrl(shareToken)
    
    try {
      await navigator.clipboard.writeText(url)
      return true
    } catch {
      return false
    }
  }

  return {
    isExporting,
    isCreatingShare,
    shares,
    shareError,
    exportSession,
    shareSession,
    loadShares,
    removeShare,
    getShareUrl,
    copyShareUrl,
  }
}