<script setup lang="ts">
import { ref, computed } from 'vue'
import type { SessionGroup } from '@/api/agent'
import { DButton, DModal } from '@/components/base'
import IconFolder from '~icons/mdi/folder'
import IconPlus from '~icons/mdi/plus'
import IconPencil from '~icons/mdi/pencil'
import IconDelete from '~icons/mdi/delete'

const props = defineProps<{
  groups: SessionGroup[]
  selectedGroupId?: string | null
}>()

const emit = defineEmits<{
  select: [groupId: string | null]
  create: [input: { name: string; description: string; color: string }]
  update: [groupId: string, input: Partial<SessionGroup>]
  delete: [groupId: string]
}>()

const showCreateModal = ref(false)
const showEditModal = ref(false)
const editingGroup = ref<SessionGroup | null>(null)

const form = ref({
  name: '',
  description: '',
  color: '#6366f1',
})

const colorOptions = [
  { value: '#6366f1', label: 'Indigo' },
  { value: '#8b5cf6', label: 'Purple' },
  { value: '#ec4899', label: 'Pink' },
  { value: '#ef4444', label: 'Red' },
  { value: '#f97316', label: 'Orange' },
  { value: '#eab308', label: 'Yellow' },
  { value: '#22c55e', label: 'Green' },
  { value: '#14b8a6', label: 'Teal' },
  { value: '#3b82f6', label: 'Blue' },
]

function handleSelect(groupId: string | null) {
  emit('select', groupId)
}

function openCreateModal() {
  form.value = { name: '', description: '', color: '#6366f1' }
  showCreateModal.value = true
}

function openEditModal(group: SessionGroup) {
  editingGroup.value = group
  form.value = {
    name: group.name,
    description: group.description,
    color: group.color,
  }
  showEditModal.value = true
}

function handleCreate() {
  if (!form.value.name.trim()) return
  emit('create', { ...form.value })
  showCreateModal.value = false
}

function handleUpdate() {
  if (!editingGroup.value || !form.value.name.trim()) return
  emit('update', editingGroup.value.id, { ...form.value })
  showEditModal.value = false
  editingGroup.value = null
}

function handleDelete(groupId: string) {
  if (confirm('确定要删除此分组吗？分组内的对话不会被删除。')) {
    emit('delete', groupId)
  }
}
</script>

<template>
  <div class="session-groups">
    <div class="groups-header">
      <span class="header-title">分组</span>
      <DButton size="sm" variant="ghost" title="新建分组" @click="openCreateModal">
        <IconPlus class="h-4 w-4" />
      </DButton>
    </div>

    <div class="groups-list">
      <button
        class="group-item"
        :class="{ active: !selectedGroupId }"
        @click="handleSelect(null)"
      >
        <IconFolder class="group-icon" />
        <span class="group-name">全部对话</span>
      </button>

      <button
        v-for="group in groups"
        :key="group.id"
        class="group-item"
        :class="{ active: group.id === selectedGroupId }"
        @click="handleSelect(group.id)"
      >
        <span
          class="group-color"
          :style="{ backgroundColor: group.color }"
        />
        <span class="group-name">{{ group.name }}</span>
        <div class="group-actions" @click.stop>
          <DButton
            size="sm"
            variant="ghost"
            title="编辑"
            @click="openEditModal(group)"
          >
            <IconPencil class="h-3.5 w-3.5" />
          </DButton>
          <DButton
            size="sm"
            variant="ghost"
            title="删除"
            @click="handleDelete(group.id)"
          >
            <IconDelete class="h-3.5 w-3.5" />
          </DButton>
        </div>
      </button>
    </div>

    <DModal :visible="showCreateModal" title="新建分组" @close="showCreateModal = false">
      <form class="group-form" @submit.prevent="handleCreate">
        <div class="form-field">
          <label class="field-label">名称</label>
          <input
            v-model="form.name"
            type="text"
            class="field-input"
            placeholder="输入分组名称"
            required
          />
        </div>
        <div class="form-field">
          <label class="field-label">描述</label>
          <input
            v-model="form.description"
            type="text"
            class="field-input"
            placeholder="可选"
          />
        </div>
        <div class="form-field">
          <label class="field-label">颜色</label>
          <div class="color-picker">
            <button
              v-for="color in colorOptions"
              :key="color.value"
              type="button"
              class="color-option"
              :class="{ active: form.color === color.value }"
              :style="{ backgroundColor: color.value }"
              :title="color.label"
              @click="form.color = color.value"
            />
          </div>
        </div>
        <div class="form-actions">
          <DButton type="button" variant="ghost" @click="showCreateModal = false">
            取消
          </DButton>
          <DButton type="submit" :disabled="!form.name.trim()">
            创建
          </DButton>
        </div>
      </form>
    </DModal>

    <DModal :visible="showEditModal" title="编辑分组" @close="showEditModal = false">
      <form class="group-form" @submit.prevent="handleUpdate">
        <div class="form-field">
          <label class="field-label">名称</label>
          <input
            v-model="form.name"
            type="text"
            class="field-input"
            required
          />
        </div>
        <div class="form-field">
          <label class="field-label">描述</label>
          <input
            v-model="form.description"
            type="text"
            class="field-input"
          />
        </div>
        <div class="form-field">
          <label class="field-label">颜色</label>
          <div class="color-picker">
            <button
              v-for="color in colorOptions"
              :key="color.value"
              type="button"
              class="color-option"
              :class="{ active: form.color === color.value }"
              :style="{ backgroundColor: color.value }"
              @click="form.color = color.value"
            />
          </div>
        </div>
        <div class="form-actions">
          <DButton type="button" variant="ghost" @click="showEditModal = false">
            取消
          </DButton>
          <DButton type="submit" :disabled="!form.name.trim()">
            保存
          </DButton>
        </div>
      </form>
    </DModal>
  </div>
</template>

<style scoped>
.session-groups {
  display: flex;
  flex-direction: column;
}

.groups-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 0.75rem;
}

.header-title {
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-on-surface-variant);
}

.groups-list {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.group-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: none;
  border-radius: 0.375rem;
  background: transparent;
  cursor: pointer;
  text-align: left;
  transition: background 0.15s ease;
}

.group-item:hover {
  background: var(--surface-container);
}

.group-item.active {
  background: oklch(from var(--color-primary) l c h / 0.1);
}

.group-icon {
  width: 1rem;
  height: 1rem;
  color: var(--color-on-surface-variant);
}

.group-color {
  width: 0.75rem;
  height: 0.75rem;
  border-radius: 0.25rem;
  flex-shrink: 0;
}

.group-name {
  flex: 1;
  font-size: 0.8125rem;
  color: var(--color-on-background);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.group-actions {
  display: flex;
  gap: 0.125rem;
  opacity: 0;
  transition: opacity 0.15s ease;
}

.group-item:hover .group-actions {
  opacity: 1;
}

.group-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.field-label {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-on-background);
}

.field-input {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--color-surface-container-high);
  border-radius: 0.5rem;
  background: var(--surface-container-low);
  font-size: 0.875rem;
  color: var(--color-on-background);
  outline: none;
  transition: border-color 0.15s ease;
}

.field-input:focus {
  border-color: var(--color-primary);
}

.color-picker {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.color-option {
  width: 1.5rem;
  height: 1.5rem;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  transition: transform 0.15s ease, border-color 0.15s ease;
}

.color-option:hover {
  transform: scale(1.1);
}

.color-option.active {
  border-color: var(--color-on-background);
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 0.5rem;
}
</style>