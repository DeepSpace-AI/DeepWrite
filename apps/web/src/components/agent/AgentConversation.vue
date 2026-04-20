<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AgentConversationSidebar from './AgentConversationSidebar.vue'
import AgentSelector from './AgentSelector.vue'
import AgentMemoryPanel from './AgentMemoryPanel.vue'
import WorkspaceSelector from './WorkspaceSelector.vue'
import { DButton } from '@/components/base'
import { ChatMessageList, ChatInput } from '@/components/chat'
import { SessionShareDialog } from '@/components/session'
import { useChatStore } from '@/stores/chat'
import { useSessionStore } from '@/stores/session'
import { useSessionActions } from '@/composables/useSessionActions'
import { useSessionEvents } from '@/composables/useSessionEvents'
import {
  listAgents,
  type Agent,
  type Message,
} from '@/api/agent'
import { ApiError } from '@/api/http'
import IconMenu from '~icons/mdi/menu'
import IconCog from '~icons/mdi/cog'
import IconFolder from '~icons/mdi/folder-outline'

const props = withDefaults(defineProps<{
  agentId?: string
  workspaceId?: string
  contextDocs?: Array<{ id: string; name: string }>
  contextRefs?: Array<{ id: string; title: string }>
  embedded?: boolean
}>(), {
  embedded: false,
})

const emit = defineEmits<{
  'session-change': [sessionId: string]
  'message-send': [message: Message]
  'context-update': [docs: typeof props.contextDocs, refs: typeof props.contextRefs]
}>()

const { t } = useI18n()

const isLoading = ref(false)
const error = ref('')

const agents = ref<Agent[]>([])
const currentAgent = ref<Agent | null>(null)

const isSidebarCollapsed = ref(false)
const showToolsPanel = ref(false)
const selectedWorkspaceId = ref<string | null>(null)

const chatStore = useChatStore()
const sessionStore = useSessionStore()
const sessionActions = useSessionActions()

const showShareDialog = ref(false)
const sharingSessionId = ref<string | null>(null)
const currentShare = ref<ReturnType<typeof sessionActions.shareSession> extends Promise<infer T> ? T : never>(null)

const effectiveWorkspaceId = computed(() => {
  return selectedWorkspaceId.value || props.workspaceId || null
})

async function loadAgents() {
  try {
    const res = await listAgents()
    agents.value = res.items || []
    if (props.agentId) {
      currentAgent.value = agents.value.find(a => a.id === props.agentId) || null
    } else if (agents.value.length > 0) {
      currentAgent.value = agents.value[0]!
    }
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Failed to load agents'
  }
}

async function selectSession(sessionId: string) {
  try {
    isLoading.value = true
    await sessionStore.selectSession(sessionId)
    chatStore.setSession(sessionId)
    chatStore.loadFromHistory(sessionStore.currentMessages)
    emit('session-change', sessionId)
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Failed to load session'
  } finally {
    isLoading.value = false
  }
}

async function startNewSession() {
  if (!currentAgent.value) return
  
  try {
    isLoading.value = true
    const session = await sessionStore.createNewSession({
      agent_id: currentAgent.value.id,
      workspace_id: effectiveWorkspaceId.value || undefined,
      context_docs: props.contextDocs ? JSON.stringify(props.contextDocs) : undefined,
      context_refs: props.contextRefs ? JSON.stringify(props.contextRefs) : undefined,
    })
    chatStore.setSession(session.id)
    chatStore.clearMessages()
    emit('session-change', session.id)
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Failed to create session'
  } finally {
    isLoading.value = false
  }
}

function selectAgent(agent: Agent) {
  currentAgent.value = agent
  sessionStore.clearCurrentSession()
  chatStore.setSession(null)
  chatStore.clearMessages()
}

async function handleSendMessage(content: string) {
  if (!content.trim() || !currentAgent.value || chatStore.isStreaming) return
  
  error.value = ''
  
  try {
    if (!sessionStore.currentSession) {
      const session = await sessionStore.createNewSession({
        agent_id: currentAgent.value.id,
        workspace_id: effectiveWorkspaceId.value || undefined,
        context_docs: props.contextDocs ? JSON.stringify(props.contextDocs) : undefined,
        context_refs: props.contextRefs ? JSON.stringify(props.contextRefs) : undefined,
      })
      chatStore.setSession(session.id)
      emit('session-change', session.id)
    }
    
    await chatStore.sendMessage(content)
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Failed to send message'
  }
}

function handleToggleReasoning(id: string) {
  chatStore.toggleReasoningCollapsed(id)
}

async function handleEditMessage(messageId: string) {
  const message = chatStore.messages.find(m => m.id === messageId)
  if (!message) return
  
  const textPart = message.parts.find(p => p.type === 'text')
  if (!textPart || textPart.type !== 'text') return
  
  chatStore.editMessage(messageId, textPart.text)
  chatStore.deleteMessagesAfter(messageId)
  await chatStore.sendMessage(textPart.text)
}

async function handleRegenerate(messageId: string) {
  await chatStore.regenerateFromMessage(messageId)
}

function handleDeleteMessage(messageId: string) {
  chatStore.deleteMessage(messageId)
}

async function handleExport(sessionId: string, format: 'markdown' | 'json') {
  await sessionActions.exportSession(sessionId, format)
}

async function handleShare(sessionId: string) {
  sharingSessionId.value = sessionId
  showShareDialog.value = true
  currentShare.value = await sessionActions.shareSession(sessionId)
}

function handleCloseShareDialog() {
  showShareDialog.value = false
  sharingSessionId.value = null
  currentShare.value = null
}

function handleWorkspaceSelect(workspaceId: string | null) {
  selectedWorkspaceId.value = workspaceId
}

async function handleSearch(query: string) {
  await sessionStore.searchSessions(query)
}

const { connect: connectEvents, disconnect: disconnectEvents } = useSessionEvents((event) => {
  console.log('[AgentConversation] Received event:', event)
  
  if (event.type === 'title_updated') {
    const title = event.data?.title as string
    if (title) {
      sessionStore.updateSessionTitle(event.session_id, title)
    }
  }
})

onMounted(async () => {
  await loadAgents()
  await sessionStore.fetchSessions()
  await sessionStore.fetchArchivedSessions()
  connectEvents()
})

onUnmounted(() => {
  disconnectEvents()
})

watch(() => props.agentId, (newId) => {
  if (newId && agents.value.length > 0) {
    currentAgent.value = agents.value.find(a => a.id === newId) || null
  }
})
</script>

<template>
  <section class="agent-chat" :class="{ embedded }">
    <button
      class="sidebar-toggle"
      :class="{ collapsed: isSidebarCollapsed }"
      @click="isSidebarCollapsed = !isSidebarCollapsed"
    >
      <IconMenu class="h-5 w-5" />
    </button>

    <aside
      class="chat-sidebar"
      :class="{ collapsed: isSidebarCollapsed }"
    >
      <div class="sidebar-header">
        <AgentSelector
          :agents="agents"
          :current-agent="currentAgent"
          @select="selectAgent"
        />
      </div>
      
      <AgentConversationSidebar
        :sessions="sessionStore.sessions"
        :archived-sessions="sessionStore.archivedSessions"
        :current-session-id="sessionStore.currentSession?.id"
        :collapsed="false"
        @select="selectSession"
        @new="startNewSession"
        @pin="sessionStore.pinExistingSession($event)"
        @unpin="sessionStore.unpinExistingSession($event)"
        @archive="sessionStore.archiveExistingSession($event)"
        @unarchive="sessionStore.unarchiveExistingSession($event)"
        @delete="sessionStore.deleteExistingSession($event)"
        @export="handleExport"
        @share="handleShare"
        @search="handleSearch"
      />
    </aside>

    <main class="chat-main">
      <div class="chat-header">
        <div class="header-left">
          <span v-if="currentAgent" class="agent-name">{{ currentAgent.name }}</span>
        </div>
        <div class="header-right">
          <WorkspaceSelector
            v-if="!embedded"
            :selected-id="effectiveWorkspaceId"
            @select="handleWorkspaceSelect"
          />
          <DButton
            size="sm"
            variant="ghost"
            :title="t('agents.tools')"
            @click="showToolsPanel = !showToolsPanel"
          >
            <IconCog class="h-5 w-5" />
          </DButton>
        </div>
      </div>

      <ChatMessageList
        :messages="chatStore.messages"
        :is-streaming="chatStore.isStreaming"
        :agent-name="currentAgent?.name"
        show-actions
        @toggle-reasoning="handleToggleReasoning"
        @edit-message="handleEditMessage"
        @regenerate="handleRegenerate"
        @delete-message="handleDeleteMessage"
      />

      <ChatInput
        :disabled="!currentAgent"
        :is-streaming="chatStore.isStreaming"
        @send="handleSendMessage"
        @abort="chatStore.abort"
      />
    </main>

    <Transition name="slide-left">
      <aside v-if="showToolsPanel" class="tools-panel">
        <div class="panel-header">
          <h3>{{ t('agents.tools') }}</h3>
          <button class="close-btn" @click="showToolsPanel = false">
            <IconMenu class="h-4 w-4" />
          </button>
        </div>

        <div class="panel-section">
          <h4 class="section-title">
            <IconFolder class="h-4 w-4" />
            {{ t('agents.workspace') }}
          </h4>
          <WorkspaceSelector
            :selected-id="effectiveWorkspaceId"
            layout="card"
            @select="handleWorkspaceSelect"
          />
        </div>

        <div v-if="effectiveWorkspaceId" class="panel-section">
          <h4 class="section-title">{{ t('agents.activeDocs') }}</h4>
          <div v-if="contextDocs?.length" class="context-list">
            <div v-for="doc in contextDocs" :key="doc.id" class="context-item">
              {{ doc.name }}
            </div>
          </div>
          <p v-else class="empty-hint">{{ t('agents.noContextDocs') }}</p>
        </div>

        <div v-if="sessionStore.currentSession" class="panel-section">
          <AgentMemoryPanel
            :agent-id="currentAgent?.id || ''"
            :workspace-id="sessionStore.currentSession.workspace_id || undefined"
            :session-id="sessionStore.currentSession.id"
            compact
          />
        </div>
      </aside>
    </Transition>

    <SessionShareDialog
      :share="currentShare"
      :is-creating="sessionActions.isCreatingShare.value"
      :error="sessionActions.shareError.value"
      @close="handleCloseShareDialog"
      @create="handleShare(sharingSessionId!)"
    />
  </section>
</template>

<style scoped>
.agent-chat {
  display: flex;
  height: 100%;
  background: var(--surface);
}

.agent-chat.embedded {
  border-radius: 0.75rem;
  overflow: hidden;
  box-shadow: 0 4px 24px oklch(0 0 0 / 0.1);
}

.sidebar-toggle {
  position: absolute;
  left: 0.5rem;
  top: 0.5rem;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 0.5rem;
  border: none;
  background: var(--surface-container);
  color: var(--color-on-surface-variant);
  cursor: pointer;
  transition: all 0.2s ease;
}

.sidebar-toggle:hover {
  background: var(--surface-container-high);
}

.sidebar-toggle.collapsed {
  left: 0.5rem;
}

.chat-sidebar {
  width: 16rem;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--color-surface-container-high);
  background: var(--surface-container-low);
  transition: width 0.25s ease, opacity 0.25s ease;
  overflow: hidden;
}

.chat-sidebar.collapsed {
  width: 0;
  opacity: 0;
}

.sidebar-header {
  padding: 0.75rem;
  border-bottom: 1px solid var(--color-surface-container-high);
}

.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  position: relative;
}

.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  padding-left: 3.5rem;
  border-bottom: 1px solid var(--color-surface-container-high);
  background: var(--surface-container);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.agent-name {
  font-weight: 600;
  color: var(--color-on-background);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.tools-panel {
  width: 20rem;
  flex-shrink: 0;
  border-left: 1px solid var(--color-surface-container-high);
  background: var(--surface-container-low);
  overflow-y: auto;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem;
  border-bottom: 1px solid var(--color-surface-container-high);
}

.panel-header h3 {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-on-background);
}

.close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.75rem;
  height: 1.75rem;
  border-radius: 0.375rem;
  border: none;
  background: transparent;
  color: var(--color-on-surface-variant);
  cursor: pointer;
}

.close-btn:hover {
  background: var(--surface-container-high);
}

.panel-section {
  padding: 1rem;
  border-bottom: 1px solid var(--color-surface-container-high);
}

.section-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-on-surface-variant);
}

.context-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.context-item {
  padding: 0.5rem 0.75rem;
  border-radius: 0.375rem;
  background: var(--surface-container);
  font-size: 0.875rem;
  color: var(--color-on-background);
}

.empty-hint {
  font-size: 0.875rem;
  color: var(--color-on-surface-variant);
  opacity: 0.7;
}

.slide-left-enter-active,
.slide-left-leave-active {
  transition: all 0.25s ease;
}

.slide-left-enter-from,
.slide-left-leave-to {
  width: 0;
  opacity: 0;
}
</style>