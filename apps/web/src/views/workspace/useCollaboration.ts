import { ref, reactive, computed, onBeforeUnmount } from 'vue'
import * as Y from 'yjs'
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

  const isReady = computed(() => state.isConnected && yjsDoc.value !== null && provider.value !== null)

  /**
   * Initialize and connect to collaboration
   */
  async function connect() {
    if (state.isConnecting || state.isConnected) {
      return
    }

    try {
      state.isConnecting = true
      state.error = null

      // Fetch collab token from backend
      const tokenData = await getCollabToken(documentId)
      state.readOnly = tokenData.read_only

      // Determine WebSocket URL
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const host = window.location.host
      const wsUrl = `${protocol}//${host}${tokenData.ws_path}/${documentId}/collab/ws?token=${tokenData.token}&name=${encodeURIComponent(userName)}`

      // Create Yjs doc and provider
      const doc = new Y.Doc()
      const prov = new GatewayYjsProvider(doc, wsUrl, userName)

      // Create shared text for content
      const text = doc.getText('content')

      // Connect
      prov.connect()

      yjsDoc.value = doc
      provider.value = prov
      yText.value = text
      state.isConnected = true
      state.isConnecting = false

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

      // Cleanup on disconnect
      const originalDisconnect = prov.disconnect.bind(prov)
      prov.disconnect = () => {
        awareness.off('change', updateRemoteUsers)
        originalDisconnect()
      }
    } catch (err) {
      state.error = err instanceof Error ? err.message : 'Failed to connect to collaboration'
      state.isConnecting = false
      console.error('[useCollaboration] Connection error:', err)
    }
  }

  /**
   * Disconnect from collaboration
   */
  function disconnect() {
    if (provider.value) {
      provider.value.disconnect()
      provider.value = null
    }
    if (yjsDoc.value) {
      yjsDoc.value.destroy()
      yjsDoc.value = null
    }
    yText.value = null
    state.isConnected = false
    state.error = null
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
