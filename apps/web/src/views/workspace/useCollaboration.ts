import { ref, reactive, computed, onBeforeUnmount } from 'vue'
import * as Y from 'yjs'
import { gatewayBase } from '@/api/http'
import { getCollabToken } from '@/api/document'
import { GatewayYjsProvider } from '@/collab/gatewayProvider'

export interface RemoteUserState {
  clientId: number
  name: string
  color?: string
}

export interface CollabState {
  isConnecting: boolean
  isConnected: boolean
  error: string | null
  readOnly: boolean
}

const RECONNECT_BASE_DELAY_MS = 1000
const RECONNECT_MAX_DELAY_MS = 15000
const RECONNECT_MAX_ATTEMPTS = 8

export function useCollaboration(documentId: string, userName: string) {
  // Yjs document and provider
  const yjsDoc = ref<Y.Doc | null>(null)
  const provider = ref<GatewayYjsProvider | null>(null)
  const yText = ref<Y.Text | null>(null)

  // Remote users state
  const remoteUsers = ref<RemoteUserState[]>([])

  // Connection state
  const state = reactive<CollabState>({
    isConnecting: false,
    isConnected: false,
    error: null,
    readOnly: false,
  })

  let disconnectRequested = false
  let cleanupAwareness: (() => void) | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectAttempts = 0
  let connectSequence = 0

  const isReady = computed(() => state.isConnected && yjsDoc.value !== null && provider.value !== null)

  function clearReconnectTimer() {
    if (!reconnectTimer) return
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }

  function resetReconnectState() {
    clearReconnectTimer()
    reconnectAttempts = 0
  }

  function teardownConnection({ destroyDoc }: { destroyDoc: boolean }) {
    cleanupAwareness?.()
    cleanupAwareness = null

    if (provider.value) {
      provider.value.disconnect()
      provider.value = null
    }

    yText.value = null
    remoteUsers.value = []
    state.isConnected = false
    state.isConnecting = false

    if (destroyDoc && yjsDoc.value) {
      yjsDoc.value.destroy()
      yjsDoc.value = null
    }
  }

  function scheduleReconnect(reason: string) {
    if (disconnectRequested || reconnectTimer || state.isConnected || state.isConnecting) {
      return
    }
    if (reconnectAttempts >= RECONNECT_MAX_ATTEMPTS) {
      state.error = 'Collaboration disconnected. Auto-reconnect limit reached.'
      return
    }

    const delay = Math.min(
      RECONNECT_MAX_DELAY_MS,
      RECONNECT_BASE_DELAY_MS * (2 ** reconnectAttempts),
    )
    reconnectAttempts += 1
    state.error = `${reason}. Reconnecting in ${Math.round(delay / 1000)}s...`

    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      void connect({ isReconnect: true })
    }, delay)
  }

  /**
   * Initialize and connect to collaboration
   */
  async function connect(options?: { isReconnect?: boolean }) {
    const isReconnect = options?.isReconnect === true

    if (state.isConnecting || state.isConnected) {
      return
    }

    try {
      if (!isReconnect) {
        disconnectRequested = false
        resetReconnectState()
      }

      state.isConnecting = true
      state.error = null
      const thisConnectSequence = ++connectSequence

      if (isReconnect && provider.value) {
        teardownConnection({ destroyDoc: false })
      }

      // Fetch collab token from backend
      const tokenData = await getCollabToken(documentId)
      state.readOnly = tokenData.read_only

      // Determine WebSocket URL
      const wsPath = tokenData.ws_path.replace(':id', encodeURIComponent(documentId))
      const gatewayUrl = new URL(gatewayBase)
      gatewayUrl.protocol = gatewayUrl.protocol === 'https:' ? 'wss:' : 'ws:'
      gatewayUrl.pathname = wsPath
      gatewayUrl.search = new URLSearchParams({
        token: tokenData.token,
      }).toString()
      const wsUrl = gatewayUrl.toString()

      // Create Yjs doc and provider
      const doc = yjsDoc.value ?? new Y.Doc()
      const prov = new GatewayYjsProvider(doc, wsUrl, userName, undefined, {
        onOpen: () => {
          if (thisConnectSequence !== connectSequence) {
            return
          }
          state.isConnected = true
          state.isConnecting = false
          state.error = null
          resetReconnectState()
        },
        onClose: () => {
          if (thisConnectSequence !== connectSequence) {
            return
          }
          state.isConnected = false
          state.isConnecting = false
          remoteUsers.value = []
          if (!disconnectRequested) {
            scheduleReconnect('Collaboration connection closed')
          }
        },
        onError: () => {
          if (thisConnectSequence !== connectSequence) {
            return
          }
          state.isConnected = false
          state.isConnecting = false
          if (!disconnectRequested) {
            scheduleReconnect('Collaboration connection error')
          }
        },
      })

      // Create shared text for content
      const text = doc.getText('content')

      yjsDoc.value = doc
      provider.value = prov
      yText.value = text

      // Connect
      await prov.connect()

      if (thisConnectSequence !== connectSequence) {
        prov.disconnect()
        return
      }

      // Setup awareness listener for remote users
      const awareness = prov.awareness
      const updateRemoteUsers = () => {
        const users: RemoteUserState[] = []
        const states = awareness.getStates()

        states.forEach((clientState, clientId) => {
          // Skip self
          if (clientId === doc.clientID) {
            return
          }
          const user = clientState?.user as { name?: string } | undefined
          if (user?.name) {
            users.push({
              clientId,
              name: user.name,
              color: clientState?.color as string | undefined,
            })
          }
        })

        remoteUsers.value = users
      }

      // Initial sync of awareness states
      updateRemoteUsers()

      // Listen to awareness changes
      awareness.on('change', updateRemoteUsers)

      cleanupAwareness = () => {
        awareness.off('change', updateRemoteUsers)
      }
    } catch (err) {
      if (disconnectRequested) {
        return
      }
      state.error = err instanceof Error ? err.message : 'Failed to connect to collaboration'
      state.isConnected = false
      state.isConnecting = false
      teardownConnection({ destroyDoc: false })
      scheduleReconnect('Failed to connect to collaboration')
      console.error('[useCollaboration] Connection error:', err)
    }
  }

  /**
   * Disconnect from collaboration
   */
  function disconnect() {
    disconnectRequested = true
    clearReconnectTimer()
    connectSequence += 1
    teardownConnection({ destroyDoc: true })
    state.error = null
    reconnectAttempts = 0
  }

  /**
   * Cleanup on unmount
   */
  onBeforeUnmount(() => {
    disconnect()
  })

  return {
    // State
    state,
    isReady,
    remoteUsers,

    // Resources
    yjsDoc,
    provider,
    yText,

    // Methods
    connect,
    disconnect,
  }
}
