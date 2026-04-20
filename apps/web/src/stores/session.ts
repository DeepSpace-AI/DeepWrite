import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import {
  listSessions,
  getSession,
  createSession,
  updateSession,
  deleteSession,
  archiveSession,
  unarchiveSession,
  pinSession,
  unpinSession,
  type Session,
  type SessionListFilter,
  type CreateSessionInput,
} from '@/api/agent'
import { ApiError } from '@/api/http'
import type { Message } from '@/api/agent'

export const useSessionStore = defineStore('session', () => {
  const sessions = ref<Session[]>([])
  const archivedSessions = ref<Session[]>([])
  const currentSession = ref<Session | null>(null)
  const currentMessages = ref<Message[]>([])
  
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  const pinnedSessions = computed(() => 
    sessions.value.filter(s => s.pinned && !s.archived)
  )

  const recentSessions = computed(() => 
    sessions.value.filter(s => !s.pinned && !s.archived)
  )

  async function fetchSessions(filter?: SessionListFilter) {
    isLoading.value = true
    error.value = null
    try {
      const res = await listSessions({ ...filter, archived: false })
      sessions.value = res.items || []
    } catch (e) {
      error.value = e instanceof ApiError ? e.message : 'Failed to load sessions'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function fetchArchivedSessions() {
    try {
      const res = await listSessions({ archived: true })
      archivedSessions.value = res.items || []
    } catch (e) {
      console.error('Failed to load archived sessions:', e)
    }
  }

  async function selectSession(sessionId: string) {
    isLoading.value = true
    error.value = null
    try {
      const res = await getSession(sessionId)
      currentSession.value = res.session
      currentMessages.value = res.messages || []
    } catch (e) {
      error.value = e instanceof ApiError ? e.message : 'Failed to load session'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function createNewSession(input: CreateSessionInput) {
    const session = await createSession(input)
    sessions.value.unshift(session)
    currentSession.value = session
    currentMessages.value = []
    return session
  }

  async function updateExistingSession(sessionId: string, input: Partial<Session>) {
    const session = await updateSession(sessionId, input)
    
    const index = sessions.value.findIndex(s => s.id === sessionId)
    if (index >= 0) {
      sessions.value[index] = session
    }
    
    const archivedIndex = archivedSessions.value.findIndex(s => s.id === sessionId)
    if (archivedIndex >= 0) {
      archivedSessions.value[archivedIndex] = session
    }
    
    if (currentSession.value?.id === sessionId) {
      currentSession.value = session
    }
    
    return session
  }

  async function deleteExistingSession(sessionId: string) {
    await deleteSession(sessionId)
    
    sessions.value = sessions.value.filter(s => s.id !== sessionId)
    archivedSessions.value = archivedSessions.value.filter(s => s.id !== sessionId)
    
    if (currentSession.value?.id === sessionId) {
      currentSession.value = null
      currentMessages.value = []
    }
  }

  async function archiveExistingSession(sessionId: string) {
    const session = await archiveSession(sessionId)
    
    sessions.value = sessions.value.filter(s => s.id !== sessionId)
    archivedSessions.value.unshift(session)
    
    if (currentSession.value?.id === sessionId) {
      currentSession.value = session
    }
    
    return session
  }

  async function unarchiveExistingSession(sessionId: string) {
    const session = await unarchiveSession(sessionId)
    
    archivedSessions.value = archivedSessions.value.filter(s => s.id !== sessionId)
    sessions.value.unshift(session)
    
    if (currentSession.value?.id === sessionId) {
      currentSession.value = session
    }
    
    return session
  }

  async function pinExistingSession(sessionId: string) {
    const session = await pinSession(sessionId)
    
    const index = sessions.value.findIndex(s => s.id === sessionId)
    if (index >= 0) {
      sessions.value[index] = session
    }
    
    if (currentSession.value?.id === sessionId) {
      currentSession.value = session
    }
    
    return session
  }

  async function unpinExistingSession(sessionId: string) {
    const session = await unpinSession(sessionId)
    
    const index = sessions.value.findIndex(s => s.id === sessionId)
    if (index >= 0) {
      sessions.value[index] = session
    }
    
    if (currentSession.value?.id === sessionId) {
      currentSession.value = session
    }
    
    return session
  }

  async function searchSessions(query: string) {
    if (!query.trim()) {
      await fetchSessions()
      return
    }
    
    await fetchSessions({ search: query })
  }

  function clearCurrentSession() {
    currentSession.value = null
    currentMessages.value = []
  }

  function updateSessionTitle(sessionId: string, title: string) {
    const index = sessions.value.findIndex(s => s.id === sessionId)
    if (index >= 0) {
      sessions.value[index] = { ...sessions.value[index]!, title }
    }
    
    if (currentSession.value?.id === sessionId) {
      currentSession.value = { ...currentSession.value, title }
    }
  }

  function reset() {
    sessions.value = []
    archivedSessions.value = []
    currentSession.value = null
    currentMessages.value = []
    error.value = null
  }

  return {
    sessions,
    archivedSessions,
    currentSession,
    currentMessages,
    isLoading,
    error,
    pinnedSessions,
    recentSessions,
    fetchSessions,
    fetchArchivedSessions,
    selectSession,
    createNewSession,
    updateExistingSession,
    updateSessionTitle,
    deleteExistingSession,
    archiveExistingSession,
    unarchiveExistingSession,
    pinExistingSession,
    unpinExistingSession,
    searchSessions,
    clearCurrentSession,
    reset,
  }
})