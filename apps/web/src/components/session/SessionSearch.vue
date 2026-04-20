<script setup lang="ts">
import { ref, watch } from 'vue'
import IconSearch from '~icons/mdi/magnify'
import IconClose from '~icons/mdi/close'

const props = defineProps<{
  modelValue?: string
  placeholder?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  search: [query: string]
  clear: []
}>()

const inputValue = ref(props.modelValue || '')
let debounceTimer: ReturnType<typeof setTimeout> | null = null

watch(() => props.modelValue, (val) => {
  inputValue.value = val || ''
})

function handleInput() {
  emit('update:modelValue', inputValue.value)
  
  if (debounceTimer) {
    clearTimeout(debounceTimer)
  }
  
  debounceTimer = setTimeout(() => {
    emit('search', inputValue.value)
  }, 300)
}

function handleClear() {
  inputValue.value = ''
  emit('update:modelValue', '')
  emit('clear')
  emit('search', '')
}
</script>

<template>
  <div class="session-search">
    <IconSearch class="search-icon" />
    <input
      v-model="inputValue"
      type="text"
      class="search-input"
      :placeholder="placeholder || '搜索对话...'"
      @input="handleInput"
    />
    <button
      v-if="inputValue"
      type="button"
      class="clear-btn"
      @click="handleClear"
    >
      <IconClose class="h-4 w-4" />
    </button>
  </div>
</template>

<style scoped>
.session-search {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  background: var(--surface-container);
  border-radius: 0.5rem;
  border: 1px solid transparent;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.session-search:focus-within {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 2px oklch(from var(--color-primary) l c h / 0.1);
}

.search-icon {
  flex-shrink: 0;
  width: 1.125rem;
  height: 1.125rem;
  color: var(--color-on-surface-variant);
  opacity: 0.6;
}

.search-input {
  flex: 1;
  min-width: 0;
  padding: 0;
  background: transparent;
  border: none;
  outline: none;
  font-size: 0.8125rem;
  color: var(--color-on-background);
}

.search-input::placeholder {
  color: var(--color-on-surface-variant);
  opacity: 0.6;
}

.clear-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.25rem;
  height: 1.25rem;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: var(--surface-container-high);
  color: var(--color-on-surface-variant);
  cursor: pointer;
  transition: background 0.15s ease;
}

.clear-btn:hover {
  background: var(--surface-container-highest);
}
</style>