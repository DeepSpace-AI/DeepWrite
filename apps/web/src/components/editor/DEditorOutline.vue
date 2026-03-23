<script setup lang="ts">
import { computed } from 'vue'
import type { Editor } from '@tiptap/vue-3'
import { useDocumentOutline, type OutlineItem } from './useDocumentOutline'

interface Props {
  editor: Editor | null
  visible: boolean
}

const props = defineProps<Props>()

const emit = defineEmits<{
  close: []
}>()

const { outline, hasHeadings, scrollToHeading, scrollToTop } = useDocumentOutline(props.editor)

const indentWidth = computed(() => {
  return 16
})

function handleSelect(item: OutlineItem) {
  scrollToHeading(item.pos)
  emit('close')
}

function handleScrollToTop() {
  scrollToTop()
  emit('close')
}

function getLevelLabel(level: number) {
  return `H${level}`
}
</script>

<template>
  <div
    v-if="visible"
    class="d-editor-outline paper-card absolute left-0 top-0 z-30 flex h-full flex-col rounded-md"
    style="width: 260px;"
  >
    <div class="flex items-center justify-between px-4 py-3">
      <h3 class="text-sm font-semibold text-pretty">文档大纲</h3>
      <button
        type="button"
        class="btn-tertiary rounded-md px-2 py-1 text-xs"
        aria-label="关闭"
        @click="emit('close')"
      >
        ✕
      </button>
    </div>

    <div class="flex-1 overflow-y-auto p-2">
      <div v-if="!hasHeadings" class="py-8 text-center text-sm text-pretty-muted">
        暂无标题
      </div>

      <div v-else class="space-y-0.5">
        <button
          type="button"
          class="d-editor-outline-item btn btn-sm h-auto w-full justify-start rounded-md py-1.5 text-left text-pretty-secondary"
          @click="handleScrollToTop"
        >
          <span class="text-pretty-muted">↑ 顶部</span>
        </button>

        <button
          v-for="(item, index) in outline"
          :key="index"
          type="button"
          class="d-editor-outline-item btn btn-sm h-auto w-full justify-start rounded-md py-1.5 text-left text-pretty-secondary"
          :style="{ paddingLeft: `${(item.level - 1) * indentWidth + 8}px` }"
          :title="item.text"
          @click="handleSelect(item)"
        >
          <span class="badge badge-xs mr-1.5 shrink-0" :class="`badge-${item.level <= 1 ? 'primary' : item.level === 2 ? 'secondary' : 'ghost'}`">
            {{ getLevelLabel(item.level) }}
          </span>
          <span class="truncate text-sm">{{ item.text }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.d-editor-outline-item:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: -2px;
}
</style>
