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
import DEditorFloatingToolbar from './DEditorFloatingToolbar.vue'
import DEditorOutline from './DEditorOutline.vue'
import DEditorFooter from './DEditorFooter.vue'
import DEditorFindReplace from './DEditorFindReplace.vue'
import DEditorVersionHistory from './DEditorVersionHistory.vue'
import { buildKeyboardShortcuts } from './useKeyboardShortcuts'
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

const floatingToolbarVisible = ref(false)
const floatingToolbarX = ref(0)
const floatingToolbarY = ref(0)
let hideFloatingToolbarTimer: number | null = null

const outlineVisible = ref(false)
const findReplaceVisible = ref(false)
const versionHistoryVisible = ref(false)

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
const editorContainerRef: HTMLElement | null = null

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

function handleKeyDown(event: KeyboardEvent) {
  const currentEditor = editor.value
  if (!currentEditor) return

  const isMac = /Mac|iPod|iPhone|iPad/.test(navigator.platform)
  const mod = isMac ? event.metaKey : event.ctrlKey

  if (mod && event.key === '\\') {
    event.preventDefault()
    outlineVisible.value = !outlineVisible.value
    return
  }

  if (mod && event.key.toLowerCase() === 'f') {
    event.preventDefault()
    findReplaceVisible.value = !findReplaceVisible.value
    if (findReplaceVisible.value) {
      outlineVisible.value = false
    }
    return
  }

  if (mod && event.shiftKey && event.key.toLowerCase() === 's') {
    event.preventDefault()
    versionHistoryVisible.value = !versionHistoryVisible.value
    if (versionHistoryVisible.value) {
      outlineVisible.value = false
      findReplaceVisible.value = false
    }
    return
  }

  const shortcuts = buildKeyboardShortcuts(currentEditor)

  for (const shortcut of shortcuts) {
    const modifiers = shortcut.modifiers || []
    const needsMod = modifiers.includes('ctrl') || modifiers.includes('meta')
    const needsAlt = modifiers.includes('alt')
    const needsShift = modifiers.includes('shift')

    if (shortcut.key.toLowerCase() === event.key.toLowerCase() || shortcut.key === event.key) {
      const modMatch = !needsMod || mod
      const altMatch = !needsAlt || event.altKey
      const shiftMatch = !needsShift || event.shiftKey

      if (modMatch && altMatch && shiftMatch) {
        event.preventDefault()
        shortcut.action(currentEditor)
        return
      }
    }
  }
}

function attachKeyboardListener() {
  if (editorContainerRef) {
    editorContainerRef.addEventListener('keydown', handleKeyDown)
  }
}

function detachKeyboardListener() {
  if (editorContainerRef) {
    editorContainerRef.removeEventListener('keydown', handleKeyDown)
  }
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

    const currentEditor = editor.value
    if (!currentEditor) return

    const { anchor, head } = currentEditor.state.selection
    const hasSelection = anchor !== head

    if (hasSelection) {
      try {
        const coords = currentEditor.view.coordsAtPos(head)
        floatingToolbarX.value = coords.left
        floatingToolbarY.value = coords.top - 48
        floatingToolbarVisible.value = true
        slashMenuVisible.value = false
      } catch {
        floatingToolbarVisible.value = false
      }
    } else {
      if (hideFloatingToolbarTimer) {
        clearTimeout(hideFloatingToolbarTimer)
      }
      hideFloatingToolbarTimer = window.setTimeout(() => {
        floatingToolbarVisible.value = false
      }, 200)
    }
  },
  onBlur: () => {
    emit('local-selection-change', null)
    if (hideFloatingToolbarTimer) {
      clearTimeout(hideFloatingToolbarTimer)
    }
    hideFloatingToolbarTimer = window.setTimeout(() => {
      floatingToolbarVisible.value = false
    }, 200)
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
    attachKeyboardListener()
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
  detachKeyboardListener()
  window.removeEventListener('resize', queueRemoteCursorUpdate)
})
</script>

<template>
  <div ref="editorContainerRef" class="d-editor flex h-full min-h-0 flex-col bg-[var(--surface-base)]">
    <DEditorToolbar :editor="editor ?? null" :tools="props.tools" />

    <div ref="editorBodyRef" class="d-editor-body relative min-h-0 flex-1 overflow-hidden">
      <DEditorOutline
        :editor="editor ?? null"
        :visible="outlineVisible"
        class="absolute left-4 top-4 z-20"
        @close="outlineVisible = false"
      />

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
        prose-headings:font-title prose-headings:text-[var(--text-primary)]
        prose-p:text-[var(--text-primary)] prose-strong:text-[var(--text-primary)] prose-em:text-[var(--text-primary)]
        prose-a:text-primary prose-a:no-underline hover:prose-a:underline
        prose-code:text-secondary prose-code:before:content-none prose-code:after:content-none
        prose-pre:bg-[var(--surface-overlay)] prose-pre:text-[var(--text-primary)]
        prose-blockquote:border-l-primary prose-blockquote:text-[color:color-mix(in_oklab,var(--text-primary)_80%,transparent)]
        prose-hr:border-[var(--surface-sunken)] prose-li:marker:text-primary"
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

    <DEditorFloatingToolbar
      :editor="editor ?? null"
      :visible="floatingToolbarVisible"
      :x="floatingToolbarX"
      :y="floatingToolbarY"
      @close="floatingToolbarVisible = false"
    />

    <DEditorFindReplace
      :editor="editor ?? null"
      :visible="findReplaceVisible"
      @close="findReplaceVisible = false"
    />

    <DEditorVersionHistory
      :editor="editor ?? null"
      :visible="versionHistoryVisible"
      @close="versionHistoryVisible = false"
    />

    <DEditorFooter :editor="editor ?? null" />
  </div>
</template>
