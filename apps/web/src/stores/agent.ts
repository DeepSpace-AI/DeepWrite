import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import {
  listAgents,
  listMyAgents,
  getAgent,
  createAgent,
  updateAgent,
  deleteAgent,
  activateAgent,
  listSessions,
  getSession,
  createSession,
  updateSession,
  deleteSession,
  archiveSession,
  createMessage,
  listMessages,
  type Agent,
  type Session,
  type Message,
} from '@/api/agent'
import { ApiError } from '@/api/http'

export const useAgentStore = defineStore('agent', () => {
  const agents = ref<Agent[]>([])
  const myAgents = ref<Agent[]>([])
  const currentAgent = ref<Agent | null>(null)
  
  const sessions = ref<Session[]>([])
  const currentSession = ref<Session | null>(null)
  const messages = ref<Message[]>([])
  
  const isLoading = ref(false)
  const isStreaming = ref(false)
  const error = ref<string | null>(null)

  const availableAgents = computed(() => agents.value.filter(a => a.enabled))

  async function fetchAgents() {
    isLoading.value = true
    error.value = null
    try {
      const res = await listAgents()
      agents.value = res.items || []
    } catch (e) {
      error.value = e instanceof ApiError ? e.message : 'Failed to load agents'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function fetchMyAgents() {
    try {
      const res = await listMyAgents()
      myAgents.value = res.items || []
    } catch (e) {
      console.error('Failed to load my agents:', e)
    }
  }

  async function selectAgent(agentId: string) {
    const agent = agents.value.find(a => a.id === agentId) || 
                  myAgents.value.find(a => a.id === agentId)
    if (agent) {
      currentAgent.value = agent
      await activateAgent(agentId)
    } else {
      try {
        currentAgent.value = await getAgent(agentId)
      } catch (e) {
        error.value = e instanceof ApiError ? e.message : 'Agent not found'
      }
    }
  }

  async function createNewAgent(input: Parameters<typeof createAgent>[0]) {
    const agent = await createAgent(input)
    myAgents.value.unshift(agent)
    return agent
  }

  async function updateExistingAgent(agentId: string, input: Partial<Agent>) {
    const agent = await updateAgent(agentId, input)
    const index = myAgents.value.findIndex(a => a.id === agentId)
    if (index !== -1) {
      myAgents.value[index] = agent
    }
    if (currentAgent.value?.id === agentId) {
      currentAgent.value = agent
    }
    return agent
  }

  async function deleteExistingAgent(agentId: string) {
    await deleteAgent(agentId)
    myAgents.value = myAgents.value.filter(a => a.id !== agentId)
    if (currentAgent.value?.id === agentId) {
      currentAgent.value = null
    }
  }

  async function fetchSessions() {
    try {
      const res = await listSessions()
      sessions.value = res.items || []
    } catch (e) {
      console.error('Failed to load sessions:', e)
    }
  }

  async function selectSession(sessionId: string) {
    isLoading.value = true
    error.value = null
    try {
      const res = await getSession(sessionId)
      currentSession.value = res.session
      messages.value = res.messages || []
    } catch (e) {
      error.value = e instanceof ApiError ? e.message : 'Failed to load session'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function createNewSession(input: Parameters<typeof createSession>[0]) {
    const session = await createSession(input)
    sessions.value.unshift(session)
    currentSession.value = session
    messages.value = []
    return session
  }

  async function updateCurrentSession(input: Partial<Session>) {
    if (!currentSession.value) return
    const session = await updateSession(currentSession.value.id, input)
    currentSession.value = session
    const index = sessions.value.findIndex(s => s.id === session.id)
    if (index !== -1) {
      sessions.value[index] = session
    }
    return session
  }

  async function deleteCurrentSession() {
    if (!currentSession.value) return
    await deleteSession(currentSession.value.id)
    sessions.value = sessions.value.filter(s => s.id !== currentSession.value?.id)
    currentSession.value = null
    messages.value = []
  }

  async function archiveCurrentSession() {
    if (!currentSession.value) return
    const session = await archiveSession(currentSession.value.id)
    currentSession.value = session
    const index = sessions.value.findIndex(s => s.id === session.id)
    if (index !== -1) {
      sessions.value[index] = session
    }
    return session
  }

  async function sendMessage(content: string) {
    if (!currentSession.value) return
    
    const userMsg = await createMessage(currentSession.value.id, {
      role: 'user',
      content,
    })
    messages.value.push(userMsg)
    
    return userMsg
  }

  async function loadMoreMessages(offset: number) {
    if (!currentSession.value) return
    const res = await listMessages(currentSession.value.id, { offset, limit: 20 })
    messages.value = [...res.items, ...messages.value]
  }

  function clearError() {
    error.value = null
  }

  function reset() {
    agents.value = []
    myAgents.value = []
    currentAgent.value = null
    sessions.value = []
    currentSession.value = null
    messages.value = []
    error.value = null
  }

  return {
    agents,
    myAgents,
    currentAgent,
    sessions,
    currentSession,
    messages,
    isLoading,
    isStreaming,
    error,
    availableAgents,
    fetchAgents,
    fetchMyAgents,
    selectAgent,
    createNewAgent,
    updateExistingAgent,
    deleteExistingAgent,
    fetchSessions,
    selectSession,
    createNewSession,
    updateCurrentSession,
    deleteCurrentSession,
    archiveCurrentSession,
    sendMessage,
    loadMoreMessages,
    clearError,
    reset,
  }
})