<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Editor } from '@tiptap/vue-3'
import type { EditorTool } from './types'
import IconBold from '~icons/mdi/format-bold'
import IconItalic from '~icons/mdi/format-italic'
import IconUnderline from '~icons/mdi/format-underline'
import IconStrike from '~icons/mdi/format-strikethrough-variant'
import IconHighlight from '~icons/mdi/marker'
import IconTextColor from '~icons/mdi/format-color-text'
import IconH1 from '~icons/mdi/format-header-1'
import IconH2 from '~icons/mdi/format-header-2'
import IconParagraph from '~icons/mdi/format-paragraph'
import IconAlignLeft from '~icons/mdi/format-align-left'
import IconAlignCenter from '~icons/mdi/format-align-center'
import IconAlignRight from '~icons/mdi/format-align-right'
import IconAlignJustify from '~icons/mdi/format-align-justify'
import IconListBullet from '~icons/mdi/format-list-bulleted'
import IconListNumber from '~icons/mdi/format-list-numbered'
import IconChecklist from '~icons/mdi/format-list-checks'
import IconQuote from '~icons/mdi/format-quote-close'
import IconCodeBlock from '~icons/mdi/code-braces'
import IconInlineCode from '~icons/mdi/code-tags'
import IconDivider from '~icons/mdi/minus'
import IconClear from '~icons/mdi/eraser'
import IconUndo from '~icons/mdi/undo'
import IconRedo from '~icons/mdi/redo'
import IconLink from '~icons/mdi/link-variant'
import IconImage from '~icons/mdi/image'
import IconTable from '~icons/mdi/table'
import IconH3 from '~icons/mdi/format-header-3'
import IconH4 from '~icons/mdi/format-header-4'
import IconTableRowPlus from '~icons/mdi/table-row-plus-after'
import IconTableRowRemove from '~icons/mdi/table-row-remove'
import IconTableColumnPlus from '~icons/mdi/table-column-plus-after'
import IconTableColumnRemove from '~icons/mdi/table-column-remove'
import IconTableRemove from '~icons/mdi/table-remove'

interface Props {
  editor: Editor | null
  tools: EditorTool[]
}

type ToolGroup = 'style' | 'heading' | 'block' | 'action'

interface ToolItem {
  key: EditorTool
  group: ToolGroup
  tip: string
  label: string
  icon: unknown
}

const props = defineProps<Props>()

const linkModalOpen = ref(false)
const imageModalOpen = ref(false)
const linkUrl = ref('')
const imageUrl = ref('')

const openLinkModal = () => {
  if (!props.editor) return
  linkUrl.value = props.editor.getAttributes('link').href ?? ''
  linkModalOpen.value = true
}

const openImageModal = () => {
  if (!props.editor) return
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

const run = (action: () => void) => {
  if (!props.editor) return
  action()
}

const can = (action: (editor: Editor) => boolean) => {
  if (!props.editor) return false
  return action(props.editor)
}

const toolSet = computed(() => new Set(props.tools))
const hasTool = (tool: EditorTool) => toolSet.value.has(tool)

const toolGroups: ToolGroup[] = ['style', 'heading', 'block', 'action']

const toolItems: ToolItem[] = [
  { key: 'bold', group: 'style', tip: '加粗', label: '加粗', icon: IconBold },
  { key: 'italic', group: 'style', tip: '斜体', label: '斜体', icon: IconItalic },
  { key: 'underline', group: 'style', tip: '下划线', label: '下划线', icon: IconUnderline },
  { key: 'strike', group: 'style', tip: '删除线', label: '删除线', icon: IconStrike },
  { key: 'highlight', group: 'style', tip: '高亮', label: '高亮', icon: IconHighlight },
  { key: 'textPrimary', group: 'style', tip: '主色文字', label: '主色文字', icon: IconTextColor },
  { key: 'textDefault', group: 'style', tip: '默认文字色', label: '默认文字色', icon: IconTextColor },
  { key: 'inlineCode', group: 'style', tip: '行内代码', label: '行内代码', icon: IconInlineCode },
  { key: 'heading1', group: 'heading', tip: '一级标题', label: '一级标题', icon: IconH1 },
  { key: 'heading2', group: 'heading', tip: '二级标题', label: '二级标题', icon: IconH2 },
  { key: 'heading3', group: 'heading', tip: '三级标题', label: '三级标题', icon: IconH3 },
  { key: 'heading4', group: 'heading', tip: '四级标题', label: '四级标题', icon: IconH4 },
  { key: 'paragraph', group: 'heading', tip: '正文', label: '正文', icon: IconParagraph },
  { key: 'alignLeft', group: 'heading', tip: '左对齐', label: '左对齐', icon: IconAlignLeft },
  { key: 'alignCenter', group: 'heading', tip: '居中对齐', label: '居中对齐', icon: IconAlignCenter },
  { key: 'alignRight', group: 'heading', tip: '右对齐', label: '右对齐', icon: IconAlignRight },
  { key: 'alignJustify', group: 'heading', tip: '两端对齐', label: '两端对齐', icon: IconAlignJustify },
  { key: 'bulletList', group: 'block', tip: '无序列表', label: '无序列表', icon: IconListBullet },
  { key: 'orderedList', group: 'block', tip: '有序列表', label: '有序列表', icon: IconListNumber },
  { key: 'taskList', group: 'block', tip: '任务列表', label: '任务列表', icon: IconChecklist },
  { key: 'blockquote', group: 'block', tip: '引用', label: '引用', icon: IconQuote },
  { key: 'codeBlock', group: 'block', tip: '代码块', label: '代码块', icon: IconCodeBlock },
  { key: 'horizontalRule', group: 'block', tip: '分割线', label: '分割线', icon: IconDivider },
  { key: 'link', group: 'block', tip: '链接', label: '链接', icon: IconLink },
  { key: 'image', group: 'block', tip: '图片', label: '图片', icon: IconImage },
  { key: 'table', group: 'block', tip: '表格', label: '表格', icon: IconTable },
  { key: 'tableAddRowAfter', group: 'block', tip: '向下插入行', label: '向下插入行', icon: IconTableRowPlus },
  { key: 'tableAddRowBefore', group: 'block', tip: '向上插入行', label: '向上插入行', icon: IconTableRowPlus },
  { key: 'tableDeleteRow', group: 'block', tip: '删除行', label: '删除行', icon: IconTableRowRemove },
  { key: 'tableAddColumnAfter', group: 'block', tip: '向右插入列', label: '向右插入列', icon: IconTableColumnPlus },
  { key: 'tableAddColumnBefore', group: 'block', tip: '向左插入列', label: '向左插入列', icon: IconTableColumnPlus },
  { key: 'tableDeleteColumn', group: 'block', tip: '删除列', label: '删除列', icon: IconTableColumnRemove },
  { key: 'tableDeleteTable', group: 'block', tip: '删除表格', label: '删除表格', icon: IconTableRemove },
  { key: 'clear', group: 'action', tip: '清除格式', label: '清除格式', icon: IconClear },
  { key: 'undo', group: 'action', tip: '撤销', label: '撤销', icon: IconUndo },
  { key: 'redo', group: 'action', tip: '重做', label: '重做', icon: IconRedo },
]

const visibleToolsByGroup = computed<Record<ToolGroup, ToolItem[]>>(() => ({
  style: toolItems.filter((item) => item.group === 'style' && hasTool(item.key)),
  heading: toolItems.filter((item) => item.group === 'heading' && hasTool(item.key)),
  block: toolItems.filter((item) => item.group === 'block' && hasTool(item.key)),
  action: toolItems.filter((item) => item.group === 'action' && hasTool(item.key)),
}))

const isToolActive = (tool: EditorTool) => {
  if (!props.editor) return false
  switch (tool) {
    case 'bold':
      return props.editor.isActive('bold')
    case 'italic':
      return props.editor.isActive('italic')
    case 'strike':
      return props.editor.isActive('strike')
    case 'underline':
      return props.editor.isActive('underline')
    case 'highlight':
      return props.editor.isActive('highlight')
    case 'textPrimary':
      return props.editor.isActive('textStyle', { color: 'var(--color-primary)' })
    case 'textDefault':
      return !props.editor.isActive('textStyle', { color: 'var(--color-primary)' })
    case 'inlineCode':
      return props.editor.isActive('code')
    case 'heading1':
      return props.editor.isActive('heading', { level: 1 })
    case 'heading2':
      return props.editor.isActive('heading', { level: 2 })
    case 'heading3':
      return props.editor.isActive('heading', { level: 3 })
    case 'heading4':
      return props.editor.isActive('heading', { level: 4 })
    case 'paragraph':
      return props.editor.isActive('paragraph')
    case 'alignLeft':
      return props.editor.isActive({ textAlign: 'left' })
    case 'alignCenter':
      return props.editor.isActive({ textAlign: 'center' })
    case 'alignRight':
      return props.editor.isActive({ textAlign: 'right' })
    case 'alignJustify':
      return props.editor.isActive({ textAlign: 'justify' })
    case 'bulletList':
      return props.editor.isActive('bulletList')
    case 'orderedList':
      return props.editor.isActive('orderedList')
    case 'taskList':
      return props.editor.isActive('taskList')
    case 'blockquote':
      return props.editor.isActive('blockquote')
    case 'codeBlock':
      return props.editor.isActive('codeBlock')
    default:
      return false
  }
}

const canRunTool = (tool: EditorTool) => {
  switch (tool) {
    case 'bold':
      return can((e) => e.can().chain().focus().toggleBold().run())
    case 'italic':
      return can((e) => e.can().chain().focus().toggleItalic().run())
    case 'strike':
      return can((e) => e.can().chain().focus().toggleStrike().run())
    case 'underline':
      return can((e) => e.can().chain().focus().toggleUnderline().run())
    case 'highlight':
      return can((e) => e.can().chain().focus().toggleHighlight().run())
    case 'textPrimary':
      return can((e) => e.can().chain().focus().setColor('var(--color-primary)').run())
    case 'textDefault':
      return can((e) => e.can().chain().focus().unsetColor().run())
    case 'inlineCode':
      return can((e) => e.can().chain().focus().toggleCode().run())
    case 'heading1':
      return can((e) => e.can().chain().focus().toggleHeading({ level: 1 }).run())
    case 'heading2':
      return can((e) => e.can().chain().focus().toggleHeading({ level: 2 }).run())
    case 'heading3':
      return can((e) => e.can().chain().focus().toggleHeading({ level: 3 }).run())
    case 'heading4':
      return can((e) => e.can().chain().focus().toggleHeading({ level: 4 }).run())
    case 'paragraph':
      return can((e) => e.can().chain().focus().setParagraph().run())
    case 'alignLeft':
      return can((e) => e.can().chain().focus().setTextAlign('left').run())
    case 'alignCenter':
      return can((e) => e.can().chain().focus().setTextAlign('center').run())
    case 'alignRight':
      return can((e) => e.can().chain().focus().setTextAlign('right').run())
    case 'alignJustify':
      return can((e) => e.can().chain().focus().setTextAlign('justify').run())
    case 'bulletList':
      return can((e) => e.can().chain().focus().toggleBulletList().run())
    case 'orderedList':
      return can((e) => e.can().chain().focus().toggleOrderedList().run())
    case 'taskList':
      return can((e) => e.can().chain().focus().toggleTaskList().run())
    case 'blockquote':
      return can((e) => e.can().chain().focus().toggleBlockquote().run())
    case 'codeBlock':
      return can((e) => e.can().chain().focus().toggleCodeBlock().run())
    case 'horizontalRule':
      return can((e) => e.can().chain().focus().setHorizontalRule().run())
    case 'clear':
      return can((e) => e.can().chain().focus().clearNodes().unsetAllMarks().run())
    case 'link':
      return true
    case 'image':
      return true
    case 'table':
      return true
    case 'tableAddRowAfter':
      return true
    case 'tableAddRowBefore':
      return true
    case 'tableDeleteRow':
      return true
    case 'tableAddColumnAfter':
      return true
    case 'tableAddColumnBefore':
      return true
    case 'tableDeleteColumn':
      return true
    case 'tableDeleteTable':
      return true
    case 'undo':
      return can((e) => e.can().chain().focus().undo().run())
    case 'redo':
      return can((e) => e.can().chain().focus().redo().run())
    default:
      return false
  }
}

const runTool = (tool: EditorTool) => {
  switch (tool) {
    case 'bold':
      run(() => props.editor?.chain().focus().toggleBold().run())
      break
    case 'italic':
      run(() => props.editor?.chain().focus().toggleItalic().run())
      break
    case 'strike':
      run(() => props.editor?.chain().focus().toggleStrike().run())
      break
    case 'underline':
      run(() => props.editor?.chain().focus().toggleUnderline().run())
      break
    case 'highlight':
      run(() => props.editor?.chain().focus().toggleHighlight().run())
      break
    case 'textPrimary':
      run(() => props.editor?.chain().focus().setColor('var(--color-primary)').run())
      break
    case 'textDefault':
      run(() => props.editor?.chain().focus().unsetColor().run())
      break
    case 'inlineCode':
      run(() => props.editor?.chain().focus().toggleCode().run())
      break
    case 'heading1':
      run(() => props.editor?.chain().focus().toggleHeading({ level: 1 }).run())
      break
    case 'heading2':
      run(() => props.editor?.chain().focus().toggleHeading({ level: 2 }).run())
      break
    case 'heading3':
      run(() => props.editor?.chain().focus().toggleHeading({ level: 3 }).run())
      break
    case 'heading4':
      run(() => props.editor?.chain().focus().toggleHeading({ level: 4 }).run())
      break
    case 'paragraph':
      run(() => props.editor?.chain().focus().setParagraph().run())
      break
    case 'alignLeft':
      run(() => props.editor?.chain().focus().setTextAlign('left').run())
      break
    case 'alignCenter':
      run(() => props.editor?.chain().focus().setTextAlign('center').run())
      break
    case 'alignRight':
      run(() => props.editor?.chain().focus().setTextAlign('right').run())
      break
    case 'alignJustify':
      run(() => props.editor?.chain().focus().setTextAlign('justify').run())
      break
    case 'bulletList':
      run(() => props.editor?.chain().focus().toggleBulletList().run())
      break
    case 'orderedList':
      run(() => props.editor?.chain().focus().toggleOrderedList().run())
      break
    case 'taskList':
      run(() => props.editor?.chain().focus().toggleTaskList().run())
      break
    case 'blockquote':
      run(() => props.editor?.chain().focus().toggleBlockquote().run())
      break
    case 'codeBlock':
      run(() => props.editor?.chain().focus().toggleCodeBlock().run())
      break
    case 'horizontalRule':
      run(() => props.editor?.chain().focus().setHorizontalRule().run())
      break
    case 'clear':
      run(() => props.editor?.chain().focus().clearNodes().unsetAllMarks().run())
      break
    case 'link': {
      openLinkModal()
      break
    }
    case 'image': {
      openImageModal()
      break
    }
    case 'table': {
      const rows = prompt('输入行数 (默认 3):', '3')
      const cols = prompt('输入列数 (默认 3):', '3')
      if (rows && cols) {
        run(() =>
          props.editor
            ?.chain()
            .focus()
            .insertTable({ rows: parseInt(rows, 10), cols: parseInt(cols, 10), withHeaderRow: true })
            .run(),
        )
      }
      break
    }
    case 'tableAddRowAfter':
      run(() => props.editor?.chain().focus().addRowAfter().run())
      break
    case 'tableAddRowBefore':
      run(() => props.editor?.chain().focus().addRowBefore().run())
      break
    case 'tableDeleteRow':
      run(() => props.editor?.chain().focus().deleteRow().run())
      break
    case 'tableAddColumnAfter':
      run(() => props.editor?.chain().focus().addColumnAfter().run())
      break
    case 'tableAddColumnBefore':
      run(() => props.editor?.chain().focus().addColumnBefore().run())
      break
    case 'tableDeleteColumn':
      run(() => props.editor?.chain().focus().deleteColumn().run())
      break
    case 'tableDeleteTable':
      run(() => props.editor?.chain().focus().deleteTable().run())
      break
    case 'undo':
      run(() => props.editor?.chain().focus().undo().run())
      break
    case 'redo':
      run(() => props.editor?.chain().focus().redo().run())
      break
    default:
      break
  }
}
</script>

<template>
  <div>
    <div class="d-editor-toolbar flex flex-wrap items-center gap-3 border-b border-base-300 p-3">
      <div v-for="group in toolGroups" :key="group" v-show="visibleToolsByGroup[group].length > 0" class="join">
        <div v-for="tool in visibleToolsByGroup[group]" :key="tool.key" class="tooltip tooltip-bottom z-20" :data-tip="tool.tip">
          <button
            type="button"
            class="d-editor-btn btn btn-sm btn-ghost join-item"
            :class="isToolActive(tool.key) ? 'btn-primary' : ''"
            :aria-label="tool.label"
            :disabled="!canRunTool(tool.key)"
            @click="runTool(tool.key)"
          >
            <component :is="tool.icon" class="size-4" />
          </button>
        </div>
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
