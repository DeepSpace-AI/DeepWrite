import { http, unwrapResponse } from '@/api/http'

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
  created_at: string
  updated_at: string
}

export interface AgentListResult {
  items: Agent[]
}

export interface CreateAgentInput {
  name: string
  description?: string
  category?: string
  icon_url?: string
  system_prompt: string
  default_model?: string
  skills?: string
  tools?: string
  temperature?: number
  max_tokens?: number
}

export interface UpdateAgentInput {
  name?: string
  description?: string
  category?: string
  icon_url?: string
  system_prompt?: string
  default_model?: string
  skills?: string
  tools?: string
  temperature?: number
  max_tokens?: number
}

export async function listOfficialAgents() {
  const res = await http.get('/admin/agents')
  return unwrapResponse<AgentListResult>(res)
}

export async function listPublicAgents() {
  const res = await http.get('/admin/agents/public')
  return unwrapResponse<AgentListResult>(res)
}

export async function getAgent(agentId: string) {
  const res = await http.get(`/admin/agents/${encodeURIComponent(agentId)}`)
  return unwrapResponse<Agent>(res)
}

export async function createAgent(input: CreateAgentInput) {
  const res = await http.post('/admin/agents', input)
  return unwrapResponse<Agent>(res)
}

export async function updateAgent(agentId: string, input: UpdateAgentInput) {
  const res = await http.put(`/admin/agents/${encodeURIComponent(agentId)}`, input)
  return unwrapResponse<Agent>(res)
}

export async function deleteAgent(agentId: string) {
  const res = await http.delete(`/admin/agents/${encodeURIComponent(agentId)}`)
  return unwrapResponse<{ id: string; deleted: boolean }>(res)
}

export async function updateAgentStatus(agentId: string, enabled: boolean) {
  const res = await http.put(`/admin/agents/${encodeURIComponent(agentId)}/status`, { enabled })
  return unwrapResponse<Agent>(res)
}

export async function moderateAgent(agentId: string, enabled: boolean) {
  const res = await http.put(`/admin/agents/${encodeURIComponent(agentId)}/moderate`, { enabled })
  return unwrapResponse<Agent>(res)
}