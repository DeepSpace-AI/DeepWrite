<script setup lang="ts">
import { ref, computed } from 'vue'
import IconSend from '~icons/mdi/send'
import IconStop from '~icons/mdi/stop'

const props = defineProps<{
  disabled?: boolean
  isStreaming?: boolean
  placeholder?: string
}>()

const emit = defineEmits<{
  send: [content: string]
  abort: []
}>()

const inputRef = ref<HTMLTextAreaElement | null>(null)
const inputValue = ref('')

const canSend = computed(() => {
  return inputValue.value.trim() && !props.disabled && !props.isStreaming
})

function handleSubmit() {
  if (!canSend.value) return
  emit('send', inputValue.value.trim())
  inputValue.value = ''
  focusInput()
}

function handleKeyDown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSubmit()
  }
}

function handleAbort() {
  emit('abort')
}

function focusInput() {
  inputRef.value?.focus()
}

defineExpose({
  focusInput,
})
</script>

<template>
  <div class="chat-input">
    <div class="input-container">
      <textarea
        ref="inputRef"
        v-model="inputValue"
        :disabled="disabled"
        :placeholder="placeholder || 'Type a message...'"
        rows="1"
        class="input-field"
        @keydown="handleKeyDown"
      />
      
      <div class="input-actions">
        <button
          v-if="isStreaming"
          type="button"
          class="action-btn abort"
          @click="handleAbort"
        >
          <IconStop class="action-icon" />
        </button>
        <button
          v-else
          type="button"
          class="action-btn send"
          :disabled="!canSend"
          @click="handleSubmit"
        >
          <IconSend class="action-icon" />
        </button>
      </div>
    </div>
    
    <p class="input-hint">
      Press <kbd>Enter</kbd> to send, <kbd>Shift + Enter</kbd> for new line
    </p>
  </div>
</template>

<style scoped>
.chat-input {
  padding: 1rem 1.5rem 1.5rem;
}

.input-container {
  display: flex;
  align-items: flex-end;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background: var(--surface-container);
  border-radius: 1rem;
  border: 1px solid var(--color-surface-container-high);
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.input-container:focus-within {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 2px oklch(from var(--color-primary) l c h / 0.1);
}

.input-field {
  flex: 1;
  min-height: 1.5rem;
  max-height: 10rem;
  padding: 0;
  background: transparent;
  border: none;
  outline: none;
  resize: none;
  font-size: 0.9375rem;
  line-height: 1.5;
  color: var(--color-on-background);
}

.input-field::placeholder {
  color: var(--color-on-surface-variant);
  opacity: 0.6;
}

.input-field:disabled {
  opacity: 0.5;
}

.input-actions {
  display: flex;
  align-items: center;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  padding: 0;
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-btn.send {
  background: var(--color-primary);
  color: var(--color-on-primary);
}

.action-btn.send:hover:not(:disabled) {
  filter: brightness(1.1);
}

.action-btn.send:disabled {
  background: var(--surface-container-high);
  color: var(--color-on-surface-variant);
  opacity: 0.5;
  cursor: not-allowed;
}

.action-btn.abort {
  background: var(--color-error);
  color: white;
}

.action-btn.abort:hover {
  filter: brightness(1.1);
}

.action-icon {
  width: 1.25rem;
  height: 1.25rem;
}

.input-hint {
  margin: 0.5rem 0 0;
  font-size: 0.6875rem;
  color: var(--color-on-surface-variant);
  opacity: 0.6;
  text-align: center;
}

.input-hint kbd {
  display: inline-block;
  padding: 0.125rem 0.375rem;
  background: var(--surface-container-high);
  border-radius: 0.25rem;
  font-family: inherit;
  font-size: 0.625rem;
}
</style>