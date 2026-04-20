<script setup lang="ts">
import { computed } from 'vue'
import IconBrain from '~icons/mdi/brain'
import IconChevronDown from '~icons/mdi/chevron-down'
import IconChevronRight from '~icons/mdi/chevron-right'

const props = defineProps<{
  text: string
  collapsed?: boolean
}>()

const emit = defineEmits<{
  toggle: []
}>()

const previewText = computed(() => {
  if (!props.collapsed) return ''
  const maxLen = 100
  return props.text.length > maxLen 
    ? props.text.slice(0, maxLen) + '...'
    : props.text
})
</script>

<template>
  <div class="reasoning-part" :class="{ collapsed }">
    <button class="reasoning-header" @click="emit('toggle')">
      <IconBrain class="header-icon" />
      <span class="header-title">Thinking</span>
      <IconChevronDown v-if="!collapsed" class="chevron" />
      <IconChevronRight v-else class="chevron" />
    </button>
    
    <div v-if="collapsed && previewText" class="reasoning-preview">
      {{ previewText }}
    </div>
    
    <div v-if="!collapsed" class="reasoning-content">
      <p class="reasoning-text">{{ text }}</p>
    </div>
  </div>
</template>

<style scoped>
.reasoning-part {
  width: 100%;
  margin-bottom: 0.75rem;
  border-radius: 0.5rem;
  background: var(--surface-container-low);
  border: 1px solid var(--color-surface-container-high);
  overflow: hidden;
  box-sizing: border-box;
}

.reasoning-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.5rem 0.75rem;
  background: transparent;
  border: none;
  cursor: pointer;
  color: var(--color-on-surface-variant);
  font-size: 0.8125rem;
  font-weight: 500;
  transition: background 0.15s ease;
}

.reasoning-header:hover {
  background: var(--surface-container);
}

.header-icon {
  width: 1rem;
  height: 1rem;
  color: var(--color-tertiary);
}

.header-title {
  flex: 1;
  text-align: left;
}

.chevron {
  width: 1rem;
  height: 1rem;
  opacity: 0.6;
  transition: transform 0.2s ease;
}

.reasoning-preview {
  padding: 0 0.75rem 0.5rem;
  font-size: 0.75rem;
  color: var(--color-on-surface-variant);
  opacity: 0.7;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.reasoning-content {
  padding: 0 0.75rem 0.75rem;
  border-top: 1px solid var(--color-surface-container-high);
}

.reasoning-text {
  margin: 0;
  padding-top: 0.75rem;
  font-size: 0.875rem;
  line-height: 1.6;
  color: var(--color-on-surface-variant);
  white-space: pre-wrap;
}
</style>