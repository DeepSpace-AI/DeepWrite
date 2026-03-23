<script setup lang="ts">
import { computed } from 'vue'
import type { Editor } from '@tiptap/vue-3'
import { useWordCount } from './useWordCount'

interface Props {
  editor: Editor | null
}

const props = defineProps<Props>()

const { charCount, wordCount, paragraphCount, formattedReadingTime } = useWordCount(props.editor)

const charDisplay = computed(() => {
  if (charCount.value >= 10000) {
    return `${(charCount.value / 1000).toFixed(1)}k`
  }
  return charCount.value.toLocaleString()
})

const wordDisplay = computed(() => {
  if (wordCount.value >= 10000) {
    return `${(wordCount.value / 1000).toFixed(1)}k`
  }
  return wordCount.value.toLocaleString()
})
</script>

<template>
  <div class="d-editor-footer flex items-center justify-between bg-[var(--surface-overlay)] px-4 py-1.5 text-xs text-pretty-secondary">
    <div class="flex items-center gap-4">
      <span title="字符数">{{ charDisplay }} 字符</span>
      <span title="词数">{{ wordDisplay }} 词</span>
      <span title="段落数">{{ paragraphCount }} 段落</span>
    </div>
    <div class="flex items-center gap-4">
      <span title="预计阅读时间">{{ formattedReadingTime }}</span>
    </div>
  </div>
</template>
