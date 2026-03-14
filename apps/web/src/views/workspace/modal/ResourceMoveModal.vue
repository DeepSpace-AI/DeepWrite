<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { VisibleFolderNode } from '@/views/workspace/types'

const ROOT_VALUE = '__root__'

const props = withDefaults(defineProps<{
  modelValue: boolean
  title: string
  subtitle?: string
  rootLabel: string
  destinationLabel: string
  confirmText: string
  visibleFolders: VisibleFolderNode[]
  currentFolderId?: string | null
  submitting?: boolean
  errorMessage?: string
}>(), {
  subtitle: '',
  currentFolderId: null,
  submitting: false,
  errorMessage: '',
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'submit', payload: { folderId: string | null }): void
}>()

const localFolderId = ref(ROOT_VALUE)

const visible = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})

watch(
  () => props.modelValue,
  (open) => {
    if (!open) return
    localFolderId.value = props.currentFolderId || ROOT_VALUE
  },
)

function folderOptionLabel(node: VisibleFolderNode) {
  return `${'-- '.repeat(node.depth)}${node.folder.name}`
}

function handleClose() {
  visible.value = false
}

function handleSubmit() {
  emit('submit', {
    folderId: localFolderId.value === ROOT_VALUE ? null : localFolderId.value,
  })
}
</script>

<template>
  <dialog class="modal" :class="{ 'modal-open': visible }" @close="handleClose">
    <div class="modal-box rounded-sm border border-base-300">
      <h3 class="text-lg font-semibold text-base-content">{{ title }}</h3>
      <p v-if="subtitle" class="mt-1 text-sm text-base-content/60">{{ subtitle }}</p>

      <div class="mt-4 space-y-4">
        <label class="fieldset">
          <legend class="fieldset-legend text-sm">{{ destinationLabel }}</legend>
          <select v-model="localFolderId" class="select select-bordered w-full rounded-sm">
            <option :value="ROOT_VALUE">{{ rootLabel }}</option>
            <option v-for="node in visibleFolders" :key="node.folder.id" :value="node.folder.id">
              {{ folderOptionLabel(node) }}
            </option>
          </select>
        </label>

        <p v-if="errorMessage" class="rounded-sm border border-error/30 bg-error/10 px-3 py-2 text-sm text-error">
          {{ errorMessage }}
        </p>
      </div>

      <div class="modal-action">
        <button type="button" class="btn btn-ghost rounded-sm" :disabled="submitting" @click="handleClose">取消</button>
        <button type="button" class="btn btn-primary rounded-sm" :disabled="submitting" @click="handleSubmit">
          {{ submitting ? '处理中...' : confirmText }}
        </button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button @click="handleClose">close</button>
    </form>
  </dialog>
</template>
