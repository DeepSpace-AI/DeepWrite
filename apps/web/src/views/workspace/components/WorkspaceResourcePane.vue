<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import IconChevronDown from '~icons/mdi/chevron-down'
import IconChevronRight from '~icons/mdi/chevron-right'
import IconDeleteOutline from '~icons/mdi/delete-outline'
import IconFileDocumentOutline from '~icons/mdi/file-document-outline'
import IconFileMoveOutline from '~icons/mdi/file-move-outline'
import IconFileOutline from '~icons/mdi/file-outline'
import IconFilePlusOutline from '~icons/mdi/file-document-plus-outline'
import IconFolderOutline from '~icons/mdi/folder-outline'
import IconFolderPlusOutline from '~icons/mdi/folder-plus-outline'
import IconPencilOutline from '~icons/mdi/pencil-outline'
import IconTrayArrowUp from '~icons/mdi/tray-arrow-up'
import IconEyeOutline from '~icons/mdi/eye-outline'
import WorkspaceEntryModal from '@/views/workspace/modal/WorkspaceEntryModal.vue'
import ResourceMoveModal from '@/views/workspace/modal/ResourceMoveModal.vue'
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
  uploadStatusMessage?: string
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
  (e: 'move-document', payload: { documentId: string; folderId: string | null }): void
  (e: 'delete-document', documentId: string): void
  (e: 'upload-files', files: File[]): void
  (e: 'rename-file', payload: { fileId: string; fileName: string }): void
  (e: 'move-file', payload: { fileId: string; folderId: string | null }): void
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
const renameFileOpen = ref(false)
const moveDocumentOpen = ref(false)
const moveFileOpen = ref(false)
const pendingCreateFolderSubmit = ref(false)
const pendingRenameFolderSubmit = ref(false)
const pendingCreateDocumentSubmit = ref(false)
const pendingRenameDocumentSubmit = ref(false)
const pendingRenameFileSubmit = ref(false)
const pendingMoveDocumentSubmit = ref(false)
const pendingMoveFileSubmit = ref(false)
const contextMenuVisible = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const contextMenuType = ref<'area' | 'folder' | 'document' | 'file'>('area')
const contextTargetFolderId = ref<string | null>(null)
const contextMenuFolder = ref<WorkspaceFolder | null>(null)
const contextMenuDocument = ref<WorkspaceDocument | null>(null)
const contextMenuFile = ref<WorkspaceFile | null>(null)

const editingFolder = ref<WorkspaceFolder | null>(null)
const editingDocument = ref<WorkspaceDocument | null>(null)
const editingFile = ref<WorkspaceFile | null>(null)
const movingDocument = ref<WorkspaceDocument | null>(null)
const movingFile = ref<WorkspaceFile | null>(null)

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
  contextMenuType.value = 'area'
  contextTargetFolderId.value = folderId
  contextMenuX.value = event.clientX
  contextMenuY.value = event.clientY
  contextMenuVisible.value = true
}

function openFolderContextMenu(event: MouseEvent, folder: WorkspaceFolder) {
  event.preventDefault()
  event.stopPropagation()
  contextMenuType.value = 'folder'
  contextMenuFolder.value = folder
  contextMenuX.value = event.clientX
  contextMenuY.value = event.clientY
  contextMenuVisible.value = true
}

function openDocumentContextMenu(event: MouseEvent, document: WorkspaceDocument) {
  event.preventDefault()
  event.stopPropagation()
  contextMenuType.value = 'document'
  contextMenuDocument.value = document
  contextMenuX.value = event.clientX
  contextMenuY.value = event.clientY
  contextMenuVisible.value = true
}

function openFileContextMenu(event: MouseEvent, file: WorkspaceFile) {
  event.preventDefault()
  event.stopPropagation()
  contextMenuType.value = 'file'
  contextMenuFile.value = file
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

function handleContextCreateSubfolder() {
  const folder = contextMenuFolder.value
  closeContextMenu()
  if (folder) emit('select-folder', folder.id)
  nextTick(() => openFolderCreateModal())
}

function handleContextNewDocumentInFolder() {
  const folder = contextMenuFolder.value
  closeContextMenu()
  if (folder) emit('select-folder', folder.id)
  nextTick(() => openDocumentCreateModal())
}

function handleContextRenameFolder() {
  if (!contextMenuFolder.value) return
  const folder = contextMenuFolder.value
  closeContextMenu()
  openFolderRenameModal(folder)
}

function handleContextDeleteFolder() {
  if (!contextMenuFolder.value) return
  const folder = contextMenuFolder.value
  closeContextMenu()
  confirmDeleteFolder(folder)
}

function handleContextOpenDocument() {
  if (!contextMenuDocument.value) return
  const doc = contextMenuDocument.value
  closeContextMenu()
  emit('open-document', doc)
}

function handleContextRenameDocument() {
  if (!contextMenuDocument.value) return
  const doc = contextMenuDocument.value
  closeContextMenu()
  openDocumentRenameModal(doc)
}

function handleContextMoveDocument() {
  if (!contextMenuDocument.value) return
  const doc = contextMenuDocument.value
  closeContextMenu()
  openDocumentMoveModal(doc)
}

function handleContextDeleteDocument() {
  if (!contextMenuDocument.value) return
  const doc = contextMenuDocument.value
  closeContextMenu()
  confirmDeleteDocument(doc)
}

function handleContextRenameFile() {
  if (!contextMenuFile.value) return
  const file = contextMenuFile.value
  closeContextMenu()
  openFileRenameModal(file)
}

function handleContextMoveFile() {
  if (!contextMenuFile.value) return
  const file = contextMenuFile.value
  closeContextMenu()
  openFileMoveModal(file)
}

function handleContextPreviewFile() {
  if (!contextMenuFile.value) return
  const fileId = contextMenuFile.value.id
  closeContextMenu()
  emit('preview-file', fileId)
}

function handleContextDeleteFile() {
  if (!contextMenuFile.value) return
  const file = contextMenuFile.value
  closeContextMenu()
  confirmDeleteFile(file.id, file.file_name)
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

function openDocumentMoveModal(document: WorkspaceDocument) {
  movingDocument.value = document
  moveDocumentOpen.value = true
}

function openFileRenameModal(file: WorkspaceFile) {
  editingFile.value = file
  renameFileOpen.value = true
}

function openFileMoveModal(file: WorkspaceFile) {
  movingFile.value = file
  moveFileOpen.value = true
}

function handleRenameFolderSubmit(payload: { name: string; description: string }) {
  if (!editingFolder.value) return
  pendingRenameFolderSubmit.value = true
  emit('rename-folder', {
    folderId: editingFolder.value.id,
    ...payload,
  })
}

function handleRenameDocumentSubmit(payload: { name: string }) {
  if (!editingDocument.value) return
  pendingRenameDocumentSubmit.value = true
  emit('rename-document', {
    documentId: editingDocument.value.id,
    title: payload.name,
  })
}

function handleRenameFileSubmit(payload: { name: string }) {
  if (!editingFile.value) return
  pendingRenameFileSubmit.value = true
  emit('rename-file', {
    fileId: editingFile.value.id,
    fileName: payload.name,
  })
}

function handleMoveDocumentSubmit(payload: { folderId: string | null }) {
  if (!movingDocument.value) return
  pendingMoveDocumentSubmit.value = true
  emit('move-document', {
    documentId: movingDocument.value.id,
    folderId: payload.folderId,
  })
}

function handleMoveFileSubmit(payload: { folderId: string | null }) {
  if (!movingFile.value) return
  pendingMoveFileSubmit.value = true
  emit('move-file', {
    fileId: movingFile.value.id,
    folderId: payload.folderId,
  })
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

function confirmDeleteDocument(document: WorkspaceDocument) {
  if (!window.confirm(t('workspace.detail.resources.confirmDeleteDocument', { name: document.title }))) return
  emit('delete-document', document.id)
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

    if (pendingRenameFolderSubmit.value) {
      if (!props.errorMessage) {
        renameFolderOpen.value = false
      }
      pendingRenameFolderSubmit.value = false
    }

    if (pendingCreateDocumentSubmit.value) {
      if (!props.errorMessage) {
        createDocumentOpen.value = false
      }
      pendingCreateDocumentSubmit.value = false
    }

    if (pendingRenameDocumentSubmit.value) {
      if (!props.errorMessage) {
        renameDocumentOpen.value = false
      }
      pendingRenameDocumentSubmit.value = false
    }

    if (pendingRenameFileSubmit.value) {
      if (!props.errorMessage) {
        renameFileOpen.value = false
      }
      pendingRenameFileSubmit.value = false
    }

    if (pendingMoveDocumentSubmit.value) {
      if (!props.errorMessage) {
        moveDocumentOpen.value = false
      }
      pendingMoveDocumentSubmit.value = false
    }

    if (pendingMoveFileSubmit.value) {
      if (!props.errorMessage) {
        moveFileOpen.value = false
      }
      pendingMoveFileSubmit.value = false
    }
  },
)
</script>

<template>
  <aside class="paper-panel relative h-full min-h-0 transition-all duration-300 flex flex-col" @click="closeContextMenu">
    <div class="bg-(--surface-overlay) px-4 py-3">
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

    <div v-if="errorMessage" class="bg-error/8 px-4 py-2 text-xs text-error">
      {{ errorMessage }}
    </div>

    <div
      v-if="uploadStatusMessage"
      class="px-4 py-2 text-xs"
      :class="uploadingFiles ? 'bg-info/10 text-info' : 'bg-success/10 text-success'
      "
    >
      <div class="flex items-center gap-2">
        <span v-if="uploadingFiles" class="loading loading-spinner loading-xs" />
        <span>{{ uploadStatusMessage }}</span>
      </div>
    </div>

    <div class="min-h-0 flex-1 overflow-hidden">
      <section class="bg-(--surface-overlay)/45 px-2 py-2" @contextmenu="openContextMenu($event, selectedFolderId)">
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
              class="flex items-center gap-1 rounded-sm pr-1"
                :class="selectedFolderId === node.folder.id ? 'bg-primary/8' : 'hover:bg-(--surface-overlay)'"
              @contextmenu.stop="openFolderContextMenu($event, node.folder)"
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
            </div>
          </div>
        </div>
      </section>

      <section class="flex min-h-0 flex-1 flex-col">
        <div class="bg-(--surface-overlay) px-3 py-2">
          <div class="tabs tabs-box rounded-sm bg-(--surface-sunken) p-1">
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

          <div v-else-if="!currentFolderDocuments.length" class="paper-panel-embedded rounded-sm px-3 py-5 text-xs text-base-content/50">
            {{ t('workspace.detail.resources.emptyDocuments') }}
          </div>

          <div v-else class="space-y-2">
            <div
              v-for="document in currentFolderDocuments"
              :key="document.id"
              class="cursor-pointer rounded-sm no-line px-3 py-2"
              :class="activeDocumentId === document.id ? 'bg-primary/10 text-base-content' : 'bg-(--surface-raised) hover:bg-(--surface-overlay)'"
              @click="emit('open-document', document)"
              @contextmenu.stop="openDocumentContextMenu($event, document)"
            >
              <div class="flex items-center gap-2">
                <IconFileDocumentOutline class="size-4 shrink-0 text-base-content/65" />
                <p class="truncate text-sm font-medium text-base-content">{{ document.title }}</p>
              </div>
              <p class="mt-1 text-[11px] text-base-content/50">
                {{ t('workspace.detail.resources.documentMeta', { time: formatTime(document.updated_at), version: document.current_version || 1 }) }}
              </p>
            </div>
          </div>
        </div>

        <div v-else class="min-h-0 flex-1 overflow-y-auto px-3 py-3">
          <div class="mb-3 flex items-center justify-between gap-2">
            <p class="text-xs text-base-content/55">{{ t('workspace.detail.resources.fileSelectionHint') }}</p>
            <button
              type="button"
              class="btn btn-xs btn-ghost rounded-sm"
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

          <div v-else-if="!currentFolderFiles.length" class="paper-panel-embedded rounded-sm px-3 py-5 text-xs text-base-content/50">
            {{ t('workspace.detail.resources.emptyFiles') }}
          </div>

          <div v-else class="space-y-2">
            <div v-for="file in currentFolderFiles" :key="file.id" class="rounded-sm no-line bg-(--surface-raised) px-3 py-2 cursor-pointer hover:bg-(--surface-overlay)" @contextmenu.stop="openFileContextMenu($event, file)" @dblclick="emit('preview-file', file.id)">
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
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>

    <Teleport to="body">
      <div
        v-if="contextMenuVisible"
        class="paper-panel fixed z-50 min-w-44 overflow-hidden py-1 text-sm"
        :style="{ left: `${contextMenuX}px`, top: `${contextMenuY}px` }"
        @click.stop
      >
        <!-- Area context menu -->
        <template v-if="contextMenuType === 'area'">
          <button type="button" class="flex w-full items-center gap-2 px-3 py-1.5 hover:bg-(--surface-overlay) disabled:opacity-50" :disabled="submitting" @click="openCreateFolderFromContext">
            <IconFolderPlusOutline class="size-4 text-base-content/65" />
            {{ t('workspace.detail.resources.createFolder') }}
          </button>
          <button type="button" class="flex w-full items-center gap-2 px-3 py-1.5 hover:bg-(--surface-overlay) disabled:opacity-50" :disabled="submitting" @click="openCreateDocumentFromContext">
            <IconFilePlusOutline class="size-4 text-base-content/65" />
            {{ t('workspace.detail.resources.createDocument') }}
          </button>
        </template>

        <!-- Folder context menu -->
        <template v-else-if="contextMenuType === 'folder'">
          <button type="button" class="flex w-full items-center gap-2 px-3 py-1.5 hover:bg-(--surface-overlay) disabled:opacity-50" :disabled="submitting" @click="handleContextCreateSubfolder">
            <IconFolderPlusOutline class="size-4 text-base-content/65" />
            {{ t('workspace.detail.resources.actionCreateSubfolder') }}
          </button>
          <button type="button" class="flex w-full items-center gap-2 px-3 py-1.5 hover:bg-(--surface-overlay) disabled:opacity-50" :disabled="submitting" @click="handleContextNewDocumentInFolder">
            <IconFilePlusOutline class="size-4 text-base-content/65" />
            {{ t('workspace.detail.resources.actionNewDocumentHere') }}
          </button>
          <div class="my-1 h-px bg-base-content/10" />
          <button type="button" class="flex w-full items-center gap-2 px-3 py-1.5 hover:bg-(--surface-overlay) disabled:opacity-50" :disabled="submitting" @click="handleContextRenameFolder">
            <IconPencilOutline class="size-4 text-base-content/65" />
            {{ t('workspace.detail.resources.actionRename') }}
          </button>
          <div class="my-1 h-px bg-base-content/10" />
          <button type="button" class="flex w-full items-center gap-2 px-3 py-1.5 text-error hover:bg-error/8 disabled:opacity-50" :disabled="submitting" @click="handleContextDeleteFolder">
            <IconDeleteOutline class="size-4" />
            {{ t('workspace.detail.resources.actionDelete') }}
          </button>
        </template>

        <!-- Document context menu -->
        <template v-else-if="contextMenuType === 'document'">
          <button type="button" class="flex w-full items-center gap-2 px-3 py-1.5 hover:bg-(--surface-overlay)" @click="handleContextOpenDocument">
            <IconFileDocumentOutline class="size-4 text-base-content/65" />
            {{ t('workspace.detail.resources.actionOpen') }}
          </button>
          <div class="my-1 h-px bg-base-content/10" />
          <button type="button" class="flex w-full items-center gap-2 px-3 py-1.5 hover:bg-(--surface-overlay) disabled:opacity-50" :disabled="submitting" @click="handleContextRenameDocument">
            <IconPencilOutline class="size-4 text-base-content/65" />
            {{ t('workspace.detail.resources.actionRename') }}
          </button>
          <button type="button" class="flex w-full items-center gap-2 px-3 py-1.5 hover:bg-(--surface-overlay) disabled:opacity-50" :disabled="submitting" @click="handleContextMoveDocument">
            <IconFileMoveOutline class="size-4 text-base-content/65" />
            {{ t('workspace.detail.resources.actionMove') }}
          </button>
          <div class="my-1 h-px bg-base-content/10" />
          <button type="button" class="flex w-full items-center gap-2 px-3 py-1.5 text-error hover:bg-error/8 disabled:opacity-50" :disabled="submitting" @click="handleContextDeleteDocument">
            <IconDeleteOutline class="size-4" />
            {{ t('workspace.detail.resources.actionDelete') }}
          </button>
        </template>

        <!-- File context menu -->
        <template v-else-if="contextMenuType === 'file'">
          <button type="button" class="flex w-full items-center gap-2 px-3 py-1.5 hover:bg-(--surface-overlay)" @click="handleContextPreviewFile">
            <IconEyeOutline class="size-4 text-base-content/65" />
            {{ t('workspace.detail.resources.actionPreview') }}
          </button>
          <div class="my-1 h-px bg-base-content/10" />
          <button type="button" class="flex w-full items-center gap-2 px-3 py-1.5 hover:bg-(--surface-overlay) disabled:opacity-50" :disabled="submitting" @click="handleContextRenameFile">
            <IconPencilOutline class="size-4 text-base-content/65" />
            {{ t('workspace.detail.resources.actionRename') }}
          </button>
          <button type="button" class="flex w-full items-center gap-2 px-3 py-1.5 hover:bg-(--surface-overlay) disabled:opacity-50" :disabled="submitting" @click="handleContextMoveFile">
            <IconFileMoveOutline class="size-4 text-base-content/65" />
            {{ t('workspace.detail.resources.actionMove') }}
          </button>
          <div class="my-1 h-px bg-base-content/10" />
          <button type="button" class="flex w-full items-center gap-2 px-3 py-1.5 text-error hover:bg-error/8 disabled:opacity-50" :disabled="submitting" @click="handleContextDeleteFile">
            <IconDeleteOutline class="size-4" />
            {{ t('workspace.detail.resources.actionDelete') }}
          </button>
        </template>
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
      @submit="handleRenameFolderSubmit"
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
      @submit="handleRenameDocumentSubmit"
    />

    <WorkspaceEntryModal
      v-model="renameFileOpen"
      :title="t('workspace.detail.resources.renameFileTitle')"
      :subtitle="t('workspace.detail.resources.renameFileSubtitle')"
      :name-label="t('workspace.detail.resources.nameLabel')"
      :name-placeholder="t('workspace.detail.resources.fileNamePlaceholder')"
      :submit-text="t('workspace.detail.resources.confirmRenameFile')"
      :submitting="submitting"
      :error-message="errorMessage"
      :initial-name="editingFile?.file_name || ''"
      @submit="handleRenameFileSubmit"
    />

    <ResourceMoveModal
      v-model="moveDocumentOpen"
      :title="t('workspace.detail.resources.moveDocumentTitle')"
      :subtitle="t('workspace.detail.resources.moveDocumentSubtitle', { name: movingDocument?.title || '' })"
      :root-label="t('workspace.detail.resources.rootFolder')"
      :destination-label="t('workspace.detail.resources.targetFolderLabel')"
      :confirm-text="t('workspace.detail.resources.confirmMoveDocument')"
      :visible-folders="visibleFolders"
      :current-folder-id="movingDocument?.folder_id || null"
      :submitting="submitting"
      :error-message="errorMessage"
      @submit="handleMoveDocumentSubmit"
    />

    <ResourceMoveModal
      v-model="moveFileOpen"
      :title="t('workspace.detail.resources.moveFileTitle')"
      :subtitle="t('workspace.detail.resources.moveFileSubtitle', { name: movingFile?.file_name || '' })"
      :root-label="t('workspace.detail.resources.rootFolder')"
      :destination-label="t('workspace.detail.resources.targetFolderLabel')"
      :confirm-text="t('workspace.detail.resources.confirmMoveFile')"
      :visible-folders="visibleFolders"
      :current-folder-id="movingFile?.folder_id || null"
      :submitting="submitting"
      :error-message="errorMessage"
      @submit="handleMoveFileSubmit"
    />
  </aside>
</template>



