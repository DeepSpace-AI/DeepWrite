import { http, unwrapResponse } from '@/api/http'
import type {
  Reference,
  Collection,
  CollectionTree,
  CreateReferenceInput,
  UpdateReferenceInput,
  ListReferencesParams,
  ListReferencesResponse,
  DoiLookupResponse,
  CreateCollectionInput,
  UpdateCollectionInput,
} from '@/types/reference'

export async function listReferences(params: ListReferencesParams = {}): Promise<ListReferencesResponse> {
  const res = await http.get('/references', {
    params: {
      q: params.q || undefined,
      type: params.type || undefined,
      year_from: params.year_from,
      year_to: params.year_to,
      collection_id: params.collection_id || undefined,
      starred: params.starred,
      status: params.status || undefined,
      workspace_id: params.workspace_id || undefined,
      sort_by: params.sort_by || 'created_at',
      sort_order: params.sort_order || 'desc',
      page: params.page || 1,
      page_size: params.page_size || 20,
    },
  })
  return unwrapResponse<ListReferencesResponse>(res)
}

export async function getReference(id: string): Promise<Reference> {
  const res = await http.get(`/references/${id}`)
  return unwrapResponse<Reference>(res)
}

export async function createReference(input: CreateReferenceInput): Promise<Reference> {
  const res = await http.post('/references', input)
  return unwrapResponse<Reference>(res)
}

export async function updateReference(id: string, input: UpdateReferenceInput): Promise<Reference> {
  const res = await http.put(`/references/${id}`, input)
  return unwrapResponse<Reference>(res)
}

export async function deleteReference(id: string): Promise<void> {
  const res = await http.delete(`/references/${id}`)
  unwrapResponse<unknown>(res)
}

export async function attachReferenceFile(id: string, fileId: string): Promise<void> {
  const res = await http.post(`/references/${id}/file`, { file_id: fileId })
  unwrapResponse<unknown>(res)
}

export async function detachReferenceFile(id: string): Promise<void> {
  const res = await http.delete(`/references/${id}/file`)
  unwrapResponse<unknown>(res)
}

export async function lookupDOI(doi: string): Promise<DoiLookupResponse> {
  const res = await http.get('/references/lookup/doi', {
    params: { doi },
  })
  return unwrapResponse<DoiLookupResponse>(res)
}

export async function searchCrossref(query: string): Promise<{ items: Reference[]; total: number }> {
  const res = await http.get('/references/search', {
    params: { q: query },
  })
  return unwrapResponse<{ items: Reference[]; total: number }>(res)
}

export async function listCollections(workspaceId?: string): Promise<CollectionTree[]> {
  const res = await http.get('/collections', {
    params: {
      workspace_id: workspaceId || undefined,
    },
  })
  return unwrapResponse<CollectionTree[]>(res)
}

export async function getCollection(id: string): Promise<Collection> {
  const res = await http.get(`/collections/${id}`)
  return unwrapResponse<Collection>(res)
}

export async function createCollection(input: CreateCollectionInput): Promise<Collection> {
  const res = await http.post('/collections', input)
  return unwrapResponse<Collection>(res)
}

export async function updateCollection(id: string, input: UpdateCollectionInput): Promise<Collection> {
  const res = await http.put(`/collections/${id}`, input)
  return unwrapResponse<Collection>(res)
}

export async function deleteCollection(id: string): Promise<void> {
  const res = await http.delete(`/collections/${id}`)
  unwrapResponse<unknown>(res)
}

export async function setCollectionReferences(collectionId: string, referenceIds: string[]): Promise<void> {
  const res = await http.put(`/collections/${collectionId}/references`, {
    reference_ids: referenceIds,
  })
  unwrapResponse<unknown>(res)
}

export async function addReferenceToCollection(collectionId: string, referenceId: string): Promise<void> {
  const res = await http.post(`/collections/${collectionId}/references/${referenceId}`)
  unwrapResponse<unknown>(res)
}

export async function removeReferenceFromCollection(collectionId: string, referenceId: string): Promise<void> {
  const res = await http.delete(`/collections/${collectionId}/references/${referenceId}`)
  unwrapResponse<unknown>(res)
}

export interface PdfExtractResponse {
  task_id?: string
  status?: string
  extracted?: PdfExtractResult
  error?: string
}

export interface PdfExtractResult {
  file_id: string
  status: string
  extracted_at: string
  text_length: number
  metadata?: Record<string, unknown>
  raw_text_preview?: string
  doi?: string
  crossref?: Record<string, unknown>
  reference?: {
    doi: string
    title: string
    authors: Array<{ family?: string; given?: string; literal?: string; orcid?: string }>
    year?: number
    source?: string
    type: string
    abstract?: string
    volume?: string
    issue?: string
    pages?: string
    publisher?: string
    url?: string
  }
}

export interface TaskStatusResponse {
  task_id: string
  status: string
  ready: boolean
  result?: PdfExtractResult
  error?: string
}

export async function extractPdfFromReference(
  referenceId: string,
  options: { wait?: boolean } = {}
): Promise<PdfExtractResponse> {
  const res = await http.post(`/references/${referenceId}/extract-pdf`, {
    wait: options.wait ?? false,
  })
  return unwrapResponse<PdfExtractResponse>(res)
}

export async function getExtractionTaskStatus(taskId: string): Promise<TaskStatusResponse> {
  const res = await http.get(`/references/tasks/${taskId}`)
  return unwrapResponse<TaskStatusResponse>(res)
}

export async function waitForExtractionTask(
  taskId: string,
  pollInterval = 2000,
  maxAttempts = 30
): Promise<TaskStatusResponse> {
  for (let i = 0; i < maxAttempts; i++) {
    const status = await getExtractionTaskStatus(taskId)
    if (status.ready) {
      return status
    }
    await new Promise(resolve => setTimeout(resolve, pollInterval))
  }
  throw new Error('Task polling timed out')
}

export interface ImportResult {
  imported: number
  skipped: number
  errors?: Array<{ line?: number; key?: string; message: string }>
  items: Reference[]
}

export async function importBibtex(content: string): Promise<ImportResult> {
  const res = await http.post('/references/import/bibtex', { content })
  return unwrapResponse<ImportResult>(res)
}

export async function importRIS(content: string): Promise<ImportResult> {
  const res = await http.post('/references/import/ris', { content })
  return unwrapResponse<ImportResult>(res)
}

export async function exportBibtex(ids?: string[]): Promise<string> {
  const res = await http.get('/references/export/bibtex', {
    params: ids ? { ids } : undefined,
    responseType: 'text',
  })
  return res.data
}

export async function exportRIS(ids?: string[]): Promise<string> {
  const res = await http.get('/references/export/ris', {
    params: ids ? { ids } : undefined,
    responseType: 'text',
  })
  return res.data
}

export async function exportCSLJSON(ids?: string[]): Promise<unknown[]> {
  const res = await http.get('/references/export/csl', {
    params: ids ? { ids } : undefined,
  })
  return unwrapResponse<unknown[]>(res)
}

export function downloadFile(content: string, filename: string, mimeType: string): void {
  const blob = new Blob([content], { type: mimeType })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

export interface AISearchResult {
  id: string
  title: string
  authors: Array<{ family?: string; given?: string; literal?: string }>
  year?: number
  source?: string
  doi?: string
  url?: string
  abstract?: string
  type: string
  citation_count?: number
  open_access_url?: string
  provider: string
}

export interface AISearchResponse {
  items: AISearchResult[]
  total: number
  provider: string
}

export async function aiSearch(
  query: string,
  provider: 'all' | 'semanticscholar' | 'crossref' = 'all'
): Promise<AISearchResponse> {
  const res = await http.get('/references/ai-search', {
    params: { q: query, provider },
  })
  return unwrapResponse<AISearchResponse>(res)
}

export async function searchByDOI(dois: string[]): Promise<AISearchResponse> {
  const res = await http.post('/references/ai-search/doi', { dois })
  return unwrapResponse<AISearchResponse>(res)
}

export interface BatchImportResult {
  imported: number
  failed: number
  items: Reference[]
  errors?: string[]
}

export async function batchImportFromSearch(items: AISearchResult[]): Promise<BatchImportResult> {
  const res = await http.post('/references/ai-search/import', { items })
  return unwrapResponse<BatchImportResult>(res)
}

export interface UploadPdfResponse {
  file_id: string
  url: string
  task_id?: string
  reference?: Reference
}

export interface PdfUploadResult {
  reference: Reference
  task_id?: string
}

export async function uploadPdfForReference(referenceId: string, file: File): Promise<UploadPdfResponse> {
  const formData = new FormData()
  formData.append('file', file)
  const res = await http.post(`/references/${referenceId}/pdf`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
  return unwrapResponse<UploadPdfResponse>(res)
}

export async function uploadPdfAndCreateReference(file: File): Promise<PdfUploadResult> {
  const formData = new FormData()
  formData.append('file', file)
  const res = await http.post('/references/pdf/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
  const result = unwrapResponse<UploadPdfResponse & { reference?: Reference }>(res)
  
  if (!result.reference) {
    throw new Error('Failed to create reference from PDF')
  }
  
  return {
    reference: result.reference,
    task_id: result.task_id,
  }
}