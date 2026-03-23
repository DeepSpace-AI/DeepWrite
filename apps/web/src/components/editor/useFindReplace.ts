import { ref, computed, watch } from 'vue'
import type { Editor } from '@tiptap/vue-3'

export interface FindResult {
  index: number
  from: number
  to: number
}

export function useFindReplace(editor: Editor | null) {
  const query = ref('')
  const replacement = ref('')
  const isSearching = ref(false)
  const matchCase = ref(false)
  const wholeWord = ref(false)
  const results = ref<FindResult[]>([])
  const currentIndex = ref(0)

  const hasResults = computed(() => results.value.length > 0)
  const hasMore = computed(() => currentIndex.value < results.value.length - 1)
  const hasPrev = computed(() => currentIndex.value > 0)
  const resultText = computed(() => {
    if (!hasResults.value) return ''
    return `${currentIndex.value + 1} / ${results.value.length}`
  })

  function search() {
    results.value = []
    currentIndex.value = 0
    isSearching.value = true

    if (!editor || !query.value.trim()) {
      isSearching.value = false
      return
    }

    const doc = editor.state.doc
    const docSize = doc.content.size

    let foundIndex = 0
    doc.descendants((node, pos) => {
      if (!node.isText) return true

      const text = node.text || ''
      const searchTerm = matchCase.value ? query.value : query.value.toLowerCase()
      const nodeText = matchCase.value ? text : text.toLowerCase()

      let searchPos = 0
      while (searchPos < nodeText.length) {
        const foundPos = nodeText.indexOf(searchTerm, searchPos)
        if (foundPos === -1) break

        if (wholeWord.value) {
          const charBefore = (foundPos > 0 ? nodeText.charAt(foundPos - 1) : ' ') || ' '
          const charAfter = (foundPos + searchTerm.length < nodeText.length ? nodeText.charAt(foundPos + searchTerm.length) : ' ') || ' '
          if (/\w/.test(charBefore) || /\w/.test(charAfter)) {
            searchPos = foundPos + 1
            continue
          }
        }

        const from = pos + foundPos
        const to = from + searchTerm.length

        if (from >= 0 && to <= docSize) {
          results.value.push({ index: foundIndex, from, to })
          foundIndex++
        }

        searchPos = foundPos + 1
      }
      return true
    })

    isSearching.value = false
  }

  function next() {
    if (!hasMore.value) {
      currentIndex.value = 0
    } else {
      currentIndex.value++
    }
    scrollToCurrent()
  }

  function prev() {
    if (!hasPrev.value) {
      currentIndex.value = results.value.length - 1
    } else {
      currentIndex.value--
    }
    scrollToCurrent()
  }

  function scrollToCurrent() {
    if (!editor || !hasResults.value) return
    const result = results.value[currentIndex.value]
    if (result) {
      editor.commands.setTextSelection(result.from)
      editor.commands.scrollIntoView()
    }
  }

  function replace() {
    if (!editor || !hasResults.value) return
    const result = results.value[currentIndex.value]
    if (result) {
      editor.chain().focus().deleteRange({ from: result.from, to: result.to }).insertContentAt(result.from, replacement.value).run()
      search()
    }
  }

  function replaceAll() {
    if (!editor || !query.value.trim()) return
    const content = editor.getText()
    const newContent = content.split(query.value).join(replacement.value)
    editor.chain().focus().setContent(newContent).run()
    search()
  }

  function close() {
    query.value = ''
    replacement.value = ''
    results.value = []
    currentIndex.value = 0
  }

  watch(query, () => {
    if (query.value.length >= 1) {
      search()
    } else {
      results.value = []
    }
  })

  watch(matchCase, () => {
    if (query.value.length >= 1) {
      search()
    }
  })

  watch(wholeWord, () => {
    if (query.value.length >= 1) {
      search()
    }
  })

  return {
    query,
    replacement,
    matchCase,
    wholeWord,
    results,
    currentIndex,
    hasResults,
    hasMore,
    hasPrev,
    resultText,
    search,
    next,
    prev,
    replace,
    replaceAll,
    close,
  }
}
