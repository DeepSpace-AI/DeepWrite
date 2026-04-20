<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Memory } from '@/api/agent'
import {
  createMemory,
  updateMemory,
  type CreateMemoryInput,
} from '@/api/agent'
import { ApiError } from '@/api/http'
import { DModal, DInput, DButton } from '@/components/base'

const props = defineProps<{
  visible: boolean
  memory?: Memory | null
  agentId: string
  workspaceId?: string
  mode: 'create' | 'edit'
}>()

const emit = defineEmits<{
  close: []
  saved: [memory: Memory]
}>()

const { t } = useI18n()

const isSaving = ref(false)
const error = ref('')

const form = ref({
  type: 'fact',
  content: '',
  summary: '',
  importance: 0.5,
})

const memoryTypes = [
  { value: 'preference', label: t('agents.memoryTypes.preference'), description: t('agents.memoryTypes.preferenceDesc') },
  { value: 'fact', label: t('agents.memoryTypes.fact'), description: t('agents.memoryTypes.factDesc') },
  { value: 'task', label: t('agents.memoryTypes.task'), description: t('agents.memoryTypes.taskDesc') },
  { value: 'insight', label: t('agents.memoryTypes.insight'), description: t('agents.memoryTypes.insightDesc') },
]

const isValid = computed(() => {
  return form.value.content.trim().length >= 10
})

watch(() => props.memory, (memory) => {
  if (memory && props.mode === 'edit') {
    form.value = {
      type: memory.type || 'fact',
      content: memory.content || '',
      summary: memory.summary || '',
      importance: memory.importance || 0.5,
    }
  } else {
    resetForm()
  }
}, { immediate: true })

function resetForm() {
  form.value = {
    type: 'fact',
    content: '',
    summary: '',
    importance: 0.5,
  }
  error.value = ''
}

async function handleSave() {
  if (!isValid.value || isSaving.value) return

  isSaving.value = true
  error.value = ''

  try {
    let memory: Memory
    
    if (props.mode === 'edit' && props.memory) {
      memory = await updateMemory(props.memory.id, {
        content: form.value.content,
        summary: form.value.summary,
        importance: form.value.importance,
      })
    } else {
      const input: CreateMemoryInput = {
        agent_id: props.agentId,
        type: form.value.type,
        content: form.value.content,
        summary: form.value.summary,
        importance: form.value.importance,
      }
      if (props.workspaceId) {
        input.workspace_id = props.workspaceId
      }
      memory = await createMemory(input)
    }
    
    emit('saved', memory)
    resetForm()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Failed to save memory'
  } finally {
    isSaving.value = false
  }
}

function handleClose() {
  resetForm()
  emit('close')
}
</script>

<template>
  <DModal
    :visible="visible"
    :title="mode === 'edit' ? t('agents.editMemory') : t('agents.createMemory')"
    size="lg"
    @close="handleClose"
  >
    <div v-if="error" class="mb-4 rounded-md bg-(--color-error)/10 p-3 text-sm text-(--color-error)">
      {{ error }}
    </div>

    <form class="space-y-4" @submit.prevent="handleSave">
      <div v-if="mode === 'create'">
        <label class="label-sm mb-2 block text-(--color-on-surface-variant)">
          {{ t('agents.form.memoryType') }}
        </label>
        <div class="grid grid-cols-2 gap-2">
          <button
            v-for="mt in memoryTypes"
            :key="mt.value"
            type="button"
            class="memory-type-btn rounded-md border p-3 text-left transition-colors"
            :class="form.type === mt.value
              ? 'border-(--color-primary) bg-(--color-primary)/10'
              : 'border-(--color-surface-container-high) hover:bg-(--surface-container)'"
            @click="form.type = mt.value"
          >
            <p class="text-sm font-medium text-(--color-on-background)">{{ mt.label }}</p>
            <p class="text-xs text-(--color-on-surface-variant)">{{ mt.description }}</p>
          </button>
        </div>
      </div>

      <div>
        <label class="label-sm mb-1 block text-(--color-on-surface-variant)">
          {{ t('agents.form.memoryContent') }} *
        </label>
        <textarea
          v-model="form.content"
          rows="4"
          class="d-input d-input-bordered d-input-md w-full"
          :placeholder="t('agents.form.memoryContentPlaceholder')"
        />
        <p class="mt-1 text-xs text-(--color-on-surface-variant)">
          {{ form.content.length }} {{ t('agents.characters') }}
        </p>
      </div>

      <div>
        <label class="label-sm mb-1 block text-(--color-on-surface-variant)">
          {{ t('agents.form.memorySummary') }}
        </label>
        <DInput
          v-model="form.summary"
          variant="bordered"
          :placeholder="t('agents.form.memorySummaryPlaceholder')"
        />
      </div>

      <div>
        <label class="label-sm mb-1 block text-(--color-on-surface-variant)">
          {{ t('agents.form.importance') }}: {{ form.importance.toFixed(1) }}
        </label>
        <input
          v-model.number="form.importance"
          type="range"
          min="0"
          max="1"
          step="0.1"
          class="w-full accent-(--color-primary)"
        />
        <div class="flex justify-between text-xs text-(--color-on-surface-variant)">
          <span>{{ t('agents.low') }}</span>
          <span>{{ t('agents.high') }}</span>
        </div>
      </div>
    </form>

    <template #actions>
      <DButton variant="ghost" @click="handleClose">
        {{ t('common.cancel') }}
      </DButton>
      <DButton variant="primary" :disabled="!isValid" :loading="isSaving" @click="handleSave">
        {{ isSaving ? t('common.saving') : t('common.save') }}
      </DButton>
    </template>
  </DModal>
</template>

<style scoped>
.memory-type-btn {
  cursor: pointer;
}
</style>