<script setup lang="ts">
import { watch, ref } from 'vue'
import { EditorContent, useEditor } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Underline from '@tiptap/extension-underline'
import { TextStyle } from '@tiptap/extension-text-style'
import Color from '@tiptap/extension-color'
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
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '<p>开始创作吧 ✍️</p>',
  modelJson: null,
  tools: (): EditorTool[] => [
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
}>()

// Slash menu state
const slashMenuVisible = ref(false)
const slashMenuX = ref(0)
const slashMenuY = ref(0)
const slashQuery = ref('')

const editor = useEditor({
  extensions: [
    StarterKit,
    Underline,
    TextStyle,
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
  onUpdate: ({ editor: currentEditor }) => {
    emit('update:modelValue', currentEditor.getHTML())
    emit('update:modelJson', currentEditor.getJSON() as Record<string, unknown>)

    // 检测 Slash 命令
    const { $from } = currentEditor.state.selection
    const text = currentEditor.state.doc.textBetween(Math.max(0, $from.pos - 50), $from.pos)
    const match = text.match(/\/(\w*)$/)

    if (match) {
      slashQuery.value = match[1] || ''
      slashMenuVisible.value = true

      // 获取光标位置
      const coords = currentEditor.view.coordsAtPos($from.pos)
      slashMenuX.value = coords.left
      slashMenuY.value = coords.top
    } else {
      slashMenuVisible.value = false
    }
  },
})

watch(
  () => props.modelJson,
  (value) => {
    const currentEditor = editor.value
    if (!currentEditor || !value) return
    if (JSON.stringify(value) === JSON.stringify(currentEditor.getJSON())) return
    currentEditor.commands.setContent(value)
  },
)

watch(
  () => props.modelValue,
  (value) => {
    const currentEditor = editor.value
    if (!currentEditor) return
    if (value === currentEditor.getHTML()) return
    currentEditor.commands.setContent(value)
  },
)

</script>

<template>
  <div class="d-editor flex h-full min-h-0 flex-col bg-base-100">
    <DEditorToolbar :editor="editor ?? null" :tools="props.tools" />

    <div class="d-editor-body relative min-h-0 flex-1 overflow-hidden">
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

    <!-- Slash Menu -->
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
