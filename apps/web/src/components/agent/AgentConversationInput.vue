<script setup lang="ts">
import { ref, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Agent } from '@/api/agent'
import IconSend from '~icons/mdi/send'
import IconAttach from '~icons/mdi/paperclip'

const props = defineProps<{
  disabled?: boolean
  agent?: Agent | null
}>()

const emit = defineEmits<{
  send: [content: string]
}>()

const { t } = useI18n()
const inputMessage = ref('')
const textareaRef = ref<HTMLTextAreaElement | null>(null)

function adjustHeight() {
  nextTick(() => {
    if (textareaRef.value) {
      textareaRef.value.style.height = 'auto'
      textareaRef.value.style.height = Math.min(textareaRef.value.scrollHeight, 200) + 'px'
    }
  })
}

function handleSend() {
  const content = inputMessage.value.trim()
  if (!content || props.disabled) return
  
  emit('send', content)
  inputMessage.value = ''
  adjustHeight()
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
}

watch(inputMessage, adjustHeight)
</script>

<template>
  <div class="input-area">
    <div class="input-container">
      <div class="input-row">
        <div class="input-field">
          <label class="input-label">
            {{ agent ? t('agents.queryLabelWith', { name: agent.name }) : t('agents.queryLabel') }}
          </label>
          <textarea
            ref="textareaRef"
            v-model="inputMessage"
            class="input-textarea"
            :placeholder="t('agents.queryPlaceholder')"
            rows="1"
            :disabled="disabled"
            @keydown="handleKeydown"
          />
        </div>
        <div class="input-actions">
          <button
            class="attach-btn"
            :disabled="disabled"
            :title="t('agents.attachFile')"
          >
            <IconAttach class="h-5 w-5" />
          </button>
          <button
            class="send-btn"
            :class="{ disabled: disabled || !inputMessage.trim() }"
            :disabled="disabled || !inputMessage.trim()"
            @click="handleSend"
          >
            <IconSend class="h-5 w-5" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.input-area {
  border-top: 1px solid var(--color-surface-container-high);
  background: linear-gradient(to top, var(--surface) 70%, transparent);
  padding: 1rem;
}

.input-container {
  max-width: 48rem;
  margin: 0 auto;
  border-radius: 0.75rem;
  background: var(--surface-container-lowest);
  padding: 1rem;
  box-shadow: 0 8px 32px oklch(0.28 0.008 105 / 0.08);
}

.input-row {
  display: flex;
  align-items: flex-end;
  gap: 0.75rem;
}

.input-field {
  flex: 1;
  min-width: 0;
}

.input-label {
  display: block;
  margin-bottom: 0.25rem;
  margin-left: 0.25rem;
  font-size: 0.75rem;
  color: var(--color-on-surface-variant);
}

.input-textarea {
  width: 100%;
  padding: 0.25rem;
  background: transparent;
  border: none;
  outline: none;
  resize: none;
  font-size: 0.9375rem;
  line-height: 1.5;
  color: var(--color-on-background);
}

.input-textarea::placeholder {
  color: var(--color-on-surface-variant);
  opacity: 0.6;
}

.input-textarea:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.input-actions {
  display: flex;
  gap: 0.5rem;
  padding-bottom: 0.25rem;
}

.attach-btn,
.send-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
  border: none;
  cursor: pointer;
  transition: all 0.15s ease;
}

.attach-btn {
  padding: 0.5rem;
  background: transparent;
  color: var(--color-on-surface-variant);
}

.attach-btn:hover:not(:disabled) {
  background: var(--surface-container-high);
}

.send-btn {
  padding: 0.625rem;
  background: linear-gradient(135deg, var(--color-primary), oklch(0.45 0.1 30));
  color: white;
}

.send-btn:hover:not(:disabled) {
  transform: scale(1.05);
}

.attach-btn:disabled,
.send-btn.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>