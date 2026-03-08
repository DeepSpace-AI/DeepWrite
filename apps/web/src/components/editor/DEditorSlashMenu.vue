<script setup lang="ts">
import { computed, ref, watch, onBeforeUnmount } from 'vue'
import type { Editor } from '@tiptap/vue-3'
import IconTable from '~icons/mdi/table'
import IconCodeBlock from '~icons/mdi/code-braces'
import IconQuote from '~icons/mdi/format-quote-close'
import IconH1 from '~icons/mdi/format-header-1'
import IconH2 from '~icons/mdi/format-header-2'
import IconH3 from '~icons/mdi/format-header-3'
import IconH4 from '~icons/mdi/format-header-4'
import IconListBullet from '~icons/mdi/format-list-bulleted'
import IconListNumber from '~icons/mdi/format-list-numbered'
import IconChecklist from '~icons/mdi/format-list-checks'
import IconDivider from '~icons/mdi/minus'
import IconImage from '~icons/mdi/image'
import IconLink from '~icons/mdi/link'

interface Props {
  editor: Editor | null
  visible: boolean
  x: number
  y: number
  query: string
}

const emit = defineEmits<{
  close: []
}>()

const props = defineProps<Props>()

const selectedIndex = ref(0)
const linkModalOpen = ref(false)
const imageModalOpen = ref(false)
const linkUrl = ref('')
const imageUrl = ref('')

interface SlashCommand {
  key: string
  label: string
  icon: unknown
  action?: (editor: Editor) => void
  needsInput?: 'link' | 'image'
}

const removeSlashTrigger = () => {
  if (!props.editor) return
  const from = props.editor.state.selection.$from.pos - (props.query.length + 1)
  const to = props.editor.state.selection.$from.pos
  props.editor.chain().focus().deleteRange({ from, to }).run()
}

const openLinkModal = () => {
  if (!props.editor) return
  linkUrl.value = props.editor.getAttributes('link').href ?? ''
  linkModalOpen.value = true
}

const openImageModal = () => {
  imageUrl.value = ''
  imageModalOpen.value = true
}

const closeLinkModal = () => {
  linkModalOpen.value = false
}

const closeImageModal = () => {
  imageModalOpen.value = false
}

const submitLink = () => {
  const url = linkUrl.value.trim()
  if (!props.editor) return
  if (!url) {
    props.editor.chain().focus().extendMarkRange('link').unsetLink().run()
    closeLinkModal()
    return
  }
  props.editor.chain().focus().extendMarkRange('link').setLink({ href: url }).run()
  closeLinkModal()
}

const submitImage = () => {
  const url = imageUrl.value.trim()
  if (!props.editor || !url) return
  props.editor.chain().focus().setImage({ src: url }).run()
  closeImageModal()
}

const slashCommands: SlashCommand[] = [
  { key: 'h1', label: '一级标题', icon: IconH1, action: (e: Editor) => e.chain().focus().toggleHeading({ level: 1 }).run() },
  { key: 'h2', label: '二级标题', icon: IconH2, action: (e: Editor) => e.chain().focus().toggleHeading({ level: 2 }).run() },
  { key: 'h3', label: '三级标题', icon: IconH3, action: (e: Editor) => e.chain().focus().toggleHeading({ level: 3 }).run() },
  { key: 'h4', label: '四级标题', icon: IconH4, action: (e: Editor) => e.chain().focus().toggleHeading({ level: 4 }).run() },
  { key: 'bulletList', label: '无序列表', icon: IconListBullet, action: (e: Editor) => e.chain().focus().toggleBulletList().run() },
  { key: 'orderedList', label: '有序列表', icon: IconListNumber, action: (e: Editor) => e.chain().focus().toggleOrderedList().run() },
  { key: 'taskList', label: '任务列表', icon: IconChecklist, action: (e: Editor) => e.chain().focus().toggleTaskList().run() },
  { key: 'codeBlock', label: '代码块', icon: IconCodeBlock, action: (e: Editor) => e.chain().focus().toggleCodeBlock().run() },
  { key: 'blockquote', label: '引用块', icon: IconQuote, action: (e: Editor) => e.chain().focus().toggleBlockquote().run() },
  { key: 'table', label: '表格', icon: IconTable, action: (e: Editor) => {
    const rows = prompt('行数 (默认 3):', '3')
    const cols = prompt('列数 (默认 3):', '3')
    if (rows && cols) {
      e.chain().focus().insertTable({ rows: parseInt(rows, 10), cols: parseInt(cols, 10), withHeaderRow: true }).run()
    }
  } },
  { key: 'hr', label: '分割线', icon: IconDivider, action: (e: Editor) => e.chain().focus().setHorizontalRule().run() },
  { key: 'image', label: '图片', icon: IconImage, needsInput: 'image' },
  { key: 'link', label: '链接', icon: IconLink, needsInput: 'link' },
]

const filteredCommands = computed(() => {
  if (!props.query) return slashCommands
  return slashCommands.filter((cmd) =>
    cmd.label.toLowerCase().includes(props.query.toLowerCase()) ||
    cmd.key.toLowerCase().includes(props.query.toLowerCase())
  )
})

const handleSelect = (index: number) => {
  if (!props.editor) return
  const cmd = filteredCommands.value[index]
  if (cmd) {
    removeSlashTrigger()

    if (cmd.needsInput === 'link') {
      emit('close')
      openLinkModal()
      return
    }

    if (cmd.needsInput === 'image') {
      emit('close')
      openImageModal()
      return
    }

    cmd.action?.(props.editor)
    emit('close')
  }
}

const handleKeyDown = (e: KeyboardEvent) => {
  if (!props.visible || filteredCommands.value.length === 0) return
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    selectedIndex.value = (selectedIndex.value + 1) % filteredCommands.value.length
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    selectedIndex.value = (selectedIndex.value - 1 + filteredCommands.value.length) % filteredCommands.value.length
  } else if (e.key === 'Enter') {
    e.preventDefault()
    handleSelect(selectedIndex.value)
  } else if (e.key === 'Escape') {
    e.preventDefault()
    emit('close')
  }
}

const bindKeyboard = () => {
  window.addEventListener('keydown', handleKeyDown)
}

const unbindKeyboard = () => {
  window.removeEventListener('keydown', handleKeyDown)
}

watch(
  () => props.visible,
  (visible) => {
    if (visible) {
      selectedIndex.value = 0
      bindKeyboard()
      return
    }
    unbindKeyboard()
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  unbindKeyboard()
})
</script>

<template>
  <div>
    <div
      v-if="props.visible && filteredCommands.length > 0"
      class="d-editor-slash-menu fixed z-50 bg-base-100 border border-base-300 rounded-lg shadow-lg max-h-80 overflow-y-auto"
      :style="{ left: `${props.x}px`, top: `${props.y}px` }"
    >
      <div
        v-for="(cmd, index) in filteredCommands"
        :key="cmd.key"
        class="d-editor-slash-item flex items-center gap-3 px-4 py-3 cursor-pointer hover:bg-base-200 transition-colors"
        :class="index === selectedIndex ? 'bg-primary text-primary-content' : ''"
        @click="handleSelect(index)"
        @mouseenter="selectedIndex = index"
      >
        <component :is="cmd.icon" class="size-5" />
        <span class="text-sm font-medium">{{ cmd.label }}</span>
      </div>
    </div>

    <div class="modal px-4" :class="{ 'modal-open': linkModalOpen }" role="dialog" aria-modal="true">
      <div class="modal-box max-w-xl p-0">
        <div class="border-b border-base-300 px-6 py-4">
          <h3 class="text-lg font-semibold text-base-content">编辑链接</h3>
          <p class="mt-1 text-sm text-base-content/70">输入完整 URL，留空后保存可移除当前链接</p>
        </div>
        <form class="space-y-4 px-6 py-5" @submit.prevent="submitLink">
          <label class="form-control w-full gap-2">
            <span class="label-text text-sm">链接地址</span>
            <input
              v-model="linkUrl"
              type="url"
              class="input input-bordered w-full"
              placeholder="https://example.com"
            >
          </label>
          <div class="modal-action mt-0 border-t border-base-300 pt-4">
            <button type="button" class="btn btn-ghost" @click="closeLinkModal">取消</button>
            <button type="submit" class="btn btn-primary">保存</button>
          </div>
        </form>
      </div>
      <div class="modal-backdrop" @click="closeLinkModal" />
    </div>

    <div class="modal px-4" :class="{ 'modal-open': imageModalOpen }" role="dialog" aria-modal="true">
      <div class="modal-box max-w-xl p-0">
        <div class="border-b border-base-300 px-6 py-4">
          <h3 class="text-lg font-semibold text-base-content">插入图片</h3>
          <p class="mt-1 text-sm text-base-content/70">支持远程图片 URL，推荐使用 HTTPS 地址</p>
        </div>
        <form class="space-y-4 px-6 py-5" @submit.prevent="submitImage">
          <label class="form-control w-full gap-2">
            <span class="label-text text-sm">图片地址</span>
            <input
              v-model="imageUrl"
              type="url"
              class="input input-bordered w-full"
              placeholder="https://example.com/image.png"
            >
          </label>
          <div class="modal-action mt-0 border-t border-base-300 pt-4">
            <button type="button" class="btn btn-ghost" @click="closeImageModal">取消</button>
            <button type="submit" class="btn btn-primary" :disabled="!imageUrl.trim()">插入</button>
          </div>
        </form>
      </div>
      <div class="modal-backdrop" @click="closeImageModal" />
    </div>
  </div>
</template>

<style scoped>
.d-editor-slash-menu {
  min-width: 200px;
}
</style>
