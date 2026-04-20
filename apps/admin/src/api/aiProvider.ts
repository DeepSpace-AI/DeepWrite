import { http, unwrapResponse } from '@/api/http'

export interface AIProviderModel {
  id: string
  provider_id?: string
  model: string
  provider: string
  request_model: string
  supports_chat_completions: boolean
  supports_chat_responses: boolean
  supports_embeddings: boolean
  supports_rerank: boolean
  supports_audio_speech: boolean
  supports_audio_transcriptions: boolean
  supports_models: boolean
  base_url: string
  organization: string
  chat_completions_path: string
  chat_responses_path: string
  embeddings_path: string
  rerank_path: string
  audio_speech_path: string
  audio_transcriptions_path: string
  models_path: string
  extra_headers: Record<string, string>
  enabled: boolean
  has_api_key: boolean
  api_key_masked: string
  created_at: string
  updated_at: string
}

export interface AIProviderListResult {
  items: AIProviderModel[]
  limit: number
  offset: number
}

export interface AIProviderVendor {
  id: string
  name: string
  base_url: string
  organization: string
  chat_completions_path: string
  chat_responses_path: string
  embeddings_path: string
  rerank_path: string
  audio_speech_path: string
  audio_transcriptions_path: string
  models_path: string
  extra_headers: Record<string, string>
  enabled: boolean
  has_api_key: boolean
  api_key_masked: string
  created_at: string
  updated_at: string
}

export interface AIProviderVendorListResult {
  items: AIProviderVendor[]
  limit: number
  offset: number
}

export interface AIProviderVendorInput {
  provider: string
  base_url: string
  api_key: string
  organization?: string
  chat_completions_path?: string
  chat_responses_path?: string
  embeddings_path?: string
  rerank_path?: string
  audio_speech_path?: string
  audio_transcriptions_path?: string
  models_path?: string
  extra_headers?: Record<string, string>
  enabled?: boolean
}

export interface DiscoveredProviderModel {
  model: string
  name: string
  owned_by: string
}

export interface VendorDetailResult {
  provider: AIProviderVendor
  models: AIProviderModel[]
  limit: number
  offset: number
}

export interface DiscoverVendorModelsResult {
  provider: AIProviderVendor
  items: DiscoveredProviderModel[]
}

export interface CreateModelByVendorInput {
  model: string
  request_model?: string
  supports_chat_completions?: boolean
  supports_chat_responses?: boolean
  supports_embeddings?: boolean
  supports_rerank?: boolean
  supports_audio_speech?: boolean
  supports_audio_transcriptions?: boolean
  supports_models?: boolean
  enabled?: boolean
}

export interface AIProviderInput {
  model: string
  provider: string
  request_model: string
  base_url: string
  api_key?: string
  organization?: string
  chat_completions_path?: string
  chat_responses_path?: string
  embeddings_path?: string
  rerank_path?: string
  audio_speech_path?: string
  audio_transcriptions_path?: string
  models_path?: string
  extra_headers?: Record<string, string>
  supports_chat_completions?: boolean
  supports_chat_responses?: boolean
  supports_embeddings?: boolean
  supports_rerank?: boolean
  supports_audio_speech?: boolean
  supports_audio_transcriptions?: boolean
  supports_models?: boolean
  enabled?: boolean
}

export async function listAIProviders(params?: { enabled?: boolean; limit?: number; offset?: number }) {
  const res = await http.get('/ai/providers', {
    params: {
      enabled: typeof params?.enabled === 'boolean' ? params.enabled : undefined,
      limit: params?.limit || 100,
      offset: params?.offset || 0,
    },
  })
  return unwrapResponse<AIProviderListResult>(res)
}

export async function getAIProvider(model: string) {
  const res = await http.get(`/ai/providers/${encodeURIComponent(model)}`)
  return unwrapResponse<AIProviderModel>(res)
}

export async function createAIProvider(input: AIProviderInput) {
  const res = await http.post('/ai/providers', input)
  return unwrapResponse<AIProviderModel>(res)
}

export async function updateAIProvider(model: string, input: Partial<AIProviderInput>) {
  const res = await http.put(`/ai/providers/${encodeURIComponent(model)}`, input)
  return unwrapResponse<AIProviderModel>(res)
}

export async function deleteAIProvider(model: string) {
  const res = await http.delete(`/ai/providers/${encodeURIComponent(model)}`)
  return unwrapResponse<{ model: string; deleted: boolean }>(res)
}

export async function batchToggleAIProviders(models: string[], enabled: boolean) {
  const tasks = models.map(model => updateAIProvider(model, { enabled }))
  return Promise.all(tasks)
}

export async function listVendors(params?: { enabled?: boolean; limit?: number; offset?: number }) {
  const res = await http.get('/ai/providers/vendors', {
    params: {
      enabled: typeof params?.enabled === 'boolean' ? params.enabled : undefined,
      limit: params?.limit || 100,
      offset: params?.offset || 0,
    },
  })
  return unwrapResponse<AIProviderVendorListResult>(res)
}

export async function createVendor(input: AIProviderVendorInput) {
  const res = await http.post('/ai/providers/vendors', input)
  return unwrapResponse<AIProviderVendor>(res)
}

export async function updateVendorEnabled(providerId: string, enabled: boolean) {
  const res = await http.patch(`/ai/providers/vendors/${encodeURIComponent(providerId)}/enabled`, { enabled })
  return unwrapResponse<AIProviderVendor>(res)
}

export async function getVendorDetail(providerId: string, params?: { enabled?: boolean; limit?: number; offset?: number }) {
  const res = await http.get(`/ai/providers/vendors/${encodeURIComponent(providerId)}`, {
    params: {
      enabled: typeof params?.enabled === 'boolean' ? params.enabled : undefined,
      limit: params?.limit || 100,
      offset: params?.offset || 0,
    },
  })
  return unwrapResponse<VendorDetailResult>(res)
}

export async function discoverVendorModels(providerId: string) {
  const res = await http.get(`/ai/providers/vendors/${encodeURIComponent(providerId)}/discover-models`)
  return unwrapResponse<DiscoverVendorModelsResult>(res)
}

export async function createModelByVendor(providerId: string, input: CreateModelByVendorInput) {
  const res = await http.post(`/ai/providers/vendors/${encodeURIComponent(providerId)}/models`, input)
  return unwrapResponse<AIProviderModel>(res)
}

export interface TestConnectionResult {
  success: boolean
  message: string
  latency_ms?: number
  error_detail?: string
}

export async function testVendorConnection(providerId: string) {
  const res = await http.post(`/ai/providers/vendors/${encodeURIComponent(providerId)}/test-connection`)
  return unwrapResponse<TestConnectionResult>(res)
}

export async function testModelConnection(model: string) {
  const res = await http.post(`/ai/providers/${encodeURIComponent(model)}/test-connection`)
  return unwrapResponse<TestConnectionResult>(res)
}
