<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  variant?: 'primary' | 'secondary' | 'tertiary' | 'ghost' | 'danger'
  size?: 'xs' | 'sm' | 'md' | 'lg'
  disabled?: boolean
  loading?: boolean
}>(), {
  variant: 'ghost',
  size: 'md',
  disabled: false,
  loading: false,
})

const emit = defineEmits<{
  click: [event: MouseEvent]
}>()

const classes = computed(() => [
  'd-btn',
  `d-btn-${props.variant}`,
  `d-btn-${props.size}`,
  {
    'd-btn-disabled': props.disabled || props.loading,
    'd-btn-loading': props.loading,
  },
])

function handleClick(e: MouseEvent) {
  if (props.disabled || props.loading) return
  emit('click', e)
}
</script>

<template>
  <button
    :class="classes"
    :disabled="disabled || loading"
    @click="handleClick"
  >
    <span v-if="loading" class="loading loading-spinner loading-xs mr-1" />
    <slot />
  </button>
</template>

<style scoped>
.d-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.375rem;
  font-family: var(--font-sans);
  font-weight: 500;
  transition: all 0.15s ease;
  cursor: pointer;
  border: none;
  outline: none;
  white-space: nowrap;
}

.d-btn:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.d-btn-primary {
  background: var(--color-primary);
  color: var(--color-on-primary);
}
.d-btn-primary:hover:not(.d-btn-disabled) {
  background: var(--color-primary-hover);
}

.d-btn-secondary {
  background: var(--color-secondary);
  color: var(--color-on-secondary);
}
.d-btn-secondary:hover:not(.d-btn-disabled) {
  background: var(--color-secondary-hover);
}

.d-btn-tertiary {
  background: transparent;
  color: var(--color-primary);
}
.d-btn-tertiary:hover:not(.d-btn-disabled) {
  background: var(--color-primary);
  color: var(--color-on-primary);
}

.d-btn-ghost {
  background: transparent;
  color: var(--color-on-surface-variant);
}
.d-btn-ghost:hover:not(.d-btn-disabled) {
  background: var(--surface-container-high);
}

.d-btn-danger {
  background: transparent;
  color: var(--color-error);
}
.d-btn-danger:hover:not(.d-btn-disabled) {
  background: var(--color-error);
  color: var(--color-on-error);
}

.d-btn-xs {
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  line-height: 1rem;
}

.d-btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.875rem;
  line-height: 1.25rem;
}

.d-btn-md {
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  line-height: 1.25rem;
}

.d-btn-lg {
  padding: 0.625rem 1.25rem;
  font-size: 1rem;
  line-height: 1.5rem;
}

.d-btn-disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.d-btn-loading {
  cursor: wait;
}
</style>