<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { Agent } from '@/api/agent'
import AgentMemoryPanel from './AgentMemoryPanel.vue'

const props = defineProps<{
  agent?: Agent | null
  workspaceId: string
  sessionId?: string
  contextDocs?: Array<{ id: string; name: string }>
  contextRefs?: Array<{ id: string; title: string }>
}>()

const emit = defineEmits<{
  'context-update': [docs: typeof props.contextDocs, refs: typeof props.contextRefs]
}>()

const { t } = useI18n()
</script>

<template>
  <div class="tools-panel space-y-6 p-4">
    <div>
      <h2 class="label-sm font-bold text-(--color-on-background)">{{ t('agents.agentTitle') }}</h2>
      <p class="label-sm mt-1 text-(--color-on-surface-variant)">
        {{ agent?.description || t('agents.agentSubtitle') }}
      </p>
    </div>

    <nav class="tabs flex gap-6 border-b border-(--color-surface-container-high) py-2">
      <button class="tab label-sm font-semibold text-(--color-tertiary)">
        {{ t('agents.tabs.context') }}
      </button>
      <button class="tab label-sm text-(--color-on-surface-variant)">
        {{ t('agents.tabs.history') }}
      </button>
    </nav>

    <div>
      <h4 class="label-sm mb-3 uppercase tracking-widest text-(--color-on-surface-variant)">
        {{ t('agents.activeDocs') }}
      </h4>
      <div v-if="contextDocs && contextDocs.length > 0" class="space-y-2">
        <div
          v-for="doc in contextDocs"
          :key="doc.id"
          class="doc-card rounded-md bg-(--surface-container-lowest) p-3"
        >
          <span class="text-sm font-medium text-(--color-tertiary)">{{ doc.name }}</span>
        </div>
      </div>
      <p v-else class="text-sm text-(--color-on-surface-variant)">
        {{ t('agents.noContextDocs') }}
      </p>
    </div>

    <div>
      <h4 class="label-sm mb-3 uppercase tracking-widest text-(--color-on-surface-variant)">
        {{ t('agents.activeRefs') }}
      </h4>
      <div v-if="contextRefs && contextRefs.length > 0" class="space-y-2">
        <div
          v-for="ref in contextRefs"
          :key="ref.id"
          class="ref-card rounded-md bg-(--surface-container-lowest) p-3"
        >
          <span class="text-sm font-medium text-(--color-on-background)">{{ ref.title }}</span>
        </div>
      </div>
      <p v-else class="text-sm text-(--color-on-surface-variant)">
        {{ t('agents.noContextRefs') }}
      </p>
    </div>

    <AgentMemoryPanel
      v-if="agent"
      :agent-id="agent.id"
      :workspace-id="workspaceId"
      :session-id="sessionId"
    />
  </div>
</template>

<style scoped>
.tab {
  position: relative;
  padding-bottom: 0.5rem;
}

.tab.font-semibold::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 2px;
  background-color: var(--color-tertiary);
}
</style>