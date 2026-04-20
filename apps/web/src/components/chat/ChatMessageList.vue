<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import type { ChatMessage as ChatMessageType } from '@/types/chat'
import ChatMessageComponent from './ChatMessage.vue'
import IconAutoAwesome from '~icons/mdi/auto-fix'

const props = defineProps<{
  messages: ChatMessageType[]
  isStreaming?: boolean
  agentName?: string
  showActions?: boolean
}>()

const emit = defineEmits<{
  'toggle-reasoning': [id: string]
  'edit-message': [id: string]
  'regenerate': [id: string]
  'delete-message': [id: string]
}>()

const containerRef = ref<HTMLElement | null>(null)
const shouldAutoScroll = ref(true)

function scrollToBottom(smooth = true) {
  if (!containerRef.value) return
  containerRef.value.scrollTo({
    top: containerRef.value.scrollHeight,
    behavior: smooth ? 'smooth' : 'instant',
  })
}

function handleScroll() {
  if (!containerRef.value) return
  const { scrollTop, scrollHeight, clientHeight } = containerRef.value
  const isNearBottom = scrollHeight - scrollTop - clientHeight < 100
  shouldAutoScroll.value = isNearBottom
}

watch(
  () => props.messages.length,
  async () => {
    if (shouldAutoScroll.value) {
      await nextTick()
      scrollToBottom()
    }
  }
)

watch(
  () => props.isStreaming,
  async (streaming) => {
    if (streaming && shouldAutoScroll.value) {
      await nextTick()
      scrollToBottom(false)
    }
  }
)

defineExpose({
  scrollToBottom,
})
</script>

<template>
  <div ref="containerRef" class="chat-message-list" @scroll="handleScroll">
    <div class="messages-container">
      <div v-if="messages.length === 0" class="empty-state">
        <div class="empty-icon">
          <IconAutoAwesome />
        </div>
        <h3 class="empty-title">{{ agentName || 'Assistant' }}</h3>
        <p class="empty-description">
          Start a conversation to get help with your tasks.
        </p>
      </div>

      <ChatMessageComponent
        v-for="message in messages"
        :key="message.id"
        :message="message"
        :agent-name="agentName"
        :show-actions="showActions && !isStreaming"
        @toggle-reasoning="emit('toggle-reasoning', message.id)"
        @edit="emit('edit-message', message.id)"
        @regenerate="emit('regenerate', message.id)"
        @delete="emit('delete-message', message.id)"
      />
    </div>
  </div>
</template>

<style scoped>
.chat-message-list {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
}

.messages-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  padding: 1.5rem 1rem;
  max-width: 48rem;
  margin: 0 auto;
}

@media (min-width: 80rem) {
  .messages-container {
    max-width: 56rem;
  }
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  text-align: center;
}

.empty-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 4rem;
  height: 4rem;
  margin-bottom: 1rem;
  border-radius: 1rem;
  background: var(--color-primary-container);
  color: var(--color-primary);
}

.empty-icon svg {
  width: 2rem;
  height: 2rem;
}

.empty-title {
  margin: 0 0 0.5rem;
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--color-on-background);
}

.empty-description {
  margin: 0;
  max-width: 20rem;
  font-size: 0.875rem;
  color: var(--color-on-surface-variant);
}
</style>