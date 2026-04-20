<script setup lang="ts">
import { computed, ref } from 'vue'
import type { ToolCallPart } from '@/types/chat'
import IconWrench from '~icons/mdi/wrench'
import IconCheck from '~icons/mdi/check-circle'
import IconAlert from '~icons/mdi/alert-circle'
import IconLoading from '~icons/mdi/loading'
import IconChevronDown from '~icons/mdi/chevron-down'
import IconChevronRight from '~icons/mdi/chevron-right'

const props = defineProps<{
  name: string
  args: Record<string, unknown>
  result?: unknown
  status: 'pending' | 'running' | 'done' | 'error'
}>()

const expanded = ref(false)

const statusIcon = computed(() => {
  switch (props.status) {
    case 'done': return IconCheck
    case 'error': return IconAlert
    case 'running': return IconLoading
    default: return IconWrench
  }
})

const statusClass = computed(() => {
  switch (props.status) {
    case 'done': return 'status-success'
    case 'error': return 'status-error'
    case 'running': return 'status-running'
    default: return 'status-pending'
  }
})

const formattedArgs = computed(() => {
  return JSON.stringify(props.args, null, 2)
})

const formattedResult = computed(() => {
  if (!props.result) return ''
  if (typeof props.result === 'string') return props.result
  return JSON.stringify(props.result, null, 2)
})
</script>

<template>
  <div class="tool-call" :class="statusClass">
    <button class="tool-header" @click="expanded = !expanded">
      <component :is="statusIcon" class="status-icon" :class="{ spinning: status === 'running' }" />
      <span class="tool-name">{{ name }}</span>
      <IconChevronDown v-if="expanded" class="chevron" />
      <IconChevronRight v-else class="chevron" />
    </button>

    <div v-if="expanded" class="tool-content">
      <div class="content-section">
        <span class="section-label">参数</span>
        <pre class="code-block">{{ formattedArgs }}</pre>
      </div>

      <div v-if="result" class="content-section">
        <span class="section-label">结果</span>
        <pre class="code-block result">{{ formattedResult }}</pre>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tool-call {
  margin: 0.5rem 0;
  border-radius: 0.5rem;
  border: 1px solid var(--color-surface-container-high);
  overflow: hidden;
}

.tool-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.625rem 0.75rem;
  background: var(--surface-container-low);
  border: none;
  cursor: pointer;
  font-size: 0.8125rem;
  transition: background 0.15s ease;
}

.tool-header:hover {
  background: var(--surface-container);
}

.status-icon {
  width: 1rem;
  height: 1rem;
  flex-shrink: 0;
}

.status-success .status-icon {
  color: var(--color-primary);
}

.status-error .status-icon {
  color: var(--color-error);
}

.status-running .status-icon {
  color: var(--color-tertiary);
}

.status-pending .status-icon {
  color: var(--color-on-surface-variant);
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.tool-name {
  flex: 1;
  text-align: left;
  font-weight: 500;
  color: var(--color-on-background);
}

.chevron {
  width: 1rem;
  height: 1rem;
  color: var(--color-on-surface-variant);
  opacity: 0.6;
}

.tool-content {
  padding: 0.5rem 0.75rem 0.75rem;
  border-top: 1px solid var(--color-surface-container-high);
}

.content-section {
  margin-top: 0.5rem;
}

.content-section:first-child {
  margin-top: 0;
}

.section-label {
  display: block;
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-on-surface-variant);
  margin-bottom: 0.25rem;
}

.code-block {
  margin: 0;
  padding: 0.5rem;
  background: var(--surface-container);
  border-radius: 0.25rem;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 0.75rem;
  line-height: 1.5;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-word;
  color: var(--color-on-background);
}

.code-block.result {
  background: oklch(0.45 0.02 150 / 0.1);
  border-left: 2px solid var(--color-primary);
}
</style>