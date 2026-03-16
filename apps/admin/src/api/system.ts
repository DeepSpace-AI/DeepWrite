import { gatewayBase, http, unwrapResponse } from '@/api/http'
import { cacheManager, CACHE_TTL } from '@/utils/cache'

const CACHE_KEY_SYSTEM_SETTINGS = 'deepwrite_admin_system_settings'

export interface SystemSettings {
  appName: string
  appEnv: string
  defaultProvider: string
  strictAdminMode: boolean
  maintenanceMode: boolean
}

export interface GatewayHealth {
  message: string
}

export type RuntimeServiceKey = 'gateway' | 'web' | 'ai' | 'worker'

export interface RuntimeServiceStatus {
  key: RuntimeServiceKey
  name: string
  status: 'online' | 'offline' | 'unknown'
  detail: string
  endpoint: string
}

const defaultSettings: SystemSettings = {
  appName: 'DeepWrite Admin',
  appEnv: 'production',
  defaultProvider: 'openai-compatible',
  strictAdminMode: true,
  maintenanceMode: false,
}

export function getCachedSystemSettings(): SystemSettings {
  return cacheManager.get<SystemSettings>(CACHE_KEY_SYSTEM_SETTINGS) || defaultSettings
}

export function saveSystemSettings(settings: SystemSettings) {
  cacheManager.set(CACHE_KEY_SYSTEM_SETTINGS, settings, CACHE_TTL.SEVEN_DAYS)
}

async function probeEndpoint(url: string, timeoutMs = 5000): Promise<{ status: 'online' | 'offline'; detail: string }> {
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), timeoutMs)
  try {
    const res = await fetch(url, { signal: controller.signal })
    if (!res.ok) {
      return { status: 'offline', detail: `HTTP ${res.status}` }
    }
    return { status: 'online', detail: 'ok' }
  } catch (error) {
    if (error instanceof Error && error.name === 'AbortError') {
      return { status: 'offline', detail: 'timeout' }
    }
    return { status: 'offline', detail: 'unreachable' }
  } finally {
    clearTimeout(timer)
  }
}

export async function fetchGatewayHealth(): Promise<GatewayHealth> {
  const url = new URL('/api/v1/', gatewayBase).toString()
  const res = await fetch(url)
  if (!res.ok) {
    return { message: `Gateway ${res.status}` }
  }
  const json = await res.json() as { data?: { message?: string } }
  return { message: json.data?.message || 'ok' }
}

export async function fetchCurrentAdminProfile(): Promise<{ id: string; email: string; role: string }> {
  const res = await http.get('/auth/me')
  return unwrapResponse<{ id: string; email: string; role: string }>(res)
}

export async function fetchRuntimeStatuses(): Promise<RuntimeServiceStatus[]> {
  const gatewayUrl = new URL('/api/v1/', gatewayBase).toString()
  const webUrl = window.location.origin
  const aiBase = ((import.meta.env.VITE_AI_BASE_URL as string | undefined)?.trim() || '').replace(/\/+$/, '')
  const workerBase = ((import.meta.env.VITE_WORKER_BASE_URL as string | undefined)?.trim() || '').replace(/\/+$/, '')
  const aiUrl = aiBase ? `${aiBase}/health` : ''
  const workerUrl = workerBase ? `${workerBase}/health` : ''

  const [gatewayProbe, webProbe, aiProbe, workerProbe] = await Promise.all([
    probeEndpoint(gatewayUrl),
    probeEndpoint(webUrl),
    aiUrl ? probeEndpoint(aiUrl) : Promise.resolve({ status: 'unknown' as const, detail: '未配置 VITE_AI_BASE_URL' }),
    workerUrl ? probeEndpoint(workerUrl) : Promise.resolve({ status: 'unknown' as const, detail: '未配置 VITE_WORKER_BASE_URL' }),
  ])

  return [
    { key: 'gateway', name: 'Gateway', status: gatewayProbe.status, detail: gatewayProbe.detail, endpoint: gatewayUrl },
    { key: 'web', name: 'Web', status: webProbe.status, detail: webProbe.detail, endpoint: webUrl },
    { key: 'ai', name: 'AI', status: aiProbe.status, detail: aiProbe.detail, endpoint: aiUrl || '-' },
    { key: 'worker', name: 'Worker', status: workerProbe.status, detail: workerProbe.detail, endpoint: workerUrl || '-' },
  ]
}
