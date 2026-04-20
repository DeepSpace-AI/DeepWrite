<script setup lang="ts">
import { ref, watch, computed, nextTick } from 'vue'
import type { AIProviderModel } from '@/api/aiProvider'
import { gatewayBase } from '@/api/http'
import { useAuthStore } from '@/stores/auth'
import IconSend from '~icons/mdi/send'
import IconClose from '~icons/mdi/close'
import IconDelete from '~icons/mdi/delete'
import IconCog from '~icons/mdi/cog'
import IconChevronDown from '~icons/mdi/chevron-down'
import IconChevronUp from '~icons/mdi/chevron-up'
import IconBrain from '~icons/mdi/brain'

interface Props {
  open: boolean
  model: AIProviderModel | null
}

const props = defineProps<Props>()
const emit = defineEmits(['close'])

const authStore = useAuthStore()

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  reasoning_content?: string
  timestamp: Date
}

const messages = ref<Message[]>([])
const inputMessage = ref('')
const isLoading = ref(false)
const streamingContent = ref('')
const streamingReasoning = ref('')
const showSettings = ref(false)
const temperature = ref(0.7)
const maxTokens = ref(2048)
const stats = ref<{
  latencyMs: number
  promptTokens: number
  completionTokens: number
  reasoningTokens?: number
} | null>(null)
const error = ref('')
const expandedReasoning = ref<Set<string>>(new Set())

const messagesContainer = ref<HTMLElement | null>(null)

const canSend = computed(() => {
  return inputMessage.value.trim() && !isLoading.value && props.model?.enabled
})

function generateId(): string {
  return Math.random().toString(36).substring(2, 15)
}

async function scrollToBottom() {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

function toggleReasoning(messageId: string) {
  if (expandedReasoning.value.has(messageId)) {
    expandedReasoning.value.delete(messageId)
  } else {
    expandedReasoning.value.add(messageId)
  }
}

async function sendMessage() {
  if (!canSend.value) return

  const userMessage = inputMessage.value.trim()
  inputMessage.value = ''
  error.value = ''
  stats.value = null

  messages.value.push({
    id: generateId(),
    role: 'user',
    content: userMessage,
    timestamp: new Date(),
  })

  isLoading.value = true
  streamingContent.value = ''
  streamingReasoning.value = ''

  const startTime = performance.now()

  try {
    console.log('[ModelTestDrawer] Sending request to:', `${gatewayBase}/api/v1/ai/chat`)
    console.log('[ModelTestDrawer] Model ID:', props.model?.id)
    
    const response = await fetch(`${gatewayBase}/api/v1/ai/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${authStore.accessToken}`,
      },
      body: JSON.stringify({
        model: props.model?.id,
        messages: messages.value
          .filter(m => m.role === 'user' || m.role === 'assistant')
          .map(m => ({ role: m.role, content: m.content })),
        temperature: temperature.value,
        max_tokens: maxTokens.value,
        stream: true,
      }),
    })

    console.log('[ModelTestDrawer] Response status:', response.status)
    
    if (!response.ok) {
      const errorText = await response.text()
      console.error('[ModelTestDrawer] Error response:', errorText)
      let errorMessage = `HTTP ${response.status}`
      try {
        const errorJson = JSON.parse(errorText)
        if (errorJson.error?.message) {
          errorMessage = errorJson.error.message
        } else if (errorJson.message) {
          errorMessage = errorJson.message
        } else {
          errorMessage = errorText.slice(0, 500)
        }
      } catch {
        errorMessage = errorText.slice(0, 500) || `HTTP ${response.status}`
      }
      throw new Error(errorMessage)
    }

    const reader = response.body?.getReader()
    if (!reader) {
      throw new Error('No response body')
    }

    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) {
        console.log('[ModelTestDrawer] Stream done')
        break
      }

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (!line.startsWith('data: ')) continue

        const data = line.slice(6)
        if (data === '[DONE]') {
          console.log('[ModelTestDrawer] Received [DONE]')
          break
        }

        try {
          const parsed = JSON.parse(data)
          
          if (parsed.error) {
            throw new Error(parsed.error.message || 'Unknown error')
          }

          const delta = parsed.choices?.[0]?.delta
          
          if (delta?.reasoning_content) {
            streamingReasoning.value += delta.reasoning_content
            scrollToBottom()
          }
          
          if (delta?.content) {
            streamingContent.value += delta.content
            scrollToBottom()
          }

          if (parsed.usage) {
            stats.value = {
              latencyMs: Math.round(performance.now() - startTime),
              promptTokens: parsed.usage.prompt_tokens || 0,
              completionTokens: parsed.usage.completion_tokens || 0,
              reasoningTokens: parsed.usage.completion_tokens_details?.reasoning_tokens || 0,
            }
          }
        } catch (e) {
          if (e instanceof SyntaxError) {
            console.warn('[ModelTestDrawer] Failed to parse:', data)
            continue
          }
          throw e
        }
      }
    }

    if (streamingContent.value || streamingReasoning.value) {
      messages.value.push({
        id: generateId(),
        role: 'assistant',
        content: streamingContent.value,
        reasoning_content: streamingReasoning.value || undefined,
        timestamp: new Date(),
      })
      streamingContent.value = ''
      streamingReasoning.value = ''
    }

    if (!stats.value) {
      stats.value = {
        latencyMs: Math.round(performance.now() - startTime),
        promptTokens: 0,
        completionTokens: 0,
      }
    }
    
    console.log('[ModelTestDrawer] Completed, stats:', stats.value)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Unknown error'
    console.error('[ModelTestDrawer] Chat error:', e)
  } finally {
    isLoading.value = false
    scrollToBottom()
  }
}

function clearMessages() {
  messages.value = []
  streamingContent.value = ''
  streamingReasoning.value = ''
  stats.value = null
  error.value = ''
  expandedReasoning.value.clear()
}

function closeDrawer() {
  emit('close')
}

watch(() => props.open, (newVal) => {
  if (!newVal) {
    streamingContent.value = ''
    streamingReasoning.value = ''
    isLoading.value = false
  }
})

watch(() => props.model, () => {
  clearMessages()
})
</script>

<template>
  <Teleport to="body">
    <Transition name="drawer">
      <div v-if="open" class="fixed inset-0 z-50">
        <div class="absolute inset-0 bg-black/50" @click="closeDrawer" />
        
        <aside class="absolute right-0 top-0 h-full w-[480px] max-w-full bg-[var(--bg-elevated)] shadow-2xl flex flex-col">
          <header class="flex items-center justify-between border-b border-[var(--border-subtle)] px-5 py-4">
            <div>
              <h2 class="text-heading text-base">模型测试</h2>
              <p v-if="model" class="text-muted mt-0.5 text-sm">{{ model.model }}</p>
            </div>
            <button
              class="flex h-8 w-8 items-center justify-center rounded-lg text-[var(--text-muted)] transition-colors hover:bg-[var(--surface-hover)]"
              @click="closeDrawer"
            >
              <IconClose class="h-5 w-5" />
            </button>
          </header>

          <div class="flex-1 overflow-hidden flex flex-col">
            <div v-if="!model?.enabled" class="flex-1 flex items-center justify-center">
              <p class="text-muted text-sm">请选择一个已启用的模型</p>
            </div>

            <template v-else>
              <div ref="messagesContainer" class="flex-1 overflow-y-auto p-4 space-y-4">
                <div v-if="!messages.length && !streamingContent && !streamingReasoning" class="flex flex-col items-center justify-center h-full text-center">
                  <div class="w-16 h-16 rounded-2xl bg-[var(--bg-base)] flex items-center justify-center mb-4">
                    <svg class="w-8 h-8 text-[var(--text-muted)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                    </svg>
                  </div>
                  <p class="text-secondary text-sm">发送消息开始测试</p>
                  <p class="text-muted mt-1 text-xs">支持 SSE 流式输出与深度思考</p>
                </div>

                <div v-for="msg in messages" :key="msg.id" class="flex" :class="msg.role === 'user' ? 'justify-end' : 'justify-start'">
                  <div class="max-w-[85%] space-y-2">
                    <div
                      v-if="msg.reasoning_content"
                      class="rounded-xl border border-[var(--border-subtle)] bg-[var(--bg-base)] overflow-hidden"
                    >
                      <button
                        class="w-full flex items-center gap-2 px-3 py-2 text-xs font-medium text-[var(--text-secondary)] hover:bg-[var(--surface-hover)] transition-colors"
                        @click="toggleReasoning(msg.id)"
                      >
                        <IconBrain class="w-4 h-4 text-[var(--accent-secondary)]" />
                        <span>深度思考</span>
                        <IconChevronDown 
                          class="w-4 h-4 ml-auto transition-transform"
                          :class="{ 'rotate-180': expandedReasoning.has(msg.id) }"
                        />
                      </button>
                      <div 
                        v-show="expandedReasoning.has(msg.id)"
                        class="px-3 py-2 text-xs text-[var(--text-muted)] whitespace-pre-wrap break-words border-t border-[var(--border-subtle)] max-h-64 overflow-y-auto"
                      >
                        {{ msg.reasoning_content }}
                      </div>
                    </div>
                    
                    <div
                      class="rounded-2xl px-4 py-2.5 text-sm"
                      :class="msg.role === 'user' 
                        ? 'bg-[var(--accent-primary)] text-white' 
                        : 'bg-[var(--bg-base)] text-[var(--text-primary)]'"
                    >
                      <p class="whitespace-pre-wrap break-words">{{ msg.content }}</p>
                    </div>
                  </div>
                </div>

                <div v-if="streamingReasoning || streamingContent" class="flex justify-start">
                  <div class="max-w-[85%] space-y-2">
                    <div
                      v-if="streamingReasoning"
                      class="rounded-xl border border-[var(--border-subtle)] bg-[var(--bg-base)] overflow-hidden"
                    >
                      <div class="flex items-center gap-2 px-3 py-2 text-xs font-medium text-[var(--text-secondary)]">
                        <IconBrain class="w-4 h-4 text-[var(--accent-secondary)]" />
                        <span>深度思考中...</span>
                        <span class="loading loading-spinner loading-xs" />
                      </div>
                      <div class="px-3 py-2 text-xs text-[var(--text-muted)] whitespace-pre-wrap break-words border-t border-[var(--border-subtle)] max-h-64 overflow-y-auto">
                        {{ streamingReasoning }}
                      </div>
                    </div>
                    
                    <div v-if="streamingContent" class="rounded-2xl bg-[var(--bg-base)] px-4 py-2.5 text-sm">
                      <p class="whitespace-pre-wrap break-words">{{ streamingContent }}</p>
                      <span v-if="!streamingReasoning" class="inline-block w-2 h-4 ml-1 bg-[var(--accent-primary)] animate-pulse" />
                    </div>
                  </div>
                </div>

                <div v-if="error" class="flex justify-center">
                  <div class="rounded-xl bg-[var(--glow-warm)] border border-[var(--accent-error)] px-4 py-2 text-sm text-[var(--accent-error)]">
                    {{ error }}
                  </div>
                </div>
              </div>

              <div class="border-t border-[var(--border-subtle)] p-4 space-y-3">
                <div v-if="stats" class="flex items-center gap-4 text-xs text-[var(--text-muted)]">
                  <span>延迟: {{ stats.latencyMs }}ms</span>
                  <span v-if="stats.promptTokens">输入: {{ stats.promptTokens }} tokens</span>
                  <span v-if="stats.completionTokens">输出: {{ stats.completionTokens }} tokens</span>
                  <span v-if="stats.reasoningTokens">思考: {{ stats.reasoningTokens }} tokens</span>
                </div>

                <div>
                  <button
                    class="flex items-center gap-1.5 text-xs text-[var(--text-muted)] hover:text-[var(--text-primary)]"
                    @click="showSettings = !showSettings"
                  >
                    <IconCog class="w-4 h-4" />
                    <span>参数配置</span>
                    <component :is="showSettings ? IconChevronUp : IconChevronDown" class="w-3 h-3" />
                  </button>
                  
                  <div v-if="showSettings" class="mt-3 grid grid-cols-2 gap-4 rounded-xl bg-[var(--bg-base)] p-3">
                    <div>
                      <label class="text-label mb-1.5 block text-xs">Temperature</label>
                      <input
                        v-model.number="temperature"
                        type="range"
                        min="0"
                        max="2"
                        step="0.1"
                        class="w-full h-2 bg-[var(--border-subtle)] rounded-lg appearance-none cursor-pointer"
                      />
                      <p class="text-mono text-xs text-[var(--text-muted)] mt-1">{{ temperature }}</p>
                    </div>
                    <div>
                      <label class="text-label mb-1.5 block text-xs">Max Tokens</label>
                      <input
                        v-model.number="maxTokens"
                        type="number"
                        min="1"
                        max="32000"
                        class="tech-input w-full px-3 py-1.5 text-xs"
                      />
                    </div>
                  </div>
                </div>

                <div class="flex gap-2">
                  <button
                    class="btn-ghost-tech flex items-center gap-1.5 px-3 py-2 text-xs"
                    @click="clearMessages"
                    :disabled="!messages.length"
                  >
                    <IconDelete class="w-4 h-4" />
                    清空
                  </button>
                  <div class="flex-1 flex gap-2">
                    <input
                      v-model="inputMessage"
                      type="text"
                      class="tech-input flex-1 px-4 py-2 text-sm"
                      placeholder="输入消息..."
                      :disabled="isLoading"
                      @keydown.enter="sendMessage"
                    />
                    <button
                      class="btn-tech flex items-center gap-1.5 px-4 py-2 text-sm"
                      :disabled="!canSend"
                      @click="sendMessage"
                    >
                      <span v-if="isLoading" class="loading loading-spinner loading-xs" />
                      <IconSend v-else class="w-4 h-4" />
                      发送
                    </button>
                  </div>
                </div>
              </div>
            </template>
          </div>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.drawer-enter-active,
.drawer-leave-active {
  transition: opacity 0.2s ease;
}

.drawer-enter-active aside,
.drawer-leave-active aside {
  transition: transform 0.3s ease;
}

.drawer-enter-from,
.drawer-leave-to {
  opacity: 0;
}

.drawer-enter-from aside,
.drawer-leave-to aside {
  transform: translateX(100%);
}
</style>