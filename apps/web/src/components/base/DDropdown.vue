<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const props = withDefaults(defineProps<{
  align?: 'left' | 'right'
}>(), {
  align: 'right',
})

const isOpen = ref(false)
const triggerRef = ref<HTMLElement | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)

const position = ref({ top: 0, left: 0 })

function updatePosition() {
  if (!triggerRef.value) return
  const rect = triggerRef.value.getBoundingClientRect()
  
  position.value = {
    top: rect.bottom + 4,
    left: props.align === 'right' ? rect.right : rect.left,
  }
}

function toggle() {
  if (!isOpen.value) {
    updatePosition()
  }
  isOpen.value = !isOpen.value
}

function close() {
  isOpen.value = false
}

function handleClickOutside(e: MouseEvent) {
  const target = e.target as Node
  if (
    triggerRef.value &&
    dropdownRef.value &&
    !triggerRef.value.contains(target) &&
    !dropdownRef.value.contains(target)
  ) {
    close()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  window.addEventListener('scroll', updatePosition, true)
  window.addEventListener('resize', updatePosition)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('scroll', updatePosition, true)
  window.removeEventListener('resize', updatePosition)
})
</script>

<template>
  <div class="dropdown">
    <div ref="triggerRef" class="dropdown-trigger" @click.stop="toggle">
      <slot name="trigger" />
    </div>
    
    <Teleport to="body">
      <Transition name="dropdown">
        <div
          v-if="isOpen"
          ref="dropdownRef"
          class="dropdown-menu"
          :class="[`align-${align}`]"
          :style="{
            top: `${position.top}px`,
            left: align === 'left' ? `${position.left}px` : 'auto',
            right: align === 'right' ? `calc(100vw - ${position.left}px)` : 'auto',
          }"
          @click.stop
        >
          <slot name="content" :close="close" />
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.dropdown {
  position: relative;
  display: inline-flex;
}

.dropdown-trigger {
  display: inline-flex;
  cursor: pointer;
}

.dropdown-menu {
  position: fixed;
  z-index: 1000;
  min-width: 160px;
  max-width: 240px;
  padding: 0.375rem 0;
  background: var(--surface-container);
  border: 1px solid var(--color-surface-container-high);
  border-radius: 0.5rem;
  box-shadow: 0 4px 12px oklch(0 0 0 / 0.15);
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>