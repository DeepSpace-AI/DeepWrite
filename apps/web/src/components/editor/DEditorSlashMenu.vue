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

type CommandGroup = 'structure' | 'content' | 'media'

interface SlashCommand {
  key: string
  label: string
  icon: unknown
  shortcut?: string
  group: CommandGroup
  action?: (editor: Editor) => void
  needsInput?: 'link' | 'image'
}

const STORAGE_KEY_RECENT = 'deepwrite:editor:recent-slash-commands'
const MAX_RECENT = 3

const getRecentCommands = (): string[] => {
  try {
    const raw = localStorage.getItem(STORAGE_KEY_RECENT)
    if (!raw) return []
    return JSON.parse(raw) as string[]
  } catch {
    return []
  }
}

const addRecentCommand = (key: string) => {
  try {
    let recent = getRecentCommands().filter((k) => k !== key)
    recent.unshift(key)
    recent = recent.slice(0, MAX_RECENT)
    localStorage.setItem(STORAGE_KEY_RECENT, JSON.stringify(recent))
  } catch {
    // Ignore storage errors
  }
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
  { key: 'h1', label: '一级标题', icon: IconH1, shortcut: '/h1', group: 'structure', action: (e: Editor) => e.chain().focus().toggleHeading({ level: 1 }).run() },
  { key: 'h2', label: '二级标题', icon: IconH2, shortcut: '/h2', group: 'structure', action: (e: Editor) => e.chain().focus().toggleHeading({ level: 2 }).run() },
  { key: 'h3', label: '三级标题', icon: IconH3, shortcut: '/h3', group: 'structure', action: (e: Editor) => e.chain().focus().toggleHeading({ level: 3 }).run() },
  { key: 'h4', label: '四级标题', icon: IconH4, shortcut: '/h4', group: 'structure', action: (e: Editor) => e.chain().focus().toggleHeading({ level: 4 }).run() },
  { key: 'bulletList', label: '无序列表', icon: IconListBullet, shortcut: '/bullet', group: 'structure', action: (e: Editor) => e.chain().focus().toggleBulletList().run() },
  { key: 'orderedList', label: '有序列表', icon: IconListNumber, shortcut: '/ordered', group: 'structure', action: (e: Editor) => e.chain().focus().toggleOrderedList().run() },
  { key: 'taskList', label: '任务列表', icon: IconChecklist, shortcut: '/task', group: 'structure', action: (e: Editor) => e.chain().focus().toggleTaskList().run() },
  { key: 'blockquote', label: '引用块', icon: IconQuote, shortcut: '/quote', group: 'structure', action: (e: Editor) => e.chain().focus().toggleBlockquote().run() },
  { key: 'codeBlock', label: '代码块', icon: IconCodeBlock, shortcut: '/code', group: 'content', action: (e: Editor) => e.chain().focus().toggleCodeBlock().run() },
  { key: 'table', label: '表格', icon: IconTable, shortcut: '/table', group: 'content', action: (e: Editor) => {
    const rows = prompt('行数 (默认 3):', '3')
    const cols = prompt('列数 (默认 3):', '3')
    if (rows && cols) {
      e.chain().focus().insertTable({ rows: parseInt(rows, 10), cols: parseInt(cols, 10), withHeaderRow: true }).run()
    }
  } },
  { key: 'hr', label: '分割线', icon: IconDivider, shortcut: '/hr', group: 'content', action: (e: Editor) => e.chain().focus().setHorizontalRule().run() },
  { key: 'image', label: '图片', icon: IconImage, shortcut: '/image', group: 'media', needsInput: 'image' },
  { key: 'link', label: '链接', icon: IconLink, shortcut: '/link', group: 'media', needsInput: 'link' },
]

const groupLabels: Record<CommandGroup, string> = {
  structure: '结构',
  content: '内容',
  media: '媒体',
}

const filteredCommands = computed(() => {
  const query = props.query.toLowerCase().trim()
  let filtered = query
    ? slashCommands.filter((cmd) =>
        cmd.label.toLowerCase().includes(query) ||
        cmd.key.toLowerCase().includes(query) ||
        (cmd.shortcut && cmd.shortcut.toLowerCase().includes(query))
      )
    : [...slashCommands]

  if (!query) {
    const recent = getRecentCommands()
    if (recent.length > 0) {
      const recentCmds = recent
        .map((key) => filtered.find((cmd) => cmd.key === key))
        .filter((cmd): cmd is SlashCommand => cmd !== undefined)
      const otherCmds = filtered.filter((cmd) => !recent.includes(cmd.key))
      filtered = [...recentCmds, ...otherCmds]
    }
  }

  return filtered
})

const groupedCommands = computed(() => {
  const groups: { group: CommandGroup; label: string; items: SlashCommand[] }[] = []
  const groupMap = new Map<CommandGroup, SlashCommand[]>()

  for (const cmd of filteredCommands.value) {
    const list = groupMap.get(cmd.group) || []
    list.push(cmd)
    groupMap.set(cmd.group, list)
  }

  const order: CommandGroup[] = ['structure', 'content', 'media']
  for (const group of order) {
    const items = groupMap.get(group)
    if (items && items.length > 0) {
      groups.push({ group, label: groupLabels[group], items })
    }
  }

  return groups
})

const getGlobalIndex = (groupIndex: number, itemIndex: number) => {
  let idx = 0
  for (let i = 0; i < groupIndex; i++) {
    const group = groupedCommands.value[i]
    if (group) {
      idx += group.items.length
    }
  }
  return idx + itemIndex
}

const handleSelect = (index: number) => {
  if (!props.editor) return
  const cmd = filteredCommands.value[index]
  if (cmd) {
    addRecentCommand(cmd.key)
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
      class="d-editor-slash-menu paper-card fixed z-50 max-h-80 min-w-[220px] max-w-[280px] overflow-y-auto rounded-md"
      :style="{ left: `${props.x}px`, top: `${props.y}px` }"
    >
      <template v-for="(group, groupIndex) in groupedCommands" :key="group.group">
        <div class="bg-[var(--surface-overlay)] px-3 py-1.5 text-[11px] font-medium uppercase tracking-wider text-pretty-muted">
          {{ group.label }}
        </div>
        <div
          v-for="(cmd, itemIndex) in group.items"
          :key="cmd.key"
          class="d-editor-slash-item flex cursor-pointer items-center gap-3 px-3 py-2 text-pretty-secondary transition-colors hover:bg-[var(--surface-overlay)]"
          :class="getGlobalIndex(groupIndex, itemIndex) === selectedIndex ? 'bg-primary text-primary-content' : ''"
          @click="handleSelect(getGlobalIndex(groupIndex, itemIndex))"
          @mouseenter="selectedIndex = getGlobalIndex(groupIndex, itemIndex)"
        >
          <component :is="cmd.icon" class="size-4 shrink-0" />
          <span class="text-sm flex-1">{{ cmd.label }}</span>
          <span v-if="cmd.shortcut" class="text-[10px] opacity-60 font-mono">{{ cmd.shortcut }}</span>
        </div>
      </template>
    </div>

    <div class="modal px-4" :class="{ 'modal-open': linkModalOpen }" role="dialog" aria-modal="true">
      <div class="modal-box paper-card max-w-xl rounded-md p-0">
        <div class="px-6 py-4">
          <h3 class="text-lg font-semibold text-pretty">编辑链接</h3>
          <p class="mt-1 text-sm text-pretty-secondary">输入完整 URL，留空后保存可移除当前链接</p>
        </div>
        <form class="space-y-4 px-6 py-5" @submit.prevent="submitLink">
          <label class="form-control w-full gap-2">
            <span class="label-text text-sm text-pretty-secondary">链接地址</span>
            <input
              v-model="linkUrl"
              type="url"
              class="glass-input input input-bordered w-full rounded-md"
              placeholder="https://example.com"
            >
          </label>
          <div class="mt-0 flex justify-end gap-3 pt-4">
            <button type="button" class="btn-tertiary rounded-md px-4 py-2 text-sm" @click="closeLinkModal">取消</button>
            <button type="submit" class="btn-primary-vellum rounded-md px-4 py-2 text-sm">保存</button>
          </div>
        </form>
      </div>
      <div class="modal-backdrop" @click="closeLinkModal" />
    </div>

    <div class="modal px-4" :class="{ 'modal-open': imageModalOpen }" role="dialog" aria-modal="true">
      <div class="modal-box paper-card max-w-xl rounded-md p-0">
        <div class="px-6 py-4">
          <h3 class="text-lg font-semibold text-pretty">插入图片</h3>
          <p class="mt-1 text-sm text-pretty-secondary">支持远程图片 URL，推荐使用 HTTPS 地址</p>
        </div>
        <form class="space-y-4 px-6 py-5" @submit.prevent="submitImage">
          <label class="form-control w-full gap-2">
            <span class="label-text text-sm text-pretty-secondary">图片地址</span>
            <input
              v-model="imageUrl"
              type="url"
              class="glass-input input input-bordered w-full rounded-md"
              placeholder="https://example.com/image.png"
            >
          </label>
          <div class="mt-0 flex justify-end gap-3 pt-4">
            <button type="button" class="btn-tertiary rounded-md px-4 py-2 text-sm" @click="closeImageModal">取消</button>
            <button type="submit" class="btn-primary-vellum rounded-md px-4 py-2 text-sm" :disabled="!imageUrl.trim()">插入</button>
          </div>
        </form>
      </div>
      <div class="modal-backdrop" @click="closeImageModal" />
    </div>
  </div>
</template>
