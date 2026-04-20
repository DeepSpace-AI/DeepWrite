<script setup lang="ts">
import { watch, onMounted, onUnmounted } from 'vue'

const props = withDefaults(defineProps<{
  visible: boolean
  title?: string
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full'
  closable?: boolean
}>(), {
  size: 'md',
  closable: true,
})

const emit = defineEmits<{
  close: []
}>()

const sizeClass: Record<string, string> = {
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
  xl: 'max-w-2xl',
  full: 'max-w-4xl',
}

function handleClose() {
  emit('close')
}

function handleBackdropClick() {
  if (props.closable) {
    handleClose()
  }
}

function handleEscape(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.closable && props.visible) {
    handleClose()
  }
}

watch(() => props.visible, (visible) => {
  if (visible) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})

onMounted(() => {
  document.addEventListener('keydown', handleEscape)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleEscape)
  document.body.style.overflow = ''
})
</script>

<template>
  <Teleport to="body">
    <div
      class="d-modal"
      :class="{ 'd-modal-open': visible }"
      role="dialog"
      aria-modal="true"
    >
      <div class="d-modal-backdrop" @click="handleBackdropClick" />
      <div class="d-modal-box paper-card" :class="sizeClass[size]">
        <div v-if="title || closable" class="d-modal-header">
          <h3 v-if="title" class="d-modal-title">{{ title }}</h3>
          <button
            v-if="closable"
            class="d-modal-close"
            aria-label="Close"
            @click="handleClose"
          >
            <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="d-modal-content">
          <slot />
        </div>
        <div v-if="$slots.actions" class="d-modal-actions">
          <slot name="actions" />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.d-modal {
  position: fixed;
  inset: 0;
  z-index: 999;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  opacity: 0;
  visibility: hidden;
  transition: opacity 0.2s ease, visibility 0.2s ease;
}

.d-modal-open {
  opacity: 1;
  visibility: visible;
}

.d-modal-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
}

.d-modal-box {
  position: relative;
  width: 100%;
  max-height: calc(100vh - 4rem);
  overflow-y: auto;
  border-radius: 0.5rem;
  background: var(--surface-container-lowest);
  box-shadow: var(--shadow-lg);
}

.d-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--color-surface-container-high);
}

.d-modal-title {
  font-family: var(--font-serif);
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-on-background);
}

.d-modal-close {
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.375rem;
  padding: 0.5rem;
  color: var(--color-on-surface-variant);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: background 0.15s ease;
}

.d-modal-close:hover {
  background: var(--surface-container-high);
}

.d-modal-content {
  padding: 1.5rem;
}

.d-modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--color-surface-container-high);
}
</style>