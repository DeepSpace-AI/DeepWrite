<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAgentStore } from '@/stores/agent'
import { DModal, DButton } from '@/components/base'
import IconResearch from '~icons/mdi/beaker'
import IconWrite from '~icons/mdi/pencil-outline'
import IconData from '~icons/mdi/database'
import IconPublish from '~icons/mdi/book-open-page-variant'
import IconGeneral from '~icons/mdi/assistant'
import IconCheck from '~icons/mdi/check'
import IconSparkles from '~icons/mdi/auto-fix'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  close: []
  created: [agent: ReturnType<typeof useAgentStore>['currentAgent']['value']]
}>()

const { t } = useI18n()
const agentStore = useAgentStore()

const currentStep = ref(0)
const isSaving = ref(false)
const error = ref('')

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
    desc: t('agents.categoryDesc.research')
  },
  { 
    value: 'writing', 
    label: t('agents.categories.writing'), 
    icon: IconWrite,
    color: 'sage',
    desc: t('agents.categoryDesc.writing')
  },
  { 
    value: 'data', 
    label: t('agents.categories.data'), 
    icon: IconData,
    color: 'amber',
    desc: t('agents.categoryDesc.data')
  },
  { 
    value: 'publishing', 
    label: t('agents.categories.publishing'), 
    icon: IconPublish,
    color: 'plum',
    desc: t('agents.categoryDesc.publishing')
  },
  { 
    value: 'general', 
    label: t('agents.categories.general'), 
    icon: IconGeneral,
    color: 'slate',
    desc: t('agents.categoryDesc.general')
  },
]

const steps = [
  { key: 'basics', title: t('agents.stepBasics'), desc: t('agents.stepBasicsDesc') },
  { key: 'prompt', title: t('agents.stepPrompt'), desc: t('agents.stepPromptDesc') },
  { key: 'advanced', title: t('agents.stepAdvanced'), desc: t('agents.stepAdvancedDesc') },
  { key: 'settings', title: t('agents.stepSettings'), desc: t('agents.stepSettingsDesc') },
]

const isValid = computed(() => {
  if (currentStep.value === 0) {
    return form.value.name.trim().length >= 2
  }
  if (currentStep.value === 1) {
    return form.value.system_prompt.trim().length >= 20
  }
  return true
})

const canProceed = computed(() => isValid.value && currentStep.value < steps.length - 1)
const canComplete = computed(() => isValid.value && currentStep.value === steps.length - 1)

function nextStep() {
  if (canProceed.value) {
    currentStep.value++
  }
}

function prevStep() {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

function resetForm() {
  currentStep.value = 0
  form.value = {
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
  }
  error.value = ''
}

async function handleCreate() {
  if (!canComplete.value || isSaving.value) return

  isSaving.value = true
  error.value = ''

  try {
    const agent = await agentStore.createNewAgent({
      name: form.value.name,
      system_prompt: form.value.system_prompt,
      description: form.value.description,
      category: form.value.category,
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
    emit('created', agent)
    resetForm()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to create agent'
  } finally {
    isSaving.value = false
  }
}

function handleClose() {
  resetForm()
  emit('close')
}

watch(() => props.visible, (visible) => {
  if (!visible) {
    resetForm()
  }
})
</script>

<template>
  <DModal
    :visible="visible"
    :title="t('agents.createAgent')"
    size="xl"
    @close="handleClose"
  >
    <div class="creator-content">
      <div class="steps-indicator">
        <div
          v-for="(step, index) in steps"
          :key="step.key"
          class="step-item"
          :class="{ active: index === currentStep, completed: index < currentStep }"
        >
          <div class="step-number">
            <IconCheck v-if="index < currentStep" class="h-3 w-3" />
            <span v-else>{{ index + 1 }}</span>
          </div>
          <div class="step-info">
            <span class="step-title">{{ step.title }}</span>
          </div>
        </div>
      </div>

      <Transition name="fade-slide" mode="out-in">
        <div :key="currentStep" class="step-content">
          <TransitionGroup name="fade" tag="div">
            <div v-if="error" key="error" class="error-message">
              {{ error }}
            </div>
          </TransitionGroup>

          <div v-if="currentStep === 0" class="form-section">
            <div class="field">
              <label class="field-label">{{ t('agents.form.name') }} *</label>
              <input
                v-model="form.name"
                type="text"
                class="field-input"
                :placeholder="t('agents.form.namePlaceholder')"
              >
            </div>

            <div class="field">
              <label class="field-label">{{ t('agents.form.category') }}</label>
              <div class="category-grid">
                <button
                  v-for="cat in categoryOptions"
                  :key="cat.value"
                  class="category-card"
                  :class="{ selected: form.category === cat.value, [cat.color]: true }"
                  @click="form.category = cat.value"
                >
                  <div class="category-icon">
                    <component :is="cat.icon" class="h-5 w-5" />
                  </div>
                  <div class="category-text">
                    <span class="category-label">{{ cat.label }}</span>
                    <span class="category-desc">{{ cat.desc }}</span>
                  </div>
                </button>
              </div>
            </div>

            <div class="field">
              <label class="field-label">{{ t('agents.form.description') }}</label>
              <textarea
                v-model="form.description"
                rows="2"
                class="field-input field-textarea"
                :placeholder="t('agents.form.descriptionPlaceholder')"
              />
            </div>
          </div>

          <div v-else-if="currentStep === 1" class="form-section">
            <div class="field">
              <label class="field-label">{{ t('agents.form.systemPrompt') }} *</label>
              <p class="field-hint">{{ t('agents.form.systemPromptHint') }}</p>
              <textarea
                v-model="form.system_prompt"
                rows="12"
                class="field-input field-textarea field-mono"
                :placeholder="t('agents.form.systemPromptPlaceholder')"
              />
              <div class="field-counter">
                {{ form.system_prompt.length }} {{ t('agents.characters') }}
              </div>
            </div>

            <div class="prompt-suggestions">
              <button class="suggestion-btn">
                <IconSparkles class="h-4 w-4" />
                <span>{{ t('agents.generatePrompt') }}</span>
              </button>
            </div>
          </div>

          <div v-else-if="currentStep === 2" class="form-section">
            <p class="section-hint">{{ t('agents.stepAdvancedDesc') }}</p>
            
            <div class="prompt-layers">
              <div class="field">
                <label class="field-label">{{ t('agents.promptLayers.identity') }}</label>
                <p class="field-hint">{{ t('agents.promptLayers.identityDesc') }}</p>
                <textarea
                  v-model="form.identity_prompt"
                  rows="3"
                  class="field-input field-textarea field-mono"
                  placeholder="e.g. You are an AI research assistant specialized in..."
                />
              </div>

              <div class="field">
                <label class="field-label">{{ t('agents.promptLayers.capability') }}</label>
                <p class="field-hint">{{ t('agents.promptLayers.capabilityDesc') }}</p>
                <textarea
                  v-model="form.capability_prompt"
                  rows="3"
                  class="field-input field-textarea field-mono"
                  placeholder="e.g. You can: analyze literature, summarize papers..."
                />
              </div>

              <div class="field">
                <label class="field-label">{{ t('agents.promptLayers.instruction') }}</label>
                <p class="field-hint">{{ t('agents.promptLayers.instructionDesc') }}</p>
                <textarea
                  v-model="form.instruction_prompt"
                  rows="3"
                  class="field-input field-textarea field-mono"
                  placeholder="e.g. Always cite sources, use academic tone..."
                />
              </div>

              <div class="field">
                <label class="field-label">{{ t('agents.promptLayers.safety') }}</label>
                <p class="field-hint">{{ t('agents.promptLayers.safetyDesc') }}</p>
                <textarea
                  v-model="form.safety_prompt"
                  rows="3"
                  class="field-input field-textarea field-mono"
                  placeholder="e.g. Never generate harmful content..."
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

          <div v-else-if="currentStep === 3" class="form-section">
            <div class="settings-grid">
              <div class="field">
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
                <p class="field-hint">{{ t('agents.form.temperatureHint') }}</p>
              </div>

              <div class="field">
                <label class="field-label">{{ t('agents.form.maxTokens') }}</label>
                <input
                  v-model.number="form.max_tokens"
                  type="number"
                  min="1"
                  max="128000"
                  class="field-input"
                >
                <p class="field-hint">{{ t('agents.form.maxTokensHint') }}</p>
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
          </div>
        </div>
      </Transition>
    </div>

    <template #actions>
      <DButton v-if="currentStep > 0" variant="ghost" @click="prevStep">
        {{ t('common.back') }}
      </DButton>
      <div class="flex-1" />
      <DButton variant="ghost" @click="handleClose">
        {{ t('common.cancel') }}
      </DButton>
      <DButton
        v-if="canProceed"
        variant="primary"
        :disabled="!isValid"
        @click="nextStep"
      >
        {{ t('common.next') }}
      </DButton>
      <DButton
        v-if="canComplete"
        variant="primary"
        :disabled="!isValid"
        :loading="isSaving"
        @click="handleCreate"
      >
        {{ t('common.create') }}
      </DButton>
    </template>
  </DModal>
</template>

<style scoped>
.creator-content {
  min-height: 24rem;
}

.steps-indicator {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 2rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid var(--color-surface-container-high);
}

.step-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.5rem 0.75rem;
  border-radius: 0.5rem;
  transition: all 0.2s ease;
}

.step-item.active {
  background: var(--color-primary-container);
}

.step-item.completed {
  opacity: 0.7;
}

.step-number {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.5rem;
  height: 1.5rem;
  border-radius: 50%;
  background: var(--surface-container-high);
  color: var(--color-on-surface-variant);
  font-size: 0.75rem;
  font-weight: 600;
  transition: all 0.2s ease;
}

.step-item.active .step-number {
  background: var(--color-primary);
  color: white;
}

.step-item.completed .step-number {
  background: var(--color-primary);
  color: white;
}

.step-title {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-on-surface-variant);
}

.step-item.active .step-title {
  color: var(--color-on-background);
}

.step-content {
  animation: fadeIn 0.25s ease;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.field-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-on-background);
}

.field-hint {
  font-size: 0.75rem;
  color: var(--color-on-surface-variant);
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

.category-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
  gap: 0.75rem;
}

.category-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.875rem;
  border-radius: 0.625rem;
  border: 2px solid var(--color-surface-container-high);
  background: var(--surface-container-low);
  text-align: left;
  cursor: pointer;
  transition: all 0.15s ease;
}

.category-card:hover {
  border-color: var(--color-primary);
}

.category-card.selected {
  border-color: var(--color-primary);
  background: var(--color-primary-container);
}

.category-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 0.5rem;
  flex-shrink: 0;
}

.category-card.terracotta .category-icon {
  background: oklch(0.55 0.12 30 / 0.15);
  color: oklch(0.55 0.12 30);
}

.category-card.sage .category-icon {
  background: oklch(0.55 0.08 130 / 0.15);
  color: oklch(0.55 0.08 130);
}

.category-card.amber .category-icon {
  background: oklch(0.65 0.12 85 / 0.15);
  color: oklch(0.55 0.12 85);
}

.category-card.plum .category-icon {
  background: oklch(0.55 0.12 320 / 0.15);
  color: oklch(0.55 0.12 320);
}

.category-card.slate .category-icon {
  background: oklch(0.55 0.02 250 / 0.15);
  color: oklch(0.45 0.02 250);
}

.category-text {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-width: 0;
}

.category-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-on-background);
}

.category-desc {
  font-size: 0.6875rem;
  color: var(--color-on-surface-variant);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.settings-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.5rem;
}

.slider-field {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.slider {
  flex: 1;
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
  min-width: 2.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-on-background);
}

.prompt-suggestions {
  display: flex;
  gap: 0.5rem;
}

.suggestion-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.875rem;
  border-radius: 0.5rem;
  border: 1px solid var(--color-surface-container-high);
  background: var(--surface-container);
  color: var(--color-on-surface-variant);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.suggestion-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
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

.error-message {
  padding: 0.875rem;
  margin-bottom: 1rem;
  border-radius: 0.5rem;
  background: oklch(0.6 0.2 25 / 0.1);
  color: oklch(0.5 0.2 25);
  font-size: 0.875rem;
}

.section-hint {
  font-size: 0.75rem;
  color: var(--color-on-surface-variant);
  margin-bottom: 1.5rem;
}

.prompt-layers {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.section-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-on-background);
  margin-bottom: 0.75rem;
}

.inject-section {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--color-surface-container-high);
}

.inject-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.75rem;
}

.memory-section {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--color-surface-container-high);
}

.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.2s ease;
}

.fade-slide-enter-from {
  opacity: 0;
  transform: translateX(1rem);
}

.fade-slide-leave-to {
  opacity: 0;
  transform: translateX(-1rem);
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
</style>