import { ref, onUnmounted, type Ref } from 'vue'
import { gatewayBase } from '@/api/http'
import { cacheManager } from '@/utils/cache'
import type { ChatMessage, StreamDelta, MessageStatus } from '@/types/chat'
import { generateMessageId, updateMessagePart } from '@/types/chat'

export interface UseChatStreamOptions {
  sessionId: Ref<string | null>
  onMessageUpdate?: (message: ChatMessage) => void
  onError?: (error: Error) => void
  onComplete?: (message: ChatMessage) => void
}

export interface ChatStreamState {
  isStreaming: boolean
  isReasoning: boolean
  error: string | null
}

interface SSEChunk {
  choices?: Array<{
    delta?: {
      content?: string
      reasoning_content?: string
      tool_calls?: Array<{
        id: string
        function?: {
          name?: string
          arguments?: string
        }
      }>
    }
    finish_reason?: string | null
  }>
}

interface ToolCallDeltaItem {
  id: string
  function?: {
    name?: string
    arguments?: string
  }
}

type ToolCallDelta = ToolCallDeltaItem | undefined

export function useChatStream(options: UseChatStreamOptions) {
  const state = ref<ChatStreamState>({
    isStreaming: false,
    isReasoning: false,
    error: null,
  })

  let abortController: AbortController | null = null
  let currentMessage: ChatMessage | null = null
  let startTime: number = 0

  async function streamMessage(content: string): Promise<ChatMessage | null> {
    if (!options.sessionId.value || state.value.isStreaming) {
      return null
    }

    state.value.isStreaming = true
    state.value.isReasoning = false
    state.value.error = null
    startTime = Date.now()

    currentMessage = {
      id: generateMessageId(),
      role: 'assistant',
      parts: [],
      status: 'streaming',
      createdAt: new Date(),
    }

    abortController = new AbortController()

    try {
      const token = cacheManager.get<string>('deepwrite_access_token') || ''
      const response = await fetch(
        `${gatewayBase}/api/v1/sessions/${encodeURIComponent(options.sessionId.value)}/chat/stream`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
          body: JSON.stringify({ content: content.trim(), stream: true }),
          signal: abortController.signal,
        }
      )

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(`HTTP ${response.status}: ${errorText}`)
      }

      const reader = response.body?.getReader()
      if (!reader) {
        throw new Error('No response body')
      }

      await processStream(reader)
      
      if (currentMessage) {
        currentMessage.status = 'done'
        currentMessage.latency = Date.now() - startTime
        options.onComplete?.(currentMessage)
      }

      return currentMessage
    } catch (err) {
      if (err instanceof Error && err.name === 'AbortError') {
        if (currentMessage) {
          currentMessage.status = 'done'
        }
        return currentMessage
      }

      const error = err instanceof Error ? err : new Error('Stream failed')
      state.value.error = error.message
      
      if (currentMessage) {
        currentMessage.status = 'error'
        currentMessage = updateMessagePart(currentMessage, 'error', () => ({
          type: 'error',
          message: error.message,
        }))
      }
      
      options.onError?.(error)
      return currentMessage
    } finally {
      state.value.isStreaming = false
      state.value.isReasoning = false
      abortController = null
    }
  }

  async function processStream(reader: ReadableStreamDefaultReader<Uint8Array>) {
    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (!line.startsWith('data: ')) continue

        const data = line.slice(6).trim()
        if (data === '[DONE]') {
          return
        }

        try {
          const chunk: SSEChunk = JSON.parse(data)
          processChunk(chunk)
        } catch {
          // Skip invalid JSON
        }
      }
    }
  }

  function processChunk(chunk: SSEChunk) {
    if (!currentMessage) return

    const delta = chunk.choices?.[0]?.delta
    if (!delta) return

    // Handle reasoning content (e.g., DeepSeek R1)
    if (delta.reasoning_content) {
      state.value.isReasoning = true
      currentMessage = updateMessagePart(currentMessage, 'reasoning', (existing) => {
        const prev = existing?.type === 'reasoning' ? existing.text : ''
        return {
          type: 'reasoning',
          text: prev + delta.reasoning_content,
          collapsed: false,
        }
      })
    }

    // Handle regular content
    if (delta.content) {
      state.value.isReasoning = false
      currentMessage = updateMessagePart(currentMessage, 'text', (existing) => {
        const prev = existing?.type === 'text' ? existing.text : ''
        return {
          type: 'text',
          text: prev + delta.content,
        }
      })
    }

    // Handle tool calls
    if (delta.tool_calls) {
      for (const toolCall of delta.tool_calls) {
        processToolCallDelta(toolCall)
      }
    }

    options.onMessageUpdate?.(currentMessage)
  }

  function processToolCallDelta(toolCallDelta: ToolCallDelta | undefined) {
    if (!currentMessage || !toolCallDelta) return

    // Tool call handling will be expanded in Phase 3
    currentMessage = updateMessagePart(currentMessage, 'tool_call', (existing) => {
      if (existing?.type === 'tool_call') {
        return {
          ...existing,
          args: {
            ...existing.args,
            ...(toolCallDelta.function?.arguments ? JSON.parse(toolCallDelta.function.arguments) : {}),
          },
        }
      }
      return {
        type: 'tool_call',
        id: toolCallDelta.id || generateMessageId(),
        name: toolCallDelta.function?.name || 'unknown',
        args: {},
        status: 'running',
      }
    })
  }

  function abort() {
    if (abortController) {
      abortController.abort()
      abortController = null
    }
  }

  function retry(content: string): Promise<ChatMessage | null> {
    abort()
    return streamMessage(content)
  }

  onUnmounted(() => {
    abort()
  })

  return {
    state,
    streamMessage,
    abort,
    retry,
    get currentMessage() {
      return currentMessage
    },
  }
}

export function useChatStreamSimple(sessionId: Ref<string | null>) {
  const messages = ref<ChatMessage[]>([])
  
  const stream = useChatStream({
    sessionId,
    onMessageUpdate: (message) => {
      const index = messages.value.findIndex(m => m.id === message.id)
      if (index >= 0) {
        messages.value[index] = { ...message }
      }
    },
    onComplete: (message) => {
      const index = messages.value.findIndex(m => m.id === message.id)
      if (index >= 0) {
        messages.value[index] = { ...message }
      }
    },
  })

  async function sendMessage(content: string) {
    if (!content.trim() || stream.state.value.isStreaming) return

    const userMessage: ChatMessage = {
      id: generateMessageId(),
      role: 'user',
      parts: [{ type: 'text', text: content.trim() }],
      status: 'done',
      createdAt: new Date(),
    }
    messages.value.push(userMessage)

    const assistantMessage = await stream.streamMessage(content)
    if (assistantMessage) {
      const existingIndex = messages.value.findIndex(m => m.id === assistantMessage.id)
      if (existingIndex >= 0) {
        messages.value[existingIndex] = assistantMessage
      } else {
        messages.value.push(assistantMessage)
      }
    }
  }

  function clearMessages() {
    messages.value = []
  }

  function loadHistory(history: ChatMessage[]) {
    messages.value = history
  }

  return {
    messages,
    state: stream.state,
    sendMessage,
    abort: stream.abort,
    clearMessages,
    loadHistory,
  }
}