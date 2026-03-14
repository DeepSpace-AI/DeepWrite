import { computed, ref, watch, type Ref } from 'vue'
import { ApiError } from '@/api/http'
import {
  createDocument,
  deleteDocument as deleteDocumentById,
  listDocuments,
  saveDocumentVersion,
  updateDocumentMeta,
} from '@/api/document'
import {
  batchDeleteWorkspaceFiles,
  completeWorkspaceUpload,
  createWorkspaceFolder,
  deleteWorkspaceFile,
  deleteWorkspaceFolder,
  getWorkspaceFileDetail,
  listWorkspaceFiles,
  listWorkspaceFolders,
  presignWorkspaceUpload,
  updateWorkspaceFile,
  updateWorkspaceFolder,
} from '@/api/workspace'
import type { VisibleFolderNode, WorkspaceDocument, WorkspaceFile, WorkspaceFolder } from '@/views/workspace/types'

const ROOT_KEY = '__root__'

function normalizeFolderId(folderId?: string | null) {
  const value = folderId?.trim()
  return value ? value : null
}

function isRootFolder(folder: WorkspaceFolder) {
  return !normalizeFolderId(folder.parent_id)
}

function parentKey(parentId?: string | null) {
  return normalizeFolderId(parentId) || ROOT_KEY
}

function emptyDocumentContent() {
  return { type: 'doc', content: [] as unknown[] }
}

function sortByUpdatedDesc<T extends { updated_at?: string }>(items: T[]) {
  return [...items].sort((left, right) => {
    const leftTime = new Date(left.updated_at || 0).getTime()
    const rightTime = new Date(right.updated_at || 0).getTime()
    return rightTime - leftTime
  })
}

export function useWorkspaceResources(workspaceId: Ref<string>) {
  const foldersByParent = ref<Record<string, WorkspaceFolder[]>>({})
  const loadedParents = ref<string[]>([])
  const expandedFolderIds = ref<string[]>([])
  const documents = ref<WorkspaceDocument[]>([])
  const files = ref<WorkspaceFile[]>([])
  const selectedFolderId = ref<string | null>(null)
  const selectedFileIds = ref<string[]>([])
  const errorMessage = ref('')
  const uploadStatusMessage = ref('')

  const isLoadingFolders = ref(false)
  const isLoadingDocuments = ref(false)
  const isLoadingFiles = ref(false)
  const isMutating = ref(false)
  const isUploadingFiles = ref(false)
  let uploadStatusTimer: ReturnType<typeof setTimeout> | null = null

  function clearUploadStatusTimer() {
    if (!uploadStatusTimer) return
    clearTimeout(uploadStatusTimer)
    uploadStatusTimer = null
  }

  function setUploadStatusMessage(message: string, autoClearMS = 0) {
    clearUploadStatusTimer()
    uploadStatusMessage.value = message
    if (autoClearMS > 0) {
      uploadStatusTimer = setTimeout(() => {
        uploadStatusMessage.value = ''
        uploadStatusTimer = null
      }, autoClearMS)
    }
  }

  const folderMap = computed<Record<string, WorkspaceFolder>>(() => {
    const map: Record<string, WorkspaceFolder> = {}
    Object.values(foldersByParent.value).flat().forEach((folder) => {
      map[folder.id] = folder
    })
    return map
  })

  const currentFolder = computed(() => {
    if (!selectedFolderId.value) return null
    return folderMap.value[selectedFolderId.value] || null
  })

  const selectedFolderPath = computed<WorkspaceFolder[]>(() => {
    if (!selectedFolderId.value) return []

    const path: WorkspaceFolder[] = []
    const visited = new Set<string>()
    let cursor: string | null = selectedFolderId.value

    while (cursor) {
      if (visited.has(cursor)) break
      visited.add(cursor)

      const folder = folderMap.value[cursor]
      if (!folder) break

      path.unshift(folder)
      cursor = normalizeFolderId(folder.parent_id)
    }

    return path
  })

  const visibleFolders = computed<VisibleFolderNode[]>(() => {
    const flattened: VisibleFolderNode[] = []

    const visit = (parentId: string | null, depth: number) => {
      const children = foldersByParent.value[parentKey(parentId)] || []
      children.forEach((folder) => {
        const expanded = expandedFolderIds.value.includes(folder.id)
        flattened.push({ folder, depth, expanded })
        if (expanded) {
          visit(folder.id, depth + 1)
        }
      })
    }

    visit(null, 0)
    return flattened
  })

  const currentFolderDocuments = computed(() => {
    return sortByUpdatedDesc(
      documents.value.filter((doc) => normalizeFolderId(doc.folder_id) === selectedFolderId.value),
    )
  })

  const currentFolderFiles = computed(() => sortByUpdatedDesc(files.value))

  function updateError(error: unknown, fallback: string) {
    if (error instanceof ApiError && error.message) {
      errorMessage.value = error.message
      return
    }
    if (error instanceof Error && error.message) {
      errorMessage.value = error.message
      return
    }
    errorMessage.value = fallback
  }

  function clearError() {
    errorMessage.value = ''
    clearUploadStatusTimer()
    uploadStatusMessage.value = ''
  }

  function resetState() {
    foldersByParent.value = {}
    loadedParents.value = []
    expandedFolderIds.value = []
    documents.value = []
    files.value = []
    selectedFolderId.value = null
    selectedFileIds.value = []
    clearError()
  }

  function replaceFolderChildren(parentId: string | null, folders: WorkspaceFolder[]) {
    foldersByParent.value = {
      ...foldersByParent.value,
      [parentKey(parentId)]: sortByUpdatedDesc(folders),
    }
  }

  function upsertFolder(folder: WorkspaceFolder) {
    const key = parentKey(folder.parent_id)
    const siblings = foldersByParent.value[key] || []
    const exists = siblings.some((item) => item.id === folder.id)
    const next = exists
      ? siblings.map((item) => (item.id === folder.id ? folder : item))
      : [folder, ...siblings]

    replaceFolderChildren(folder.parent_id || null, next)

    Object.keys(foldersByParent.value).forEach((entryKey) => {
      if (entryKey === key) return
      const branch = foldersByParent.value[entryKey]
      if (!branch?.some((item) => item.id === folder.id)) return
      replaceFolderChildren(entryKey === ROOT_KEY ? null : entryKey, branch.filter((item) => item.id !== folder.id))
    })
  }

  function removeFolder(folderId: string) {
    const nextEntries = Object.entries(foldersByParent.value).map(([key, value]) => [key, value.filter((item) => item.id !== folderId)] as const)
    foldersByParent.value = Object.fromEntries(nextEntries)
    expandedFolderIds.value = expandedFolderIds.value.filter((id) => id !== folderId)
  }

  function upsertDocument(document: WorkspaceDocument) {
    const exists = documents.value.some((item) => item.id === document.id)
    documents.value = sortByUpdatedDesc(exists
      ? documents.value.map((item) => (item.id === document.id ? document : item))
      : [document, ...documents.value])
  }

  function upsertFiles(nextFiles: WorkspaceFile[]) {
    files.value = sortByUpdatedDesc(nextFiles)
  }

  async function loadFolderChildren(parentId: string | null) {
    if (!workspaceId.value) return

    const key = parentKey(parentId)
    if (!loadedParents.value.includes(key)) {
      isLoadingFolders.value = true
    }

    try {
      const result = await listWorkspaceFolders(workspaceId.value, {
        parent_id: parentId || undefined,
      })
      const normalized = parentId ? result : result.filter(isRootFolder)
      replaceFolderChildren(parentId, normalized)
      if (!loadedParents.value.includes(key)) {
        loadedParents.value = [...loadedParents.value, key]
      }
    } catch (error) {
      updateError(error, '加载文件夹失败')
    } finally {
      isLoadingFolders.value = false
    }
  }

  async function loadDocuments() {
    if (!workspaceId.value) return
    isLoadingDocuments.value = true
    try {
      documents.value = sortByUpdatedDesc(await listDocuments(workspaceId.value))
    } catch (error) {
      updateError(error, '加载文档失败')
    } finally {
      isLoadingDocuments.value = false
    }
  }

  async function loadFiles(folderId: string | null = selectedFolderId.value) {
    if (!workspaceId.value) return
    isLoadingFiles.value = true
    try {
      upsertFiles(await listWorkspaceFiles(workspaceId.value, {
        folder_id: folderId || undefined,
      }))
    } catch (error) {
      updateError(error, '加载文件失败')
    } finally {
      isLoadingFiles.value = false
    }
  }

  async function reloadAll() {
    if (!workspaceId.value) return
    clearError()
    await Promise.all([
      loadFolderChildren(null),
      loadDocuments(),
      loadFiles(selectedFolderId.value),
    ])
  }

  async function selectFolder(folderId: string | null) {
    selectedFolderId.value = normalizeFolderId(folderId)
    selectedFileIds.value = []
    await loadFiles(selectedFolderId.value)
  }

  async function toggleFolder(folderId: string) {
    if (expandedFolderIds.value.includes(folderId)) {
      expandedFolderIds.value = expandedFolderIds.value.filter((id) => id !== folderId)
      return
    }

    await loadFolderChildren(folderId)
    expandedFolderIds.value = [...expandedFolderIds.value, folderId]
  }

  async function createFolder(input: { name: string; description?: string; parentId?: string | null }) {
    if (!workspaceId.value) return null
    isMutating.value = true
    clearError()
    try {
      const folder = await createWorkspaceFolder(workspaceId.value, {
        name: input.name,
        description: input.description,
        parent_id: input.parentId ?? selectedFolderId.value,
      })
      upsertFolder(folder)
      if (folder.parent_id && !expandedFolderIds.value.includes(folder.parent_id)) {
        expandedFolderIds.value = [...expandedFolderIds.value, folder.parent_id]
      }
      return folder
    } catch (error) {
      updateError(error, '创建文件夹失败')
      return null
    } finally {
      isMutating.value = false
    }
  }

  async function renameFolder(folderId: string, input: { name: string; description?: string }) {
    if (!workspaceId.value) return null
    isMutating.value = true
    clearError()
    try {
      const folder = await updateWorkspaceFolder(workspaceId.value, folderId, input)
      upsertFolder(folder)
      return folder
    } catch (error) {
      updateError(error, '重命名文件夹失败')
      return null
    } finally {
      isMutating.value = false
    }
  }

  async function deleteFolder(folder: WorkspaceFolder) {
    if (!workspaceId.value) return false
    isMutating.value = true
    clearError()
    try {
      const [childFolders, folderFiles] = await Promise.all([
        listWorkspaceFolders(workspaceId.value, { parent_id: folder.id }),
        listWorkspaceFiles(workspaceId.value, { folder_id: folder.id }),
      ])
      const folderDocuments = documents.value.filter((item) => normalizeFolderId(item.folder_id) === folder.id)

      if (childFolders.length || folderFiles.length || folderDocuments.length) {
        throw new ApiError('文件夹内仍有资源，请先清空后再删除。')
      }

      await deleteWorkspaceFolder(workspaceId.value, folder.id)
      removeFolder(folder.id)
      if (selectedFolderId.value === folder.id) {
        await selectFolder(normalizeFolderId(folder.parent_id))
      }
      return true
    } catch (error) {
      updateError(error, '删除文件夹失败')
      return false
    } finally {
      isMutating.value = false
    }
  }

  async function createNewDocument(title: string) {
    if (!workspaceId.value) return null
    isMutating.value = true
    clearError()
    try {
      const document = await createDocument({
        workspace_id: workspaceId.value,
        folder_id: selectedFolderId.value,
        title,
        content_json: emptyDocumentContent(),
        tiptap_schema: 'doc',
        tiptap_schema_ver: 'v1',
      })
      upsertDocument(document)
      return document
    } catch (error) {
      updateError(error, '创建文档失败')
      return null
    } finally {
      isMutating.value = false
    }
  }

  async function renameDocument(documentId: string, title: string) {
    return updateDocument(documentId, { title })
  }

  async function moveDocument(documentId: string, folderId: string | null) {
    return updateDocument(documentId, {
      folderId,
      clearFolder: !folderId,
    })
  }

  async function updateDocument(documentId: string, input: { title?: string; folderId?: string | null; clearFolder?: boolean }) {
    isMutating.value = true
    clearError()
    try {
      const result = await updateDocumentMeta(documentId, {
        title: input.title,
        folder_id: input.folderId,
        clear_folder: input.clearFolder ?? false,
      })
      upsertDocument(result)
      return result
    } catch (error) {
      updateError(error, '更新文档失败')
      return null
    } finally {
      isMutating.value = false
    }
  }

  async function deleteDocument(documentId: string) {
    isMutating.value = true
    clearError()
    try {
      await deleteDocumentById(documentId)
      documents.value = documents.value.filter((document) => document.id !== documentId)
      return true
    } catch (error) {
      updateError(error, '删除文档失败')
      return false
    } finally {
      isMutating.value = false
    }
  }

  async function saveDocument(
    documentId: string,
    payload: { title: string; contentJson: Record<string, unknown> | null; summary?: string },
    options?: { source?: 'autosave' | 'manual' | 'snapshot' | string; snapshot?: boolean },
  ) {
    isMutating.value = true
    clearError()
    try {
      const result = await saveDocumentVersion(documentId, {
        title: payload.title,
        content_json: payload.contentJson ?? emptyDocumentContent(),
        source: options?.source || 'manual',
        snapshot: options?.snapshot ?? true,
        summary: payload.summary || '',
      })
      upsertDocument(result.document)
      return result.document
    } catch (error) {
      updateError(error, '保存文档失败')
      return null
    } finally {
      isMutating.value = false
    }
  }

  async function uploadFilesToCurrentFolder(fileList: File[] | FileList) {
    if (!workspaceId.value) return []
    const filesToUpload = Array.from(fileList)
    if (!filesToUpload.length) return []

    isUploadingFiles.value = true
    clearError()
    setUploadStatusMessage(`正在上传 0/${filesToUpload.length} 个文件...`)
    const uploaded: WorkspaceFile[] = []

    try {
      for (let i = 0; i < filesToUpload.length; i++) {
        const file = filesToUpload[i]!
        setUploadStatusMessage(`正在上传 ${i + 1}/${filesToUpload.length}：${file.name}`)

        const ticket = await presignWorkspaceUpload(workspaceId.value, {
          folder_id: selectedFolderId.value,
          file_name: file.name,
          content_type: file.type || 'application/octet-stream',
        })

        const uploadRes = await fetch(ticket.upload_url, {
          method: ticket.upload_method || 'PUT',
          headers: {
            'Content-Type': file.type || 'application/octet-stream',
          },
          body: file,
        })

        if (!uploadRes.ok) {
          throw new ApiError(`上传文件失败：${file.name}`)
        }

        const completeRes = await completeWorkspaceUpload(workspaceId.value, {
          folder_id: selectedFolderId.value,
          object_key: ticket.object_key,
          file_name: file.name,
        })

        uploaded.push({
          ...completeRes.file,
          preview_url: completeRes.preview_url,
        })
      }

      await loadFiles(selectedFolderId.value)
      if (uploaded.length > 0) {
        setUploadStatusMessage(`上传成功：${uploaded.length} 个文件`, 4000)
      }
      return uploaded
    } catch (error) {
      updateError(error, '上传文件失败')
      setUploadStatusMessage('上传失败，请重试', 4000)
      return []
    } finally {
      isUploadingFiles.value = false
    }
  }

  async function deleteFile(fileId: string) {
    if (!workspaceId.value) return false
    isMutating.value = true
    clearError()
    try {
      await deleteWorkspaceFile(workspaceId.value, fileId)
      files.value = files.value.filter((file) => file.id !== fileId)
      selectedFileIds.value = selectedFileIds.value.filter((id) => id !== fileId)
      return true
    } catch (error) {
      updateError(error, '删除文件失败')
      return false
    } finally {
      isMutating.value = false
    }
  }

  async function renameFile(fileId: string, fileName: string) {
    return updateFile(fileId, { fileName })
  }

  async function moveFile(fileId: string, folderId: string | null) {
    const updated = await updateFile(fileId, {
      folderId,
      clearFolder: !folderId,
    })
    if (updated) {
      selectedFileIds.value = selectedFileIds.value.filter((id) => id !== fileId)
    }
    return updated
  }

  async function updateFile(fileId: string, input: { fileName?: string; folderId?: string | null; clearFolder?: boolean }) {
    if (!workspaceId.value) return null
    isMutating.value = true
    clearError()
    try {
      const updated = await updateWorkspaceFile(workspaceId.value, fileId, {
        file_name: input.fileName,
        folder_id: input.folderId,
        clear_folder: input.clearFolder ?? false,
      })
      await loadFiles(selectedFolderId.value)
      return updated
    } catch (error) {
      updateError(error, '更新文件失败')
      return null
    } finally {
      isMutating.value = false
    }
  }

  async function batchDeleteFiles(fileIds: string[]) {
    if (!workspaceId.value || !fileIds.length) return false
    isMutating.value = true
    clearError()
    try {
      const result = await batchDeleteWorkspaceFiles(workspaceId.value, fileIds)
      files.value = files.value.filter((file) => !result.deleted_ids.includes(file.id))
      selectedFileIds.value = []
      if (result.failed.length) {
        throw new ApiError(result.failed[0]?.reason || '批量删除文件失败')
      }
      return true
    } catch (error) {
      updateError(error, '批量删除文件失败')
      return false
    } finally {
      isMutating.value = false
    }
  }

  async function previewFile(fileId: string) {
    if (!workspaceId.value) return
    clearError()
    try {
      const detail = await getWorkspaceFileDetail(workspaceId.value, fileId)
      if (detail.preview_url && typeof window !== 'undefined') {
        window.open(detail.preview_url, '_blank', 'noopener,noreferrer')
      }
    } catch (error) {
      updateError(error, '获取文件预览失败')
    }
  }

  watch(
    workspaceId,
    async (nextId) => {
      resetState()
      if (!nextId) return
      await reloadAll()
    },
    { immediate: true },
  )

  return {
    documents,
    files,
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
    reloadAll,
    loadFolderChildren,
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
  }
}
