import { http, unwrapResponse } from '@/api/http'
import { gatewayBase } from '@/api/http'

export interface Agent {
  id: string
  name: string
  description: string
  type: 'official' | 'user'
  category: string
  icon_url: string
  system_prompt: string
  default_model: string
  skills: Array<{ id: string; name: string }>
  tools: Array<{ id: string; name: string; enabled: boolean }>
  temperature: number
  max_tokens: number
  owner_id: string | null
  public: boolean
  rating: number
  usage_count: number
  enabled: boolean
  
  identity_prompt: string
  capability_prompt: string
  instruction_prompt: string
  safety_prompt: string
  
  inject_user_context: boolean
  inject_memory: boolean
  inject_time: boolean
  inject_workspace: boolean
  
  memory_retrieval_count: number
  memory_min_relevance: number
  
  created_at: string
  updated_at: string
}

export interface Session {
  id: string
  agent_id: string
  user_id: string
  workspace_id: string | null
  title: string
  context_docs: ContextDoc[]
  context_refs: ContextRef[]
  summary: string
  token_count: number
  status: 'active' | 'archived'
  pinned: boolean
  archived: boolean
  group_id: string | null
  tags: string[]
  last_message_at: string
  created_at: string
  updated_at: string
}

export interface Message {
  id: string
  session_id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  token_count: number
  model_used: string
  sources: MessageSource[]
  actions: MessageAction[]
  created_at: string
}

export interface ContextDoc {
  id: string
  name: string
}

export interface ContextRef {
  id: string
  title: string
}

export interface MessageSource {
  id: string
  type: string
  title: string
}

export interface MessageAction {
  id: string
  label: string
  type: string
  enabled: boolean
}

export interface Memory {
  id: string
  agent_id: string
  user_id: string
  session_id: string | null
  workspace_id: string | null
  type: 'preference' | 'fact' | 'task' | 'insight'
  content: string
  summary: string
  keywords: string[]
  source_session_id: string | null
  importance: number
  access_count: number
  expires_at: string | null
  created_at: string
  updated_at: string
}

export interface AgentListResult {
  items: Agent[]
}

export interface SessionListResult {
  items: Session[]
  limit: number
  offset: number
}

export interface MessageListResult {
  items: Message[]
  limit: number
  offset: number
}

export interface MemoryListResult {
  items: Memory[]
  limit: number
  offset: number
}

export interface CreateSessionInput {
  agent_id: string
  workspace_id?: string
  title?: string
  context_docs?: string
  context_refs?: string
}

export interface CreateMessageInput {
  role: string
  content: string
  token_count?: number
  model_used?: string
  sources?: string
  actions?: string
}

export interface CreateMemoryInput {
  agent_id: string
  session_id?: string
  workspace_id?: string
  type: string
  content: string
  summary?: string
  keywords?: string
  source_session_id?: string
  importance?: number
  expires_at?: string
}

export async function listAgents() {
  const res = await http.get('/agents')
  return unwrapResponse<AgentListResult>(res)
}

export async function listMyAgents() {
  const res = await http.get('/agents/my')
  return unwrapResponse<AgentListResult>(res)
}

export async function getAgent(agentId: string) {
  const res = await http.get(`/agents/${encodeURIComponent(agentId)}`)
  return unwrapResponse<Agent>(res)
}

export interface CreateAgentInput {
  name: string
  description?: string
  category?: string
  system_prompt: string
  identity_prompt?: string
  capability_prompt?: string
  instruction_prompt?: string
  safety_prompt?: string
  inject_user_context?: boolean
  inject_memory?: boolean
  inject_time?: boolean
  inject_workspace?: boolean
  memory_retrieval_count?: number
  memory_min_relevance?: number
  temperature?: number
  max_tokens?: number
  public?: boolean
}

export async function createAgent(input: CreateAgentInput) {
  const res = await http.post('/agents', input)
  return unwrapResponse<Agent>(res)
}

export async function updateAgent(agentId: string, input: Partial<Agent>) {
  const res = await http.put(`/agents/${encodeURIComponent(agentId)}`, input)
  return unwrapResponse<Agent>(res)
}

export async function deleteAgent(agentId: string) {
  const res = await http.delete(`/agents/${encodeURIComponent(agentId)}`)
  return unwrapResponse<{ id: string; deleted: boolean }>(res)
}

export async function activateAgent(agentId: string) {
  const res = await http.post(`/agents/${encodeURIComponent(agentId)}/activate`)
  return unwrapResponse<{ activated: boolean }>(res)
}

export interface SessionListFilter {
  agent_id?: string
  search?: string
  pinned?: boolean
  archived?: boolean
  group_id?: string
  limit?: number
  offset?: number
}

export async function listSessions(params?: SessionListFilter) {
  const res = await http.get('/sessions', { params })
  return unwrapResponse<SessionListResult>(res)
}

export async function getSession(sessionId: string) {
  const res = await http.get(`/sessions/${encodeURIComponent(sessionId)}`)
  return unwrapResponse<{ session: Session; messages: Message[] }>(res)
}

export async function createSession(input: CreateSessionInput) {
  const res = await http.post('/sessions', input)
  return unwrapResponse<Session>(res)
}

export async function updateSession(sessionId: string, input: Partial<Session>) {
  const res = await http.put(`/sessions/${encodeURIComponent(sessionId)}`, input)
  return unwrapResponse<Session>(res)
}

export async function deleteSession(sessionId: string) {
  const res = await http.delete(`/sessions/${encodeURIComponent(sessionId)}`)
  return unwrapResponse<{ id: string; deleted: boolean }>(res)
}

export async function generateSessionTitle(sessionId: string) {
  const res = await http.post(`/sessions/${encodeURIComponent(sessionId)}/generate-title`)
  return unwrapResponse<{ title: string }>(res)
}

export async function archiveSession(sessionId: string) {
  const res = await http.put(`/sessions/${encodeURIComponent(sessionId)}/archive`)
  return unwrapResponse<Session>(res)
}

export async function unarchiveSession(sessionId: string) {
  const res = await http.put(`/sessions/${encodeURIComponent(sessionId)}/unarchive`)
  return unwrapResponse<Session>(res)
}

export async function pinSession(sessionId: string) {
  const res = await http.put(`/sessions/${encodeURIComponent(sessionId)}/pin`)
  return unwrapResponse<Session>(res)
}

export async function unpinSession(sessionId: string) {
  const res = await http.put(`/sessions/${encodeURIComponent(sessionId)}/unpin`)
  return unwrapResponse<Session>(res)
}

export async function listMessages(sessionId: string, params?: { limit?: number; offset?: number }) {
  const res = await http.get(`/sessions/${encodeURIComponent(sessionId)}/messages`, { params })
  return unwrapResponse<MessageListResult>(res)
}

export async function createMessage(sessionId: string, input: CreateMessageInput) {
  const res = await http.post(`/sessions/${encodeURIComponent(sessionId)}/messages`, input)
  return unwrapResponse<Message>(res)
}

export async function listMemories(params: { agent_id?: string; session_id?: string; workspace_id?: string; limit?: number; offset?: number }) {
  const res = await http.get('/memories', { params })
  return unwrapResponse<MemoryListResult>(res)
}

export async function getSessionMemories(sessionId: string, params?: { limit?: number; offset?: number }) {
  const res = await http.get('/memories', { params: { session_id: sessionId, ...params } })
  return unwrapResponse<MemoryListResult>(res)
}

export async function getRecentMemories(params: { agent_id?: string; session_id?: string; workspace_id?: string }) {
  const res = await http.get('/memories/recent', { params })
  return unwrapResponse<MemoryListResult>(res)
}

export async function createMemory(input: CreateMemoryInput) {
  const res = await http.post('/memories', input)
  return unwrapResponse<Memory>(res)
}

export async function updateMemory(memoryId: string, input: Partial<Memory>) {
  const res = await http.put(`/memories/${encodeURIComponent(memoryId)}`, input)
  return unwrapResponse<Memory>(res)
}

export async function deleteMemory(memoryId: string) {
  const res = await http.delete(`/memories/${encodeURIComponent(memoryId)}`)
  return unwrapResponse<{ id: string; deleted: boolean }>(res)
}

export async function triggerMemoryExtraction(sessionId: string) {
  const res = await http.post(`/sessions/${encodeURIComponent(sessionId)}/extract-memories`)
  return unwrapResponse<{ task_id: string; task_name: string; queue: string }>(res)
}

export async function getMemoryExtractionStatus(taskId: string) {
  const res = await http.get(`/memories/tasks/${encodeURIComponent(taskId)}`)
  return unwrapResponse<{ task_id: string; status: string; ready: boolean; result?: MemoryListResult; error?: string }>(res)
}

export interface SessionGroup {
  id: string
  user_id: string
  name: string
  description: string
  color: string
  icon: string
  sort_order: number
  created_at: string
  updated_at: string
}

export interface SessionShare {
  id: string
  session_id: string
  user_id: string
  share_token: string
  title: string
  expires_at: string | null
  view_count: number
  allow_copy: boolean
  is_public: boolean
  created_at: string
  updated_at: string
}

export interface CreateSessionGroupInput {
  name: string
  description?: string
  color?: string
  icon?: string
}

export interface CreateSessionShareInput {
  session_id: string
  title?: string
  expires_in_hours?: number
  allow_copy?: boolean
  is_public?: boolean
}

export async function listSessionGroups() {
  const res = await http.get('/session-groups')
  return unwrapResponse<{ items: SessionGroup[] }>(res)
}

export async function createSessionGroup(input: CreateSessionGroupInput) {
  const res = await http.post('/session-groups', input)
  return unwrapResponse<SessionGroup>(res)
}

export async function updateSessionGroup(groupId: string, input: Partial<SessionGroup>) {
  const res = await http.put(`/session-groups/${encodeURIComponent(groupId)}`, input)
  return unwrapResponse<SessionGroup>(res)
}

export async function deleteSessionGroup(groupId: string) {
  const res = await http.delete(`/session-groups/${encodeURIComponent(groupId)}`)
  return unwrapResponse<{ id: string; deleted: boolean }>(res)
}

export async function listSessionShares() {
  const res = await http.get('/session-shares')
  return unwrapResponse<{ items: SessionShare[] }>(res)
}

export async function createSessionShare(input: CreateSessionShareInput) {
  const res = await http.post('/session-shares', input)
  return unwrapResponse<SessionShare>(res)
}

export async function deleteSessionShare(shareId: string) {
  const res = await http.delete(`/session-shares/${encodeURIComponent(shareId)}`)
  return unwrapResponse<{ id: string; deleted: boolean }>(res)
}

export async function getSharedSession(token: string) {
  const res = await http.get(`/shares/${encodeURIComponent(token)}`)
  return unwrapResponse<{ share: SessionShare; session: Session; messages: Message[] }>(res)
}

export function getSessionExportUrl(sessionId: string, format: 'markdown' | 'json'): string {
  return `${gatewayBase}/api/v1/sessions/${encodeURIComponent(sessionId)}/export/${format}`
}