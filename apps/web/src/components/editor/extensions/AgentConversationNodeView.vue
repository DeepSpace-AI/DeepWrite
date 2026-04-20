<script setup lang="ts">
import { computed, ref, watch, inject, type ComputedRef } from 'vue'
import { NodeViewWrapper, nodeViewProps } from '@tiptap/vue-3'
import { useI18n } from 'vue-i18n'
import { DButton } from '@/components/base'
import AgentConversation from '@/components/agent/AgentConversation.vue'
import IconRobot from '~icons/mdi/robot'
import IconChevronDown from '~icons/mdi/chevron-down'
import IconChevronUp from '~icons/mdi/chevron-up'
import IconDelete from '~icons/mdi/delete'
import IconDrag from '~icons/mdi/drag'

const props = defineProps(nodeViewProps)

const { t } = useI18n()

const injectedWorkspaceId = inject<ComputedRef<string>>('editorWorkspaceId', computed(() => ''))

const isCollapsed = ref(props.node.attrs.collapsed)

const agentId = computed(() => props.node.attrs.agentId)
const nodeWorkspaceId = computed(() => props.node.attrs.workspaceId)
const workspaceId = computed(() => nodeWorkspaceId.value || injectedWorkspaceId.value)
const sessionId = computed(() => props.node.attrs.sessionId)

watch(isCollapsed, (val) => {
  props.updateAttributes({ collapsed: val })
})

function toggleCollapse() {
  isCollapsed.value = !isCollapsed.value
}

function deleteNode() {
  props.deleteNode()
}

function handleSessionChange(newSessionId: string) {
  props.updateAttributes({ sessionId: newSessionId })
}
</script>

<template>
  <NodeViewWrapper
    class="agent-conversation-node my-4 rounded-lg border border-(--color-surface-container-high) bg-(--surface-container-low)"
    data-drag-handle
  >
    <div
      class="node-header flex items-center gap-2 px-4 py-2 border-b border-(--color-surface-container-high) bg-(--surface-container) cursor-grab"
      data-drag-handle
    >
      <IconDrag class="h-4 w-4 text-(--color-on-surface-variant) opacity-50" data-drag-handle />
      <IconRobot class="h-5 w-5 text-(--color-primary)" />
      <span class="text-sm font-medium text-(--color-on-background)">
        {{ t('agents.agentConversationNode') }}
      </span>
      <div class="flex-1" />
      <DButton size="xs" :title="isCollapsed ? t('common.expand') : t('common.collapse')" @click="toggleCollapse">
        <IconChevronDown v-if="isCollapsed" class="h-4 w-4" />
        <IconChevronUp v-else class="h-4 w-4" />
      </DButton>
      <DButton variant="danger" size="xs" :title="t('common.delete')" @click="deleteNode">
        <IconDelete class="h-4 w-4" />
      </DButton>
    </div>

    <div v-show="!isCollapsed" class="node-content">
      <AgentConversation
        v-if="workspaceId"
        :agent-id="agentId"
        :workspace-id="workspaceId"
        :embedded="true"
        :collapsible="false"
        @session-change="handleSessionChange"
      />
      <div v-else class="p-4 text-center text-sm text-(--color-on-surface-variant)">
        {{ t('agents.selectWorkspaceFirst') }}
      </div>
    </div>
  </NodeViewWrapper>
</template>

<style scoped>
.agent-conversation-node {
  position: relative;
}

.agent-conversation-node:hover .node-header {
  background-color: var(--surface-container-high);
}

.node-content {
  min-height: 300px;
  max-height: 500px;
  overflow: hidden;
}
</style>