import * as Y from 'yjs'
import { Awareness, applyAwarenessUpdate, encodeAwarenessUpdate } from 'y-protocols/awareness'

const MESSAGE_SYNC = 0
const MESSAGE_AWARENESS = 1

const SYNC_STEP1 = 0
const SYNC_STEP2 = 1
const SYNC_UPDATE = 2
const COLLAB_DEBUG = import.meta.env.DEV || (import.meta.env.VITE_COLLAB_DEBUG === 'true')

function debugLog(event: string, fields?: Record<string, unknown>) {
  if (!COLLAB_DEBUG) return
  if (fields) {
    console.info('[collab][provider]', event, fields)
    return
  }
  console.info('[collab][provider]', event)
}

function encodeVarUint(value: number): Uint8Array {
  const bytes: number[] = []
  let v = value >>> 0
  while (v >= 0x80) {
    bytes.push((v & 0x7f) | 0x80)
    v >>>= 7
  }
  bytes.push(v)
  return new Uint8Array(bytes)
}

function decodeVarUint(buffer: Uint8Array, start: number): { value: number, next: number } {
  let num = 0
  let shift = 0
  let pos = start

  while (pos < buffer.length) {
    const b = buffer[pos]
    if (b === undefined) {
      throw new Error('Invalid varuint frame')
    }
    num |= (b & 0x7f) << shift
    pos += 1

    if ((b & 0x80) === 0) {
      return { value: num >>> 0, next: pos }
    }
    shift += 7
  }

  throw new Error('Invalid varuint frame')
}

function concatBytes(...parts: Uint8Array[]): Uint8Array {
  const total = parts.reduce((acc, part) => acc + part.length, 0)
  const out = new Uint8Array(total)
  let offset = 0
  for (const part of parts) {
    out.set(part, offset)
    offset += part.length
  }
  return out
}

function buildSyncFrame(subtype: number, payload: Uint8Array): Uint8Array {
  return concatBytes(encodeVarUint(MESSAGE_SYNC), encodeVarUint(subtype), payload)
}

function buildAwarenessFrame(payload: Uint8Array): Uint8Array {
  return concatBytes(encodeVarUint(MESSAGE_AWARENESS), payload)
}

interface GatewayYjsProviderHooks {
  onOpen?: () => void
  onClose?: () => void
  onError?: (error: Event) => void
}

export class GatewayYjsProvider {
  public readonly doc: Y.Doc
  public readonly awareness: Awareness

  private ws: WebSocket | null = null
  private readonly wsUrl: string
  private readonly name: string
  private readonly hooks?: GatewayYjsProviderHooks
  private connected = false
  private awarenessHeartbeatTimer: number | null = null

  constructor(doc: Y.Doc, wsUrl: string, name: string, awareness?: Awareness, hooks?: GatewayYjsProviderHooks) {
    this.doc = doc
    this.wsUrl = wsUrl
    this.name = name
    this.awareness = awareness ?? new Awareness(doc)
    this.hooks = hooks

    this.doc.on('update', (update: Uint8Array, origin: unknown) => {
      if (origin === this || !this.connected) {
        return
      }
      this.send(buildSyncFrame(SYNC_UPDATE, update))
    })

    this.awareness.on('update', ({ added, updated, removed }: { added: number[]; updated: number[]; removed: number[] }, origin: unknown) => {
      if (origin === this || !this.connected) {
        return
      }
      const clients = [...added, ...updated, ...removed]
      const encoded = encodeAwarenessUpdate(this.awareness, clients)
      this.send(buildAwarenessFrame(encoded))
    })
  }

  connect(): Promise<void> {
    if (this.connected) {
      return Promise.resolve()
    }

    return new Promise((resolve, reject) => {
      let settled = false

      this.ws = new WebSocket(this.wsUrl)
      this.ws.binaryType = 'arraybuffer'
      debugLog('ws_connect_start', { wsUrl: this.wsUrl, clientId: this.doc.clientID, name: this.name })

      this.ws.onopen = () => {
        this.connected = true
        debugLog('ws_open', { wsUrl: this.wsUrl, clientId: this.doc.clientID })
        this.awareness.setLocalStateField('user', {
          name: this.name,
        })
        this.startAwarenessHeartbeat()
        this.send(buildSyncFrame(SYNC_STEP1, new Uint8Array()))
        this.hooks?.onOpen?.()

        if (!settled) {
          settled = true
          resolve()
        }
      }

      this.ws.onmessage = (event: MessageEvent<ArrayBuffer>) => {
        const frame = new Uint8Array(event.data)
        debugLog('ws_message', { bytes: frame.length })
        this.handleFrame(frame)
      }

      this.ws.onclose = () => {
        this.connected = false
        this.stopAwarenessHeartbeat()
        debugLog('ws_close', { wsUrl: this.wsUrl, clientId: this.doc.clientID })
        this.hooks?.onClose?.()

        if (!settled) {
          settled = true
          reject(new Error('Collaboration connection closed before ready'))
        }
      }

      this.ws.onerror = (error) => {
        this.connected = false
        this.stopAwarenessHeartbeat()
        debugLog('ws_error', { wsUrl: this.wsUrl, clientId: this.doc.clientID })
        this.hooks?.onError?.(error)

        if (!settled) {
          settled = true
          reject(new Error('Failed to establish collaboration connection'))
        }
      }
    })
  }

  disconnect() {
    this.stopAwarenessHeartbeat()
    this.awareness.setLocalState(null)
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.connected = false
  }

  setLocalSelection(anchor: number, head: number) {
    this.awareness.setLocalStateField('selection', { anchor, head })
  }

  private send(payload: Uint8Array) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      debugLog('ws_send_dropped', {
        reason: !this.ws ? 'missing_socket' : `state_${this.ws.readyState}`,
        bytes: payload.length,
      })
      return
    }
    let msgType = -1
    let subtype = -1
    try {
      const decoded = decodeVarUint(payload, 0)
      msgType = decoded.value
      if (msgType === MESSAGE_SYNC) {
        subtype = decodeVarUint(payload, decoded.next).value
      }
    } catch {
      // Ignore decode failure in diagnostics.
    }
    debugLog('ws_send', { msgType, subtype, bytes: payload.length, clientId: this.doc.clientID })
    this.ws.send(payload)
  }

  private startAwarenessHeartbeat() {
    this.stopAwarenessHeartbeat()
    this.awarenessHeartbeatTimer = window.setInterval(() => {
      if (!this.connected) {
        return
      }
      const localState = this.awareness.getLocalState()
      if (!localState) {
        return
      }
      const encoded = encodeAwarenessUpdate(this.awareness, [this.doc.clientID])
      this.send(buildAwarenessFrame(encoded))
    }, 15000)
  }

  private stopAwarenessHeartbeat() {
    if (this.awarenessHeartbeatTimer !== null) {
      window.clearInterval(this.awarenessHeartbeatTimer)
      this.awarenessHeartbeatTimer = null
    }
  }

  private handleFrame(frame: Uint8Array) {
    let idx = 0
    const msg = decodeVarUint(frame, idx)
    idx = msg.next
    debugLog('ws_frame', { msgType: msg.value, bytes: frame.length, clientId: this.doc.clientID })

    if (msg.value === MESSAGE_SYNC) {
      const syncType = decodeVarUint(frame, idx)
      idx = syncType.next
      const payload = frame.slice(idx)
      debugLog('ws_sync_frame', { syncType: syncType.value, payloadBytes: payload.length, clientId: this.doc.clientID })

      if (syncType.value === SYNC_STEP1) {
        const fullState = Y.encodeStateAsUpdate(this.doc)
        this.send(buildSyncFrame(SYNC_STEP2, fullState))
        return
      }

      if (syncType.value === SYNC_STEP2 || syncType.value === SYNC_UPDATE) {
        if (payload.length > 0) {
          Y.applyUpdate(this.doc, payload, this)
        }
      }
      return
    }

    if (msg.value === MESSAGE_AWARENESS) {
      const payload = frame.slice(idx)
      debugLog('ws_awareness_frame', { payloadBytes: payload.length, clientId: this.doc.clientID })
      applyAwarenessUpdate(this.awareness, payload, this)
    }
  }
}
