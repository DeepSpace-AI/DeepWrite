<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Session } from '@/api/agent'
import { DButton } from '@/components/base'
import { SessionList, SessionSearch } from '@/components/session'
import IconPlus from '~icons/mdi/plus'
import IconArchive from '~icons/mdi/archive-outline'
import IconInbox from '~icons/mdi/inbox'

defineProps<{
  sessions: Session[]
  archivedSessions?: Session[]
  currentSessionId?: string
  collapsed?: boolean
}>()

const emit = defineEmits<{
  select: [sessionId: string]
  new: []
  pin: [sessionId: string]
  unpin: [sessionId: string]
  archive: [sessionId: string]
  unarchive: [sessionId: string]
  delete: [sessionId: string]
  export: [sessionId: string, format: 'markdown' | 'json']
  share: [sessionId: string]
  search: [query: string]
}>()

const { t } = useI18n()
const showArchived = ref(false)
const searchQuery = ref('')

function handleSearch(query: string) {
  searchQuery.value = query
  emit('search', query)
}

function handleClearSearch() {
  searchQuery.value = ''
  emit('search', '')
}
</script>

<template>
  <div class="sidebar-content">
    <div class="sidebar-header">
      <DButton size="sm" :title="t('agents.newSession')" @click="emit('new')">
        <IconPlus class="h-5 w-5" />
      </DButton>
      
      <DButton
        v-if="!showArchived"
        size="sm"
        variant="ghost"
        title="归档"
        @click="showArchived = true"
      >
        <IconArchive class="h-5 w-5" />
      </DButton>
      
      <DButton
        v-else
        size="sm"
        variant="ghost"
        title="返回"
        @click="showArchived = false"
      >
        <IconInbox class="h-5 w-5" />
      </DButton>
    </div>

    <SessionSearch
      v-if="!showArchived"
      class="search-wrapper"
      :model-value="searchQuery"
      @search="handleSearch"
      @clear="handleClearSearch"
    />

    <div class="session-list-wrapper">
      <SessionList
        v-if="!showArchived"
        :sessions="sessions"
        :current-session-id="currentSessionId"
        @select="emit('select', $event)"
        @pin="emit('pin', $event)"
        @unpin="emit('unpin', $event)"
        @archive="emit('archive', $event)"
        @delete="emit('delete', $event)"
        @export="(id, fmt) => emit('export', id, fmt)"
        @share="emit('share', $event)"
      />
      
      <SessionList
        v-else
        :sessions="archivedSessions || []"
        :current-session-id="currentSessionId"
        show-archived
        @select="emit('select', $event)"
        @unarchive="emit('unarchive', $event)"
        @delete="emit('delete', $event)"
      />
    </div>
  </div>
</template>

<style scoped>
.sidebar-content {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 0.5rem;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.search-wrapper {
  margin-bottom: 0.5rem;
}

.session-list-wrapper {
  flex: 1;
  overflow-y: auto;
}
</style>