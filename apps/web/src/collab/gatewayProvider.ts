import * as Y from 'yjs'
import { Awareness, applyAwarenessUpdate, encodeAwarenessUpdate } from 'y-protocols/awareness'

const MESSAGE_SYNC = 0
const MESSAGE_AWARENESS = 1

const SYNC_STEP1 = 0
const SYNC_STEP2 = 1
const SYNC_UPDATE = 2

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

export class GatewayYjsProvider {
  public readonly doc: Y.Doc
  public readonly awareness: Awareness

  private ws: WebSocket | null = null
  private readonly wsUrl: string
  private readonly name: string
  private connected = false
  private awarenessHeartbeatTimer: number | null = null

  constructor(doc: Y.Doc, wsUrl: string, name: string, awareness?: Awareness) {
    this.doc = doc
    this.wsUrl = wsUrl
    this.name = name
    this.awareness = awareness ?? new Awareness(doc)

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

  connect() {
    this.ws = new WebSocket(this.wsUrl)
    this.ws.binaryType = 'arraybuffer'

    this.ws.onopen = () => {
      this.connected = true
      this.awareness.setLocalStateField('user', {
        name: this.name,
      })
      this.startAwarenessHeartbeat()
      this.send(buildSyncFrame(SYNC_STEP1, new Uint8Array()))
    }

    this.ws.onmessage = (event: MessageEvent<ArrayBuffer>) => {
      const frame = new Uint8Array(event.data)
      this.handleFrame(frame)
    }

    this.ws.onclose = () => {
      this.connected = false
      this.stopAwarenessHeartbeat()
    }

    this.ws.onerror = () => {
      this.connected = false
      this.stopAwarenessHeartbeat()
    }
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
      return
    }
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

    if (msg.value === MESSAGE_SYNC) {
      const syncType = decodeVarUint(frame, idx)
      idx = syncType.next
      const payload = frame.slice(idx)

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
      applyAwarenessUpdate(this.awareness, payload, this)
    }
  }
}
