<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  variant?: 'glass' | 'ghost' | 'bordered' | 'filled'
  size?: 'sm' | 'md' | 'lg'
  placeholder?: string
  disabled?: boolean
  error?: boolean
  type?: string
  modelValue?: string
}>(), {
  variant: 'glass',
  size: 'md',
  disabled: false,
  error: false,
  type: 'text',
  modelValue: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  focus: [event: FocusEvent]
  blur: [event: FocusEvent]
}>()

const classes = computed(() => [
  'd-input',
  `d-input-${props.variant}`,
  `d-input-${props.size}`,
  {
    'd-input-disabled': props.disabled,
    'd-input-error': props.error,
  },
])

function handleInput(e: Event) {
  const target = e.target as HTMLInputElement
  emit('update:modelValue', target.value)
}
</script>

<template>
  <input
    :class="classes"
    :type="type"
    :placeholder="placeholder"
    :disabled="disabled"
    :value="modelValue"
    @input="handleInput"
    @focus="emit('focus', $event)"
    @blur="emit('blur', $event)"
  >
</template>

<style scoped>
.d-input {
  width: 100%;
  border-radius: 0.375rem;
  font-family: var(--font-sans);
  color: var(--color-on-background);
  transition: all 0.15s ease;
  outline: none;
}

.d-input::placeholder {
  color: var(--color-on-surface-variant);
  opacity: 0.6;
}

.d-input:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.d-input-glass {
  background: var(--surface-container);
  border: none;
  border-left: 2px solid transparent;
}
.d-input-glass:focus {
  background: var(--surface-overlay);
  border-left-color: var(--color-primary);
}

.d-input-ghost {
  background: transparent;
  border: 1px solid transparent;
}
.d-input-ghost:hover:not(.d-input-disabled) {
  background: var(--surface-container);
}
.d-input-ghost:focus {
  background: var(--surface-overlay);
  border-color: var(--color-surface-container-high);
}

.d-input-bordered {
  background: transparent;
  border: 1px solid var(--color-surface-container-high);
}
.d-input-bordered:hover:not(.d-input-disabled) {
  border-color: var(--border-subtle);
}
.d-input-bordered:focus {
  border-color: var(--color-primary);
}

.d-input-filled {
  background: var(--surface-container);
  border: none;
}
.d-input-filled:hover:not(.d-input-disabled) {
  background: var(--surface-container-high);
}
.d-input-filled:focus {
  background: var(--surface-overlay);
}

.d-input-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.875rem;
  line-height: 1.25rem;
}

.d-input-md {
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  line-height: 1.25rem;
}

.d-input-lg {
  padding: 0.625rem 1rem;
  font-size: 1rem;
  line-height: 1.5rem;
}

.d-input-disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background: var(--surface-container-high);
}

.d-input-error {
  border-color: var(--color-error);
}
.d-input-error:focus {
  border-color: var(--color-error);
  outline-color: var(--color-error);
}
</style>