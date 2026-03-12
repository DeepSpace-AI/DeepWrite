<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import IconChevronDown from '~icons/mdi/chevron-down'
import IconChevronRight from '~icons/mdi/chevron-right'
import IconDeleteOutline from '~icons/mdi/delete-outline'
import IconFileDocumentOutline from '~icons/mdi/file-document-outline'
import IconFileOutline from '~icons/mdi/file-outline'
import IconFilePlusOutline from '~icons/mdi/file-document-plus-outline'
import IconFolderOutline from '~icons/mdi/folder-outline'
import IconFolderPlusOutline from '~icons/mdi/folder-plus-outline'
import IconPencilOutline from '~icons/mdi/pencil-outline'
import IconTrayArrowUp from '~icons/mdi/tray-arrow-up'
import IconEyeOutline from '~icons/mdi/eye-outline'
import WorkspaceEntryModal from '@/views/workspace/modal/WorkspaceEntryModal.vue'
import type { VisibleFolderNode, WorkspaceDocument, WorkspaceFile, WorkspaceFolder } from '@/views/workspace/types'

const props = defineProps<{
  visibleFolders: VisibleFolderNode[]
  selectedFolderId: string | null
  selectedFolder: WorkspaceFolder | null
  selectedFolderPath: WorkspaceFolder[]
  currentFolderDocuments: WorkspaceDocument[]
  currentFolderFiles: WorkspaceFile[]
  selectedFileIds: string[]
  activeDocumentId?: string
  loadingFolders: boolean
  loadingDocuments: boolean
  loadingFiles: boolean
  uploadingFiles: boolean
  submitting: boolean
  errorMessage?: string
}>()

const emit = defineEmits<{
  (e: 'select-root'): void
  (e: 'select-folder', folderId: string): void
  (e: 'toggle-folder', folderId: string): void
  (e: 'create-folder', payload: { name: string; description: string; parentId: string | null }): void
  (e: 'rename-folder', payload: { folderId: string; name: string; description: string }): void
  (e: 'delete-folder', folder: WorkspaceFolder): void
  (e: 'create-document', payload: { title: string }): void
  (e: 'open-document', document: WorkspaceDocument): void
  (e: 'rename-document', payload: { documentId: string; title: string }): void
  (e: 'upload-files', files: File[]): void
  (e: 'delete-file', fileId: string): void
  (e: 'preview-file', fileId: string): void
  (e: 'batch-delete-files', fileIds: string[]): void
  (e: 'update:selectedFileIds', value: string[]): void
}>()

const { t } = useI18n()

const activeTab = ref<'documents' | 'files'>('documents')
const fileInput = ref<HTMLInputElement | null>(null)

const createFolderOpen = ref(false)
const renameFolderOpen = ref(false)
const createDocumentOpen = ref(false)
const renameDocumentOpen = ref(false)
const pendingCreateFolderSubmit = ref(false)
const pendingCreateDocumentSubmit = ref(false)
const contextMenuVisible = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const contextTargetFolderId = ref<string | null>(null)

const editingFolder = ref<WorkspaceFolder | null>(null)
const editingDocument = ref<WorkspaceDocument | null>(null)

const selectedFileIdSet = computed(() => new Set(props.selectedFileIds))

const currentFolderLabel = computed(() => props.selectedFolder?.name || t('workspace.detail.resources.rootFolder'))
const breadcrumbPath = computed(() => props.selectedFolderPath || [])

function formatTime(value?: string) {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(date)
}

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
}

function openFolderCreateModal() {
  createFolderOpen.value = true
}

function closeContextMenu() {
  contextMenuVisible.value = false
}

function openContextMenu(event: MouseEvent, folderId: string | null) {
  event.preventDefault()
  event.stopPropagation()
  contextTargetFolderId.value = folderId
  contextMenuX.value = event.clientX
  contextMenuY.value = event.clientY
  contextMenuVisible.value = true
}

function openCreateFolderFromContext() {
  const target = contextTargetFolderId.value
  closeContextMenu()
  if (target) {
    emit('select-folder', target)
  }
  else {
    emit('select-root')
  }
  nextTick(() => {
    openFolderCreateModal()
  })
}

function openCreateDocumentFromContext() {
  const target = contextTargetFolderId.value
  closeContextMenu()
  if (target) {
    emit('select-folder', target)
  }
  else {
    emit('select-root')
  }
  nextTick(() => {
    openDocumentCreateModal()
  })
}

function handleCreateFolderSubmit(payload: { name: string; description: string }) {
  pendingCreateFolderSubmit.value = true
  emit('create-folder', {
    ...payload,
    parentId: props.selectedFolderId,
  })
}

function openFolderRenameModal(folder: WorkspaceFolder) {
  editingFolder.value = folder
  renameFolderOpen.value = true
}

function openDocumentCreateModal() {
  createDocumentOpen.value = true
}

function handleCreateDocumentSubmit(payload: { name: string }) {
  pendingCreateDocumentSubmit.value = true
  emit('create-document', { title: payload.name })
}

function openDocumentRenameModal(document: WorkspaceDocument) {
  editingDocument.value = document
  renameDocumentOpen.value = true
}

function handleFilePick(event: Event) {
  const target = event.target as HTMLInputElement | null
  const files = target?.files ? Array.from(target.files) : []
  if (files.length) {
    emit('upload-files', files)
  }
  if (target) {
    target.value = ''
  }
}

function toggleFileSelection(fileId: string, checked: boolean) {
  const next = checked
    ? [...props.selectedFileIds, fileId]
    : props.selectedFileIds.filter((id) => id !== fileId)

  emit('update:selectedFileIds', Array.from(new Set(next)))
}

function handleFileCheckboxChange(fileId: string, event: Event) {
  const target = event.target as HTMLInputElement | null
  toggleFileSelection(fileId, !!target?.checked)
}

function confirmDeleteFolder(folder: WorkspaceFolder) {
  if (!window.confirm(t('workspace.detail.resources.confirmDeleteFolder', { name: folder.name }))) return
  emit('delete-folder', folder)
}

function confirmDeleteFile(fileId: string, fileName: string) {
  if (!window.confirm(t('workspace.detail.resources.confirmDeleteFile', { name: fileName }))) return
  emit('delete-file', fileId)
}

function confirmBatchDeleteFiles() {
  if (!props.selectedFileIds.length) return
  if (!window.confirm(t('workspace.detail.resources.confirmBatchDeleteFiles', { count: props.selectedFileIds.length }))) return
  emit('batch-delete-files', props.selectedFileIds)
}

watch(
  () => props.submitting,
  (submitting, previous) => {
    if (!previous || submitting) return

    if (pendingCreateFolderSubmit.value) {
      if (!props.errorMessage) {
        createFolderOpen.value = false
      }
      pendingCreateFolderSubmit.value = false
    }

    if (pendingCreateDocumentSubmit.value) {
      if (!props.errorMessage) {
        createDocumentOpen.value = false
      }
      pendingCreateDocumentSubmit.value = false
    }
  },
)
</script>

<template>
  <aside class="relative h-full min-h-0 rounded-sm border border-base-300 bg-base-100 shadow-sm transition-all duration-300 flex flex-col" @click="closeContextMenu">
    <div class="border-b border-base-300 px-4 py-3">
      <div class="flex items-start justify-between gap-3">
        <div>
          <h3 class="text-sm font-semibold text-base-content">{{ t('workspace.detail.resources.title') }}</h3>
          <div class="mt-1 flex items-center gap-1 text-xs text-base-content/55">
            <button type="button" class="hover:text-base-content" @click="emit('select-root')">
              {{ t('workspace.detail.resources.rootFolder') }}
            </button>
            <template v-for="(folder, index) in breadcrumbPath" :key="folder.id">
              <span class="text-base-content/35">/</span>
              <button
                type="button"
                class="truncate hover:text-base-content"
                :class="index === breadcrumbPath.length - 1 ? 'pointer-events-none font-medium text-base-content' : ''"
                @click="index === breadcrumbPath.length - 1 ? undefined : emit('select-folder', folder.id)"
              >
                {{ folder.name }}
              </button>
            </template>
          </div>
        </div>
        <div class="flex items-center gap-1">
          <div class="tooltip tooltip-bottom" :data-tip="t('workspace.detail.resources.createDocument')">
            <button type="button" class="btn btn-ghost btn-xs rounded-sm" @click="openDocumentCreateModal()">
              <IconFilePlusOutline class="size-4" />
            </button>
          </div>
          <div class="tooltip tooltip-bottom" :data-tip="t('workspace.detail.resources.createFolder')">
            <button type="button" class="btn btn-ghost btn-xs rounded-sm" @click="openFolderCreateModal()">
              <IconFolderPlusOutline class="size-4" />
            </button>
          </div>
          <div class="tooltip tooltip-bottom" :data-tip="t('workspace.detail.resources.uploadFiles')">
            <button type="button" class="btn btn-ghost btn-xs rounded-sm" :disabled="uploadingFiles" @click="fileInput?.click()">
              <IconTrayArrowUp class="size-4" />
            </button>
          </div>
        </div>
      </div>

      <div class="mt-3 flex flex-wrap gap-2">
        <button type="button" class="btn btn-xs rounded-sm" :class="!selectedFolderId ? 'btn-primary' : 'btn-ghost'" @click="emit('select-root')">
          {{ t('workspace.detail.resources.rootFolder') }}
        </button>
      </div>
    </div>

    <div v-if="errorMessage" class="border-b border-base-300 bg-error/8 px-4 py-2 text-xs text-error">
      {{ errorMessage }}
    </div>

    <div class="min-h-0 flex-1 overflow-hidden">
      <section class="border-b border-base-300 px-2 py-2" @contextmenu="openContextMenu($event, selectedFolderId)">
        <div class="max-h-52 overflow-y-auto pr-1">
          <div v-if="loadingFolders && !visibleFolders.length" class="flex items-center gap-2 px-2 py-6 text-xs text-base-content/50">
            <span class="loading loading-spinner loading-xs" />
            {{ t('workspace.detail.resources.loadingFolders') }}
          </div>

          <div v-else-if="!visibleFolders.length" class="px-2 py-6 text-xs text-base-content/50">
            {{ t('workspace.detail.resources.emptyFolders') }}
          </div>

          <div v-else class="space-y-1">
            <div
              v-for="node in visibleFolders"
              :key="node.folder.id"
              class="group flex items-center gap-1 rounded-sm pr-1"
              :class="selectedFolderId === node.folder.id ? 'bg-primary/8' : 'hover:bg-base-200/60'"
              @contextmenu="openContextMenu($event, node.folder.id)"
            >
              <button
                type="button"
                class="btn btn-ghost btn-xs rounded-sm px-0"
                :style="{ marginLeft: `${node.depth * 14}px` }"
                @click="emit('toggle-folder', node.folder.id)"
              >
                <component :is="node.expanded ? IconChevronDown : IconChevronRight" class="size-4 text-base-content/55" />
              </button>

              <button
                type="button"
                class="flex min-w-0 flex-1 items-center gap-2 rounded-sm px-1.5 py-1.5 text-left"
                @click="emit('select-folder', node.folder.id)"
              >
                <IconFolderOutline class="size-4 shrink-0 text-base-content/65" />
                <span class="truncate text-sm text-base-content">{{ node.folder.name }}</span>
              </button>

              <div class="flex shrink-0 items-center gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
                <button type="button" class="btn btn-ghost btn-xs rounded-sm" :disabled="submitting" @click="openFolderRenameModal(node.folder)">
                  <IconPencilOutline class="size-3.5" />
                </button>
                <button type="button" class="btn btn-ghost btn-xs rounded-sm text-error" :disabled="submitting" @click="confirmDeleteFolder(node.folder)">
                  <IconDeleteOutline class="size-3.5" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="flex min-h-0 flex-1 flex-col">
        <div class="border-b border-base-300 px-3 py-2">
          <div class="tabs tabs-box rounded-sm bg-base-200 p-1">
            <button type="button" class="tab rounded-sm" :class="activeTab === 'documents' ? 'tab-active' : ''" @click="activeTab = 'documents'">
              {{ t('workspace.detail.resources.documentsTab') }}
            </button>
            <button type="button" class="tab rounded-sm" :class="activeTab === 'files' ? 'tab-active' : ''" @click="activeTab = 'files'">
              {{ t('workspace.detail.resources.filesTab') }}
            </button>
          </div>
        </div>

        <div v-if="activeTab === 'documents'" class="min-h-0 flex-1 overflow-y-auto px-3 py-3">
          <div v-if="loadingDocuments" class="flex items-center gap-2 py-6 text-xs text-base-content/50">
            <span class="loading loading-spinner loading-xs" />
            {{ t('workspace.detail.resources.loadingDocuments') }}
          </div>

          <div v-else-if="!currentFolderDocuments.length" class="rounded-sm border border-dashed border-base-300 px-3 py-5 text-xs text-base-content/50">
            {{ t('workspace.detail.resources.emptyDocuments') }}
          </div>

          <div v-else class="space-y-2">
            <div
              v-for="document in currentFolderDocuments"
              :key="document.id"
              class="rounded-sm border border-base-300 px-3 py-2"
              :class="activeDocumentId === document.id ? 'border-primary bg-primary/6' : 'bg-base-100 hover:bg-base-200/50'"
            >
              <div class="flex items-start justify-between gap-2">
                <button type="button" class="min-w-0 flex-1 text-left" @click="emit('open-document', document)">
                  <div class="flex items-center gap-2">
                    <IconFileDocumentOutline class="size-4 shrink-0 text-base-content/65" />
                    <p class="truncate text-sm font-medium text-base-content">{{ document.title }}</p>
                  </div>
                  <p class="mt-1 text-[11px] text-base-content/50">
                    {{ t('workspace.detail.resources.documentMeta', { time: formatTime(document.updated_at), version: document.current_version || 1 }) }}
                  </p>
                </button>

                <button type="button" class="btn btn-ghost btn-xs rounded-sm" :disabled="submitting" @click="openDocumentRenameModal(document)">
                  <IconPencilOutline class="size-3.5" />
                </button>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="min-h-0 flex-1 overflow-y-auto px-3 py-3">
          <div class="mb-3 flex items-center justify-between gap-2">
            <p class="text-xs text-base-content/55">{{ t('workspace.detail.resources.fileSelectionHint') }}</p>
            <button
              type="button"
              class="btn btn-xs btn-outline rounded-sm"
              :disabled="!selectedFileIds.length || submitting"
              @click="confirmBatchDeleteFiles"
            >
              {{ t('workspace.detail.resources.batchDeleteFiles', { count: selectedFileIds.length }) }}
            </button>
          </div>

          <div v-if="loadingFiles" class="flex items-center gap-2 py-6 text-xs text-base-content/50">
            <span class="loading loading-spinner loading-xs" />
            {{ t('workspace.detail.resources.loadingFiles') }}
          </div>

          <div v-else-if="!currentFolderFiles.length" class="rounded-sm border border-dashed border-base-300 px-3 py-5 text-xs text-base-content/50">
            {{ t('workspace.detail.resources.emptyFiles') }}
          </div>

          <div v-else class="space-y-2">
            <div v-for="file in currentFolderFiles" :key="file.id" class="rounded-sm border border-base-300 bg-base-100 px-3 py-2">
              <div class="flex items-start gap-2">
                <input
                  type="checkbox"
                  class="checkbox checkbox-xs mt-1 rounded-xs"
                  :checked="selectedFileIdSet.has(file.id)"
                  @change="handleFileCheckboxChange(file.id, $event)"
                />

                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <IconFileOutline class="size-4 shrink-0 text-base-content/65" />
                    <p class="truncate text-sm font-medium text-base-content">{{ file.file_name }}</p>
                  </div>
                  <p class="mt-1 text-[11px] text-base-content/50">
                    {{ t('workspace.detail.resources.fileMeta', { size: formatSize(file.size), time: formatTime(file.updated_at) }) }}
                  </p>
                </div>

                <div class="flex shrink-0 items-center gap-1">
                  <button type="button" class="btn btn-ghost btn-xs rounded-sm" @click="emit('preview-file', file.id)">
                    <IconEyeOutline class="size-3.5" />
                  </button>
                  <button type="button" class="btn btn-ghost btn-xs rounded-sm text-error" :disabled="submitting" @click="confirmDeleteFile(file.id, file.file_name)">
                    <IconDeleteOutline class="size-3.5" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>

    <Teleport to="body">
      <div
        v-if="contextMenuVisible"
        class="fixed z-50 min-w-36 rounded-sm border border-base-300 bg-base-100 p-1 shadow-lg"
        :style="{ left: `${contextMenuX}px`, top: `${contextMenuY}px` }"
        @click.stop
      >
        <button type="button" class="btn btn-ghost btn-sm w-full justify-start rounded-sm" :disabled="submitting" @click="openCreateFolderFromContext">
          {{ t('workspace.detail.resources.createFolder') }}
        </button>
        <button type="button" class="btn btn-ghost btn-sm w-full justify-start rounded-sm" :disabled="submitting" @click="openCreateDocumentFromContext">
          {{ t('workspace.detail.resources.createDocument') }}
        </button>
      </div>
    </Teleport>

    <input ref="fileInput" type="file" class="hidden" multiple @change="handleFilePick" />

    <WorkspaceEntryModal
      v-model="createFolderOpen"
      :title="t('workspace.detail.resources.createFolderTitle')"
      :subtitle="t('workspace.detail.resources.createFolderSubtitle', { name: currentFolderLabel })"
      :name-label="t('workspace.detail.resources.nameLabel')"
      :name-placeholder="t('workspace.detail.resources.folderNamePlaceholder')"
      :description-label="t('workspace.detail.resources.descriptionLabel')"
      :description-placeholder="t('workspace.detail.resources.folderDescriptionPlaceholder')"
      :submit-text="t('workspace.detail.resources.confirmCreateFolder')"
      :show-description="true"
      :submitting="submitting"
      :error-message="errorMessage"
      @submit="handleCreateFolderSubmit"
    />

    <WorkspaceEntryModal
      v-model="renameFolderOpen"
      :title="t('workspace.detail.resources.renameFolderTitle')"
      :subtitle="t('workspace.detail.resources.renameFolderSubtitle')"
      :name-label="t('workspace.detail.resources.nameLabel')"
      :name-placeholder="t('workspace.detail.resources.folderNamePlaceholder')"
      :description-label="t('workspace.detail.resources.descriptionLabel')"
      :description-placeholder="t('workspace.detail.resources.folderDescriptionPlaceholder')"
      :submit-text="t('workspace.detail.resources.confirmRenameFolder')"
      :show-description="true"
      :submitting="submitting"
      :error-message="errorMessage"
      :initial-name="editingFolder?.name || ''"
      :initial-description="editingFolder?.description || ''"
      @submit="editingFolder && emit('rename-folder', { folderId: editingFolder.id, ...$event })"
    />

    <WorkspaceEntryModal
      v-model="createDocumentOpen"
      :title="t('workspace.detail.resources.createDocumentTitle')"
      :subtitle="t('workspace.detail.resources.createDocumentSubtitle', { name: currentFolderLabel })"
      :name-label="t('workspace.detail.resources.nameLabel')"
      :name-placeholder="t('workspace.detail.resources.documentNamePlaceholder')"
      :submit-text="t('workspace.detail.resources.confirmCreateDocument')"
      :submitting="submitting"
      :error-message="errorMessage"
      @submit="handleCreateDocumentSubmit($event)"
    />

    <WorkspaceEntryModal
      v-model="renameDocumentOpen"
      :title="t('workspace.detail.resources.renameDocumentTitle')"
      :subtitle="t('workspace.detail.resources.renameDocumentSubtitle')"
      :name-label="t('workspace.detail.resources.nameLabel')"
      :name-placeholder="t('workspace.detail.resources.documentNamePlaceholder')"
      :submit-text="t('workspace.detail.resources.confirmRenameDocument')"
      :submitting="submitting"
      :error-message="errorMessage"
      :initial-name="editingDocument?.title || ''"
      @submit="editingDocument && emit('rename-document', { documentId: editingDocument.id, title: $event.name })"
    />
  </aside>
</template>
