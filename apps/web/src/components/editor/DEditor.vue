<script setup lang="ts">
import { watch, ref, nextTick, onBeforeUnmount, onMounted } from 'vue'
import { EditorContent, useEditor } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Underline from '@tiptap/extension-underline'
import { TextStyle } from '@tiptap/extension-text-style'
import Color from '@tiptap/extension-color'
import FontFamily from '@tiptap/extension-font-family'
import Highlight from '@tiptap/extension-highlight'
import TextAlign from '@tiptap/extension-text-align'
import TaskList from '@tiptap/extension-task-list'
import TaskItem from '@tiptap/extension-task-item'
import Link from '@tiptap/extension-link'
import Image from '@tiptap/extension-image'
import { Table } from '@tiptap/extension-table'
import { TableRow } from '@tiptap/extension-table-row'
import { TableCell } from '@tiptap/extension-table-cell'
import { TableHeader } from '@tiptap/extension-table-header'
import '@/assets/editor.css'
import DEditorToolbar from './DEditorToolbar.vue'
import DEditorSlashMenu from './DEditorSlashMenu.vue'
import type { EditorTool } from './types'

interface Props {
  modelValue?: string
  modelJson?: Record<string, unknown> | null
  tools?: EditorTool[]
  readOnly?: boolean
  remoteCursors?: Array<{ clientId: number; name: string; avatarUrl?: string; color?: string; selection?: { anchor: number; head: number } }>
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '<p>开始创作吧 ✍️</p>',
  modelJson: null,
  readOnly: false,
  remoteCursors: () => [],
  tools: (): EditorTool[] => [
    'fontFamily',
    'textColor',
    'bgColor',
    'bold',
    'italic',
    'underline',
    'strike',
    'highlight',
    'textPrimary',
    'textDefault',
    'inlineCode',
    'heading1',
    'heading2',
    'heading3',
    'heading4',
    'paragraph',
    'alignLeft',
    'alignCenter',
    'alignRight',
    'alignJustify',
    'bulletList',
    'orderedList',
    'taskList',
    'blockquote',
    'codeBlock',
    'horizontalRule',
    'link',
    'image',
    'table',
    'clear',
    'undo',
    'redo',
  ],
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'update:modelJson': [value: Record<string, unknown>]
  'local-selection-change': [value: { anchor: number; head: number } | null]
}>()

const slashMenuVisible = ref(false)
const slashMenuX = ref(0)
const slashMenuY = ref(0)
const slashQuery = ref('')
const editorBodyRef = ref<HTMLElement | null>(null)

interface RemoteCursorMarker {
  clientId: number
  name: string
  avatarUrl?: string
  color: string
  top: number
  left: number
}

const remoteCursorMarkers = ref<RemoteCursorMarker[]>([])
let cursorUpdateTimer: number | null = null
let scrollContainer: HTMLElement | null = null

function clampSelectionPos(pos: number, size: number) {
  return Math.max(1, Math.min(pos, size + 1))
}

function queueRemoteCursorUpdate() {
  if (cursorUpdateTimer !== null) {
    window.cancelAnimationFrame(cursorUpdateTimer)
  }
  cursorUpdateTimer = window.requestAnimationFrame(() => {
    cursorUpdateTimer = null
    updateRemoteCursorMarkers()
  })
}

function updateRemoteCursorMarkers() {
  const currentEditor = editor.value
  const body = editorBodyRef.value
  const scroller = scrollContainer
  if (!currentEditor || !body || !scroller) {
    remoteCursorMarkers.value = []
    return
  }

  const containerRect = body.getBoundingClientRect()
  const docSize = currentEditor.state.doc.content.size
  const nextMarkers: RemoteCursorMarker[] = []

  props.remoteCursors.forEach((cursor) => {
    const selection = cursor.selection
    if (!selection) return

    const pos = clampSelectionPos(selection.head, docSize)
    try {
      const coords = currentEditor.view.coordsAtPos(pos)
      nextMarkers.push({
        clientId: cursor.clientId,
        name: cursor.name,
        avatarUrl: cursor.avatarUrl,
        color: cursor.color || '#3b82f6',
        top: coords.top - containerRect.top + scroller.scrollTop,
        left: coords.left - containerRect.left + scroller.scrollLeft,
      })
    } catch {
      // Ignore invalid selection positions from stale awareness state.
    }
  })

  remoteCursorMarkers.value = nextMarkers
}

function emitLocalSelection() {
  const currentEditor = editor.value
  if (!currentEditor) {
    emit('local-selection-change', null)
    return
  }
  const { anchor, head } = currentEditor.state.selection
  emit('local-selection-change', { anchor, head })
}

function detachScrollListener() {
  if (!scrollContainer) return
  scrollContainer.removeEventListener('scroll', queueRemoteCursorUpdate)
  scrollContainer = null
}

function attachScrollListener() {
  detachScrollListener()
  const body = editorBodyRef.value
  if (!body) return
  scrollContainer = body.querySelector('.d-editor-content') as HTMLElement | null
  scrollContainer?.addEventListener('scroll', queueRemoteCursorUpdate, { passive: true })
}

const editor = useEditor({
  extensions: [
    StarterKit,
    Underline,
    TextStyle,
    FontFamily,
    Color,
    Highlight.configure({ multicolor: true }),
    TextAlign.configure({
      types: ['heading', 'paragraph'],
      defaultAlignment: 'left',
    }),
    TaskList,
    TaskItem.configure({ nested: true }),
    Link.configure({
      openOnClick: false,
      autolink: true,
      linkOnPaste: true,
    }),
    Image.configure({
      allowBase64: true,
    }),
    Table.configure({
      resizable: false,
    }),
    TableRow,
    TableCell,
    TableHeader,
  ],
  content: props.modelJson ?? props.modelValue,
  editable: !props.readOnly,
  onUpdate: ({ editor: currentEditor }) => {
    emit('update:modelValue', currentEditor.getHTML())
    emit('update:modelJson', currentEditor.getJSON() as Record<string, unknown>)

    const { $from } = currentEditor.state.selection
    const text = currentEditor.state.doc.textBetween(Math.max(0, $from.pos - 50), $from.pos)
    const match = text.match(/\/(\w*)$/)

    if (match) {
      slashQuery.value = match[1] || ''
      slashMenuVisible.value = true
      const coords = currentEditor.view.coordsAtPos($from.pos)
      slashMenuX.value = coords.left
      slashMenuY.value = coords.top
    } else {
      slashMenuVisible.value = false
    }

    queueRemoteCursorUpdate()
  },
  onSelectionUpdate: () => {
    emitLocalSelection()
    queueRemoteCursorUpdate()
  },
  onBlur: () => {
    emit('local-selection-change', null)
  },
})

watch(
  () => props.modelJson,
  (value) => {
    const currentEditor = editor.value
    if (!currentEditor || !value) return
    if (JSON.stringify(value) === JSON.stringify(currentEditor.getJSON())) return
    currentEditor.commands.setContent(value)
    queueRemoteCursorUpdate()
  },
)

watch(
  () => props.readOnly,
  (readOnly) => {
    const currentEditor = editor.value
    if (!currentEditor) return
    currentEditor.setEditable(!readOnly)
  },
)

watch(
  () => props.modelValue,
  (value) => {
    const currentEditor = editor.value
    if (!currentEditor) return
    if (value === currentEditor.getHTML()) return
    currentEditor.commands.setContent(value)
    queueRemoteCursorUpdate()
  },
)

watch(
  () => props.remoteCursors,
  () => {
    queueRemoteCursorUpdate()
  },
  { deep: true },
)

watch(
  () => editor.value,
  async (currentEditor) => {
    if (!currentEditor) return
    await nextTick()
    attachScrollListener()
    emitLocalSelection()
    queueRemoteCursorUpdate()
  },
  { immediate: true },
)

onMounted(() => {
  attachScrollListener()
  window.addEventListener('resize', queueRemoteCursorUpdate)
})

onBeforeUnmount(() => {
  if (cursorUpdateTimer !== null) {
    window.cancelAnimationFrame(cursorUpdateTimer)
    cursorUpdateTimer = null
  }
  detachScrollListener()
  window.removeEventListener('resize', queueRemoteCursorUpdate)
})
</script>

<template>
  <div class="d-editor flex h-full min-h-0 flex-col bg-base-100">
    <DEditorToolbar :editor="editor ?? null" :tools="props.tools" />

    <div ref="editorBodyRef" class="d-editor-body relative min-h-0 flex-1 overflow-hidden">
      <div class="d-editor-remote-cursor-layer" aria-hidden="true">
        <div
          v-for="marker in remoteCursorMarkers"
          :key="marker.clientId"
          class="d-editor-remote-cursor"
          :style="{
            top: `${marker.top}px`,
            left: `${marker.left}px`,
            '--cursor-color': marker.color,
          }"
        >
          <span class="d-editor-remote-cursor-line" />
          <div class="d-editor-remote-cursor-card">
            <img v-if="marker.avatarUrl" :src="marker.avatarUrl" :alt="marker.name" class="d-editor-remote-cursor-avatar" />
            <span v-else class="d-editor-remote-cursor-avatar d-editor-remote-cursor-avatar-fallback">{{ marker.name.slice(0, 1).toUpperCase() }}</span>
            <span class="d-editor-remote-cursor-name">{{ marker.name }}</span>
          </div>
        </div>
      </div>

      <EditorContent
        v-if="editor"
        :editor="editor"
        class="d-editor-content prose prose-sm md:prose-base h-full max-w-none
        prose-headings:font-title prose-headings:text-base-content
        prose-p:text-base-content prose-strong:text-base-content prose-em:text-base-content
        prose-a:text-primary prose-a:no-underline hover:prose-a:underline
        prose-code:text-secondary prose-code:before:content-none prose-code:after:content-none
        prose-pre:bg-base-200 prose-pre:text-base-content
        prose-blockquote:border-l-primary prose-blockquote:text-base-content/80
        prose-hr:border-base-300 prose-li:marker:text-primary"
      />
    </div>

    <DEditorSlashMenu
      :editor="editor ?? null"
      :visible="slashMenuVisible"
      :x="slashMenuX"
      :y="slashMenuY"
      :query="slashQuery"
      @close="slashMenuVisible = false"
    />
  </div>
</template>
