export type ReferenceType =
  | 'article'
  | 'book'
  | 'book-chapter'
  | 'conference'
  | 'thesis'
  | 'report'
  | 'web'
  | 'preprint'
  | 'unknown'

export interface Author {
  family: string
  given?: string
  suffix?: string
  literal?: string
  orcid?: string
}

export interface Reference {
  id: string
  owner_id: string
  workspace_id?: string

  title: string
  authors: Author[]
  year?: number
  source?: string
  doi?: string
  isbn?: string
  url?: string
  abstract?: string
  keywords?: string[]
  type: ReferenceType

  volume?: string
  issue?: string
  pages?: string
  publisher?: string
  language?: string

  file_id?: string
  citation_key?: string
  bibtex_raw?: string
  metadata?: Record<string, unknown>

  starred: boolean
  status: 'active' | 'reviewed' | 'archived'

  created_at: string
  updated_at: string
}

export interface Collection {
  id: string
  owner_id: string
  workspace_id?: string
  parent_id?: string

  name: string
  description?: string
  color?: string
  icon?: string
  sort_order: number

  created_at: string
  updated_at: string
}

export interface CollectionTree extends Collection {
  children?: CollectionTree[]
  reference_count?: number
}

export interface ReferenceNote {
  id: string
  reference_id: string
  user_id: string

  title?: string
  content: string
  page_from?: number
  page_to?: number

  created_at: string
  updated_at: string
}

export interface CreateReferenceInput {
  title: string
  authors?: Author[]
  year?: number
  source?: string
  doi?: string
  isbn?: string
  url?: string
  abstract?: string
  keywords?: string[]
  type?: ReferenceType
  volume?: string
  issue?: string
  pages?: string
  publisher?: string
  language?: string
  file_id?: string
  workspace_id?: string
  citation_key?: string
  bibtex_raw?: string
  metadata?: Record<string, unknown>
}

export interface UpdateReferenceInput {
  title?: string
  authors?: Author[]
  year?: number
  source?: string
  doi?: string
  isbn?: string
  url?: string
  abstract?: string
  keywords?: string[]
  type?: ReferenceType
  volume?: string
  issue?: string
  pages?: string
  publisher?: string
  language?: string
  file_id?: string
  workspace_id?: string
  citation_key?: string
  bibtex_raw?: string
  metadata?: Record<string, unknown>
  starred?: boolean
  status?: 'active' | 'reviewed' | 'archived'
}

export interface ListReferencesParams {
  q?: string
  type?: string
  year_from?: number
  year_to?: number
  collection_id?: string
  starred?: boolean
  status?: string
  workspace_id?: string
  sort_by?: string
  sort_order?: 'asc' | 'desc'
  page?: number
  page_size?: number
}

export interface ListReferencesResponse {
  items: Reference[]
  total: number
  page: number
  page_size: number
}

export interface DoiLookupResponse {
  found: boolean
  reference?: Reference
  error?: string
}

export interface CreateCollectionInput {
  name: string
  description?: string
  parent_id?: string
  color?: string
  icon?: string
  workspace_id?: string
}

export interface UpdateCollectionInput {
  name?: string
  description?: string
  parent_id?: string
  color?: string
  icon?: string
  sort_order?: number
}

export const REFERENCE_TYPE_LABELS: Record<ReferenceType, string> = {
  article: '期刊论文',
  book: '书籍',
  'book-chapter': '书籍章节',
  conference: '会议论文',
  thesis: '学位论文',
  report: '技术报告',
  web: '网页',
  preprint: '预印本',
  unknown: '未知类型',
}

export const REFERENCE_STATUS_LABELS: Record<string, string> = {
  active: '进行中',
  reviewed: '已审阅',
  archived: '已归档',
}

export function formatAuthors(authors: Author[], maxCount = 3): string {
  if (!authors || authors.length === 0) return ''

  const formatAuthor = (a: Author): string => {
    if (a.literal) return a.literal
    if (a.family && a.given) return `${a.family} ${a.given.charAt(0)}.`
    return a.family || ''
  }

  if (authors.length <= maxCount) {
    return authors.map(formatAuthor).join(', ')
  }

  return `${authors.slice(0, maxCount).map(formatAuthor).join(', ')} et al.`
}

export function generateCitationKey(ref: Partial<Reference>): string {
  const parts: string[] = []

  if (ref.authors && ref.authors.length > 0) {
    const firstAuthor = ref.authors[0]
    const family = firstAuthor.family || firstAuthor.literal || ''
    if (family) {
      parts.push(family.toLowerCase().replace(/[^a-z]/g, ''))
    }
  }

  if (ref.year) {
    parts.push(String(ref.year))
  }

  if (ref.title) {
    const firstWord = ref.title.split(/\s+/)[0]?.toLowerCase().replace(/[^a-z]/g, '')
    if (firstWord && firstWord.length > 2) {
      parts.push(firstWord)
    }
  }

  return parts.join('') || 'untitled'
}