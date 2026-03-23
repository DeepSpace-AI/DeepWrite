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
import { useUserStore } from '@/stores/user'
import { useWorkspaceResources } from '@/views/workspace/useWorkspaceResources'
import { useCollaboration } from '@/views/workspace/useCollaboration'
import type { Collaborator, MainPaneType, WorkspaceDocument } from '@/views/workspace/types'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const userStore = useUserStore()

const workspaceId = computed(() => String(route.params.id || ''))
const routeDocumentId = computed(() => {
  const raw = route.query.doc
  if (typeof raw === 'string') return raw.trim()
  if (Array.isArray(raw)) return String(raw[0] || '').trim()
  return ''
})
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
  uploadStatusMessage,
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
  moveDocument,
  deleteDocument,
  saveDocument,
  uploadFilesToCurrentFolder,
  renameFile,
  moveFile,
  deleteFile,
  batchDeleteFiles,
  previewFile,
  currentPDFfile,
  currentPDFPreviewUrl,
  currentFileAnnotations,
  isLoadingAnnotations,
  closePDFReader,
  addAnnotation,
  updateAnnotation,
  removeAnnotation,
} = useWorkspaceResources(workspaceId)

// Collaboration setup
const currentCollaboration = ref<ReturnType<typeof useCollaboration> | null>(null)
const collabError = ref('')
const isUpdatingFromYjs = ref(false)

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
const remoteCollaborators = computed(() => {
  return currentCollaboration.value?.remoteUsers || []
})

const isReadOnly = computed(() => {
  return currentCollaboration.value?.state.readOnly ?? false
})
const isCollabConnecting = computed(() => currentCollaboration.value?.state.isConnecting ?? false)
const isCollabConnected = computed(() => currentCollaboration.value?.state.isConnected ?? false)
const collabUserName = computed(() => {
  const displayName = userStore.user?.displayName?.trim()
  if (displayName) return displayName

  const emailName = userStore.user?.email?.split('@')[0]?.trim()
  if (emailName) return emailName

  return 'Anonymous'
})
const headerCollaborators = computed<Collaborator[]>(() => {
  if (!selectedDocument.value || !isCollabConnected.value) return []

  const collaborators: Collaborator[] = []
  const currentUserId = userStore.user?.id?.trim() || 'local-user'
  collaborators.push({
    id: currentUserId,
    name: collabUserName.value,
    avatarUrl: userStore.user?.avatarUrl || '',
  })

  remoteCollaborators.value.forEach((person) => {
    collaborators.push({
      id: `remote-${person.clientId}`,
      name: person.name,
      avatarUrl: person.avatarUrl || '',
      color: person.color,
    })
  })

  return collaborators
})
const isWritingMode = computed(() => mainPaneType.value === 'writing')
const showWritingEditor = computed(() => isWritingMode.value && !!selectedDocument.value && isEditing.value && !currentPDFfile.value)
const showPDFReader = computed(() => isWritingMode.value && !!currentPDFfile.value)
const isSubmitting = computed(() => isMutating.value || isSavingDraft.value)

const layoutClass = computed(() => {
  if (!isWritingMode.value) {
    return 'grid-cols-1 xl:grid-cols-[64px_minmax(0,1fr)]'
  }
  if (!showWritingEditor.value && !showPDFReader.value) {
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

  const collabMapStr = currentCollaboration.value?.yContentMap?.get('content_json') || ''
  const collabJson = (() => {
    if (!collabMapStr) return editorJson.value
    try {
      return JSON.parse(collabMapStr) as Record<string, unknown>
    } catch {
      return editorJson.value
    }
  })()

  const currentHash = serializeContent(collabJson)
  if (!currentHash || currentHash === lastRemoteSavedHash.value) return

  isAutoSaving.value = true
  autoSaveError.value = ''
  const saved = await saveDocument(
    selectedDocument.value.id,
    {
      title: selectedDocument.value.title,
      contentJson: collabJson,
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
  const collaboration = useCollaboration(doc.id, {
    name: collabUserName.value,
    avatarUrl: userStore.user?.avatarUrl || '',
  })
  currentCollaboration.value = collaboration
  collaboration.connect().catch((err) => {
    collabError.value = err instanceof Error ? err.message : 'Failed to connect'
    console.error('[beginEditDocument] Collab error:', err)
  })
}

function syncDocumentQuery(documentId: string) {
  const nextDocId = documentId.trim()
  const currentDocId = routeDocumentId.value
  if (nextDocId === currentDocId) return

  const nextQuery = { ...route.query } as Record<string, string | string[] | undefined>
  if (nextDocId) {
    nextQuery.doc = nextDocId
  } else {
    delete nextQuery.doc
  }

  router.replace({ query: nextQuery })
}

function openDocumentFromRoute(documentId: string) {
  const nextDocId = documentId.trim()
  if (!nextDocId) return

  const matchedDocument = documents.value.find((doc) => doc.id === nextDocId)
  if (!matchedDocument) return
  if (selectedDocId.value === matchedDocument.id && isEditing.value) return

  beginEditDocument(matchedDocument)
}

function handleCollabSelectionChange(selection: { anchor: number, head: number } | null) {
  const prov = currentCollaboration.value?.provider
  if (!prov) {
    return
  }

  if (!selection) {
    prov.clearLocalSelection()
    return
  }

  prov.setLocalSelection(selection.anchor, selection.head)
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

function handleClosePDFReader() {
  closePDFReader()
}

async function handleAnnotationCreate(annotation: {
  type: 'highlight' | 'note' | 'drawing'
  page: number
  rect_x: number
  rect_y: number
  rect_width: number
  rect_height: number
  color?: string
  content?: string
  paths?: any[]
}) {
  await addAnnotation({
    type: annotation.type,
    page: annotation.page,
    rect_x: annotation.rect_x,
    rect_y: annotation.rect_y,
    rect_width: annotation.rect_width,
    rect_height: annotation.rect_height,
    color: annotation.color,
    content: annotation.content,
    paths: annotation.paths ? JSON.stringify(annotation.paths) : undefined,
  })
}

async function handleAnnotationUpdate(annotationId: string, updates: {
  rect_x?: number
  rect_y?: number
  rect_width?: number
  rect_height?: number
  color?: string
  content?: string
  paths?: string
}) {
  await updateAnnotation(annotationId, updates)
}

async function handleAnnotationDelete(annotationId: string) {
  await removeAnnotation(annotationId)
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

async function handleDeleteDocument(documentId: string) {
  const deletingCurrent = selectedDocId.value === documentId
  const success = await deleteDocument(documentId)
  if (success && deletingCurrent) {
    selectedDocId.value = ''
    editorJson.value = null
    editorContent.value = '<p></p>'
    exitEditor()
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

    // Sync editor to Yjs (prefer atomic map field over raw text stream)
    const collaboration = currentCollaboration.value
    const yText = collaboration?.yText
    const contentMap = collaboration?.yContentMap
    if (collaboration?.state.isConnected) {
      const editorContentStr = JSON.stringify(editorJson.value ?? {})
      const mapContentStr = (contentMap?.get('content_json') || '').toString()
      const yjsContentStr = mapContentStr || yText?.toString() || ''

      // Only update if content actually changed
      if (editorContentStr !== yjsContentStr) {
        try {
          // Keep delete+insert in a single transaction to avoid transient invalid states.
          collaboration?.yjsDoc?.transact(() => {
            contentMap?.set('content_json', editorContentStr)
            if (yText && yText.toString() !== editorContentStr) {
              yText.delete(0, yText.length)
              yText.insert(0, editorContentStr)
            }
          }, 'editor-sync')
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
  (_, __, onCleanup) => {
    clearDbAutosaveTimer()
    clearLocalDraftTimer()

    // Setup Yjs sync when collaboration is ready
    const collaboration = currentCollaboration.value
    const yText = collaboration?.yText
    const contentMap = collaboration?.yContentMap
    if (collaboration?.state.isConnected && selectedDocument.value) {

      // Sync initial content: Yjs -> Editor
      // Prefer atomic map payload, fallback to legacy text payload for compatibility.
      const yjsContent = (contentMap?.get('content_json') || '').toString() || yText?.toString() || ''
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
        collaboration?.yjsDoc?.transact(() => {
          const initialContent = JSON.stringify(editorJson.value)
          contentMap?.set('content_json', initialContent)
          if (yText && yText.toString() !== initialContent) {
            yText.delete(0, yText.length)
            yText.insert(0, initialContent)
          }
        }, 'editor-bootstrap')
      }

      // Listen to Yjs updates
      const handleYjsUpdate = () => {
        try {
          const yjsStr = (contentMap?.get('content_json') || '').toString() || yText?.toString() || ''
          if (!yjsStr) {
            return
          }
          const parsed = JSON.parse(yjsStr) as Record<string, unknown>
          isUpdatingFromYjs.value = true
          editorJson.value = parsed
          isUpdatingFromYjs.value = false
        } catch {
          console.warn('[sync] Failed to parse Yjs content on update')
        }
      }

      if (contentMap) {
        contentMap.observe(handleYjsUpdate)
      }
      if (yText) {
        yText.observe(handleYjsUpdate)
      }

      // Cleanup observer on unmount or editor change
      onCleanup(() => {
        if (contentMap) {
          contentMap.unobserve(handleYjsUpdate)
        }
        if (yText) {
          yText.unobserve(handleYjsUpdate)
        }
      })
    }
  },
)

watch(
  () => selectedDocId.value,
  (documentId) => {
    clearDbAutosaveTimer()
    clearLocalDraftTimer()
    syncDocumentQuery(documentId)
  },
)

watch(
  () => [routeDocumentId.value, documents.value.length] as const,
  ([documentId]) => {
    if (!documentId) return
    openDocumentFromRoute(documentId)
  },
  { immediate: true },
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
      <aside class="paper-panel h-full min-h-0 transition-all duration-300">
        <div class="flex h-full flex-col items-center px-2 py-3">
          <button type="button" class="btn btn-square btn-sm rounded-sm" :title="t('workspace.backToList')" @click="backToList">
            <IconArrowLeft class="h-4 w-4" />
          </button>

          <div class="my-3 h-px w-8 bg-base-content/15" />

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
        :upload-status-message="uploadStatusMessage"
        :show-writing-editor="showWritingEditor"
        :selected-document="selectedDocument"
        :editor-content="editorContent"
        :editor-json="editorJson"
        :header-document="headerDocument"
        :header-collaborators="headerCollaborators"
        :remote-collaborators="remoteCollaborators"
        :is-auto-saving="isAutoSaving"
        :read-only="isReadOnly"
        :collab-error="collabError || currentCollaboration?.state.error || ''"
        :is-collab-connecting="isCollabConnecting"
        :is-collab-connected="isCollabConnected"
        :auto-save-error="autoSaveError"
        :last-local-save-at="lastLocalSaveAt"
        :last-cloud-save-at="lastCloudSaveAt"
        :workspace-id="workspaceId"
        :show-p-d-f-reader="showPDFReader"
        :current-p-d-f-file="currentPDFfile"
        :current-p-d-f-preview-url="currentPDFPreviewUrl"
        :current-file-annotations="currentFileAnnotations"
        :is-loading-annotations="isLoadingAnnotations"
        @select-root="selectFolder(null)"
        @select-folder="selectFolder($event)"
        @toggle-folder="toggleFolder($event)"
        @create-folder="createFolder($event)"
        @rename-folder="renameFolder($event.folderId, { name: $event.name, description: $event.description })"
        @delete-folder="deleteFolder($event)"
        @create-document="handleCreateDocument"
        @begin-edit="beginEditDocument"
        @rename-document="renameDocument($event.documentId, $event.title)"
        @move-document="moveDocument($event.documentId, $event.folderId)"
        @delete-document="handleDeleteDocument($event)"
        @upload-files="uploadFilesToCurrentFolder($event)"
        @rename-file="renameFile($event.fileId, $event.fileName)"
        @move-file="moveFile($event.fileId, $event.folderId)"
        @delete-file="deleteFile($event)"
        @preview-file="previewFile($event)"
        @batch-delete-files="batchDeleteFiles($event)"
        @update:selected-file-ids="handleSelectedFileIdsUpdate($event)"
        @update:editor-content="editorContent = $event"
        @update:editor-json="editorJson = $event"
        @collab-selection-change="handleCollabSelectionChange($event)"
        @save-draft="saveDraft"
        @exit-editor="exitEditor"
        @close-pdf-reader="handleClosePDFReader"
        @annotation-create="handleAnnotationCreate"
        @annotation-update="handleAnnotationUpdate"
        @annotation-delete="handleAnnotationDelete"
      />

      <LiteratureSearchPane v-else-if="mainPaneType === 'literature'" />
      <DataAnalysisPane v-else-if="mainPaneType === 'analysis'" />
      <WorkspaceSettingsPane v-else />
    </section>
  </section>
</template>
