import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { ChatMessage, MessageStatus } from '@/types/chat'
import { generateMessageId, updateMessagePart } from '@/types/chat'
import { gatewayBase } from '@/api/http'
import { cacheManager } from '@/utils/cache'
import type { Message } from '@/api/agent'

export const useChatStore = defineStore('chat', () => {
  const messages = ref<ChatMessage[]>([])
  const isStreaming = ref(false)
  const isReasoning = ref(false)
  const error = ref<string | null>(null)
  const currentSessionId = ref<string | null>(null)

  let abortController: AbortController | null = null
  let streamingMessageId: string | null = null

  const isEmpty = computed(() => messages.value.length === 0)
  const lastMessage = computed(() => messages.value[messages.value.length - 1] || null)

  function addUserMessage(content: string): ChatMessage {
    const message: ChatMessage = {
      id: generateMessageId(),
      role: 'user',
      parts: [{ type: 'text', text: content.trim() }],
      status: 'done',
      createdAt: new Date(),
    }
    messages.value.push(message)
    return message
  }

  function addPendingAssistantMessage(): ChatMessage {
    const message: ChatMessage = {
      id: generateMessageId(),
      role: 'assistant',
      parts: [],
      status: 'pending',
      createdAt: new Date(),
    }
    messages.value.push(message)
    streamingMessageId = message.id
    return message
  }

  function updateStreamingMessage(updater: (msg: ChatMessage) => ChatMessage) {
    if (!streamingMessageId) return
    const index = messages.value.findIndex(m => m.id === streamingMessageId)
    if (index >= 0) {
      const updated = updater(messages.value[index]!)
      messages.value[index] = updated
    }
  }

  function setMessageStatus(id: string, status: MessageStatus) {
    const index = messages.value.findIndex(m => m.id === id)
    if (index >= 0) {
      const msg = messages.value[index]!
      msg.status = status
    }
  }

  async function sendMessage(content: string): Promise<void> {
    if (!content.trim() || !currentSessionId.value || isStreaming.value) {
      return
    }

    error.value = null
    addUserMessage(content)
    addPendingAssistantMessage()

    isStreaming.value = true
    abortController = new AbortController()

    try {
      const token = cacheManager.get<string>('deepwrite_access_token') || ''
      const response = await fetch(
        `${gatewayBase}/api/v1/sessions/${encodeURIComponent(currentSessionId.value)}/chat/stream`,
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
      
      if (streamingMessageId) {
        setMessageStatus(streamingMessageId, 'done')
      }
    } catch (err) {
      if (err instanceof Error && err.name === 'AbortError') {
        if (streamingMessageId) {
          setMessageStatus(streamingMessageId, 'done')
        }
        return
      }

      const errorMsg = err instanceof Error ? err.message : 'Stream failed'
      error.value = errorMsg
      
      if (streamingMessageId) {
        updateStreamingMessage((msg) => 
          updateMessagePart(msg, 'error', () => ({
            type: 'error',
            message: errorMsg,
          }))
        )
        setMessageStatus(streamingMessageId, 'error')
      }
    } finally {
      isStreaming.value = false
      isReasoning.value = false
      abortController = null
      streamingMessageId = null
    }
  }

  async function processStream(reader: ReadableStreamDefaultReader<Uint8Array>) {
    const decoder = new TextDecoder()
    let buffer = ''
    let textContent = ''
    let reasoningContent = ''

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
          const chunk = JSON.parse(data)
          const delta = chunk.choices?.[0]?.delta

          if (delta?.reasoning_content) {
            isReasoning.value = true
            reasoningContent += delta.reasoning_content
            updateStreamingMessage((msg) => 
              updateMessagePart(msg, 'reasoning', () => ({
                type: 'reasoning',
                text: reasoningContent,
                collapsed: false,
              }))
            )
          }

          if (delta?.content) {
            isReasoning.value = false
            textContent += delta.content
            updateStreamingMessage((msg) => 
              updateMessagePart(msg, 'text', () => ({
                type: 'text',
                text: textContent,
              }))
            )
          }
        } catch {
          // Skip invalid JSON
        }
      }
    }
  }

  function abort() {
    if (abortController) {
      abortController.abort()
      abortController = null
    }
  }

  function clearMessages() {
    messages.value = []
    error.value = null
    streamingMessageId = null
  }

  function setSession(sessionId: string | null) {
    if (currentSessionId.value !== sessionId) {
      clearMessages()
    }
    currentSessionId.value = sessionId
  }

  function loadFromHistory(historyMessages: Message[]) {
    messages.value = historyMessages.map(m => ({
      id: m.id,
      role: m.role as 'user' | 'assistant' | 'system',
      parts: [{ type: 'text' as const, text: m.content }],
      status: 'done' as const,
      createdAt: new Date(m.created_at),
      modelUsed: m.model_used,
      tokenCount: m.token_count,
    }))
  }

  function editMessage(id: string, newContent: string) {
    const index = messages.value.findIndex(m => m.id === id)
    if (index >= 0) {
      const msg = messages.value[index]!
      messages.value[index] = updateMessagePart(
        msg,
        'text',
        () => ({ type: 'text', text: newContent })
      )
    }
  }

  function deleteMessage(id: string) {
    const index = messages.value.findIndex(m => m.id === id)
    if (index >= 0) {
      messages.value.splice(index, 1)
    }
  }

  function deleteMessagesAfter(id: string) {
    const index = messages.value.findIndex(m => m.id === id)
    if (index >= 0) {
      messages.value = messages.value.slice(0, index + 1)
    }
  }

  async function regenerateFromMessage(messageId: string): Promise<void> {
    const index = messages.value.findIndex(m => m.id === messageId)
    if (index < 0) return

    const message = messages.value[index]
    if (!message || message.role !== 'user') return

    const textPart = message.parts.find(p => p.type === 'text')
    if (!textPart || textPart.type !== 'text') return

    messages.value = messages.value.slice(0, index + 1)
    
    await sendMessage(textPart.text)
  }

  function toggleReasoningCollapsed(id: string) {
    const index = messages.value.findIndex(m => m.id === id)
    if (index >= 0) {
      const msg = messages.value[index]!
      const part = msg.parts.find(p => p.type === 'reasoning')
      if (part && part.type === 'reasoning') {
        part.collapsed = !part.collapsed
      }
    }
  }

  return {
    messages,
    isStreaming,
    isReasoning,
    error,
    currentSessionId,
    isEmpty,
    lastMessage,
    sendMessage,
    abort,
    clearMessages,
    setSession,
    loadFromHistory,
    editMessage,
    deleteMessage,
    deleteMessagesAfter,
    regenerateFromMessage,
    toggleReasoningCollapsed,
  }
})