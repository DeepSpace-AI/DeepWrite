import { http, unwrapResponse } from '@/api/http'
import type { WorkspaceDocument } from '@/views/workspace/types'

export interface CreateDocumentInput {
  workspace_id: string
  folder_id?: string | null
  title: string
  content_json?: Record<string, unknown>
  tiptap_schema?: string
  tiptap_schema_ver?: string
}

export interface SaveDocumentVersionInput {
  title?: string
  content_json?: Record<string, unknown> | null
  source?: 'autosave' | 'manual' | 'snapshot' | string
  snapshot?: boolean
  summary?: string
}

export interface SaveDocumentVersionResult {
  document: WorkspaceDocument
}

export interface DocumentVersion {
  id: string
  document_id: string
  workspace_id: string
  version: number
  title: string
  content_json: Record<string, unknown> | null
  source: string
  snapshot: boolean
  summary?: string
  created_by?: string
  created_at: string
}

export async function listDocuments(workspaceId: string, limit = 200, offset = 0): Promise<WorkspaceDocument[]> {
  const res = await http.get('/documents', {
    params: {
      workspace_id: workspaceId,
      limit,
      offset,
    },
  })
  return unwrapResponse<WorkspaceDocument[]>(res)
}

export async function createDocument(input: CreateDocumentInput): Promise<WorkspaceDocument> {
  const res = await http.post('/documents', {
    workspace_id: input.workspace_id,
    folder_id: input.folder_id || undefined,
    title: input.title,
    content_json: input.content_json,
    tiptap_schema: input.tiptap_schema,
    tiptap_schema_ver: input.tiptap_schema_ver,
  })
  return unwrapResponse<WorkspaceDocument>(res)
}

export async function getDocumentById(documentId: string): Promise<WorkspaceDocument> {
  const res = await http.get(`/documents/${documentId}`)
  return unwrapResponse<WorkspaceDocument>(res)
}

export async function saveDocumentVersion(documentId: string, input: SaveDocumentVersionInput): Promise<SaveDocumentVersionResult> {
  const res = await http.put(`/documents/${documentId}`, {
    title: input.title,
    content_json: input.content_json ?? undefined,
    source: input.source || 'manual',
    snapshot: input.snapshot ?? false,
    summary: input.summary || '',
  })
  return unwrapResponse<SaveDocumentVersionResult>(res)
}

export async function listDocumentVersions(documentId: string, limit = 50, offset = 0): Promise<DocumentVersion[]> {
  const res = await http.get(`/documents/${documentId}/versions`, {
    params: { limit, offset },
  })
  return unwrapResponse<DocumentVersion[]>(res)
}
