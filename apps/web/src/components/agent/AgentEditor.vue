<script setup lang="ts">
import { ref, computed, watch, TransitionGroup } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAgentStore, type Agent } from '@/stores/agent'
import { DModal, DButton } from '@/components/base'
import IconResearch from '~icons/mdi/beaker'
import IconWrite from '~icons/mdi/pencil-outline'
import IconData from '~icons/mdi/database'
import IconPublish from '~icons/mdi/book-open-page-variant'
import IconGeneral from '~icons/mdi/assistant'
import IconDelete from '~icons/mdi/delete'
import IconWarning from '~icons/mdi/alert-circle'
import IconChevronDown from '~icons/mdi/chevron-down'
import IconChevronUp from '~icons/mdi/chevron-up'

const props = defineProps<{
  visible: boolean
  agent: Agent | null
}>()

const emit = defineEmits<{
  close: []
  updated: [agent: Agent]
  deleted: [agentId: string]
}>()

const { t } = useI18n()
const agentStore = useAgentStore()

const isSaving = ref(false)
const isDeleting = ref(false)
const error = ref('')
const showDeleteConfirm = ref(false)
const showAdvanced = ref(false)

const form = ref({
  name: '',
  description: '',
  category: 'research',
  system_prompt: '',
  identity_prompt: '',
  capability_prompt: '',
  instruction_prompt: '',
  safety_prompt: '',
  inject_user_context: true,
  inject_memory: true,
  inject_time: true,
  inject_workspace: true,
  memory_retrieval_count: 5,
  memory_min_relevance: 0.7,
  temperature: 0.7,
  max_tokens: 4096,
  public: false,
})

const categoryOptions = [
  { 
    value: 'research', 
    label: t('agents.categories.research'), 
    icon: IconResearch,
    color: 'terracotta',
  },
  { 
    value: 'writing', 
    label: t('agents.categories.writing'), 
    icon: IconWrite,
    color: 'sage',
  },
  { 
    value: 'data', 
    label: t('agents.categories.data'), 
    icon: IconData,
    color: 'amber',
  },
  { 
    value: 'publishing', 
    label: t('agents.categories.publishing'), 
    icon: IconPublish,
    color: 'plum',
  },
  { 
    value: 'general', 
    label: t('agents.categories.general'), 
    icon: IconGeneral,
    color: 'slate',
  },
]

const isValid = computed(() => {
  return form.value.name.trim() && form.value.system_prompt.trim()
})

const hasChanges = computed(() => {
  if (!props.agent) return false
  return (
    form.value.name !== props.agent.name ||
    form.value.description !== (props.agent.description || '') ||
    form.value.category !== props.agent.category ||
    form.value.system_prompt !== props.agent.system_prompt ||
    form.value.identity_prompt !== (props.agent.identity_prompt || '') ||
    form.value.capability_prompt !== (props.agent.capability_prompt || '') ||
    form.value.instruction_prompt !== (props.agent.instruction_prompt || '') ||
    form.value.safety_prompt !== (props.agent.safety_prompt || '') ||
    form.value.inject_user_context !== props.agent.inject_user_context ||
    form.value.inject_memory !== props.agent.inject_memory ||
    form.value.inject_time !== props.agent.inject_time ||
    form.value.inject_workspace !== props.agent.inject_workspace ||
    form.value.memory_retrieval_count !== props.agent.memory_retrieval_count ||
    form.value.memory_min_relevance !== props.agent.memory_min_relevance ||
    form.value.temperature !== props.agent.temperature ||
    form.value.max_tokens !== props.agent.max_tokens ||
    form.value.public !== props.agent.public
  )
})

watch(() => props.agent, (agent) => {
  if (agent) {
    form.value = {
      name: agent.name,
      description: agent.description || '',
      category: agent.category || 'research',
      system_prompt: agent.system_prompt || '',
      identity_prompt: agent.identity_prompt || '',
      capability_prompt: agent.capability_prompt || '',
      instruction_prompt: agent.instruction_prompt || '',
      safety_prompt: agent.safety_prompt || '',
      inject_user_context: agent.inject_user_context ?? true,
      inject_memory: agent.inject_memory ?? true,
      inject_time: agent.inject_time ?? true,
      inject_workspace: agent.inject_workspace ?? true,
      memory_retrieval_count: agent.memory_retrieval_count ?? 5,
      memory_min_relevance: agent.memory_min_relevance ?? 0.7,
      temperature: agent.temperature || 0.7,
      max_tokens: agent.max_tokens || 4096,
      public: agent.public || false,
    }
  }
  error.value = ''
  showDeleteConfirm.value = false
}, { immediate: true })

async function handleUpdate() {
  if (!isValid.value || !props.agent || isSaving.value) return

  isSaving.value = true
  error.value = ''

  try {
    const updated = await agentStore.updateExistingAgent(props.agent.id, {
      name: form.value.name,
      description: form.value.description,
      category: form.value.category,
      system_prompt: form.value.system_prompt,
      identity_prompt: form.value.identity_prompt,
      capability_prompt: form.value.capability_prompt,
      instruction_prompt: form.value.instruction_prompt,
      safety_prompt: form.value.safety_prompt,
      inject_user_context: form.value.inject_user_context,
      inject_memory: form.value.inject_memory,
      inject_time: form.value.inject_time,
      inject_workspace: form.value.inject_workspace,
      memory_retrieval_count: form.value.memory_retrieval_count,
      memory_min_relevance: form.value.memory_min_relevance,
      temperature: form.value.temperature,
      max_tokens: form.value.max_tokens,
      public: form.value.public,
    })
    emit('updated', updated)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to update agent'
  } finally {
    isSaving.value = false
  }
}

async function handleDelete() {
  if (!props.agent || isDeleting.value) return
  
  if (!showDeleteConfirm.value) {
    showDeleteConfirm.value = true
    return
  }

  isDeleting.value = true
  error.value = ''

  try {
    await agentStore.deleteExistingAgent(props.agent.id)
    emit('deleted', props.agent.id)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to delete agent'
  } finally {
    isDeleting.value = false
    showDeleteConfirm.value = false
  }
}

function handleClose() {
  error.value = ''
  showDeleteConfirm.value = false
  emit('close')
}
</script>

<template>
  <DModal
    :visible="visible"
    :title="t('agents.editAgent')"
    size="xl"
    @close="handleClose"
  >
    <TransitionGroup name="fade" tag="div">
      <div v-if="error" key="error" class="error-message">
        {{ error }}
      </div>
    </TransitionGroup>

    <div class="editor-content">
      <div class="form-grid">
        <div class="field">
          <label class="field-label">{{ t('agents.form.name') }} *</label>
          <input
            v-model="form.name"
            type="text"
            class="field-input"
          >
        </div>

        <div class="field">
          <label class="field-label">{{ t('agents.form.category') }}</label>
          <div class="category-select">
            <button
              v-for="cat in categoryOptions"
              :key="cat.value"
              class="category-chip"
              :class="{ selected: form.category === cat.value, [cat.color]: true }"
              @click="form.category = cat.value"
            >
              <component :is="cat.icon" class="h-4 w-4" />
              <span>{{ cat.label }}</span>
            </button>
          </div>
        </div>
      </div>

      <div class="field">
        <label class="field-label">{{ t('agents.form.description') }}</label>
        <textarea
          v-model="form.description"
          rows="2"
          class="field-input field-textarea"
        />
      </div>

      <div class="field">
        <label class="field-label">{{ t('agents.form.systemPrompt') }} *</label>
        <textarea
          v-model="form.system_prompt"
          rows="8"
          class="field-input field-textarea field-mono"
        />
        <div class="field-counter">
          {{ form.system_prompt.length }} {{ t('agents.characters') }}
        </div>
      </div>

      <div class="settings-row">
        <div class="field field-small">
          <label class="field-label">{{ t('agents.form.temperature') }}</label>
          <div class="slider-field">
            <input
              v-model.number="form.temperature"
              type="range"
              min="0"
              max="2"
              step="0.1"
              class="slider"
            >
            <span class="slider-value">{{ form.temperature.toFixed(1) }}</span>
          </div>
        </div>

        <div class="field field-small">
          <label class="field-label">{{ t('agents.form.maxTokens') }}</label>
          <input
            v-model.number="form.max_tokens"
            type="number"
            min="1"
            max="128000"
            class="field-input"
          >
        </div>
      </div>

      <div class="field">
        <label class="checkbox-field">
          <input
            v-model="form.public"
            type="checkbox"
            class="checkbox"
          >
          <span class="checkbox-label">
            <span class="checkbox-title">{{ t('agents.form.sharePublicly') }}</span>
            <span class="checkbox-desc">{{ t('agents.form.sharePubliclyDesc') }}</span>
          </span>
        </label>
      </div>

      <div class="advanced-section">
        <button class="advanced-toggle" @click="showAdvanced = !showAdvanced">
          <span class="toggle-title">{{ t('agents.stepAdvanced') }}</span>
          <IconChevronDown v-if="!showAdvanced" class="h-4 w-4" />
          <IconChevronUp v-if="showAdvanced" class="h-4 w-4" />
        </button>

        <Transition name="slide">
          <div v-if="showAdvanced" class="advanced-content">
            <div class="prompt-layers">
              <div class="field">
                <label class="field-label">{{ t('agents.promptLayers.identity') }}</label>
                <p class="field-hint">{{ t('agents.promptLayers.identityDesc') }}</p>
                <textarea
                  v-model="form.identity_prompt"
                  rows="3"
                  class="field-input field-textarea field-mono"
                />
              </div>

              <div class="field">
                <label class="field-label">{{ t('agents.promptLayers.capability') }}</label>
                <p class="field-hint">{{ t('agents.promptLayers.capabilityDesc') }}</p>
                <textarea
                  v-model="form.capability_prompt"
                  rows="3"
                  class="field-input field-textarea field-mono"
                />
              </div>

              <div class="field">
                <label class="field-label">{{ t('agents.promptLayers.instruction') }}</label>
                <p class="field-hint">{{ t('agents.promptLayers.instructionDesc') }}</p>
                <textarea
                  v-model="form.instruction_prompt"
                  rows="3"
                  class="field-input field-textarea field-mono"
                />
              </div>

              <div class="field">
                <label class="field-label">{{ t('agents.promptLayers.safety') }}</label>
                <p class="field-hint">{{ t('agents.promptLayers.safetyDesc') }}</p>
                <textarea
                  v-model="form.safety_prompt"
                  rows="3"
                  class="field-input field-textarea field-mono"
                />
              </div>
            </div>

            <div class="inject-section">
              <h4 class="section-title">{{ t('agents.injectSettings.title') }}</h4>
              <div class="inject-grid">
                <label class="checkbox-field">
                  <input v-model="form.inject_user_context" type="checkbox" class="checkbox" />
                  <span class="checkbox-label">
                    <span class="checkbox-title">{{ t('agents.injectSettings.userContext') }}</span>
                    <span class="checkbox-desc">{{ t('agents.injectSettings.userContextDesc') }}</span>
                  </span>
                </label>

                <label class="checkbox-field">
                  <input v-model="form.inject_memory" type="checkbox" class="checkbox" />
                  <span class="checkbox-label">
                    <span class="checkbox-title">{{ t('agents.injectSettings.memory') }}</span>
                    <span class="checkbox-desc">{{ t('agents.injectSettings.memoryDesc') }}</span>
                  </span>
                </label>

                <label class="checkbox-field">
                  <input v-model="form.inject_time" type="checkbox" class="checkbox" />
                  <span class="checkbox-label">
                    <span class="checkbox-title">{{ t('agents.injectSettings.time') }}</span>
                    <span class="checkbox-desc">{{ t('agents.injectSettings.timeDesc') }}</span>
                  </span>
                </label>

                <label class="checkbox-field">
                  <input v-model="form.inject_workspace" type="checkbox" class="checkbox" />
                  <span class="checkbox-label">
                    <span class="checkbox-title">{{ t('agents.injectSettings.workspace') }}</span>
                    <span class="checkbox-desc">{{ t('agents.injectSettings.workspaceDesc') }}</span>
                  </span>
                </label>
              </div>
            </div>

            <div class="memory-section">
              <h4 class="section-title">{{ t('agents.memorySettings.retrievalCount') }}</h4>
              <div class="settings-grid">
                <div class="field">
                  <label class="field-label">{{ t('agents.memorySettings.retrievalCount') }}</label>
                  <input
                    v-model.number="form.memory_retrieval_count"
                    type="number"
                    min="1"
                    max="20"
                    class="field-input"
                  />
                  <p class="field-hint">{{ t('agents.memorySettings.retrievalCountDesc') }}</p>
                </div>

                <div class="field">
                  <label class="field-label">{{ t('agents.memorySettings.minRelevance') }}</label>
                  <div class="slider-field">
                    <input
                      v-model.number="form.memory_min_relevance"
                      type="range"
                      min="0"
                      max="1"
                      step="0.1"
                      class="slider"
                    />
                    <span class="slider-value">{{ form.memory_min_relevance.toFixed(1) }}</span>
                  </div>
                  <p class="field-hint">{{ t('agents.memorySettings.minRelevanceDesc') }}</p>
                </div>
              </div>
            </div>
          </div>
        </Transition>
      </div>

      <Transition name="slide">
        <div v-if="showDeleteConfirm" class="delete-confirm">
          <div class="delete-warning">
            <IconWarning class="h-5 w-5" />
            <span>{{ t('agents.confirmDelete', { name: props.agent?.name }) }}</span>
          </div>
        </div>
      </Transition>
    </div>

    <template #actions>
      <button
        class="delete-btn"
        :class="{ 'delete-confirm-active': showDeleteConfirm }"
        :disabled="isDeleting"
        @click="handleDelete"
      >
        <IconDelete class="h-4 w-4" />
        <span v-if="showDeleteConfirm">{{ t('common.confirmDelete') }}</span>
        <span v-else>{{ t('common.delete') }}</span>
      </button>
      <div class="flex-1" />
      <DButton variant="ghost" @click="handleClose">
        {{ t('common.cancel') }}
      </DButton>
      <DButton
        variant="primary"
        :disabled="!isValid || !hasChanges"
        :loading="isSaving"
        @click="handleUpdate"
      >
        {{ isSaving ? t('common.saving') : t('common.save') }}
      </DButton>
    </template>
  </DModal>
</template>

<style scoped>
.editor-content {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.25rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.field-small {
  flex: 0 0 auto;
}

.field-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-on-background);
}

.field-input {
  width: 100%;
  padding: 0.625rem 0.875rem;
  border-radius: 0.5rem;
  border: 1px solid var(--color-surface-container-high);
  background: var(--surface-container);
  color: var(--color-on-background);
  font-size: 0.875rem;
  transition: all 0.15s ease;
}

.field-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px oklch(0.55 0.12 30 / 0.15);
}

.field-textarea {
  resize: vertical;
  min-height: 4rem;
}

.field-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.8125rem;
  line-height: 1.6;
}

.field-counter {
  font-size: 0.75rem;
  color: var(--color-on-surface-variant);
  text-align: right;
}

.category-select {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.category-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  border-radius: 1rem;
  border: 1px solid var(--color-surface-container-high);
  background: var(--surface-container);
  color: var(--color-on-surface-variant);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.category-chip:hover {
  border-color: var(--color-primary);
}

.category-chip.selected.terracotta {
  background: oklch(0.55 0.12 30 / 0.15);
  border-color: oklch(0.55 0.12 30);
  color: oklch(0.55 0.12 30);
}

.category-chip.selected.sage {
  background: oklch(0.55 0.08 130 / 0.15);
  border-color: oklch(0.55 0.08 130);
  color: oklch(0.55 0.08 130);
}

.category-chip.selected.amber {
  background: oklch(0.65 0.12 85 / 0.15);
  border-color: oklch(0.55 0.12 85);
  color: oklch(0.55 0.12 85);
}

.category-chip.selected.plum {
  background: oklch(0.55 0.12 320 / 0.15);
  border-color: oklch(0.55 0.12 320);
  color: oklch(0.55 0.12 320);
}

.category-chip.selected.slate {
  background: oklch(0.55 0.02 250 / 0.15);
  border-color: oklch(0.45 0.02 250);
  color: oklch(0.45 0.02 250);
}

.settings-row {
  display: flex;
  gap: 1.5rem;
}

.slider-field {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.slider {
  width: 6rem;
  height: 0.375rem;
  border-radius: 0.1875rem;
  background: var(--surface-container-high);
  appearance: none;
  cursor: pointer;
}

.slider::-webkit-slider-thumb {
  appearance: none;
  width: 1rem;
  height: 1rem;
  border-radius: 50%;
  background: var(--color-primary);
  cursor: pointer;
  box-shadow: 0 2px 6px oklch(0 0 0 / 0.2);
}

.slider-value {
  min-width: 2rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-on-background);
}

.checkbox-field {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  cursor: pointer;
}

.checkbox {
  width: 1.125rem;
  height: 1.125rem;
  margin-top: 0.125rem;
  border-radius: 0.25rem;
  border: 2px solid var(--color-surface-container-high);
  accent-color: var(--color-primary);
  cursor: pointer;
}

.checkbox-label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.checkbox-title {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-on-background);
}

.checkbox-desc {
  font-size: 0.75rem;
  color: var(--color-on-surface-variant);
}

.delete-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
  border: none;
  background: transparent;
  color: var(--color-on-surface-variant);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.delete-btn:hover {
  background: oklch(0.6 0.2 25 / 0.1);
  color: oklch(0.5 0.2 25);
}

.delete-btn.delete-confirm-active {
  background: oklch(0.6 0.2 25);
  color: white;
}

.delete-confirm {
  padding: 1rem;
  border-radius: 0.5rem;
  background: oklch(0.6 0.2 25 / 0.08);
  border: 1px solid oklch(0.6 0.2 25 / 0.2);
}

.delete-warning {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: oklch(0.5 0.2 25);
  font-size: 0.875rem;
}

.error-message {
  padding: 0.875rem;
  margin-bottom: 1rem;
  border-radius: 0.5rem;
  background: oklch(0.6 0.2 25 / 0.1);
  color: oklch(0.5 0.2 25);
  font-size: 0.875rem;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.slide-enter-active,
.slide-leave-active {
  transition: all 0.2s ease;
}

.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  transform: translateY(-0.5rem);
}

.advanced-section {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--color-surface-container-high);
}

.advanced-toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0.5rem 0;
  border: none;
  background: transparent;
  cursor: pointer;
}

.toggle-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-on-surface-variant);
}

.advanced-content {
  margin-top: 1rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.prompt-layers {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.section-title {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-on-background);
  margin-bottom: 0.5rem;
}

.field-hint {
  font-size: 0.75rem;
  color: var(--color-on-surface-variant);
}

.inject-section {
  padding-top: 1rem;
  border-top: 1px solid var(--color-surface-container-high);
}

.inject-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.75rem;
}

.memory-section {
  padding-top: 1rem;
  border-top: 1px solid var(--color-surface-container-high);
}

.settings-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
}
</style>