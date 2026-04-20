import { ref, onUnmounted } from 'vue'
import { gatewayBase } from '@/api/http'
import { cacheManager } from '@/utils/cache'

export interface SessionEvent {
  session_id: string
  type: 'title_updated' | 'message_created' | 'session_archived'
  data: Record<string, any>
}

export function useSessionEvents(onEvent: (event: SessionEvent) => void) {
  const isConnected = ref(false)
  let eventSource: EventSource | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null

  function connect() {
    if (eventSource) {
      eventSource.close()
    }

    const token = cacheManager.get<string>('deepwrite_access_token')
    if (!token) {
      console.warn('[SSE] No access token available')
      return
    }

    const url = `${gatewayBase}/api/v1/events?token=${encodeURIComponent(token)}`
    
    try {
      eventSource = new EventSource(url)

      eventSource.onopen = () => {
        console.log('[SSE] Connected')
        isConnected.value = true
      }

      eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as SessionEvent
          console.log('[SSE] Event received:', data)
          onEvent(data)
        } catch (e) {
          console.warn('[SSE] Failed to parse event:', e)
        }
      }

      eventSource.onerror = (e) => {
        console.error('[SSE] Error:', e)
        isConnected.value = false
        eventSource?.close()
        eventSource = null
        scheduleReconnect()
      }
    } catch (e) {
      console.error('[SSE] Failed to connect:', e)
      scheduleReconnect()
    }
  }

  function scheduleReconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
    }
    reconnectTimer = setTimeout(() => {
      console.log('[SSE] Attempting to reconnect...')
      connect()
    }, 5000)
  }

  function disconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
    isConnected.value = false
  }

  onUnmounted(() => {
    disconnect()
  })

  return {
    isConnected,
    connect,
    disconnect,
  }
}