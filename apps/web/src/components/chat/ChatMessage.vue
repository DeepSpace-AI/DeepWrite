<script setup lang="ts">
import { computed, ref } from 'vue'
import type { ChatMessage } from '@/types/chat'
import type { ToolCallPart } from '@/types/chat'
import ChatMessageText from './ChatMessageText.vue'
import ChatMessageReasoning from './ChatMessageReasoning.vue'
import ChatMessageError from './ChatMessageError.vue'
import ChatMessageToolCall from './ChatMessageToolCall.vue'
import IconAutoAwesome from '~icons/mdi/auto-fix'
import IconUser from '~icons/mdi/account'
import IconPencil from '~icons/mdi/pencil'
import IconRefresh from '~icons/mdi/refresh'
import IconDelete from '~icons/mdi/delete'

const props = defineProps<{
  message: ChatMessage
  agentName?: string
  showActions?: boolean
}>()

const emit = defineEmits<{
  'toggle-reasoning': []
  'edit': []
  'regenerate': []
  'delete': []
}>()

const isUser = computed(() => props.message.role === 'user')
const isAssistant = computed(() => props.message.role === 'assistant')
const isStreaming = computed(() => props.message.status === 'streaming')
const isDone = computed(() => props.message.status === 'done')
const hasError = computed(() => props.message.parts.some(p => p.type === 'error'))

const showActionButtons = computed(() => 
  props.showActions && isDone.value && !isStreaming.value
)

function handleToggleReasoning() {
  emit('toggle-reasoning')
}

function getToolCallPart(part: unknown): ToolCallPart | null {
  if (part && typeof part === 'object' && (part as ToolCallPart).type === 'tool_call') {
    return part as ToolCallPart
  }
  return null
}
</script>

<template>
  <div class="chat-message" :class="[message.role, { streaming: isStreaming, error: hasError }]">
    <div v-if="isAssistant" class="message-header">
      <IconAutoAwesome class="header-icon" />
      <span class="header-name">{{ agentName || 'Assistant' }}</span>
    </div>

    <div class="message-body">
      <div class="message-content">
        <template v-for="(part, index) in message.parts" :key="index">
          <ChatMessageReasoning
            v-if="part.type === 'reasoning'"
            :text="part.text"
            :collapsed="part.collapsed"
            @toggle="handleToggleReasoning"
          />
          <ChatMessageText
            v-else-if="part.type === 'text'"
            :text="part.text"
            :is-streaming="isStreaming"
          />
          <ChatMessageToolCall
            v-else-if="part.type === 'tool_call'"
            :name="(part as ToolCallPart).name"
            :args="(part as ToolCallPart).args"
            :result="(part as ToolCallPart).result"
            :status="(part as ToolCallPart).status"
          />
          <ChatMessageError
            v-else-if="part.type === 'error'"
            :message="part.message"
          />
        </template>

        <div v-if="isStreaming && message.parts.length === 0" class="streaming-indicator">
          <span class="dot" />
          <span class="dot" />
          <span class="dot" />
        </div>
      </div>

      <div v-if="showActionButtons" class="message-actions">
        <button
          v-if="isUser"
          class="action-btn"
          title="编辑"
          @click="emit('edit')"
        >
          <IconPencil class="action-icon" />
        </button>
        <button
          v-if="isAssistant"
          class="action-btn"
          title="重新生成"
          @click="emit('regenerate')"
        >
          <IconRefresh class="action-icon" />
        </button>
        <button
          class="action-btn"
          title="删除"
          @click="emit('delete')"
        >
          <IconDelete class="action-icon" />
        </button>
      </div>
    </div>

    <div v-if="isUser" class="message-footer">
      <IconUser class="footer-icon" />
    </div>
  </div>
</template>

<style scoped>
.chat-message {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.chat-message.user {
  align-items: flex-end;
}

.chat-message.assistant {
  align-items: flex-start;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.header-icon {
  width: 1rem;
  height: 1rem;
  color: var(--color-primary);
}

.header-name {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-primary);
}

.message-body {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
}

.message-content {
  display: flex;
  flex-direction: column;
  max-width: 85%;
  padding: 0.75rem 1rem;
  border-radius: 0.75rem;
  line-height: 1.6;
}

.message-content > * {
  width: 100%;
}

.chat-message.user .message-content {
  background: var(--surface-container-lowest);
  border-left: 3px solid var(--color-primary);
}

.chat-message.assistant .message-content {
  background: transparent;
}

.chat-message.error .message-content {
  border: 1px solid var(--color-error);
}

.message-actions {
  display: flex;
  gap: 0.25rem;
  opacity: 0;
  transition: opacity 0.15s ease;
}

.chat-message:hover .message-actions {
  opacity: 1;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.75rem;
  height: 1.75rem;
  padding: 0;
  border: none;
  border-radius: 0.375rem;
  background: transparent;
  color: var(--color-on-surface-variant);
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-btn:hover {
  background: var(--surface-container);
  color: var(--color-on-background);
}

.action-icon {
  width: 1rem;
  height: 1rem;
}

.message-footer {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.footer-icon {
  width: 1rem;
  height: 1rem;
  color: var(--color-on-surface-variant);
  opacity: 0.6;
}

.streaming-indicator {
  display: flex;
  gap: 0.25rem;
  padding: 0.5rem 0;
}

.streaming-indicator .dot {
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 50%;
  background: var(--color-primary);
  animation: bounce 1.4s infinite ease-in-out both;
}

.streaming-indicator .dot:nth-child(1) {
  animation-delay: -0.32s;
}

.streaming-indicator .dot:nth-child(2) {
  animation-delay: -0.16s;
}

@keyframes bounce {
  0%, 80%, 100% {
    transform: scale(0);
  }
  40% {
    transform: scale(1);
  }
}
</style>