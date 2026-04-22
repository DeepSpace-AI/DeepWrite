<script setup lang="ts">
import { ref } from 'vue'
import type { Editor } from '@tiptap/vue-3'
import { useVersionHistory, type Snapshot } from './useVersionHistory'

interface Props {
  editor: Editor | null
  visible: boolean
}

const props = defineProps<Props>()

const emit = defineEmits<{
  close: []
}>()

const {
  snapshots,
  lastSaved,
  createSnapshot,
  restoreSnapshot,
  deleteSnapshot,
  renameSnapshot,
  hasAutoSave,
  getAutoSavedContent,
  discardAutoSave,
  formatDate,
} = useVersionHistory(props.editor)

const isCreating = ref(false)
const newSnapshotName = ref('')
const editingId = ref<string | null>(null)
const editingName = ref('')

function handleCreate() {
  if (isCreating.value) {
    if (newSnapshotName.value.trim()) {
      createSnapshot(newSnapshotName.value.trim())
    } else {
      createSnapshot()
    }
    newSnapshotName.value = ''
    isCreating.value = false
  } else {
    isCreating.value = true
  }
}

function handleRestore(snapshot: Snapshot) {
  if (confirm(`确定要恢复到"${snapshot.name}"吗？当前未保存的更改将丢失。`)) {
    restoreSnapshot(snapshot)
    emit('close')
  }
}

function handleDelete(snapshot: Snapshot) {
  if (confirm(`确定要删除快照"${snapshot.name}"吗？此操作不可撤销。`)) {
    deleteSnapshot(snapshot.id)
  }
}

function startRename(snapshot: Snapshot) {
  editingId.value = snapshot.id
  editingName.value = snapshot.name
}

function handleRename(snapshot: Snapshot) {
  if (editingName.value.trim()) {
    renameSnapshot(snapshot.id, editingName.value.trim())
  }
  editingId.value = null
  editingName.value = ''
}

function handleDiscardAutoSave() {
  if (confirm('确定要丢弃自动保存的草稿吗？此操作不可撤销。')) {
    discardAutoSave()
  }
}

function handleRestoreAutoSave() {
  if (!confirm('确定要恢复自动保存的草稿吗？当前内容将被替换。')) return
  const content = getAutoSavedContent()
  if (content && props.editor) {
    props.editor.commands.setContent(content)
    emit('close')
  }
}
</script>

<template>
  <div
    v-if="visible"
    class="d-editor-history paper-card absolute right-4 top-14 z-30 w-80 rounded-md"
  >
    <div class="px-4 py-3">
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-semibold text-pretty">版本历史</h3>
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

    <div class="max-h-96 overflow-y-auto p-3 space-y-3">
      <div v-if="hasAutoSave()" class="rounded-md bg-[var(--glow-primary)] p-3">
        <div class="flex items-start justify-between gap-2">
          <div>
            <p class="text-sm font-medium text-primary">自动保存的草稿</p>
            <p v-if="lastSaved" class="mt-0.5 text-xs text-base-content/60">
              {{ formatDate(lastSaved) }}
            </p>
          </div>
          <div class="flex gap-1">
            <button
              type="button"
              class="btn-primary-vellum rounded-md px-2 py-1 text-xs"
              @click="handleRestoreAutoSave"
            >
              恢复
            </button>
            <button
              type="button"
              class="btn-tertiary rounded-md px-2 py-1 text-xs"
              @click="handleDiscardAutoSave"
            >
              丢弃
            </button>
          </div>
        </div>
      </div>

      <div class="space-y-1.5">
        <div class="flex items-center justify-between">
          <span class="text-xs font-medium text-base-content/70">快照列表</span>
          <button
            type="button"
            class="btn-primary-vellum rounded-md px-2 py-1 text-xs"
            @click="handleCreate"
          >
            {{ isCreating ? '保存快照' : '创建快照' }}
          </button>
        </div>

        <div v-if="isCreating" class="space-y-1.5 rounded-md bg-[var(--surface-overlay)] p-2">
          <input
            v-model="newSnapshotName"
            type="text"
            class="glass-input input input-bordered input-xs w-full rounded-md"
            placeholder="快照名称（可选）"
            @keyup.enter="handleCreate"
            @keyup.escape="isCreating = false"
          />
          <div class="flex justify-end gap-1">
            <button
              type="button"
                class="btn-tertiary rounded-md px-2 py-1 text-xs"
              @click="isCreating = false"
            >
              取消
            </button>
          </div>
        </div>
      </div>

      <div v-if="snapshots.length === 0 && !isCreating" class="py-4 text-center text-sm text-base-content/50">
        暂无快照
      </div>

      <div v-else class="space-y-1">
        <div
          v-for="snapshot in snapshots"
          :key="snapshot.id"
          class="group rounded-md bg-[var(--surface-overlay)] p-2"
        >
          <div class="flex items-start justify-between gap-2">
            <div class="min-w-0 flex-1">
              <template v-if="editingId === snapshot.id">
                <input
                  v-model="editingName"
                  type="text"
                  class="glass-input input input-bordered input-xs w-full rounded-md"
                  @keyup.enter="handleRename(snapshot)"
                  @keyup.escape="editingId = null"
                />
              </template>
              <template v-else>
                <p class="truncate text-sm font-medium">{{ snapshot.name }}</p>
                <p class="text-xs text-base-content/50">{{ formatDate(snapshot.createdAt) }}</p>
              </template>
            </div>
            <div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
              <button
                v-if="editingId !== snapshot.id"
                type="button"
                class="btn-tertiary rounded-md px-2 py-1 text-xs"
                title="重命名"
                @click="startRename(snapshot)"
              >
                ✏️
              </button>
              <button
                v-if="editingId !== snapshot.id"
                type="button"
                class="btn-primary-vellum rounded-md px-2 py-1 text-xs"
                @click="handleRestore(snapshot)"
              >
                恢复
              </button>
              <button
                v-if="editingId !== snapshot.id"
                type="button"
                class="btn-tertiary rounded-md px-2 py-1 text-xs text-error"
                title="删除"
                @click="handleDelete(snapshot)"
              >
                🗑️
              </button>
              <button
                v-if="editingId === snapshot.id"
                type="button"
                class="btn-tertiary rounded-md px-2 py-1 text-xs"
                @click="editingId = null"
              >
                取消
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="px-4 py-2">
      <p class="text-xs text-pretty-muted">
        最多保存 {{ 20 }} 个快照，自动保存间隔 30 秒
      </p>
    </div>
  </div>
</template>
