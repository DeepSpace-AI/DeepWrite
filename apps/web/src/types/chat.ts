export type MessageRole = 'user' | 'assistant' | 'system'
export type MessageStatus = 'pending' | 'streaming' | 'done' | 'error'

export interface TextPart {
  type: 'text'
  text: string
}

export interface ReasoningPart {
  type: 'reasoning'
  text: string
  collapsed?: boolean
}

export interface ToolCallPart {
  type: 'tool_call'
  id: string
  name: string
  args: Record<string, unknown>
  result?: unknown
  status: 'pending' | 'running' | 'done' | 'error'
}

export interface ImagePart {
  type: 'image'
  url: string
  alt?: string
}

export interface AudioPart {
  type: 'audio'
  url: string
  transcript?: string
}

export interface ErrorPart {
  type: 'error'
  message: string
  code?: string
}

export type MessagePart = TextPart | ReasoningPart | ToolCallPart | ImagePart | AudioPart | ErrorPart

export interface ChatMessage {
  id: string
  role: MessageRole
  parts: MessagePart[]
  status: MessageStatus
  createdAt: Date
  modelUsed?: string
  tokenCount?: number
  latency?: number
}

export interface StreamDelta {
  text?: string
  reasoning?: string
  toolCall?: {
    id: string
    name: string
    argsDelta?: string
  }
  image?: { url: string }
  audio?: { url: string }
}

export interface StreamState {
  isStreaming: boolean
  isReasoning: boolean
  currentText: string
  currentReasoning: string
  error: string | null
}

export function createTextMessage(
  role: MessageRole,
  text: string,
  options?: { id?: string; status?: MessageStatus }
): ChatMessage {
  return {
    id: options?.id || generateMessageId(),
    role,
    parts: [{ type: 'text', text }],
    status: options?.status || 'done',
    createdAt: new Date(),
  }
}

export function createReasoningMessage(
  role: MessageRole,
  reasoning: string,
  text: string = '',
  options?: { id?: string; status?: MessageStatus }
): ChatMessage {
  const parts: MessagePart[] = []
  
  if (reasoning) {
    parts.push({ type: 'reasoning', text: reasoning, collapsed: true })
  }
  if (text) {
    parts.push({ type: 'text', text })
  }
  
  return {
    id: options?.id || generateMessageId(),
    role,
    parts,
    status: options?.status || 'done',
    createdAt: new Date(),
  }
}

export function createUserMessage(text: string): ChatMessage {
  return createTextMessage('user', text, { status: 'done' })
}

export function createPendingAssistantMessage(): ChatMessage {
  return {
    id: generateMessageId(),
    role: 'assistant',
    parts: [],
    status: 'pending',
    createdAt: new Date(),
  }
}

export function generateMessageId(): string {
  return `msg-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`
}

export function getTextFromMessage(message: ChatMessage): string {
  return message.parts
    .filter((p): p is TextPart => p.type === 'text')
    .map(p => p.text)
    .join('')
}

export function getReasoningFromMessage(message: ChatMessage): string {
  const part = message.parts.find((p): p is ReasoningPart => p.type === 'reasoning')
  return part?.text || ''
}

export function updateMessagePart(
  message: ChatMessage,
  type: MessagePart['type'],
  updater: (part: MessagePart | undefined) => MessagePart | undefined
): ChatMessage {
  const existingIndex = message.parts.findIndex(p => p.type === type)
  const existing = existingIndex >= 0 ? message.parts[existingIndex] : undefined
  const updated = updater(existing)
  
  const newParts = [...message.parts]
  
  if (updated) {
    if (existingIndex >= 0) {
      newParts[existingIndex] = updated
    } else {
      newParts.push(updated)
    }
  } else if (existingIndex >= 0) {
    newParts.splice(existingIndex, 1)
  }
  
  return { ...message, parts: newParts }
}

export function hasReasoning(message: ChatMessage): boolean {
  return message.parts.some(p => p.type === 'reasoning')
}

export function hasToolCalls(message: ChatMessage): boolean {
  return message.parts.some(p => p.type === 'tool_call')
}

export function hasError(message: ChatMessage): boolean {
  return message.parts.some(p => p.type === 'error')
}