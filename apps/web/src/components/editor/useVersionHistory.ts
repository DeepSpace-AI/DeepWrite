import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import type { Editor } from '@tiptap/vue-3'

export interface Snapshot {
  id: string
  name: string
  content: string
  createdAt: number
}

const STORAGE_KEY_PREFIX = 'deepwrite:draft:'
const AUTO_SAVE_KEY = 'deepwrite:draft:auto'
const SNAPSHOTS_KEY = 'deepwrite:draft:snapshots'
const MAX_SNAPSHOTS = 20

export function useVersionHistory(editor: Editor | null, documentId?: string) {
  const snapshots = ref<Snapshot[]>([])
  const lastSaved = ref<number | null>(null)
  const hasUnsavedChanges = ref(false)
  const autoSaveEnabled = ref(true)
  const autoSaveInterval = 30000

  const storageKey = documentId ? `${STORAGE_KEY_PREFIX}${documentId}` : AUTO_SAVE_KEY

  function generateId(): string {
    return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`
  }

  function loadSnapshots(): Snapshot[] {
    try {
      const raw = localStorage.getItem(`${storageKey}:${SNAPSHOTS_KEY}`)
      if (!raw) return []
      const parsed = JSON.parse(raw) as Snapshot[]
      return Array.isArray(parsed) ? parsed : []
    } catch {
      return []
    }
  }

  function saveSnapshotsToStorage() {
    try {
      localStorage.setItem(`${storageKey}:${SNAPSHOTS_KEY}`, JSON.stringify(snapshots.value.slice(0, MAX_SNAPSHOTS)))
    } catch {
      // Storage full or unavailable
    }
  }

  function loadAutoSave(): string | null {
    try {
      return localStorage.getItem(storageKey)
    } catch {
      return null
    }
  }

  function saveAutoSave(content: string) {
    try {
      localStorage.setItem(storageKey, content)
      lastSaved.value = Date.now()
      hasUnsavedChanges.value = false
    } catch {
      // Storage full or unavailable
    }
  }

  function clearAutoSave() {
    try {
      localStorage.removeItem(storageKey)
    } catch {
      // Ignore
    }
  }

  function createSnapshot(name?: string): Snapshot {
    if (!editor) {
      throw new Error('Editor not available')
    }

    const content = editor.getHTML()
    const snapshot: Snapshot = {
      id: generateId(),
      name: name || `快照 ${snapshots.value.length + 1}`,
      content,
      createdAt: Date.now(),
    }

    snapshots.value.unshift(snapshot)
    if (snapshots.value.length > MAX_SNAPSHOTS) {
      snapshots.value = snapshots.value.slice(0, MAX_SNAPSHOTS)
    }

    saveSnapshotsToStorage()
    return snapshot
  }

  function restoreSnapshot(snapshot: Snapshot) {
    if (!editor) return
    editor.commands.setContent(snapshot.content)
    hasUnsavedChanges.value = true
  }

  function deleteSnapshot(id: string) {
    const index = snapshots.value.findIndex((s) => s.id === id)
    if (index !== -1) {
      snapshots.value.splice(index, 1)
      saveSnapshotsToStorage()
    }
  }

  function renameSnapshot(id: string, name: string) {
    const snapshot = snapshots.value.find((s) => s.id === id)
    if (snapshot) {
      snapshot.name = name
      saveSnapshotsToStorage()
    }
  }

  function getAutoSavedContent(): string | null {
    return loadAutoSave()
  }

  function hasAutoSave(): boolean {
    const content = loadAutoSave()
    return content !== null && content.length > 0
  }

  function discardAutoSave() {
    clearAutoSave()
    lastSaved.value = null
    hasUnsavedChanges.value = false
  }

  let autoSaveTimer: number | null = null

  function startAutoSave() {
    if (autoSaveTimer) return

    autoSaveTimer = window.setInterval(() => {
      if (!autoSaveEnabled.value || !editor || !hasUnsavedChanges.value) return
      const content = editor.getHTML()
      saveAutoSave(content)
    }, autoSaveInterval)
  }

  function stopAutoSave() {
    if (autoSaveTimer) {
      clearInterval(autoSaveTimer)
      autoSaveTimer = null
    }
  }

  function markDirty() {
    hasUnsavedChanges.value = true
  }

  function formatDate(timestamp: number): string {
    const date = new Date(timestamp)
    const now = new Date()
    const diff = now.getTime() - date.getTime()

    if (diff < 60000) return '刚刚'
    if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
    if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
    if (diff < 604800000) return `${Math.floor(diff / 86400000)} 天前`

    return new Intl.DateTimeFormat('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    }).format(date)
  }

  onMounted(() => {
    snapshots.value = loadSnapshots()
    const autoContent = loadAutoSave()
    if (autoContent) {
      lastSaved.value = Date.now()
    }
    startAutoSave()
  })

  onBeforeUnmount(() => {
    stopAutoSave()
    if (hasUnsavedChanges.value && editor) {
      const content = editor.getHTML()
      saveAutoSave(content)
    }
  })

  watch(
    () => editor?.state,
    () => {
      if (editor) {
        markDirty()
      }
    },
    { deep: true },
  )

  return {
    snapshots,
    lastSaved,
    hasUnsavedChanges,
    autoSaveEnabled,
    createSnapshot,
    restoreSnapshot,
    deleteSnapshot,
    renameSnapshot,
    getAutoSavedContent,
    hasAutoSave,
    discardAutoSave,
    formatDate,
    startAutoSave,
    stopAutoSave,
  }
}
