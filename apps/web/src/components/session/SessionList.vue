<script setup lang="ts">
import { computed } from 'vue'
import type { Session } from '@/api/agent'
import { DButton, DDropdown } from '@/components/base'
import IconPin from '~icons/mdi/pin'
import IconPinOff from '~icons/mdi/pin-off'
import IconArchive from '~icons/mdi/archive'
import IconArchiveOutline from '~icons/mdi/archive-outline'
import IconDelete from '~icons/mdi/delete'
import IconDownload from '~icons/mdi/download'
import IconShare from '~icons/mdi/share-variant'
import IconDotsVertical from '~icons/mdi/dots-vertical'

const props = defineProps<{
  sessions: Session[]
  currentSessionId?: string
  showArchived?: boolean
}>()

const emit = defineEmits<{
  select: [sessionId: string]
  pin: [sessionId: string]
  unpin: [sessionId: string]
  archive: [sessionId: string]
  unarchive: [sessionId: string]
  delete: [sessionId: string]
  export: [sessionId: string, format: 'markdown' | 'json']
  share: [sessionId: string]
}>()

const sortedSessions = computed(() => {
  return [...props.sessions].sort((a, b) => {
    if (a.pinned !== b.pinned) {
      return a.pinned ? -1 : 1
    }
    return new Date(b.last_message_at).getTime() - new Date(a.last_message_at).getTime()
  })
})

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  
  if (days === 0) {
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  } else if (days === 1) {
    return '昨天'
  } else if (days < 7) {
    return `${days}天前`
  } else {
    return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
  }
}

function handleSelect(session: Session) {
  emit('select', session.id)
}

function handlePinToggle(session: Session) {
  if (session.pinned) {
    emit('unpin', session.id)
  } else {
    emit('pin', session.id)
  }
}
</script>

<template>
  <div class="session-list">
    <TransitionGroup name="list">
      <div
        v-for="session in sortedSessions"
        :key="session.id"
        class="session-item"
        :class="{ active: session.id === currentSessionId, pinned: session.pinned }"
        @click="handleSelect(session)"
      >
        <div class="session-content">
          <div class="session-title">
            <IconPin v-if="session.pinned" class="pin-icon" />
            <span class="title-text">{{ session.title || '新对话' }}</span>
          </div>
          <div class="session-meta">
            <span class="session-date">{{ formatDate(session.last_message_at) }}</span>
          </div>
        </div>

        <div class="session-actions" @click.stop>
          <DDropdown align="right">
            <template #trigger>
              <button class="action-trigger">
                <IconDotsVertical class="trigger-icon" />
              </button>
            </template>
            <template #content="{ close }">
              <div class="menu-list">
                <button
                  v-if="!showArchived"
                  class="menu-item"
                  @click="handlePinToggle(session); close()"
                >
                  <IconPinOff v-if="session.pinned" class="menu-icon" />
                  <IconPin v-else class="menu-icon" />
                  <span>{{ session.pinned ? '取消置顶' : '置顶' }}</span>
                </button>
                
                <button
                  v-if="!showArchived"
                  class="menu-item"
                  @click="emit('archive', session.id); close()"
                >
                  <IconArchiveOutline class="menu-icon" />
                  <span>归档</span>
                </button>
                
                <button
                  v-if="showArchived"
                  class="menu-item"
                  @click="emit('unarchive', session.id); close()"
                >
                  <IconArchive class="menu-icon" />
                  <span>恢复</span>
                </button>

                <button
                  v-if="!showArchived"
                  class="menu-item"
                  @click="emit('export', session.id, 'markdown'); close()"
                >
                  <IconDownload class="menu-icon" />
                  <span>导出 Markdown</span>
                </button>

                <button
                  v-if="!showArchived"
                  class="menu-item"
                  @click="emit('share', session.id); close()"
                >
                  <IconShare class="menu-icon" />
                  <span>分享</span>
                </button>
                
                <div class="menu-divider" />
                
                <button
                  class="menu-item danger"
                  @click="emit('delete', session.id); close()"
                >
                  <IconDelete class="menu-icon" />
                  <span>删除</span>
                </button>
              </div>
            </template>
          </DDropdown>
        </div>
      </div>
    </TransitionGroup>

    <div v-if="sessions.length === 0" class="empty-state">
      <p>{{ showArchived ? '没有已归档的对话' : '暂无对话' }}</p>
    </div>
  </div>
</template>

<style scoped>
.session-list {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.session-item {
  position: relative;
  display: flex;
  align-items: center;
  padding: 0.625rem 0.75rem;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.session-item:hover {
  background: var(--surface-container);
}

.session-item.active {
  background: var(--surface-container-high);
}

.session-item.pinned {
  background: oklch(from var(--color-primary) l c h / 0.05);
}

.session-item.active.pinned {
  background: oklch(from var(--color-primary) l c h / 0.1);
}

.session-content {
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.session-title {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.pin-icon {
  flex-shrink: 0;
  width: 0.875rem;
  height: 0.875rem;
  color: var(--color-primary);
}

.title-text {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-on-background);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.session-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.125rem;
}

.session-date {
  font-size: 0.6875rem;
  color: var(--color-on-surface-variant);
}

.session-actions {
  position: absolute;
  right: 0.375rem;
  top: 50%;
  transform: translateY(-50%);
  opacity: 0;
  transition: opacity 0.15s ease;
}

.session-item:hover .session-actions,
.session-item.active .session-actions {
  opacity: 1;
}

.action-trigger {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.75rem;
  height: 1.75rem;
  padding: 0;
  border: none;
  border-radius: 0.375rem;
  background: transparent;
  color: var(--color-on-surface-variant);
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-trigger:hover {
  background: var(--surface-container-high);
  color: var(--color-on-background);
}

.trigger-icon {
  width: 1.125rem;
  height: 1.125rem;
}

.menu-list {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: none;
  background: transparent;
  font-size: 0.8125rem;
  color: var(--color-on-background);
  cursor: pointer;
  text-align: left;
  transition: background 0.15s ease;
}

.menu-item:hover {
  background: var(--surface-container-high);
}

.menu-item.danger {
  color: var(--color-error);
}

.menu-icon {
  width: 1rem;
  height: 1rem;
  flex-shrink: 0;
  color: var(--color-on-surface-variant);
}

.menu-item.danger .menu-icon {
  color: var(--color-error);
}

.menu-divider {
  height: 1px;
  margin: 0.25rem 0.5rem;
  background: var(--color-surface-container-high);
}

.empty-state {
  padding: 2rem;
  text-align: center;
}

.empty-state p {
  margin: 0;
  font-size: 0.875rem;
  color: var(--color-on-surface-variant);
  opacity: 0.6;
}

.list-move,
.list-enter-active,
.list-leave-active {
  transition: all 0.2s ease;
}

.list-enter-from,
.list-leave-to {
  opacity: 0;
  transform: translateX(-1rem);
}

.list-leave-active {
  position: absolute;
}
</style>