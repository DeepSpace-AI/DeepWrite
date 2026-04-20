<script setup lang="ts">
import { ref } from 'vue'
import type { Agent } from '@/api/agent'
import IconRobot from '~icons/mdi/robot'
import IconChevronDown from '~icons/mdi/chevron-down'
import IconResearch from '~icons/mdi/beaker'
import IconWrite from '~icons/mdi/pencil-outline'
import IconData from '~icons/mdi/database'
import IconPublish from '~icons/mdi/book-open-page-variant'

const props = defineProps<{
  agents: Agent[]
  currentAgent?: Agent | null
}>()

const emit = defineEmits<{
  select: [agent: Agent]
}>()

const isOpen = ref(false)

const categoryIcons: Record<string, typeof IconResearch> = {
  research: IconResearch,
  writing: IconWrite,
  data: IconData,
  publishing: IconPublish,
}

function getCategoryIcon(category: string) {
  return categoryIcons[category] || IconRobot
}

function selectAgent(agent: Agent) {
  emit('select', agent)
  isOpen.value = false
}
</script>

<template>
  <div class="agent-selector relative">
    <button
      class="selector-trigger flex items-center gap-2 rounded-md bg-(--surface-container) px-3 py-2 transition-colors hover:bg-(--surface-container-high)"
      @click="isOpen = !isOpen"
    >
      <div class="flex h-6 w-6 items-center justify-center rounded bg-(--color-primary)/10">
        <component
          :is="currentAgent ? getCategoryIcon(currentAgent.category) : IconRobot"
          class="h-4 w-4 text-(--color-primary)"
        />
      </div>
      <span class="text-sm font-medium text-(--color-on-background)">{{ currentAgent?.name || 'Select Agent' }}</span>
      <IconChevronDown
        class="h-4 w-4 text-(--color-on-surface-variant) transition-transform"
        :class="{ 'rotate-180': isOpen }"
      />
    </button>

    <div
      v-if="isOpen"
      class="selector-dropdown absolute left-0 top-full z-50 mt-1 max-h-80 w-64 overflow-y-auto rounded-md bg-(--surface-container-lowest) shadow-lg"
    >
      <div class="p-2">
        <div
          v-for="agent in agents"
          :key="agent.id"
          class="agent-option flex cursor-pointer items-center gap-2 rounded-md p-2 transition-colors hover:bg-(--surface-container)"
          :class="{ 'bg-(--color-primary-container)': agent.id === currentAgent?.id }"
          @click="selectAgent(agent)"
        >
          <div class="flex h-8 w-8 items-center justify-center rounded bg-(--color-primary)/10">
            <component
              :is="getCategoryIcon(agent.category)"
              class="h-4 w-4 text-(--color-primary)"
            />
          </div>
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-medium text-(--color-on-background)">{{ agent.name }}</p>
            <p class="truncate text-xs text-(--color-on-surface-variant)">{{ agent.description }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.selector-dropdown {
  box-shadow: 0 8px 32px oklch(0.28 0.008 105 / 0.12);
}
</style>