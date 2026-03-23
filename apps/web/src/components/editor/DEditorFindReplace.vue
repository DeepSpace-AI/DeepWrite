<script setup lang="ts">
import { computed } from 'vue'
import type { Editor } from '@tiptap/vue-3'
import { useFindReplace } from './useFindReplace'

interface Props {
  editor: Editor | null
  visible: boolean
}

const props = defineProps<Props>()

const emit = defineEmits<{
  close: []
}>()

const {
  query,
  replacement,
  matchCase,
  wholeWord,
  hasResults,
  hasMore,
  hasPrev,
  resultText,
  next,
  prev,
  replace,
  replaceAll,
} = useFindReplace(props.editor)

const hasQuery = computed(() => query.value.trim().length > 0)
</script>

<template>
  <div
    v-if="visible"
    class="d-editor-find-replace paper-card absolute right-4 top-14 z-30 w-80 rounded-md"
  >
    <div class="px-4 py-3">
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-semibold text-pretty">查找替换</h3>
        <button
          type="button"
          class="btn-tertiary rounded-md px-2 py-1 text-xs"
          aria-label="关闭"
          @click="emit('close')"
        >
          ✕
        </button>
      </div>
    </div>

    <div class="space-y-3 p-4">
      <div class="space-y-1.5">
        <label class="text-xs text-pretty-secondary">查找</label>
        <div class="flex gap-1.5">
          <input
            v-model="query"
            type="text"
            class="glass-input input input-bordered input-sm flex-1 rounded-md"
            placeholder="输入查找内容..."
            @keyup.enter="next"
          />
          <div class="join">
            <button
              type="button"
              class="join-item btn btn-sm"
              :disabled="!hasResults"
              @click="prev"
            >
              ‹
            </button>
            <button
              type="button"
              class="join-item btn btn-sm pointer-events-none"
              :disabled="!hasResults"
            >
              {{ resultText }}
            </button>
            <button
              type="button"
              class="join-item btn btn-sm"
              :disabled="!hasResults"
              @click="next"
            >
              ›
            </button>
          </div>
        </div>
      </div>

      <div class="space-y-1.5">
        <label class="text-xs text-pretty-secondary">替换为</label>
        <div class="flex gap-1.5">
          <input
            v-model="replacement"
            type="text"
            class="glass-input input input-bordered input-sm flex-1 rounded-md"
            placeholder="输入替换内容..."
          />
        </div>
      </div>

      <div class="flex items-center gap-4">
        <label class="flex items-center gap-1.5 text-xs cursor-pointer">
          <input
            v-model="matchCase"
            type="checkbox"
            class="checkbox checkbox-xs checkbox-primary"
          />
          <span>区分大小写</span>
        </label>
        <label class="flex items-center gap-1.5 text-xs cursor-pointer">
          <input
            v-model="wholeWord"
            type="checkbox"
            class="checkbox checkbox-xs checkbox-primary"
          />
          <span>全词匹配</span>
        </label>
      </div>

      <div class="flex gap-2 pt-1">
        <button
          type="button"
          class="btn-primary-vellum flex-1 rounded-md px-3 py-1.5 text-sm"
          :disabled="!hasQuery"
          @click="replace"
        >
          替换
        </button>
        <button
          type="button"
          class="btn-tertiary flex-1 rounded-md px-3 py-1.5 text-sm"
          :disabled="!hasQuery"
          @click="replaceAll"
        >
          全部替换
        </button>
      </div>

      <p v-if="query && !hasResults" class="py-2 text-center text-xs text-pretty-muted">
        未找到匹配项
      </p>
    </div>
  </div>
</template>
