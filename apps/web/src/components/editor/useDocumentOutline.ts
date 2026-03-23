import { computed, ref, watch } from 'vue'
import type { Editor } from '@tiptap/vue-3'

export interface OutlineItem {
  level: number
  text: string
  pos: number
}

export function useDocumentOutline(editor: Editor | null) {
  const outline = ref<OutlineItem[]>([])

  function extractOutline() {
    if (!editor) {
      outline.value = []
      return
    }

    const items: OutlineItem[] = []
    const doc = editor.state.doc

    doc.descendants((node, pos) => {
      if (node.type.name === 'heading') {
        const level = node.attrs.level as number
        const text = node.textContent.trim()
        if (text) {
          items.push({
            level,
            text,
            pos,
          })
        }
      }
      return true
    })

    outline.value = items
  }

  function scrollToHeading(pos: number) {
    if (!editor) return
    editor.commands.setTextSelection(pos)
    editor.commands.scrollIntoView()
  }

  function scrollToTop() {
    if (!editor) return
    editor.commands.setTextSelection(0)
    editor.commands.scrollIntoView()
  }

  watch(
    () => editor?.state,
    () => {
      extractOutline()
    },
    { immediate: true, deep: true },
  )

  const hasHeadings = computed(() => outline.value.length > 0)

  const maxLevel = computed(() => {
    if (outline.value.length === 0) return 0
    return Math.max(...outline.value.map((item) => item.level))
  })

  return {
    outline,
    hasHeadings,
    maxLevel,
    scrollToHeading,
    scrollToTop,
  }
}
