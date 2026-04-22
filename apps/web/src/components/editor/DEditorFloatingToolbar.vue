<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Editor } from '@tiptap/vue-3'
import IconBold from '~icons/mdi/format-bold'
import IconItalic from '~icons/mdi/format-italic'
import IconUnderline from '~icons/mdi/format-underline'
import IconStrike from '~icons/mdi/format-strikethrough-variant'
import IconHighlight from '~icons/mdi/marker'
import IconInlineCode from '~icons/mdi/code-tags'
import IconLink from '~icons/mdi/link-variant'
import IconClear from '~icons/mdi/eraser'

interface Props {
  editor: Editor | null
  visible: boolean
  x: number
  y: number
}

const props = defineProps<Props>()

defineEmits<{
  close: []
  'open-link-modal': []
}>()

const linkModalOpen = ref(false)
const linkUrl = ref('')

const isActive = (type: string, attrs?: Record<string, unknown>) => {
  if (!props.editor) return false
  if (attrs) {
    return props.editor.isActive(type, attrs)
  }
  return props.editor.isActive(type)
}



const toggleBold = () => props.editor?.chain().focus().toggleBold().run()
const toggleItalic = () => props.editor?.chain().focus().toggleItalic().run()
const toggleUnderline = () => props.editor?.chain().focus().toggleUnderline().run()
const toggleStrike = () => props.editor?.chain().focus().toggleStrike().run()
const toggleHighlight = () => props.editor?.chain().focus().toggleHighlight().run()
const toggleCode = () => props.editor?.chain().focus().toggleCode().run()
const clearFormat = () => props.editor?.chain().focus().clearNodes().unsetAllMarks().run()

const openLink = () => {
  if (!props.editor) return
  linkUrl.value = props.editor.getAttributes('link').href ?? ''
  linkModalOpen.value = true
}

const submitLink = () => {
  if (!props.editor) return
  const url = linkUrl.value.trim()
  if (!url) {
    props.editor.chain().focus().extendMarkRange('link').unsetLink().run()
  } else {
    props.editor.chain().focus().extendMarkRange('link').setLink({ href: url }).run()
  }
  linkModalOpen.value = false
  linkUrl.value = ''
}

const toolbarStyle = computed(() => ({
  left: `${props.x}px`,
  top: `${props.y}px`,
  opacity: props.visible ? '1' : '0',
  pointerEvents: props.visible ? ('auto' as const) : ('none' as const),
}))
</script>

<template>
  <div
    class="d-editor-floating-toolbar paper-card fixed z-50 flex items-center gap-0.5 rounded-md px-1 py-1 transition-all duration-150"
    :style="toolbarStyle"
    role="toolbar"
    aria-label="文本格式工具"
  >
    <button
      type="button"
      class="d-editor-btn btn btn-sm btn-ghost min-w-8 flex-col gap-0 px-1.5"
      :class="{ 'btn-primary': isActive('bold') }"
      title="加粗 (Ctrl+B)"
      aria-label="加粗"
      :aria-pressed="isActive('bold')"
      @click="toggleBold"
    >
      <IconBold class="size-4" />
    </button>

    <button
      type="button"
      class="d-editor-btn btn btn-sm btn-ghost min-w-8 flex-col gap-0 px-1.5"
      :class="{ 'btn-primary': isActive('italic') }"
      title="斜体 (Ctrl+I)"
      aria-label="斜体"
      @click="toggleItalic"
    >
      <IconItalic class="size-4" />
    </button>

    <button
      type="button"
      class="d-editor-btn btn btn-sm btn-ghost min-w-8 flex-col gap-0 px-1.5"
      :class="{ 'btn-primary': isActive('underline') }"
      title="下划线 (Ctrl+U)"
      aria-label="下划线"
      @click="toggleUnderline"
    >
      <IconUnderline class="size-4" />
    </button>

    <button
      type="button"
      class="d-editor-btn btn btn-sm btn-ghost min-w-8 flex-col gap-0 px-1.5"
      :class="{ 'btn-primary': isActive('strike') }"
      title="删除线 (Ctrl+Shift+X)"
      aria-label="删除线"
      @click="toggleStrike"
    >
      <IconStrike class="size-4" />
    </button>

    <div class="mx-0.5 h-5 w-px bg-[var(--surface-sunken)]" />

    <button
      type="button"
      class="d-editor-btn btn btn-sm btn-ghost min-w-8 flex-col gap-0 px-1.5"
      :class="{ 'btn-primary': isActive('highlight') }"
      title="高亮"
      aria-label="高亮"
      @click="toggleHighlight"
    >
      <IconHighlight class="size-4" />
    </button>

    <button
      type="button"
      class="d-editor-btn btn btn-sm btn-ghost min-w-8 flex-col gap-0 px-1.5"
      :class="{ 'btn-primary': isActive('code') }"
      title="行内代码"
      aria-label="行内代码"
      @click="toggleCode"
    >
      <IconInlineCode class="size-4" />
    </button>

    <div class="mx-0.5 h-5 w-px bg-[var(--surface-sunken)]" />

    <button
      type="button"
      class="d-editor-btn btn btn-sm btn-ghost min-w-8 flex-col gap-0 px-1.5"
      :class="{ 'btn-primary': isActive('link') }"
      title="链接"
      aria-label="添加链接"
      @click="openLink"
    >
      <IconLink class="size-4" />
    </button>

    <button
      type="button"
      class="d-editor-btn btn btn-sm btn-ghost min-w-8 flex-col gap-0 px-1.5"
      title="清除格式"
      aria-label="清除格式"
      @click="clearFormat"
    >
      <IconClear class="size-4" />
    </button>
  </div>

  <div class="modal px-4" :class="{ 'modal-open': linkModalOpen }" role="dialog" aria-modal="true">
    <div class="modal-box paper-card max-w-sm rounded-md p-0">
      <div class="px-5 py-4">
        <h3 class="text-base font-semibold text-pretty">编辑链接</h3>
      </div>
      <form class="space-y-4 px-5 py-4" @submit.prevent="submitLink">
        <label class="form-control w-full gap-2">
          <span class="label-text text-sm text-pretty-secondary">链接地址</span>
          <input
            v-model="linkUrl"
            type="url"
            class="glass-input input input-bordered input-sm w-full rounded-md"
            placeholder="https://example.com"
          >
        </label>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn-tertiary rounded-md px-3 py-1.5 text-sm" @click="linkModalOpen = false">取消</button>
          <button type="submit" class="btn-primary-vellum rounded-md px-3 py-1.5 text-sm">确定</button>
        </div>
      </form>
    </div>
    <div class="modal-backdrop" @click="linkModalOpen = false" />
  </div>
</template>
