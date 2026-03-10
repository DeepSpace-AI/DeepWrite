<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { EditorContent, useEditor } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Collaboration from '@tiptap/extension-collaboration'
import * as Y from 'yjs'
import { Awareness } from 'y-protocols/awareness'
import { GatewayYjsProvider } from '@/collab/gatewayProvider'
import '@/assets/editor.css'

interface IssueTokenResponse {
  token: string
  read_only: boolean
}

const route = useRoute()
const documentId = computed(() => String(route.params.id || '').trim())

const gatewayBase = (import.meta.env.VITE_GATEWAY_BASE_URL as string | undefined)?.trim() || 'http://localhost:8080'
const accessToken = ref(localStorage.getItem('deepwrite_access_token') || '')
const displayName = ref(localStorage.getItem('deepwrite_display_name') || `User-${Math.floor(Math.random() * 1000)}`)
const title = ref('')
const status = ref('未连接')
const readOnly = ref(false)
const connectionError = ref('')
const isConnecting = ref(false)

const ydoc = new Y.Doc()
const awareness = new Awareness(ydoc)
const provider = ref<GatewayYjsProvider | null>(null)
const awarenessVersion = ref(0)
let localUpdateCount = 0
let syncTimer: number | null = null

const handleAwarenessUpdate = () => {
  awarenessVersion.value += 1
}

const handleDocUpdate = (_update: Uint8Array, origin: unknown) => {
  if (origin === provider.value) {
    return
  }
  localUpdateCount += 1
  if (localUpdateCount >= 50) {
    void syncContentIfNeeded(true)
  }
}

const editor = useEditor({
  editable: true,
  extensions: [
    StarterKit,
    Collaboration.configure({
      document: ydoc,
    }),
  ],
  onSelectionUpdate: ({ editor: currentEditor }) => {
    if (!provider.value) return
    const { from, to } = currentEditor.state.selection
    provider.value.setLocalSelection(from, to)
  },
})

const members = computed(() => {
  awarenessVersion.value

  const currentProvider = provider.value
  if (!currentProvider) {
    return []
  }

  const rows: Array<{ id: number, name: string, selection: string }> = []
  currentProvider.awareness.getStates().forEach((state, clientID) => {
    const user = state.user as { name?: string } | undefined
    const selection = state.selection as { anchor?: number, head?: number } | undefined
    rows.push({
      id: Number(clientID),
      name: user?.name || `User-${clientID}`,
      selection: selection ? `${selection.anchor ?? 0}-${selection.head ?? 0}` : '无',
    })
  })

  return rows.sort((a, b) => a.id - b.id)
})

function toWsBase(httpBase: string): string {
  if (httpBase.startsWith('https://')) {
    return `wss://${httpBase.slice('https://'.length)}`
  }
  if (httpBase.startsWith('http://')) {
    return `ws://${httpBase.slice('http://'.length)}`
  }
  return httpBase
}

async function apiFetch(path: string, options: RequestInit = {}) {
  const headers = new Headers(options.headers || {})
  headers.set('Content-Type', 'application/json')
  if (accessToken.value.trim()) {
    headers.set('Authorization', `Bearer ${accessToken.value.trim()}`)
  }

  const res = await fetch(`${gatewayBase}${path}`, {
    ...options,
    headers,
  })
  const payload = await res.json()
  if (!res.ok || payload.code >= 400) {
    throw new Error(payload.message || '请求失败')
  }
  return payload.data
}

async function fetchDocument() {
  const doc = await apiFetch(`/api/v1/documents/${documentId.value}`)
  title.value = doc.title || ''
  return doc
}

async function issueToken(): Promise<IssueTokenResponse> {
  return apiFetch(`/api/v1/documents/${documentId.value}/collab-token`, {
    method: 'POST',
    body: '{}',
  })
}

async function syncContentIfNeeded(force = false) {
  const currentEditor = editor.value
  if (!currentEditor) return
  if (!force && localUpdateCount < 50) return

  const content = currentEditor.getJSON()
  localUpdateCount = 0

  await apiFetch(`/api/v1/documents/${documentId.value}/collab/content`, {
    method: 'POST',
    body: JSON.stringify({
      title: title.value,
      content_json: content,
    }),
  })
}

async function connectCollab() {
  if (!documentId.value) {
    connectionError.value = '缺少文档 ID'
    return
  }

  isConnecting.value = true
  status.value = '连接中...'
  connectionError.value = ''
  localStorage.setItem('deepwrite_access_token', accessToken.value)
  localStorage.setItem('deepwrite_display_name', displayName.value)

  try {
    const doc = await fetchDocument()
    if (doc.content_json && editor.value) {
      editor.value.commands.setContent(doc.content_json)
    }
    const token = await issueToken()
    readOnly.value = Boolean(token.read_only)
    editor.value?.setEditable(!readOnly.value)

    if (provider.value) {
      provider.value.awareness.off('update', handleAwarenessUpdate)
      provider.value.disconnect()
    }
    const wsUrl = `${toWsBase(gatewayBase)}/api/v1/documents/${documentId.value}/collab/ws?token=${encodeURIComponent(token.token)}&name=${encodeURIComponent(displayName.value)}`
    const nextProvider = new GatewayYjsProvider(ydoc, wsUrl, displayName.value, awareness)
    nextProvider.awareness.on('update', handleAwarenessUpdate)
    provider.value = nextProvider
    awarenessVersion.value += 1
    nextProvider.connect()

    if (syncTimer) {
      window.clearInterval(syncTimer)
    }
    syncTimer = window.setInterval(() => {
      void syncContentIfNeeded(true)
    }, 3000)

    status.value = readOnly.value ? '已连接（只读）' : '已连接（可编辑）'
  } catch (err) {
    connectionError.value = (err as Error).message
    status.value = '连接失败'
  } finally {
    isConnecting.value = false
  }
}

onMounted(() => {
  ydoc.on('update', handleDocUpdate)
  void connectCollab()
})

onBeforeUnmount(() => {
  if (syncTimer) {
    window.clearInterval(syncTimer)
  }
  ydoc.off('update', handleDocUpdate)
  if (provider.value) {
    provider.value.awareness.off('update', handleAwarenessUpdate)
    provider.value.disconnect()
  }
  provider.value = null
})
</script>

<template>
  <main class="mx-auto max-w-7xl p-4 md:p-8">
    <section class="mb-4 rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm">
      <div class="grid gap-3 md:grid-cols-4">
        <label class="form-control">
          <span class="label-text">文档 ID</span>
          <input :value="documentId" disabled class="input input-bordered" />
        </label>

        <label class="form-control md:col-span-2">
          <span class="label-text">Access Token（Bearer）</span>
          <input v-model="accessToken" class="input input-bordered" placeholder="粘贴登录得到的 access token" />
        </label>

        <label class="form-control">
          <span class="label-text">显示名</span>
          <input v-model="displayName" class="input input-bordered" />
        </label>
      </div>

      <div class="mt-3 flex flex-wrap items-center gap-2">
        <button class="btn btn-primary" :disabled="isConnecting" @click="connectCollab">重新连接</button>
        <span class="badge badge-outline">{{ status }}</span>
        <span v-if="readOnly" class="badge badge-warning">viewer 只读</span>
        <span v-if="connectionError" class="text-error">{{ connectionError }}</span>
      </div>
    </section>

    <section class="grid gap-4 lg:grid-cols-[1fr_280px]">
      <article class="rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm">
        <h1 class="mb-3 text-xl font-semibold">{{ title || '协作文档' }}</h1>
        <EditorContent :editor="editor" class="d-editor-content prose max-w-none min-h-105" />
      </article>

      <aside class="rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm">
        <h2 class="mb-3 font-semibold">在线成员与选区</h2>
        <ul class="space-y-2 text-sm">
          <li v-for="member in members" :key="member.id" class="rounded-lg border border-base-300 p-2">
            <div class="font-medium">{{ member.name }}</div>
            <div class="text-base-content/70">client: {{ member.id }}</div>
            <div class="text-base-content/70">selection: {{ member.selection }}</div>
          </li>
          <li v-if="members.length === 0" class="text-base-content/70">暂无在线成员</li>
        </ul>
      </aside>
    </section>
  </main>
</template>
