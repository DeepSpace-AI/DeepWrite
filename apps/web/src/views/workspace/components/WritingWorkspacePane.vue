<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import DEditor from '@/components/editor/DEditor.vue'
import WorkspaceResourcePane from '@/views/workspace/components/WorkspaceResourcePane.vue'
import PDFReaderPane from '@/views/workspace/components/PDFReaderPane.vue'
import type { Collaborator, VisibleFolderNode, WorkspaceDocument, WorkspaceFile, WorkspaceFolder, PDFAnnotation } from '@/views/workspace/types'

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
  resourceErrorMessage?: string
  uploadStatusMessage?: string
  showWritingEditor: boolean
  selectedDocument: WorkspaceDocument | null
  editorContent: string
  editorJson: Record<string, unknown> | null
  headerDocument: WorkspaceDocument | null
  headerCollaborators: Collaborator[]
  remoteCollaborators: Array<{ clientId: number; name: string; avatarUrl?: string; color?: string; selection?: { anchor: number; head: number } }>
  isAutoSaving: boolean
  readOnly?: boolean
  collabError?: string
  isCollabConnecting?: boolean
  isCollabConnected?: boolean
  autoSaveError?: string
  lastLocalSaveAt?: number | null
  lastCloudSaveAt?: number | null
  workspaceId: string
  showPDFReader: boolean
  currentPDFFile: WorkspaceFile | null
  currentPDFPreviewUrl: string
  currentFileAnnotations: PDFAnnotation[]
  isLoadingAnnotations: boolean
}>()

const emit = defineEmits<{
  (e: 'select-root'): void
  (e: 'select-folder', folderId: string): void
  (e: 'toggle-folder', folderId: string): void
  (e: 'create-folder', payload: { name: string; description: string; parentId: string | null }): void
  (e: 'rename-folder', payload: { folderId: string; name: string; description: string }): void
  (e: 'delete-folder', folder: WorkspaceFolder): void
  (e: 'create-document', payload: { title: string }): void
  (e: 'begin-edit', doc: WorkspaceDocument): void
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
  (e: 'update:editorContent', value: string): void
  (e: 'update:editorJson', value: Record<string, unknown> | null): void
  (e: 'collab-selection-change', value: { anchor: number; head: number } | null): void
  (e: 'save-draft'): void
  (e: 'exit-editor'): void
  (e: 'close-pdf-reader'): void
  (e: 'annotation-create', annotation: Partial<PDFAnnotation>): void
  (e: 'annotation-update', annotationId: string, updates: Partial<PDFAnnotation>): void
  (e: 'annotation-delete', annotationId: string): void
}>()

const { t } = useI18n()

const contentProxy = computed({
  get: () => props.editorContent,
  set: (value: string) => emit('update:editorContent', value),
})

function formatTime(value?: string) {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(date)
}

function collaboratorInitial(name: string) {
  return name.trim().slice(0, 1).toUpperCase() || 'U'
}

function formatTimestamp(value?: number | null) {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '--'
  return new Intl.DateTimeFormat('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).format(date)
}

const autoSaveStatusText = computed(() => {
  if (props.autoSaveError) {
    return t('workspace.detail.autoSaveError')
  }
  if (props.isAutoSaving) {
    return t('workspace.detail.autoSaveSaving')
  }
  if (props.lastCloudSaveAt) {
    return t('workspace.detail.autoSaveCloudAt', { time: formatTimestamp(props.lastCloudSaveAt) })
  }
  if (props.lastLocalSaveAt) {
    return t('workspace.detail.autoSaveLocalAt', { time: formatTimestamp(props.lastLocalSaveAt) })
  }
  return t('workspace.detail.autoSaveIdle')
})

const collabStatusText = computed(() => {
  if (props.collabError) {
    return props.collabError
  }
  if (props.isCollabConnecting) {
    return '协作连接中'
  }
  if (props.isCollabConnected) {
    return '协作已连接'
  }
  return '协作未连接'
})

const onlineCollaboratorCount = computed(() => props.headerCollaborators.length)

function collaboratorAvatarStyle(color?: string) {
  if (!color) return undefined
  return {
    backgroundColor: `${color}22`,
    borderColor: color,
    color,
  }
}
</script>

<template>
  <WorkspaceResourcePane
    :visible-folders="visibleFolders"
    :selected-folder-id="selectedFolderId"
    :selected-folder="selectedFolder"
    :selected-folder-path="selectedFolderPath"
    :current-folder-documents="currentFolderDocuments"
    :current-folder-files="currentFolderFiles"
    :selected-file-ids="selectedFileIds"
    :active-document-id="activeDocumentId"
    :loading-folders="loadingFolders"
    :loading-documents="loadingDocuments"
    :loading-files="loadingFiles"
    :uploading-files="uploadingFiles"
    :submitting="submitting"
    :error-message="resourceErrorMessage"
    :upload-status-message="uploadStatusMessage"
    @select-root="emit('select-root')"
    @select-folder="emit('select-folder', $event)"
    @toggle-folder="emit('toggle-folder', $event)"
    @create-folder="emit('create-folder', $event)"
    @rename-folder="emit('rename-folder', $event)"
    @delete-folder="emit('delete-folder', $event)"
    @create-document="emit('create-document', $event)"
    @open-document="emit('begin-edit', $event)"
    @rename-document="emit('rename-document', $event)"
    @move-document="emit('move-document', $event)"
    @delete-document="emit('delete-document', $event)"
    @upload-files="emit('upload-files', $event)"
    @rename-file="emit('rename-file', $event)"
    @move-file="emit('move-file', $event)"
    @delete-file="emit('delete-file', $event)"
    @preview-file="emit('preview-file', $event)"
    @batch-delete-files="emit('batch-delete-files', $event)"
    @update:selected-file-ids="emit('update:selectedFileIds', $event)"
  />

  <main v-if="showWritingEditor || showPDFReader" class="paper-panel h-full min-h-0 flex flex-col">
    <template v-if="showWritingEditor">
    <div class="bg-(--surface-overlay) px-4 py-3">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="min-w-0">
          <div class="flex items-center gap-2 text-sm">
            <span class="text-[11px] font-mono uppercase tracking-[0.2em] text-base-content/45">{{ t('workspace.detail.currentDocLabel') }}</span>
            <span class="text-base-content/35">|</span>
            <h3 class="max-w-70 truncate text-base font-semibold text-base-content">{{ selectedDocument?.title }}</h3>
          </div>
          <div class="mt-1 flex flex-wrap items-center gap-2 text-xs text-base-content/55">
            <span
              class="tooltip tooltip-top"
              :data-tip="collabStatusText"
            >
              <span
                class="inline-block h-2 w-2 rounded-full"
                :class="[
                  collabError
                    ? 'bg-error'
                    : isCollabConnecting
                      ? 'bg-warning animate-pulse'
                      : isCollabConnected
                        ? 'bg-success'
                        : 'bg-base-content/30',
                ]"
              />
            </span>
            <span>{{ collabStatusText }}</span>
            <span class="text-base-content/35">|</span>
            <span
              class="tooltip tooltip-top"
              :data-tip="autoSaveStatusText"
            >
              <span
                class="inline-block h-2 w-2 rounded-full"
                :class="[
                  autoSaveError
                    ? 'bg-error'
                    : isAutoSaving
                      ? 'bg-warning animate-pulse'
                      : (lastCloudSaveAt || lastLocalSaveAt)
                        ? 'bg-success'
                        : 'bg-base-content/30',
                ]"
              />
            </span>
            <span>{{ t('workspace.detail.updatedAtLabel') }}: {{ formatTime(headerDocument?.updated_at) }}</span>
            <span class="text-base-content/35">|</span>
            <span>{{ t('workspace.detail.resources.versionLabel') }}: {{ headerDocument?.current_version || 1 }}</span>
            <span class="text-base-content/35">|</span>
            <span>{{ t('workspace.detail.collaboratorsLabel') }}: {{ onlineCollaboratorCount }}</span>
            <span v-if="onlineCollaboratorCount > 0" class="text-base-content/35">|</span>
            <span v-if="onlineCollaboratorCount > 0" class="text-xs text-warning">
              {{ onlineCollaboratorCount }} online
            </span>
          </div>
        </div>

        <div class="flex flex-wrap items-center justify-end gap-2">
          <div class="avatar-group -space-x-3 rtl:space-x-reverse">
            <div
              v-for="person in headerCollaborators.slice(0, 5)"
              :key="person.id"
              class="avatar tooltip tooltip-bottom"
              :data-tip="`${person.name} (online)`"
            >
              <div
                class="w-7 bg-(--surface-sunken) text-[10px] text-base-content"
                :style="collaboratorAvatarStyle(person.color)"
              >
                <img v-if="person.avatarUrl" :src="person.avatarUrl" :alt="person.name" />
                <span v-else class="inline-flex h-full w-full items-center justify-center">{{ collaboratorInitial(person.name) }}</span>
              </div>
            </div>

            <div
              v-if="onlineCollaboratorCount > 5"
              class="avatar placeholder"
            >
              <div class="w-7 bg-neutral text-[10px] text-neutral-content">
                +{{ onlineCollaboratorCount - 5 }}
              </div>
            </div>
          </div>
          <button type="button" class="btn btn-sm rounded-sm" :disabled="submitting" @click="emit('save-draft')">{{ t('workspace.detail.saveDraft') }}</button>
          <button type="button" class="btn btn-sm btn-ghost rounded-sm" @click="emit('exit-editor')">{{ t('workspace.detail.exitEditor') }}</button>
        </div>
      </div>

      <div v-if="collabError" class="mt-3 rounded-sm border border-error/30 bg-error/10 px-3 py-2 text-xs text-error">
        {{ collabError }}
      </div>
    </div>
    </template>

    <div class="min-h-0 flex-1">
      <DEditor
        v-if="showWritingEditor"
        class="h-full"
        v-model="contentProxy"
        :model-json="editorJson"
        :read-only="readOnly"
        :remote-cursors="remoteCollaborators"
        @update:model-json="emit('update:editorJson', $event)"
        @local-selection-change="emit('collab-selection-change', $event)"
      />
      <PDFReaderPane
        v-else-if="showPDFReader && currentPDFFile"
        class="h-full"
        :file="currentPDFFile"
        :preview-url="currentPDFPreviewUrl"
        :annotations="currentFileAnnotations"
        :workspace-id="workspaceId"
        @close="emit('close-pdf-reader')"
        @annotation-create="emit('annotation-create', $event)"
        @annotation-update="(id, updates) => emit('annotation-update', id, updates)"
        @annotation-delete="(id) => emit('annotation-delete', id)"
      />
    </div>
  </main>

  <aside class="paper-panel h-full min-h-0 transition-all duration-300 flex flex-col">
    <div class="bg-(--surface-overlay) px-4 py-3">
      <h3 class="text-sm font-semibold text-base-content">{{ t('workspace.detail.agentTitle') }}</h3>
    </div>
    <div class="flex min-h-0 flex-1 flex-col">
      <div class="flex-1 space-y-2 overflow-y-auto px-3 py-3">
        <div
          v-for="msg in [
            { id: 'm-1', role: 'agent', text: '已就当前文档提炼 3 条结构优化建议。' },
            { id: 'm-2', role: 'user', text: '请优先优化第一章开头的背景段落。' },
            { id: 'm-3', role: 'agent', text: '建议把“问题背景”改为“行业转折点”，可增强叙事张力。' },
          ]"
          :key="msg.id"
          class="rounded-sm px-3 py-2 text-sm"
          :class="msg.role === 'agent' ? 'bg-(--surface-overlay) text-base-content' : 'bg-primary/10 text-base-content/85'"
        >
          {{ msg.text }}
        </div>
      </div>
      <div class="bg-(--surface-overlay) p-3">
        <div class="rounded-sm bg-(--surface-sunken) px-3 py-2 text-sm text-base-content/50">
          {{ t('workspace.detail.agentInputHint') }}
        </div>
      </div>
    </div>
  </aside>
</template>



