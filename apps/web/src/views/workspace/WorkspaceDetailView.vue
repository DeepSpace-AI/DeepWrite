<script setup lang="ts">
import { computed, ref, watch, onBeforeUnmount } from 'vue'
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
import { useCollaboration } from '@/views/workspace/useCollaboration'
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

// Collaboration setup
const currentCollaboration = ref<ReturnType<typeof useCollaboration> | null>(null)
const collabError = ref('')
const isUpdatingFromYjs = ref(false)

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
const isApplyingDocumentContent = ref(false)
const lastRemoteSavedHash = ref('')
const isAutoSaving = ref(false)
const autoSaveError = ref('')
const lastLocalSaveAt = ref<number | null>(null)
const lastCloudSaveAt = ref<number | null>(null)
const LOCAL_DRAFT_DELAY_MS = 800
const DB_AUTOSAVE_DELAY_MS = 5000
let localDraftTimer: ReturnType<typeof setTimeout> | null = null
let dbAutosaveTimer: ReturnType<typeof setTimeout> | null = null
const headerDocument = computed(() => selectedDocument.value)
const headerCollaborators = computed(() => {
  if (!headerDocument.value) return []
  return docCollaborators[headerDocument.value.id] || docCollaborators.sample || []
})

const remoteCollaborators = computed(() => {
  return currentCollaboration.value?.remoteUsers.value || []
})

const isReadOnly = computed(() => {
  return currentCollaboration.value?.state.readOnly ?? false
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

function serializeContent(value: Record<string, unknown> | null | undefined) {
  try {
    return JSON.stringify(value ?? emptyDocumentContent())
  } catch {
    return ''
  }
}

function localDraftKey(documentId: string) {
  return `deepwrite:draft:${workspaceId.value}:${documentId}`
}

function clearLocalDraftTimer() {
  if (!localDraftTimer) return
  clearTimeout(localDraftTimer)
  localDraftTimer = null
}

function clearDbAutosaveTimer() {
  if (!dbAutosaveTimer) return
  clearTimeout(dbAutosaveTimer)
  dbAutosaveTimer = null
}

function readLocalDraft(documentId: string) {
  try {
    const raw = window.localStorage.getItem(localDraftKey(documentId))
    if (!raw) return null
    const parsed = JSON.parse(raw) as {
      contentJson?: Record<string, unknown> | null
      contentHtml?: string
      updatedAt?: number
    }
    if (!parsed || typeof parsed !== 'object') return null
    return parsed
  } catch {
    return null
  }
}

function writeLocalDraft() {
  if (!selectedDocument.value) return
  try {
    const now = Date.now()
    window.localStorage.setItem(
      localDraftKey(selectedDocument.value.id),
      JSON.stringify({
        contentJson: editorJson.value,
        contentHtml: editorContent.value,
        updatedAt: now,
      }),
    )
    lastLocalSaveAt.value = now
  } catch {
    // Ignore storage quota/private mode errors.
  }
}

function scheduleLocalDraftSave() {
  if (!selectedDocument.value || !isEditing.value) return
  clearLocalDraftTimer()
  localDraftTimer = setTimeout(() => {
    writeLocalDraft()
    localDraftTimer = null
  }, LOCAL_DRAFT_DELAY_MS)
}

async function performDbAutosave() {
  if (!selectedDocument.value || !isEditing.value) return
  if (isSavingDraft.value || isMutating.value) {
    scheduleDbAutosave()
    return
  }

  const currentHash = serializeContent(editorJson.value)
  if (!currentHash || currentHash === lastRemoteSavedHash.value) return

  isAutoSaving.value = true
  autoSaveError.value = ''
  const saved = await saveDocument(
    selectedDocument.value.id,
    {
      title: selectedDocument.value.title,
      contentJson: editorJson.value,
    },
    {
      source: 'autosave',
      snapshot: false,
    },
  )

  if (saved) {
    selectedDocId.value = saved.id
    editorJson.value = saved.content_json ?? editorJson.value
    lastRemoteSavedHash.value = serializeContent(saved.content_json)
    lastCloudSaveAt.value = Date.now()
  }
  else {
    autoSaveError.value = t('workspace.detail.autoSaveError')
  }
  isAutoSaving.value = false
}

function scheduleDbAutosave() {
  if (!selectedDocument.value || !isEditing.value) return
  clearDbAutosaveTimer()
  dbAutosaveTimer = setTimeout(async () => {
    dbAutosaveTimer = null
    await performDbAutosave()
  }, DB_AUTOSAVE_DELAY_MS)
}

function beginEditDocument(doc: WorkspaceDocument) {
  clearLocalDraftTimer()
  clearDbAutosaveTimer()

  const remoteContent = doc.content_json ?? emptyDocumentContent()
  const remoteUpdatedAt = new Date(doc.updated_at || 0).getTime()
  const localDraft = readLocalDraft(doc.id)
  const localUpdatedAt = localDraft?.updatedAt || 0
  const useLocalDraft = !!localDraft?.contentJson && localUpdatedAt > remoteUpdatedAt

  isApplyingDocumentContent.value = true
  selectedDocId.value = doc.id
  editorContent.value = useLocalDraft ? (localDraft?.contentHtml || '<p></p>') : '<p></p>'
  editorJson.value = useLocalDraft ? (localDraft?.contentJson || remoteContent) : remoteContent
  lastRemoteSavedHash.value = serializeContent(remoteContent)
  autoSaveError.value = ''
  lastLocalSaveAt.value = localUpdatedAt || null
  lastCloudSaveAt.value = remoteUpdatedAt || null
  isEditing.value = true
  mainPaneType.value = 'writing'
  setTimeout(() => {
    isApplyingDocumentContent.value = false
  }, 0)

  // Initialize collaboration
  collabError.value = ''
  if (currentCollaboration.value) {
    currentCollaboration.value.disconnect()
  }
  currentCollaboration.value = useCollaboration(doc.id, 'Anonymous')
  currentCollaboration.value.connect().catch((err) => {
    collabError.value = err instanceof Error ? err.message : 'Failed to connect'
    console.error('[beginEditDocument] Collab error:', err)
  })
}

function exitEditor() {
  isEditing.value = false
  mainPaneType.value = 'writing'

  // Disconnect collaboration
  if (currentCollaboration.value) {
    currentCollaboration.value.disconnect()
    currentCollaboration.value = null
  }
}

function switchMainPane(type: MainPaneType) {
  mainPaneType.value = type
}

async function saveDraft() {
  if (!selectedDocument.value) return
  clearDbAutosaveTimer()
  autoSaveError.value = ''
  isSavingDraft.value = true
  const saved = await saveDocument(selectedDocument.value.id, {
    title: selectedDocument.value.title,
    contentJson: editorJson.value,
    summary: t('workspace.detail.resources.saveSummary'),
  })
  if (saved) {
    selectedDocId.value = saved.id
    editorJson.value = saved.content_json ?? editorJson.value
    lastRemoteSavedHash.value = serializeContent(saved.content_json)
    lastCloudSaveAt.value = Date.now()
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

watch(
  () => [selectedDocId.value, isEditing.value, editorJson.value] as const,
  () => {
    if (!selectedDocId.value || !isEditing.value || isApplyingDocumentContent.value || isUpdatingFromYjs.value) return

    // Sync editor to Yjs
    if (currentCollaboration.value?.yText?.value) {
      const yText = currentCollaboration.value.yText.value
      const editorContentStr = JSON.stringify(editorJson.value ?? {})
      const yjsContentStr = yText.toString()

      // Only update if content actually changed
      if (editorContentStr !== yjsContentStr) {
        try {
          // Clear and insert new content
          yText.delete(0, yText.length)
          yText.insert(0, editorContentStr)
        } catch (err) {
          console.warn('[sync] Failed to push editor to Yjs:', err)
        }
      }
    }

    scheduleLocalDraftSave()
    scheduleDbAutosave()
  },
  { deep: true },
)

watch(
  () => [selectedDocId.value, currentCollaboration.value?.state.isConnected],
  () => {
    clearDbAutosaveTimer()
    clearLocalDraftTimer()

    // Setup Yjs sync when collaboration is ready
    if (currentCollaboration.value?.state.isConnected && selectedDocument.value && currentCollaboration.value.yText?.value) {
      const yText = currentCollaboration.value.yText.value

      // Sync initial content: Yjs -> Editor
      // If Yjs has content, use it; otherwise, initialize Yjs with editor content
      const yjsContent = yText.toString()
      if (yjsContent && !isApplyingDocumentContent.value) {
        // Parse Yjs content back to JSON
        try {
          const parsed = JSON.parse(yjsContent) as Record<string, unknown>
          isUpdatingFromYjs.value = true
          editorJson.value = parsed
          editorContent.value = '<p></p>'
          isUpdatingFromYjs.value = false
        } catch {
          console.warn('[sync] Failed to parse Yjs content as JSON')
        }
      } else if (!yjsContent && editorJson.value) {
        // Initialize Yjs with current editor content
        yText.insert(0, JSON.stringify(editorJson.value))
      }

      // Listen to Yjs updates
      const handleYjsUpdate = () => {
        try {
          const yjsStr = yText.toString()
          const parsed = JSON.parse(yjsStr) as Record<string, unknown>
          isUpdatingFromYjs.value = true
          editorJson.value = parsed
          isUpdatingFromYjs.value = false
        } catch {
          console.warn('[sync] Failed to parse Yjs content on update')
        }
      }

      yText.observe(handleYjsUpdate)

      // Cleanup observer on unmount or editor change
      return () => {
        yText.unobserve(handleYjsUpdate)
      }
    }
  },
)

watch(
  () => selectedDocId.value,
  () => {
    clearDbAutosaveTimer()
    clearLocalDraftTimer()
  },
)

onBeforeUnmount(() => {
  clearDbAutosaveTimer()
  clearLocalDraftTimer()
  writeLocalDraft()

  // Cleanup collaboration
  if (currentCollaboration.value) {
    currentCollaboration.value.disconnect()
    currentCollaboration.value = null
  }
})
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
        :remote-collaborators="remoteCollaborators"
        :is-auto-saving="isAutoSaving"
          :read-only="isReadOnly"
        :auto-save-error="autoSaveError"
        :last-local-save-at="lastLocalSaveAt"
        :last-cloud-save-at="lastCloudSaveAt"
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
