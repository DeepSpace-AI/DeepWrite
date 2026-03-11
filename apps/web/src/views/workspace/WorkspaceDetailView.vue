<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import IconArrowLeft from '~icons/mdi/arrow-left'
import IconFileDocumentOutline from '~icons/mdi/file-document-outline'
import IconBookOpenPageVariantOutline from '~icons/mdi/book-open-page-variant-outline'
import IconChartBar from '~icons/mdi/chart-bar'
import IconCogOutline from '~icons/mdi/cog-outline'
import WritingWorkspacePane from '@/views/workspace/components/WritingWorkspacePane.vue'
import LiteratureSearchPane from '@/views/workspace/components/LiteratureSearchPane.vue'
import DataAnalysisPane from '@/views/workspace/components/DataAnalysisPane.vue'
import WorkspaceSettingsPane from '@/views/workspace/components/WorkspaceSettingsPane.vue'
import { useWorkspaceResources } from '@/views/workspace/useWorkspaceResources'
import type { Collaborator, MainPaneType, WorkspaceDocument } from '@/views/workspace/types'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const workspaceId = computed(() => String(route.params.id || ''))
const isEditing = ref(false)
const selectedDocId = ref('')

const {
  documents,
  selectedFolderId,
  selectedFileIds,
  currentFolder,
  selectedFolderPath,
  visibleFolders,
  currentFolderDocuments,
  currentFolderFiles,
  errorMessage,
  isLoadingFolders,
  isLoadingDocuments,
  isLoadingFiles,
  isMutating,
  isUploadingFiles,
  selectFolder,
  toggleFolder,
  createFolder,
  renameFolder,
  deleteFolder,
  createNewDocument,
  renameDocument,
  saveDocument,
  uploadFilesToCurrentFolder,
  deleteFile,
  batchDeleteFiles,
  previewFile,
} = useWorkspaceResources(workspaceId)

const docCollaborators: Record<string, Collaborator[]> = {
  sample: [
    { id: 'u-1', name: 'Lin Chen' },
    { id: 'u-2', name: 'Ava Sun' },
    { id: 'u-3', name: 'Noah Gu' },
  ],
}

const selectedDocument = computed(() => documents.value.find((doc) => doc.id === selectedDocId.value) || null)
const editorContent = ref('<p></p>')
const editorJson = ref<Record<string, unknown> | null>(null)
const mainPaneType = ref<MainPaneType>('writing')
const isSavingDraft = ref(false)
const headerDocument = computed(() => selectedDocument.value)
const headerCollaborators = computed(() => {
  if (!headerDocument.value) return []
  return docCollaborators[headerDocument.value.id] || docCollaborators.sample || []
})

const isWritingMode = computed(() => mainPaneType.value === 'writing')
const showWritingEditor = computed(() => isWritingMode.value && !!selectedDocument.value && isEditing.value)
const isSubmitting = computed(() => isMutating.value || isSavingDraft.value)

const layoutClass = computed(() => {
  if (!isWritingMode.value) {
    return 'grid-cols-1 xl:grid-cols-[64px_minmax(0,1fr)]'
  }
  if (!showWritingEditor.value) {
    return 'grid-cols-1 xl:grid-cols-[64px_minmax(0,1fr)_360px]'
  }
  return 'grid-cols-1 xl:grid-cols-[64px_260px_minmax(0,1fr)_300px]'
})

function emptyDocumentContent() {
  return { type: 'doc', content: [] as unknown[] }
}

function beginEditDocument(doc: WorkspaceDocument) {
  selectedDocId.value = doc.id
  editorContent.value = '<p></p>'
  editorJson.value = doc.content_json ?? emptyDocumentContent()
  isEditing.value = true
  mainPaneType.value = 'writing'
}

function exitEditor() {
  isEditing.value = false
  mainPaneType.value = 'writing'
}

function switchMainPane(type: MainPaneType) {
  mainPaneType.value = type
}

async function saveDraft() {
  if (!selectedDocument.value) return
  isSavingDraft.value = true
  const saved = await saveDocument(selectedDocument.value.id, {
    title: selectedDocument.value.title,
    contentJson: editorJson.value,
    summary: t('workspace.detail.resources.saveSummary'),
  })
  if (saved) {
    selectedDocId.value = saved.id
    editorJson.value = saved.content_json ?? editorJson.value
  }
  isSavingDraft.value = false
}

async function handleCreateDocument(payload: { title: string }) {
  const document = await createNewDocument(payload.title)
  if (document) {
    beginEditDocument(document)
  }
}

function handleSelectedFileIdsUpdate(fileIds: string[]) {
  selectedFileIds.value = fileIds
}

function backToList() {
  router.push({ name: 'workspace-list' })
}
</script>

<template>
  <section class="h-full min-h-0">
    <section class="grid h-full min-h-0 gap-4" :class="layoutClass">
      <aside class="h-full min-h-0 rounded-sm border border-base-300 bg-base-100 shadow-sm transition-all duration-300">
        <div class="flex h-full flex-col items-center px-2 py-3">
          <button type="button" class="btn btn-square btn-sm rounded-sm" :title="t('workspace.backToList')" @click="backToList">
            <IconArrowLeft class="h-4 w-4" />
          </button>

          <div class="my-3 h-px w-8 bg-base-300" />

          <div class="flex flex-col gap-2">
            <button
              type="button"
              class="btn btn-square btn-sm rounded-sm"
              :class="mainPaneType === 'writing' ? 'btn-primary' : 'btn-ghost'"
              :title="t('workspace.detail.navWriting')"
              @click="switchMainPane('writing')"
            >
              <IconFileDocumentOutline class="h-4 w-4" />
            </button>

            <button
              type="button"
              class="btn btn-square btn-sm rounded-sm"
              :class="mainPaneType === 'literature' ? 'btn-primary' : 'btn-ghost'"
              :title="t('workspace.detail.navLiterature')"
              @click="switchMainPane('literature')"
            >
              <IconBookOpenPageVariantOutline class="h-4 w-4" />
            </button>

            <button
              type="button"
              class="btn btn-square btn-sm rounded-sm"
              :class="mainPaneType === 'analysis' ? 'btn-primary' : 'btn-ghost'"
              :title="t('workspace.detail.navAnalysis')"
              @click="switchMainPane('analysis')"
            >
              <IconChartBar class="h-4 w-4" />
            </button>

            <button
              type="button"
              class="btn btn-square btn-sm rounded-sm"
              :class="mainPaneType === 'settings' ? 'btn-primary' : 'btn-ghost'"
              :title="t('workspace.detail.navSettings')"
              @click="switchMainPane('settings')"
            >
              <IconCogOutline class="h-4 w-4" />
            </button>
          </div>
        </div>
      </aside>

      <WritingWorkspacePane
        v-if="isWritingMode"
        :visible-folders="visibleFolders"
        :selected-folder-id="selectedFolderId"
        :selected-folder="currentFolder"
        :selected-folder-path="selectedFolderPath"
        :current-folder-documents="currentFolderDocuments"
        :current-folder-files="currentFolderFiles"
        :selected-file-ids="selectedFileIds"
        :active-document-id="selectedDocId"
        :loading-folders="isLoadingFolders"
        :loading-documents="isLoadingDocuments"
        :loading-files="isLoadingFiles"
        :uploading-files="isUploadingFiles"
        :submitting="isSubmitting"
        :resource-error-message="errorMessage"
        :show-writing-editor="showWritingEditor"
        :selected-document="selectedDocument"
        :editor-content="editorContent"
        :editor-json="editorJson"
        :header-document="headerDocument"
        :header-collaborators="headerCollaborators"
        :workspace-id="workspaceId"
        @select-root="selectFolder(null)"
        @select-folder="selectFolder($event)"
        @toggle-folder="toggleFolder($event)"
        @create-folder="createFolder($event)"
        @rename-folder="renameFolder($event.folderId, { name: $event.name, description: $event.description })"
        @delete-folder="deleteFolder($event)"
        @create-document="handleCreateDocument"
        @begin-edit="beginEditDocument"
        @rename-document="renameDocument($event.documentId, $event.title)"
        @upload-files="uploadFilesToCurrentFolder($event)"
        @delete-file="deleteFile($event)"
        @preview-file="previewFile($event)"
        @batch-delete-files="batchDeleteFiles($event)"
        @update:selected-file-ids="handleSelectedFileIdsUpdate($event)"
        @update:editor-content="editorContent = $event"
        @update:editor-json="editorJson = $event"
        @save-draft="saveDraft"
        @exit-editor="exitEditor"
      />

      <LiteratureSearchPane v-else-if="mainPaneType === 'literature'" />
      <DataAnalysisPane v-else-if="mainPaneType === 'analysis'" />
      <WorkspaceSettingsPane v-else />
    </section>
  </section>
</template>
