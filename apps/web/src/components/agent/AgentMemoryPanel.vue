<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  listMemories,
  deleteMemory,
  triggerMemoryExtraction,
  getMemoryExtractionStatus,
  type Memory,
} from '@/api/agent'
import { ApiError } from '@/api/http'
import { DButton } from '@/components/base'
import IconPlus from '~icons/mdi/plus'
import IconDelete from '~icons/mdi/delete'
import IconEdit from '~icons/mdi/pencil'
import IconRefresh from '~icons/mdi/refresh'
import IconBrain from '~icons/mdi/brain'
import IconPreference from '~icons/mdi/heart'
import IconFact from '~icons/mdi/information'
import IconTask from '~icons/mdi/checkbox-marked'
import IconInsight from '~icons/mdi/lightbulb'
import AgentMemoryEditor from './AgentMemoryEditor.vue'

const props = defineProps<{
  agentId: string
  workspaceId?: string
  sessionId?: string
  compact?: boolean
}>()

const emit = defineEmits<{
  'memory-create': [memory: Memory]
  'memory-delete': [memoryId: string]
}>()

const { t } = useI18n()

const memories = ref<Memory[]>([])
const isLoading = ref(false)
const error = ref('')
const showEditor = ref(false)
const editingMemory = ref<Memory | null>(null)
const editorMode = ref<'create' | 'edit'>('create')

const isExtracting = ref(false)
const extractionTaskId = ref<string | null>(null)
const extractionStatus = ref<string | null>(null)

const memoryTypeConfig: Record<string, { icon: typeof IconPreference; label: string; colorClass: string }> = {
  preference: { icon: IconPreference, label: t('agents.memoryTypes.preference'), colorClass: 'text-(--color-primary)' },
  fact: { icon: IconFact, label: t('agents.memoryTypes.fact'), colorClass: 'text-(--color-secondary)' },
  task: { icon: IconTask, label: t('agents.memoryTypes.task'), colorClass: 'text-(--color-tertiary)' },
  insight: { icon: IconInsight, label: t('agents.memoryTypes.insight'), colorClass: 'text-(--color-success)' },
}

const hasMemories = computed(() => memories.value.length > 0)

async function loadMemories() {
  isLoading.value = true
  error.value = ''
  try {
    const res = await listMemories({
      agent_id: props.agentId,
      workspace_id: props.workspaceId,
      limit: 20,
    })
    memories.value = res.items || []
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Failed to load memories'
  } finally {
    isLoading.value = false
  }
}

async function handleExtractMemories() {
  if (!props.sessionId || isExtracting.value) return

  isExtracting.value = true
  extractionStatus.value = 'pending'
  error.value = ''

  try {
    const result = await triggerMemoryExtraction(props.sessionId)
    extractionTaskId.value = result.task_id
    
    pollExtractionStatus(result.task_id)
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Failed to trigger extraction'
    isExtracting.value = false
  }
}

async function pollExtractionStatus(taskId: string) {
  const maxPolls = 30
  let polls = 0

  const poll = async () => {
    if (polls >= maxPolls) {
      extractionStatus.value = 'timeout'
      isExtracting.value = false
      return
    }

    polls++
    try {
      const status = await getMemoryExtractionStatus(taskId)
      extractionStatus.value = status.status

      if (status.ready) {
        isExtracting.value = false
        if (status.status === 'SUCCESS') {
          await loadMemories()
        }
        return
      }

      setTimeout(poll, 2000)
    } catch {
      extractionStatus.value = 'error'
      isExtracting.value = false
    }
  }

  poll()
}

function openCreateEditor() {
  editingMemory.value = null
  editorMode.value = 'create'
  showEditor.value = true
}

function openEditEditor(memory: Memory) {
  editingMemory.value = memory
  editorMode.value = 'edit'
  showEditor.value = true
}

async function handleDelete(memoryId: string) {
  if (!confirm(t('agents.confirmDeleteMemory'))) return

  try {
    await deleteMemory(memoryId)
    memories.value = memories.value.filter(m => m.id !== memoryId)
    emit('memory-delete', memoryId)
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Failed to delete memory'
  }
}

function handleEditorSaved(memory: Memory) {
  showEditor.value = false
  if (editorMode.value === 'create') {
    memories.value.unshift(memory)
    emit('memory-create', memory)
  } else {
    const index = memories.value.findIndex(m => m.id === memory.id)
    if (index !== -1) {
      memories.value[index] = memory
    }
  }
}

onMounted(() => loadMemories())
</script>

<template>
  <div class="memory-panel">
    <div class="mb-3 flex items-center justify-between">
      <h4 class="label-sm uppercase tracking-widest text-(--color-on-surface-variant)">
        {{ t('agents.memories') }}
        <span v-if="memories.length > 0" class="ml-1 text-(--color-primary)">({{ memories.length }})</span>
      </h4>
      <div class="flex gap-1">
        <DButton
          v-if="sessionId && !compact"
          size="xs"
          :class="{ 'animate-pulse': isExtracting }"
          :disabled="isExtracting"
          :title="t('agents.extractMemories')"
          @click="handleExtractMemories"
        >
          <IconBrain class="h-4 w-4" />
        </DButton>
        <DButton
          size="xs"
          :title="t('agents.addMemory')"
          @click="openCreateEditor"
        >
          <IconPlus class="h-4 w-4" />
        </DButton>
        <DButton
          size="xs"
          :title="t('common.refresh')"
          @click="loadMemories"
        >
          <IconRefresh class="h-4 w-4" :class="{ 'animate-spin': isLoading }" />
        </DButton>
      </div>
    </div>

    <div v-if="isExtracting" class="mb-3 rounded-md bg-(--color-primary)/10 p-2">
      <p class="text-xs text-(--color-primary)">
        {{ t('agents.extractingMemories') }}
        <span v-if="extractionStatus">({{ extractionStatus }})</span>
      </p>
    </div>

    <div v-if="error" class="mb-2 rounded-md bg-(--color-error)/10 p-2 text-xs text-(--color-error)">
      {{ error }}
    </div>

    <div v-if="isLoading" class="flex justify-center py-4">
      <span class="loading loading-spinner loading-sm" />
    </div>

    <div v-else-if="!hasMemories" class="py-4 text-center">
      <IconBrain class="mx-auto mb-2 h-8 w-8 text-(--color-on-surface-variant) opacity-30" />
      <p class="text-xs text-(--color-on-surface-variant)">{{ t('agents.noMemories') }}</p>
      <button
        v-if="sessionId"
        class="mt-2 text-xs text-(--color-primary) hover:underline"
        @click="handleExtractMemories"
      >
        {{ t('agents.extractFromConversation') }}
      </button>
    </div>

    <div v-else class="max-h-64 space-y-2 overflow-y-auto">
      <div
        v-for="memory in memories"
        :key="memory.id"
        class="memory-card group rounded-md bg-(--surface-container) p-3 transition-colors hover:bg-(--surface-container-high)"
      >
        <div class="flex items-start justify-between gap-2">
          <div class="flex items-center gap-2">
            <component
              :is="memoryTypeConfig[memory.type]?.icon || IconFact"
              class="h-4 w-4 shrink-0"
              :class="memoryTypeConfig[memory.type]?.colorClass"
            />
            <span class="text-xs font-medium text-(--color-on-surface-variant)">
              {{ memoryTypeConfig[memory.type]?.label || memory.type }}
            </span>
            <span class="text-xs text-(--color-on-surface-variant) opacity-50">
              {{ (memory.importance * 100).toFixed(0) }}%
            </span>
          </div>
          <div class="flex gap-1 opacity-0 transition-opacity group-hover:opacity-100">
            <DButton size="xs" @click="openEditEditor(memory)">
              <IconEdit class="h-3 w-3" />
            </DButton>
            <DButton variant="danger" size="xs" @click="handleDelete(memory.id)">
              <IconDelete class="h-3 w-3" />
            </DButton>
          </div>
        </div>
        <p class="mt-2 text-sm text-(--color-on-background) line-clamp-2">
          {{ memory.summary || memory.content }}
        </p>
      </div>
    </div>

    <AgentMemoryEditor
      :visible="showEditor"
      :memory="editingMemory"
      :agent-id="agentId"
      :workspace-id="workspaceId"
      :mode="editorMode"
      @close="showEditor = false"
      @saved="handleEditorSaved"
    />
  </div>
</template>

<style scoped>
.memory-card {
  position: relative;
}

.memory-card:hover .group {
  opacity: 1;
}
</style>