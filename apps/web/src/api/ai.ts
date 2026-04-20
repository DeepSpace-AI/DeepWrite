import { http, unwrapResponse, gatewayBase } from './http'

export interface ProviderModel {
  id: string
  model: string
  request_model: string
  provider_name: string
  base_url: string
  chat_completions_path: string
  api_key: string
  enabled: boolean
  max_tokens: number
  organization: string
  created_at: string
  updated_at: string
}

export interface ProviderVendor {
  id: string
  name: string
  base_url: string
  api_key: string
  enabled: boolean
  models: ProviderModel[]
  created_at: string
  updated_at: string
}

export async function listProviderVendors() {
  const res = await http.get('/ai/providers/vendors')
  return unwrapResponse<{ items: ProviderVendor[] }>(res)
}

export async function getProviderVendor(vendorId: string) {
  const res = await http.get(`/ai/providers/vendors/${encodeURIComponent(vendorId)}`)
  return unwrapResponse<ProviderVendor>(res)
}

export async function listEnabledModels() {
  const res = await http.get('/ai/providers', { params: { enabled: true } })
  return unwrapResponse<{ items: ProviderModel[] }>(res)
}

export async function testModelConnection(model: string) {
  const res = await http.post(`/ai/providers/${encodeURIComponent(model)}/test-connection`)
  return unwrapResponse<{ success: boolean; message: string }>(res)
}