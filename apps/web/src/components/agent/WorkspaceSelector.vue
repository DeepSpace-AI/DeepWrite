<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { listWorkspaces, type Workspace } from '@/api/workspace'
import IconChevronDown from '~icons/mdi/chevron-down'
import IconFolder from '~icons/mdi/folder-outline'
import IconCheck from '~icons/mdi/check'

const props = withDefaults(defineProps<{
  selectedId?: string | null
  layout?: 'dropdown' | 'card'
}>(), {
  layout: 'dropdown',
})

const emit = defineEmits<{
  select: [workspaceId: string | null]
}>()

const { t } = useI18n()

const isOpen = ref(false)
const workspaces = ref<Workspace[]>([])
const isLoading = ref(false)

const selectedWorkspace = computed(() => {
  if (!props.selectedId) return null
  return workspaces.value.find(w => w.id === props.selectedId)
})

async function loadWorkspaces() {
  isLoading.value = true
  try {
    const res = await listWorkspaces()
    workspaces.value = res.items || []
  } catch (e) {
    console.error('Failed to load workspaces:', e)
  } finally {
    isLoading.value = false
  }
}

function handleSelect(workspaceId: string | null) {
  emit('select', workspaceId)
  isOpen.value = false
}

onMounted(() => {
  loadWorkspaces()
})
</script>

<template>
  <div v-if="layout === 'dropdown'" class="workspace-selector">
    <button
      class="selector-btn"
      @click="isOpen = !isOpen"
    >
      <IconFolder class="h-4 w-4" />
      <span class="selector-label">
        {{ selectedWorkspace?.name || t('agents.selectWorkspace') }}
      </span>
      <IconChevronDown class="chevron h-4 w-4" :class="{ rotated: isOpen }" />
    </button>

    <Transition name="fade">
      <div v-if="isOpen" class="selector-dropdown">
        <button
          class="dropdown-item"
          :class="{ selected: !selectedId }"
          @click="handleSelect(null)"
        >
          <IconFolder class="h-4 w-4 opacity-50" />
          <span>{{ t('agents.noWorkspace') }}</span>
          <IconCheck v-if="!selectedId" class="check-icon h-4 w-4" />
        </button>

        <div class="dropdown-divider" />

        <div class="dropdown-scroll">
          <button
            v-for="ws in workspaces"
            :key="ws.id"
            class="dropdown-item"
            :class="{ selected: ws.id === selectedId }"
            @click="handleSelect(ws.id)"
          >
            <IconFolder class="h-4 w-4" />
            <span>{{ ws.name }}</span>
            <IconCheck v-if="ws.id === selectedId" class="check-icon h-4 w-4" />
          </button>
        </div>
      </div>
    </Transition>
  </div>

  <div v-else class="workspace-card-selector">
    <button
      class="no-workspace-card"
      :class="{ selected: !selectedId }"
      @click="handleSelect(null)"
    >
      <IconFolder class="h-5 w-5 opacity-50" />
      <span>{{ t('agents.noWorkspace') }}</span>
    </button>

    <div class="workspace-cards">
      <button
        v-for="ws in workspaces.slice(0, 3)"
        :key="ws.id"
        class="workspace-card"
        :class="{ selected: ws.id === selectedId }"
        @click="handleSelect(ws.id)"
      >
        <IconFolder class="h-4 w-4" />
        <span>{{ ws.name }}</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.workspace-selector {
  position: relative;
}

.selector-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.375rem 0.625rem;
  border-radius: 0.5rem;
  border: 1px solid var(--color-surface-container-high);
  background: var(--surface-container);
  color: var(--color-on-surface-variant);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.selector-btn:hover {
  border-color: var(--color-primary);
}

.selector-label {
  max-width: 8rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chevron {
  transition: transform 0.15s ease;
}

.chevron.rotated {
  transform: rotate(180deg);
}

.selector-dropdown {
  position: absolute;
  top: calc(100% + 0.25rem);
  right: 0;
  z-index: 50;
  min-width: 12rem;
  border-radius: 0.5rem;
  border: 1px solid var(--color-surface-container-high);
  background: var(--surface-container-lowest);
  box-shadow: 0 8px 24px oklch(0 0 0 / 0.12);
  overflow: hidden;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.625rem 0.75rem;
  border: none;
  background: transparent;
  color: var(--color-on-background);
  font-size: 0.8125rem;
  text-align: left;
  cursor: pointer;
  transition: background 0.1s ease;
}

.dropdown-item:hover {
  background: var(--surface-container);
}

.dropdown-item.selected {
  color: var(--color-primary);
}

.check-icon {
  margin-left: auto;
}

.dropdown-divider {
  height: 1px;
  background: var(--color-surface-container-high);
}

.dropdown-scroll {
  max-height: 12rem;
  overflow-y: auto;
}

.workspace-card-selector {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.no-workspace-card {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 0.75rem;
  border-radius: 0.5rem;
  border: 1px solid var(--color-surface-container-high);
  background: var(--surface-container);
  color: var(--color-on-surface-variant);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.no-workspace-card:hover,
.no-workspace-card.selected {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.workspace-cards {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
}

.workspace-card {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.625rem;
  border-radius: 0.375rem;
  border: 1px solid var(--color-surface-container-high);
  background: var(--surface-container);
  color: var(--color-on-surface-variant);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.workspace-card:hover {
  border-color: var(--color-primary);
}

.workspace-card.selected {
  border-color: var(--color-primary);
  background: var(--color-primary-container);
  color: var(--color-on-primary-container);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-0.25rem);
}
</style>