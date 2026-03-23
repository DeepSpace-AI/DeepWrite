import { computed, ref, watch } from 'vue'
import type { Editor } from '@tiptap/vue-3'

export function useWordCount(editor: Editor | null) {
  const charCount = ref(0)
  const wordCount = ref(0)
  const paragraphCount = ref(0)
  const readingTime = ref(0)

  function updateCounts() {
    if (!editor) {
      charCount.value = 0
      wordCount.value = 0
      paragraphCount.value = 0
      readingTime.value = 0
      return
    }

    const text = editor.state.doc.textContent
    charCount.value = text.length

    const words = text.trim().split(/\s+/).filter((w) => w.length > 0)
    wordCount.value = words.length

    let paras = 0
    editor.state.doc.descendants((node) => {
      if (node.type.name === 'paragraph' && node.textContent.trim()) {
        paras++
      }
    })
    paragraphCount.value = paras

    const minutes = Math.ceil(words.length / 200)
    readingTime.value = minutes
  }

  watch(
    () => editor?.state,
    () => {
      updateCounts()
    },
    { immediate: true, deep: true },
  )

  const formattedReadingTime = computed(() => {
    if (readingTime.value <= 0) return '0 分钟'
    if (readingTime.value === 1) return '1 分钟'
    return `${readingTime.value} 分钟`
  })

  return {
    charCount,
    wordCount,
    paragraphCount,
    readingTime,
    formattedReadingTime,
  }
}
