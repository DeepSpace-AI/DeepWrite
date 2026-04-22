<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useReferenceStore } from '@/stores/reference'
import { formatAuthors, REFERENCE_TYPE_LABELS, REFERENCE_STATUS_LABELS } from '@/types/reference'
import type { Reference, CollectionTree, Author } from '@/types/reference'
import * as api from '@/api/reference'
import IconSearch from '~icons/mdi/magnify'
import IconFolderSpecial from '~icons/mdi/folder-star'
import IconHistory from '~icons/mdi/history'
import IconStar from '~icons/mdi/star'
import IconStarOutline from '~icons/mdi/star-outline'
import IconFolderOutline from '~icons/mdi/folder-outline'
import IconPlus from '~icons/mdi/plus'
import IconUpload from '~icons/mdi/upload'
import IconGridView from '~icons/mdi/view-grid-outline'
import IconListView from '~icons/mdi/view-list-outline'
import IconLoading from '~icons/mdi/loading'
import IconClose from '~icons/mdi/close'
import IconCheck from '~icons/mdi/check'
import IconDelete from '~icons/mdi/delete'
import IconDoi from '~icons/mdi/identifier'
import IconDownload from '~icons/mdi/download'
import IconFileImport from '~icons/mdi/file-import'
import IconAlert from '~icons/mdi/alert-circle'
import IconCloudSearch from '~icons/mdi/cloud-search'
import IconOpenInNew from '~icons/mdi/open-in-new'

const { t } = useI18n()
const store = useReferenceStore()

const searchKeyword = ref('')
const viewMode = ref<'grid' | 'list'>('grid')
const activeCollectionId = ref<string>('all')

const showAddModal = ref(false)
const showCollectionModal = ref(false)
const showDetailPanel = ref(false)
const showImportModal = ref(false)
const showExportDropdown = ref(false)
const showAISearchModal = ref(false)
const selectedReference = ref<Reference | null>(null)
const editingCollection = ref<CollectionTree | null>(null)


const newCollectionName = ref('')
const doiInput = ref('')
const doiLooking = ref(false)
const doiFound = ref(false)
const doiPreview = ref<Partial<Reference> | null>(null)

const importTab = ref<'bibtex' | 'ris'>('bibtex')
const importContent = ref('')
const importLoading = ref(false)
const importResult = ref<api.ImportResult | null>(null)

const aiSearchQuery = ref('')
const aiSearchProvider = ref<'all' | 'semanticscholar' | 'crossref'>('all')
const aiSearchLoading = ref(false)
const aiSearchResults = ref<api.AISearchResult[]>([])
const aiSearchTotal = ref(0)
const aiSearchSelected = ref<Set<string>>(new Set())
const aiSearchImporting = ref(false)

const pdfUploading = ref(false)
const pdfFileInput = ref<HTMLInputElement | null>(null)

const manualForm = ref({
  title: '',
  authors: '' as string,
  year: undefined as number | undefined,
  source: '',
  doi: '',
  type: 'article' as Reference['type'],
})

const addTab = ref<'doi' | 'manual' | 'pdf'>('doi')

const pdfAddUploading = ref(false)
const pdfAddFileInput = ref<HTMLInputElement | null>(null)
const pdfAddResult = ref<api.PdfExtractResult | null>(null)
const pdfAddError = ref('')

const defaultCollections = computed(() => [
  { id: 'all', label: t('library.collections.all'), icon: IconFolderSpecial },
  { id: 'recent', label: t('library.collections.recent'), icon: IconHistory },
  { id: 'starred', label: t('library.collections.starred'), icon: IconStar },
])

function getStatusColor(status: string) {
  switch (status) {
    case 'reviewed':
      return 'bg-emerald-500'
    case 'archived':
      return 'bg-gray-400'
    default:
      return 'bg-blue-500'
  }
}

function getStatusLabel(status: string) {
  return REFERENCE_STATUS_LABELS[status] || status
}

function getTypeLabel(type: string) {
  return REFERENCE_TYPE_LABELS[type as keyof typeof REFERENCE_TYPE_LABELS] || type
}

function formatAuthorsList(authors: Author[]) {
  return formatAuthors(authors)
}

function cleanJatsTags(text: string): string {
  if (!text) return ''
  
  return text
    .replace(/<jats:sup>/gi, '<sup>')
    .replace(/<\/jats:sup>/gi, '</sup>')
    .replace(/<jats:sub>/gi, '<sub>')
    .replace(/<\/jats:sub>/gi, '</sub>')
    .replace(/<jats:italic>/gi, '<i>')
    .replace(/<\/jats:italic>/gi, '</i>')
    .replace(/<jats:bold>/gi, '<b>')
    .replace(/<\/jats:bold>/gi, '</b>')
    .replace(/<jats:title>/gi, '<strong>')
    .replace(/<\/jats:title>/gi, '</strong>')
    .replace(/<jats:p>/gi, '<p>')
    .replace(/<\/jats:p>/gi, '</p>')
    .replace(/<jats:sc>/gi, '<span style="font-variant: small-caps">')
    .replace(/<\/jats:sc>/gi, '</span>')
    .replace(/<jats:ext-link[^>]*>/gi, '')
    .replace(/<\/jats:ext-link>/gi, '')
    .replace(/<jats:\w+[^>]*>/gi, '')
    .replace(/<\/jats:\w+>/gi, '')
}

function handleSelectDefaultCollection(id: string) {
  activeCollectionId.value = id
  if (id === 'starred') {
    store.setFilters({ starred: true })
  } else {
    store.setFilters({ starred: undefined })
  }
  store.selectCollection(null)
  store.fetchReferences()
}

function handleSelectFolder(id: string) {
  activeCollectionId.value = id
  store.selectCollection(id)
  store.setFilters({ starred: undefined })
  store.fetchReferences()
}

async function handleToggleStar(ref: Reference, event: Event) {
  event.stopPropagation()
  try {
    await store.toggleStar(ref.id)
  } catch (e) {
    console.error('Failed to toggle star:', e)
  }
}

function handleReferenceClick(ref: Reference) {
  selectedReference.value = ref
  showDetailPanel.value = true
}

function openAddModal() {
  showAddModal.value = true
  addTab.value = 'doi'
  doiInput.value = ''
  doiPreview.value = null
  doiFound.value = false
  pdfAddResult.value = null
  pdfAddError.value = ''
  resetManualForm()
}

function closeAddModal() {
  showAddModal.value = false
  doiPreview.value = null
  doiFound.value = false
  pdfAddResult.value = null
  pdfAddError.value = ''
}

function resetManualForm() {
  manualForm.value = {
    title: '',
    authors: '',
    year: undefined,
    source: '',
    doi: '',
    type: 'article',
  }
}

async function lookupDOI() {
  if (!doiInput.value.trim()) return
  
  doiLooking.value = true
  doiFound.value = false
  doiPreview.value = null

  try {
    const result = await api.lookupDOI(doiInput.value.trim())
    if (result.found && result.reference) {
      doiFound.value = true
      doiPreview.value = result.reference
    }
  } catch (e) {
    console.error('DOI lookup failed:', e)
  } finally {
    doiLooking.value = false
  }
}

async function createFromDOI() {
  if (!doiPreview.value) return

  try {
    const authors = doiPreview.value.authors || []
    await store.createReference({
      title: doiPreview.value.title || '',
      authors,
      year: doiPreview.value.year,
      source: doiPreview.value.source,
      doi: doiPreview.value.doi,
      type: doiPreview.value.type || 'article',
      abstract: doiPreview.value.abstract,
    })
    closeAddModal()
  } catch (e) {
    console.error('Failed to create reference:', e)
  }
}

async function createFromManual() {
  if (!manualForm.value.title.trim()) return

  try {
    const authors: Author[] = manualForm.value.authors
      ? manualForm.value.authors.split(',').map(a => {
          const parts = a.trim().split(/\s+/)
          if (parts.length >= 2) {
            return { family: parts.slice(1).join(' '), given: parts[0] }
          }
          return { family: a.trim(), given: '' }
        })
      : []

    await store.createReference({
      title: manualForm.value.title,
      authors,
      year: manualForm.value.year,
      source: manualForm.value.source || undefined,
      doi: manualForm.value.doi || undefined,
      type: manualForm.value.type,
    })
    closeAddModal()
  } catch (e) {
    console.error('Failed to create reference:', e)
  }
}

function triggerPdfAddUpload() {
  pdfAddFileInput.value?.click()
}

async function handlePdfAddUpload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  pdfAddUploading.value = true
  pdfAddResult.value = null
  pdfAddError.value = ''

  try {
    const result = await api.uploadPdfAndCreateReference(file)
    
    await store.fetchReferences()
    
    pdfAddResult.value = {
      reference: {
        doi: result.reference.doi || '',
        title: result.reference.title || file.name.replace('.pdf', ''),
        authors: result.reference.authors || [],
        year: result.reference.year,
        source: result.reference.source || '',
        type: result.reference.type || 'unknown',
        abstract: result.reference.abstract || '',
        volume: result.reference.volume || '',
        issue: result.reference.issue || '',
        pages: result.reference.pages || '',
        publisher: result.reference.publisher || '',
        url: result.reference.url || '',
      },
    } as api.PdfExtractResult
    
    closeAddModal()
  } catch (e) {
    console.error('PDF upload failed:', e)
    pdfAddError.value = e instanceof Error ? e.message : 'Upload failed'
  } finally {
    pdfAddUploading.value = false
    input.value = ''
  }
}

function openCreateCollection() {
  editingCollection.value = null
  newCollectionName.value = ''
  showCollectionModal.value = true
}

function openEditCollection(col: CollectionTree) {
  editingCollection.value = col
  newCollectionName.value = col.name
  showCollectionModal.value = true
}

function closeCollectionModal() {
  showCollectionModal.value = false
  editingCollection.value = null
  newCollectionName.value = ''
}

async function saveCollection() {
  if (!newCollectionName.value.trim()) return

  try {
    if (editingCollection.value) {
      await api.updateCollection(editingCollection.value.id, { name: newCollectionName.value })
      await store.fetchCollections()
    } else {
      await store.createCollection(newCollectionName.value)
    }
    closeCollectionModal()
  } catch (e) {
    console.error('Failed to save collection:', e)
  }
}

async function deleteReference(ref: Reference) {
  if (!confirm(t('library.detail.deleteConfirm'))) return

  try {
    await store.deleteReference(ref.id)
    showDetailPanel.value = false
    selectedReference.value = null
  } catch (e) {
    console.error('Failed to delete reference:', e)
  }
}

function closeDetailPanel() {
  showDetailPanel.value = false
  selectedReference.value = null
}

function openImportModal() {
  showImportModal.value = true
  importTab.value = 'bibtex'
  importContent.value = ''
  importResult.value = null
}

function closeImportModal() {
  showImportModal.value = false
  importContent.value = ''
  importResult.value = null
}

async function handleImport() {
  if (!importContent.value.trim()) return

  importLoading.value = true
  importResult.value = null

  try {
    if (importTab.value === 'bibtex') {
      importResult.value = await api.importBibtex(importContent.value)
    } else {
      importResult.value = await api.importRIS(importContent.value)
    }

    if (importResult.value.imported > 0) {
      await store.fetchReferences()
    }
  } catch (e) {
    console.error('Import failed:', e)
    importResult.value = {
      imported: 0,
      skipped: 0,
      errors: [{ message: e instanceof Error ? e.message : 'Import failed' }],
      items: [],
    }
  } finally {
    importLoading.value = false
  }
}

async function handleExport(format: 'bibtex' | 'ris' | 'csl') {
  showExportDropdown.value = false

  try {
    const ids = store.selectedReference ? [store.selectedReference.id] : undefined

    switch (format) {
      case 'bibtex': {
        const content = await api.exportBibtex(ids)
        api.downloadFile(content, 'references.bib', 'text/plain')
        break
      }
      case 'ris': {
        const content = await api.exportRIS(ids)
        api.downloadFile(content, 'references.ris', 'text/plain')
        break
      }
      case 'csl': {
        const content = await api.exportCSLJSON(ids)
        api.downloadFile(JSON.stringify(content, null, 2), 'references.json', 'application/json')
        break
      }
    }
  } catch (e) {
    console.error('Export failed:', e)
  }
}

function openAISearchModal() {
  showAISearchModal.value = true
  aiSearchQuery.value = ''
  aiSearchResults.value = []
  aiSearchTotal.value = 0
  aiSearchSelected.value = new Set()
}

function closeAISearchModal() {
  showAISearchModal.value = false
  aiSearchQuery.value = ''
  aiSearchResults.value = []
  aiSearchTotal.value = 0
  aiSearchSelected.value = new Set()
}

async function handleAISearch() {
  if (!aiSearchQuery.value.trim()) return

  aiSearchLoading.value = true
  aiSearchResults.value = []
  aiSearchTotal.value = 0

  try {
    const result = await api.aiSearch(aiSearchQuery.value, aiSearchProvider.value)
    aiSearchResults.value = result.items
    aiSearchTotal.value = result.total
  } catch (e) {
    console.error('AI search failed:', e)
  } finally {
    aiSearchLoading.value = false
  }
}

function toggleSelectSearchResult(id: string) {
  if (aiSearchSelected.value.has(id)) {
    aiSearchSelected.value.delete(id)
  } else {
    aiSearchSelected.value.add(id)
  }
}

function selectAllSearchResults() {
  aiSearchResults.value.forEach(item => aiSearchSelected.value.add(item.id))
}

function deselectAllSearchResults() {
  aiSearchSelected.value.clear()
}

async function importSelectedResults() {
  if (aiSearchSelected.value.size === 0) return

  aiSearchImporting.value = true

  try {
    const selectedItems = aiSearchResults.value.filter(item => aiSearchSelected.value.has(item.id))
    const result = await api.batchImportFromSearch(selectedItems)
    
    if (result.imported > 0) {
      await store.fetchReferences()
    }
    
    aiSearchSelected.value.clear()
  } catch (e) {
    console.error('Batch import failed:', e)
  } finally {
    aiSearchImporting.value = false
  }
}

function triggerPdfUpload() {
  pdfFileInput.value?.click()
}

async function handlePdfUpload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file || !selectedReference.value) return

  pdfUploading.value = true
  try {
    const result = await api.uploadPdfForReference(selectedReference.value.id, file)
    if (result.reference) {
      selectedReference.value = result.reference
      await store.fetchReferences()
    }
  } catch (e) {
    console.error('PDF upload failed:', e)
  } finally {
    pdfUploading.value = false
    input.value = ''
  }
}

let searchTimeout: ReturnType<typeof setTimeout>
watch(searchKeyword, (value) => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    store.setSearchQuery(value)
    store.fetchReferences({ page: 1 })
  }, 300)
})

onMounted(async () => {
  await Promise.all([
    store.fetchReferences(),
    store.fetchCollections(),
  ])
})
</script>

<template>
  <section class="library-page flex h-[calc(100vh-4rem)]">
    <!-- Sidebar -->
    <aside class="library-sidebar w-64 flex-shrink-0 overflow-y-auto border-r border-[var(--surface-container-high)] bg-[var(--surface-container)] p-4">
      <!-- Default Collections -->
      <div class="mb-8">
        <h3 class="mb-3 text-xs font-semibold uppercase tracking-wider text-[var(--color-on-surface-variant)]">
          {{ t('library.collections.title') }}
        </h3>
        <ul class="space-y-1">
          <li v-for="col in defaultCollections" :key="col.id">
            <button
              type="button"
              class="flex w-full items-center rounded-lg px-3 py-2 text-left text-sm transition-colors"
              :class="activeCollectionId === col.id && !store.selectedCollectionId
                ? 'bg-[var(--surface-container-lowest)] text-[var(--color-on-surface)]'
                : 'text-[var(--color-on-surface-variant)] hover:bg-[var(--surface-container-high)]'"
              @click="handleSelectDefaultCollection(col.id)"
            >
              <component :is="col.icon" class="mr-3 h-4 w-4" />
              {{ col.label }}
            </button>
          </li>
        </ul>
      </div>

      <!-- User Folders -->
      <div class="mb-8">
        <div class="mb-3 flex items-center justify-between">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-[var(--color-on-surface-variant)]">
            {{ t('library.folders.title') }}
          </h3>
          <button
            type="button"
            class="rounded p-1 text-[var(--color-on-surface-variant)] hover:bg-[var(--surface-container-high)]"
            title="新建收藏集"
            @click="openCreateCollection"
          >
            <IconPlus class="h-4 w-4" />
          </button>
        </div>
        <ul v-if="store.collections && store.collections.length > 0" class="space-y-1">
          <li v-for="col in store.collections" :key="col.id" class="group relative">
            <button
              type="button"
              class="flex w-full items-center justify-between rounded-lg px-3 py-2 text-left text-sm transition-colors"
              :class="store.selectedCollectionId === col.id
                ? 'bg-[var(--surface-container-lowest)] text-[var(--color-on-surface)]'
                : 'text-[var(--color-on-surface-variant)] hover:bg-[var(--surface-container-high)]'"
              @click="handleSelectFolder(col.id)"
            >
              <div class="flex items-center">
                <IconFolderOutline class="mr-3 h-4 w-4" />
                <span>{{ col.name }}</span>
              </div>
              <span v-if="col.reference_count" class="text-xs opacity-60">
                {{ col.reference_count }}
              </span>
            </button>
            <div class="absolute right-2 top-1/2 -translate-y-1/2 hidden group-hover:flex">
              <button
                type="button"
                class="rounded p-1 text-[var(--color-on-surface-variant)] hover:bg-[var(--surface-container-high)]"
                @click.stop="openEditCollection(col)"
              >
                <IconPlus class="h-3 w-3" />
              </button>
            </div>
          </li>
        </ul>
        <p v-else class="px-3 py-2 text-sm text-[var(--color-on-surface-variant)] opacity-60">
          暂无收藏集
        </p>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="library-main flex-1 overflow-y-auto bg-[var(--surface)]">
      <!-- Header -->
      <header class="border-b border-[var(--surface-container-high)] bg-[var(--surface-container-low)] px-8 py-6">
        <div class="flex items-center justify-between">
          <div>
            <h1 class="text-3xl font-light text-[var(--color-on-background)]">
              {{ t('library.title') }}
            </h1>
            <p class="mt-1 text-sm text-[var(--color-on-surface-variant)]">
              {{ t('library.subtitle', { count: store.totalReferences }) }}
            </p>
          </div>
          <div class="flex items-center gap-3">
            <!-- AI Search -->
            <button
              type="button"
              class="flex items-center gap-2 rounded-lg bg-gradient-to-r from-[var(--color-primary)] to-purple-500 px-4 py-2 text-sm font-medium text-white shadow-md transition-all hover:shadow-lg"
              @click="openAISearchModal"
            >
              <IconCloudSearch class="h-4 w-4" />
              AI 检索
            </button>

            <!-- Import/Export -->
            <div class="flex items-center gap-2">
              <button
                type="button"
                class="flex items-center gap-2 rounded-lg border border-[var(--outline-variant)] bg-[var(--surface-container-lowest)] px-3 py-2 text-sm transition-colors hover:bg-[var(--surface-container)]"
                @click="openImportModal"
              >
                <IconFileImport class="h-4 w-4" />
                导入
              </button>
              <div class="relative">
                <button
                  type="button"
                  class="flex items-center gap-2 rounded-lg border border-[var(--outline-variant)] bg-[var(--surface-container-lowest)] px-3 py-2 text-sm transition-colors hover:bg-[var(--surface-container)]"
                  @click="showExportDropdown = !showExportDropdown"
                >
                  <IconDownload class="h-4 w-4" />
                  导出
                </button>
                <div
                  v-if="showExportDropdown"
                  class="absolute right-0 top-full z-10 mt-1 w-40 rounded-lg border border-[var(--outline-variant)] bg-[var(--surface-container-lowest)] py-1 shadow-lg"
                >
                  <button
                    type="button"
                    class="block w-full px-4 py-2 text-left text-sm hover:bg-[var(--surface-container)]"
                    @click="handleExport('bibtex')"
                  >
                    BibTeX (.bib)
                  </button>
                  <button
                    type="button"
                    class="block w-full px-4 py-2 text-left text-sm hover:bg-[var(--surface-container)]"
                    @click="handleExport('ris')"
                  >
                    RIS (.ris)
                  </button>
                  <button
                    type="button"
                    class="block w-full px-4 py-2 text-left text-sm hover:bg-[var(--surface-container)]"
                    @click="handleExport('csl')"
                  >
                    CSL JSON (.json)
                  </button>
                </div>
              </div>
            </div>

            <div class="flex rounded-lg bg-[var(--surface-container)] p-1">
              <button
                type="button"
                class="rounded-md p-1.5 transition-colors"
                :class="viewMode === 'grid' ? 'bg-[var(--surface-container-lowest)]' : ''"
                @click="viewMode = 'grid'"
              >
                <IconGridView class="h-5 w-5" />
              </button>
              <button
                type="button"
                class="rounded-md p-1.5 transition-colors"
                :class="viewMode === 'list' ? 'bg-[var(--surface-container-lowest)]' : ''"
                @click="viewMode = 'list'"
              >
                <IconListView class="h-5 w-5" />
              </button>
            </div>
            <button
              type="button"
              class="flex items-center gap-2 rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm font-medium text-[var(--color-on-primary)] transition-colors hover:opacity-90"
              @click="openAddModal"
            >
              <IconPlus class="h-4 w-4" />
              {{ t('library.add.title') }}
            </button>
          </div>
        </div>
      </header>

      <!-- Search -->
      <div class="border-b border-[var(--surface-container-high)] bg-[var(--surface-container)] px-8 py-4">
        <div class="relative max-w-xl">
          <IconSearch class="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-[var(--color-on-surface-variant)] opacity-50" />
          <input
            v-model="searchKeyword"
            type="text"
            class="w-full rounded-lg bg-[var(--surface-container-lowest)] py-2.5 pl-10 pr-4 text-sm outline-none focus:bg-[var(--surface-container)]"
            :placeholder="t('library.searchPlaceholder')"
          />
        </div>
      </div>

      <!-- Loading -->
      <div v-if="store.isLoading && (!store.references || store.references.length === 0)" class="flex items-center justify-center py-20">
        <IconLoading class="h-8 w-8 animate-spin text-[var(--color-primary)]" />
      </div>

      <!-- Error -->
      <div v-else-if="store.error" class="px-8 py-20 text-center">
        <p class="text-[var(--color-error)]">{{ store.error }}</p>
        <button
          type="button"
          class="mt-4 text-[var(--color-primary)] hover:underline"
          @click="store.fetchReferences()"
        >
          重试
        </button>
      </div>

      <!-- Empty -->
      <div v-else-if="!store.references || store.references.length === 0" class="px-8 py-20 text-center">
        <p class="text-[var(--color-on-surface-variant)]">{{ t('library.empty') }}</p>
        <button
          type="button"
          class="mt-4 rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-[var(--color-on-primary)]"
          @click="openAddModal"
        >
          {{ t('library.add.title') }}
        </button>
      </div>

      <!-- Content -->
      <div v-else class="p-8">
        <!-- Grid View -->
        <div v-if="viewMode === 'grid'" class="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
          <article
            v-for="ref in store.references"
            :key="ref.id"
            class="group cursor-pointer rounded-xl border border-[var(--surface-container-high)] bg-[var(--surface-container-lowest)] p-5 transition-all hover:border-[var(--color-primary)] hover:shadow-lg"
            @click="handleReferenceClick(ref)"
          >
            <div class="mb-3 flex items-start justify-between">
              <span class="rounded bg-[var(--surface-container-high)] px-2 py-1 text-xs text-[var(--color-on-surface-variant)]">
                {{ getTypeLabel(ref.type) }}
              </span>
              <button
                type="button"
                class="rounded p-1 transition-colors hover:bg-[var(--surface-container)]"
                @click="handleToggleStar(ref, $event)"
              >
                <IconStar v-if="ref.starred" class="h-5 w-5 text-amber-500" />
                <IconStarOutline v-else class="h-5 w-5 text-[var(--color-on-surface-variant)] opacity-40 hover:opacity-100" />
              </button>
            </div>
            <h4 class="mb-2 line-clamp-2 text-lg font-medium leading-tight">
              {{ ref.title }}
            </h4>
            <p class="mb-4 text-sm text-[var(--color-on-surface-variant)]">
              {{ formatAuthorsList(ref.authors) }}<template v-if="ref.year"> · {{ ref.year }}</template>
            </p>
            <div class="flex items-center gap-2">
              <span class="h-2 w-2 rounded-full" :class="getStatusColor(ref.status)" />
              <span class="text-xs text-[var(--color-on-surface-variant)]">
                {{ getStatusLabel(ref.status) }}
              </span>
            </div>
          </article>
        </div>

        <!-- List View -->
        <div v-else class="overflow-hidden rounded-xl border border-[var(--surface-container-high)] bg-[var(--surface-container-lowest)]">
          <table class="w-full">
            <thead class="bg-[var(--surface-container)]">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--color-on-surface-variant)]">{{ t('library.table.title') }}</th>
                <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--color-on-surface-variant)]">{{ t('library.table.authors') }}</th>
                <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--color-on-surface-variant)]">{{ t('library.table.year') }}</th>
                <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--color-on-surface-variant)]">{{ t('library.table.type') }}</th>
                <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--color-on-surface-variant)]">{{ t('library.table.status') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="ref in store.references"
                :key="ref.id"
                class="cursor-pointer border-t border-[var(--surface-container)] transition-colors hover:bg-[var(--surface-container)]"
                @click="handleReferenceClick(ref)"
              >
                <td class="px-4 py-3 font-medium">{{ ref.title }}</td>
                <td class="px-4 py-3 text-sm text-[var(--color-on-surface-variant)]">{{ formatAuthorsList(ref.authors) }}</td>
                <td class="px-4 py-3 text-sm">{{ ref.year || '-' }}</td>
                <td class="px-4 py-3">
                  <span class="rounded bg-[var(--surface-container-high)] px-2 py-0.5 text-xs">{{ getTypeLabel(ref.type) }}</span>
                </td>
                <td class="px-4 py-3">
                  <div class="flex items-center gap-2">
                    <span class="h-2 w-2 rounded-full" :class="getStatusColor(ref.status)" />
                    <span class="text-sm">{{ getStatusLabel(ref.status) }}</span>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Load More -->
        <div v-if="store.hasMore" class="mt-6 flex justify-center">
          <button
            type="button"
            class="flex items-center gap-2 rounded-lg border border-[var(--outline-variant)] bg-[var(--surface-container-lowest)] px-6 py-2 text-sm transition-colors hover:bg-[var(--surface-container)]"
            :disabled="store.isLoading"
            @click="store.fetchMoreReferences()"
          >
            <IconLoading v-if="store.isLoading" class="h-4 w-4 animate-spin" />
            {{ t('library.loadMore') }}
          </button>
        </div>
      </div>
    </main>

    <!-- Add Reference Modal -->
    <Teleport to="body">
      <div v-if="showAddModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="closeAddModal">
        <div class="w-full max-w-lg rounded-xl bg-[var(--surface-container-lowest)] p-6 shadow-xl">
          <div class="mb-4 flex items-center justify-between">
            <h2 class="text-xl font-semibold">{{ t('library.add.title') }}</h2>
            <button type="button" class="rounded p-1 hover:bg-[var(--surface-container)]" @click="closeAddModal">
              <IconClose class="h-5 w-5" />
            </button>
          </div>

          <!-- Tabs -->
          <div class="mb-4 flex border-b border-[var(--surface-container-high)]">
            <button
              type="button"
              class="px-4 py-2 text-sm font-medium transition-colors"
              :class="addTab === 'doi' ? 'border-b-2 border-[var(--color-primary)] text-[var(--color-primary)]' : 'text-[var(--color-on-surface-variant)]'"
              @click="addTab = 'doi'"
            >
              {{ t('library.add.doi') }}
            </button>
            <button
              type="button"
              class="px-4 py-2 text-sm font-medium transition-colors"
              :class="addTab === 'pdf' ? 'border-b-2 border-[var(--color-primary)] text-[var(--color-primary)]' : 'text-[var(--color-on-surface-variant)]'"
              @click="addTab = 'pdf'"
            >
              上传 PDF
            </button>
            <button
              type="button"
              class="px-4 py-2 text-sm font-medium transition-colors"
              :class="addTab === 'manual' ? 'border-b-2 border-[var(--color-primary)] text-[var(--color-primary)]' : 'text-[var(--color-on-surface-variant)]'"
              @click="addTab = 'manual'"
            >
              {{ t('library.add.manual') }}
            </button>
          </div>

          <!-- DOI Tab -->
          <div v-if="addTab === 'doi'" class="space-y-4">
            <div class="flex gap-2">
              <div class="relative flex-1">
                <IconDoi class="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-[var(--color-on-surface-variant)] opacity-50" />
                <input
                  v-model="doiInput"
                  type="text"
                  class="w-full rounded-lg bg-[var(--surface-container-lowest)] py-2.5 pl-10 pr-4 text-sm outline-none focus:bg-[var(--surface-container)]"
                  :placeholder="t('library.add.doiPlaceholder')"
                  @keyup.enter="lookupDOI"
                />
              </div>
              <button
                type="button"
                class="flex items-center gap-2 rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-[var(--color-on-primary)] disabled:opacity-50"
                :disabled="doiLooking || !doiInput.trim()"
                @click="lookupDOI"
              >
                <IconLoading v-if="doiLooking" class="h-4 w-4 animate-spin" />
                <span v-else>{{ t('library.add.doiLookup') }}</span>
              </button>
            </div>

            <!-- DOI Preview -->
            <div v-if="doiFound && doiPreview" class="rounded-lg border border-[var(--color-primary)] bg-[var(--color-primary-container)]/10 p-4">
              <div class="mb-2 flex items-center gap-2 text-sm font-medium text-[var(--color-primary)]">
                <IconCheck class="h-4 w-4" />
                {{ t('library.add.doiFound') }}
              </div>
              <h4 class="mb-1 font-medium">{{ doiPreview.title }}</h4>
              <p class="text-sm text-[var(--color-on-surface-variant)]">
                {{ formatAuthorsList(doiPreview.authors || []) }}<template v-if="doiPreview.year"> · {{ doiPreview.year }}</template>
              </p>
              <button
                type="button"
                class="mt-3 rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-[var(--color-on-primary)]"
                @click="createFromDOI"
              >
                {{ t('library.add.create') }}
              </button>
            </div>
          </div>

          <!-- PDF Upload Tab -->
          <div v-if="addTab === 'pdf'" class="space-y-4">
            <input
              ref="pdfAddFileInput"
              type="file"
              accept=".pdf"
              class="hidden"
              @change="handlePdfAddUpload"
            />
            
            <div
              class="flex cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed border-[var(--outline-variant)] p-8 transition-colors hover:border-[var(--color-primary)] hover:bg-[var(--surface-container)]"
              @click="triggerPdfAddUpload"
            >
              <IconLoading v-if="pdfAddUploading" class="h-12 w-12 animate-spin text-[var(--color-primary)]" />
              <IconUpload v-else class="h-12 w-12 text-[var(--color-on-surface-variant)] opacity-50" />
              <p class="mt-3 text-sm text-[var(--color-on-surface-variant)]">
                {{ pdfAddUploading ? '正在上传...' : '点击上传 PDF 文件' }}
              </p>
              <p class="mt-1 text-xs text-[var(--color-on-surface-variant)] opacity-60">
                支持最大 100MB 的 PDF 文件，上传后自动识别文献信息
              </p>
            </div>

            <!-- Error -->
            <div v-if="pdfAddError" class="rounded-lg border border-red-300 bg-red-50 p-4">
              <div class="flex items-center gap-2 text-sm font-medium text-red-700">
                <IconAlert class="h-4 w-4" />
                上传失败
              </div>
              <p class="mt-1 text-sm text-red-600">{{ pdfAddError }}</p>
            </div>
          </div>

          <!-- Manual Tab -->
          <div v-if="addTab === 'manual'" class="space-y-4">
            <div>
              <label class="mb-1 block text-sm font-medium">标题 *</label>
              <input
                v-model="manualForm.title"
                type="text"
                class="w-full rounded-lg bg-[var(--surface-container-lowest)] px-3 py-2 text-sm outline-none focus:bg-[var(--surface-container)]"
                placeholder="文献标题"
              />
            </div>
            <div>
              <label class="mb-1 block text-sm font-medium">作者</label>
              <input
                v-model="manualForm.authors"
                type="text"
                class="w-full rounded-lg bg-[var(--surface-container-lowest)] px-3 py-2 text-sm outline-none focus:bg-[var(--surface-container)]"
                placeholder="逗号分隔，如: John Smith, Jane Doe"
              />
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="mb-1 block text-sm font-medium">年份</label>
                <input
                  v-model.number="manualForm.year"
                  type="number"
                  class="w-full rounded-lg bg-[var(--surface-container-lowest)] px-3 py-2 text-sm outline-none focus:bg-[var(--surface-container)]"
                  placeholder="2024"
                />
              </div>
              <div>
                <label class="mb-1 block text-sm font-medium">类型</label>
                <select
                  v-model="manualForm.type"
                  class="w-full rounded-lg bg-[var(--surface-container-lowest)] px-3 py-2 text-sm outline-none focus:bg-[var(--surface-container)]"
                >
                  <option value="article">期刊论文</option>
                  <option value="book">书籍</option>
                  <option value="conference">会议论文</option>
                  <option value="thesis">学位论文</option>
                  <option value="preprint">预印本</option>
                  <option value="web">网页</option>
                </select>
              </div>
            </div>
            <div>
              <label class="mb-1 block text-sm font-medium">DOI</label>
              <input
                v-model="manualForm.doi"
                type="text"
                class="w-full rounded-lg bg-[var(--surface-container-lowest)] px-3 py-2 text-sm outline-none focus:bg-[var(--surface-container)]"
                placeholder="10.1000/xyz123"
              />
            </div>
            <div>
              <label class="mb-1 block text-sm font-medium">来源</label>
              <input
                v-model="manualForm.source"
                type="text"
                class="w-full rounded-lg bg-[var(--surface-container-lowest)] px-3 py-2 text-sm outline-none focus:bg-[var(--surface-container)]"
                placeholder="期刊名、会议名等"
              />
            </div>
            <button
              type="button"
              class="w-full rounded-lg bg-[var(--color-primary)] py-2.5 text-sm font-medium text-[var(--color-on-primary)] disabled:opacity-50"
              :disabled="!manualForm.title.trim()"
              @click="createFromManual"
            >
              {{ t('library.add.create') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Collection Modal -->
    <Teleport to="body">
      <div v-if="showCollectionModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="closeCollectionModal">
        <div class="w-full max-w-sm rounded-xl bg-[var(--surface-container-lowest)] p-6 shadow-xl">
          <h2 class="mb-4 text-lg font-semibold">
            {{ editingCollection ? t('library.folders.edit') : t('library.folders.create') }}
          </h2>
          <input
            v-model="newCollectionName"
            type="text"
            class="mb-4 w-full rounded-lg bg-[var(--surface-container-lowest)] px-3 py-2 text-sm outline-none focus:bg-[var(--surface-container)]"
            :placeholder="t('library.folders.namePlaceholder')"
            @keyup.enter="saveCollection"
          />
          <div class="flex justify-end gap-2">
            <button
              type="button"
              class="rounded-lg px-4 py-2 text-sm text-[var(--color-on-surface-variant)] hover:bg-[var(--surface-container)]"
              @click="closeCollectionModal"
            >
              取消
            </button>
            <button
              type="button"
              class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-[var(--color-on-primary)]"
              :disabled="!newCollectionName.trim()"
              @click="saveCollection"
            >
              保存
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Detail Panel -->
    <Teleport to="body">
      <div v-if="showDetailPanel && selectedReference" class="fixed inset-y-0 right-0 z-50 w-96 overflow-y-auto border-l border-[var(--surface-container-high)] bg-[var(--surface-container-lowest)] shadow-xl">
        <div class="sticky top-0 flex items-center justify-between border-b border-[var(--surface-container-high)] bg-[var(--surface-container-lowest)] p-4">
          <h3 class="font-semibold">{{ t('library.detail.title') }}</h3>
          <button type="button" class="rounded p-1 hover:bg-[var(--surface-container)]" @click="closeDetailPanel">
            <IconClose class="h-5 w-5" />
          </button>
        </div>
        <div class="p-4">
          <h4 class="mb-4 text-lg font-medium">{{ selectedReference.title }}</h4>
          
          <div class="mb-4">
            <span class="rounded bg-[var(--surface-container-high)] px-2 py-1 text-xs">
              {{ getTypeLabel(selectedReference.type) }}
            </span>
          </div>

          <dl class="space-y-3 text-sm">
            <div v-if="selectedReference.authors?.length">
              <dt class="text-[var(--color-on-surface-variant)]">作者</dt>
              <dd class="mt-1">{{ formatAuthorsList(selectedReference.authors) }}</dd>
            </div>
            <div v-if="selectedReference.year">
              <dt class="text-[var(--color-on-surface-variant)]">年份</dt>
              <dd class="mt-1">{{ selectedReference.year }}</dd>
            </div>
            <div v-if="selectedReference.source">
              <dt class="text-[var(--color-on-surface-variant)]">来源</dt>
              <dd class="mt-1">{{ selectedReference.source }}</dd>
            </div>
            <div v-if="selectedReference.doi">
              <dt class="text-[var(--color-on-surface-variant)]">DOI</dt>
              <dd class="mt-1">
                <a :href="`https://doi.org/${selectedReference.doi}`" target="_blank" class="text-[var(--color-primary)] hover:underline">
                  {{ selectedReference.doi }}
                </a>
              </dd>
            </div>
            <div v-if="selectedReference.abstract">
              <dt class="text-[var(--color-on-surface-variant)]">摘要</dt>
              <dd class="mt-1 text-[var(--color-on-surface-variant)]" v-html="cleanJatsTags(selectedReference.abstract)"></dd>
            </div>
            <div>
              <dt class="text-[var(--color-on-surface-variant)]">PDF 原文</dt>
              <dd class="mt-1">
                <span v-if="selectedReference.file_id" class="inline-flex items-center gap-1 text-emerald-600">
                  <IconCheck class="h-4 w-4" />
                  已上传
                </span>
                <span v-else class="text-[var(--color-on-surface-variant)] opacity-60">未上传</span>
              </dd>
            </div>
          </dl>

          <input
            ref="pdfFileInput"
            type="file"
            accept=".pdf"
            class="hidden"
            @change="handlePdfUpload"
          />

          <div class="mt-6 flex gap-2">
            <button
              type="button"
              class="flex flex-1 items-center justify-center gap-2 rounded-lg border border-[var(--outline-variant)] py-2 text-sm hover:bg-[var(--surface-container)] disabled:opacity-50"
              :disabled="pdfUploading"
              @click="triggerPdfUpload"
            >
              <IconLoading v-if="pdfUploading" class="h-4 w-4 animate-spin" />
              <IconUpload v-else class="h-4 w-4" />
              {{ selectedReference.file_id ? '更换 PDF' : '上传 PDF' }}
            </button>
            <button
              type="button"
              class="flex flex-1 items-center justify-center gap-2 rounded-lg border border-[var(--outline-variant)] py-2 text-sm hover:bg-[var(--surface-container)]"
              @click="handleToggleStar(selectedReference, $event)"
            >
              <IconStar v-if="selectedReference.starred" class="h-4 w-4 text-amber-500" />
              <IconStarOutline v-else class="h-4 w-4" />
              {{ selectedReference.starred ? '取消收藏' : '收藏' }}
            </button>
          </div>
          <div class="mt-2">
            <button
              type="button"
              class="flex w-full items-center justify-center gap-2 rounded-lg border border-red-300 px-4 py-2 text-sm text-red-600 hover:bg-red-50"
              @click="deleteReference(selectedReference)"
            >
              <IconDelete class="h-4 w-4" />
              删除
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Import Modal -->
    <Teleport to="body">
      <div v-if="showImportModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="closeImportModal">
        <div class="w-full max-w-2xl rounded-xl bg-[var(--surface-container-lowest)] p-6 shadow-xl">
          <div class="mb-4 flex items-center justify-between">
            <h2 class="text-xl font-semibold">导入文献</h2>
            <button type="button" class="rounded p-1 hover:bg-[var(--surface-container)]" @click="closeImportModal">
              <IconClose class="h-5 w-5" />
            </button>
          </div>

          <!-- Tabs -->
          <div class="mb-4 flex border-b border-[var(--surface-container-high)]">
            <button
              type="button"
              class="px-4 py-2 text-sm font-medium transition-colors"
              :class="importTab === 'bibtex' ? 'border-b-2 border-[var(--color-primary)] text-[var(--color-primary)]' : 'text-[var(--color-on-surface-variant)]'"
              @click="importTab = 'bibtex'"
            >
              BibTeX
            </button>
            <button
              type="button"
              class="px-4 py-2 text-sm font-medium transition-colors"
              :class="importTab === 'ris' ? 'border-b-2 border-[var(--color-primary)] text-[var(--color-primary)]' : 'text-[var(--color-on-surface-variant)]'"
              @click="importTab = 'ris'"
            >
              RIS
            </button>
          </div>

          <!-- Content Input -->
          <div class="mb-4">
            <label class="mb-1 block text-sm font-medium">
              {{ importTab === 'bibtex' ? 'BibTeX 内容' : 'RIS 内容' }}
            </label>
            <textarea
              v-model="importContent"
              class="h-64 w-full rounded-lg bg-[var(--surface-container-lowest)] p-3 font-mono text-sm outline-none focus:bg-[var(--surface-container)]"
              :placeholder="importTab === 'bibtex' ? '@article{key,\n  title = {文章标题},\n  author = {作者},\n  year = {2024},\n  journal = {期刊名}\n}' : 'TY  - JOUR\nTI  - 文章标题\nAU  - 作者\nPY  - 2024\nJO  - 期刊名\nER  -'"
            ></textarea>
          </div>

          <!-- Result -->
          <div v-if="importResult" class="mb-4 rounded-lg border p-4" :class="importResult.imported > 0 ? 'border-emerald-500 bg-emerald-50' : 'border-red-500 bg-red-50'">
            <div class="flex items-center gap-2 font-medium" :class="importResult.imported > 0 ? 'text-emerald-700' : 'text-red-700'">
              <IconCheck v-if="importResult.imported > 0" class="h-5 w-5" />
              <IconAlert v-else class="h-5 w-5" />
              成功导入 {{ importResult.imported }} 条文献
              <span v-if="importResult.skipped > 0" class="opacity-70">，跳过 {{ importResult.skipped }} 条</span>
            </div>
            <ul v-if="importResult.errors && importResult.errors.length > 0" class="mt-2 text-sm opacity-80">
              <li v-for="(err, i) in importResult.errors.slice(0, 5)" :key="i">
                {{ err.key ? `[${err.key}] ` : '' }}{{ err.message }}
              </li>
            </ul>
          </div>

          <div class="flex justify-end gap-2">
            <button
              type="button"
              class="rounded-lg px-4 py-2 text-sm text-[var(--color-on-surface-variant)] hover:bg-[var(--surface-container)]"
              @click="closeImportModal"
            >
              关闭
            </button>
            <button
              type="button"
              class="flex items-center gap-2 rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-[var(--color-on-primary)] disabled:opacity-50"
              :disabled="importLoading || !importContent.trim()"
              @click="handleImport"
            >
              <IconLoading v-if="importLoading" class="h-4 w-4 animate-spin" />
              导入
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- AI Search Modal -->
    <Teleport to="body">
      <div v-if="showAISearchModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="closeAISearchModal">
        <div class="flex h-[80vh] w-full max-w-4xl flex-col rounded-xl bg-[var(--surface-container-lowest)] p-6 shadow-xl">
          <div class="mb-4 flex items-center justify-between">
            <h2 class="text-xl font-semibold">AI 文献检索</h2>
            <button type="button" class="rounded p-1 hover:bg-[var(--surface-container)]" @click="closeAISearchModal">
              <IconClose class="h-5 w-5" />
            </button>
          </div>

          <!-- Search Input -->
          <div class="mb-4 flex gap-3">
            <div class="relative flex-1">
              <IconSearch class="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-[var(--color-on-surface-variant)] opacity-50" />
              <input
                v-model="aiSearchQuery"
                type="text"
                class="w-full rounded-lg bg-[var(--surface-container-lowest)] py-2.5 pl-10 pr-4 text-sm outline-none focus:bg-[var(--surface-container)]"
                placeholder="搜索论文标题、作者、关键词..."
                @keyup.enter="handleAISearch"
              />
            </div>
            <select
              v-model="aiSearchProvider"
              class="rounded-lg bg-[var(--surface-container-lowest)] px-3 py-2 text-sm outline-none focus:bg-[var(--surface-container)]"
            >
              <option value="all">全部来源</option>
              <option value="semanticscholar">Semantic Scholar</option>
              <option value="crossref">Crossref</option>
            </select>
            <button
              type="button"
              class="flex items-center gap-2 rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-[var(--color-on-primary)] disabled:opacity-50"
              :disabled="aiSearchLoading || !aiSearchQuery.trim()"
              @click="handleAISearch"
            >
              <IconLoading v-if="aiSearchLoading" class="h-4 w-4 animate-spin" />
              <IconCloudSearch v-else class="h-4 w-4" />
              搜索
            </button>
          </div>

          <!-- Results -->
          <div class="flex-1 overflow-y-auto">
            <div v-if="aiSearchLoading" class="flex items-center justify-center py-20">
              <IconLoading class="h-8 w-8 animate-spin text-[var(--color-primary)]" />
            </div>

            <div v-else-if="aiSearchResults.length === 0" class="flex flex-col items-center justify-center py-20 text-[var(--color-on-surface-variant)]">
              <IconCloudSearch class="mb-4 h-16 w-16 opacity-30" />
              <p>输入关键词搜索学术论文</p>
              <p class="mt-1 text-sm opacity-60">支持 Semantic Scholar 和 Crossref 数据库</p>
            </div>

            <div v-else class="space-y-3">
              <!-- Results Header -->
              <div class="flex items-center justify-between border-b border-[var(--surface-container-high)] pb-2 text-sm">
                <span class="text-[var(--color-on-surface-variant)]">
                  找到 {{ aiSearchTotal }} 篇文献
                </span>
                <div class="flex gap-2">
                  <button
                    type="button"
                    class="text-[var(--color-primary)] hover:underline"
                    @click="selectAllSearchResults"
                  >
                    全选
                  </button>
                  <button
                    type="button"
                    class="text-[var(--color-on-surface-variant)] hover:underline"
                    @click="deselectAllSearchResults"
                  >
                    取消全选
                  </button>
                </div>
              </div>

              <!-- Result Items -->
              <div
                v-for="item in aiSearchResults"
                :key="item.id"
                class="group cursor-pointer rounded-lg border p-4 transition-colors"
                :class="aiSearchSelected.has(item.id) ? 'border-[var(--color-primary)] bg-[var(--color-primary-container)]/10' : 'border-[var(--surface-container-high)] hover:bg-[var(--surface-container)]'"
                @click="toggleSelectSearchResult(item.id)"
              >
                <div class="flex items-start gap-3">
                  <div class="mt-1 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded border"
                    :class="aiSearchSelected.has(item.id) ? 'border-[var(--color-primary)] bg-[var(--color-primary)]' : 'border-[var(--outline-variant)]'">
                    <IconCheck v-if="aiSearchSelected.has(item.id)" class="h-3 w-3 text-[var(--color-on-primary)]" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <h4 class="font-medium leading-tight">{{ item.title }}</h4>
                    <p class="mt-1 text-sm text-[var(--color-on-surface-variant)]">
                      {{ item.authors.map(a => a.literal || `${a.given} ${a.family}`).join(', ') }}
                      <span v-if="item.year"> · {{ item.year }}</span>
                      <span v-if="item.source"> · {{ item.source }}</span>
                    </p>
                    <div class="mt-2 flex flex-wrap items-center gap-2 text-xs">
                      <span class="rounded bg-[var(--surface-container-high)] px-2 py-0.5">
                        {{ item.type }}
                      </span>
                      <span v-if="item.citation_count" class="text-[var(--color-on-surface-variant)]">
                        引用: {{ item.citation_count }}
                      </span>
                      <span class="text-[var(--color-on-surface-variant)] opacity-60">
                        {{ item.provider === 'semanticscholar' ? 'Semantic Scholar' : 'Crossref' }}
                      </span>
                    </div>
                    <p v-if="item.abstract" class="mt-2 line-clamp-2 text-sm text-[var(--color-on-surface-variant)] opacity-80">
                      {{ item.abstract }}
                    </p>
                    <div class="mt-2 flex gap-2">
                      <a
                        v-if="item.open_access_url"
                        :href="item.open_access_url"
                        target="_blank"
                        class="flex items-center gap-1 text-xs text-[var(--color-primary)] hover:underline"
                        @click.stop
                      >
                        <IconOpenInNew class="h-3 w-3" />
                        开放获取
                      </a>
                      <a
                        v-if="item.doi"
                        :href="`https://doi.org/${item.doi}`"
                        target="_blank"
                        class="flex items-center gap-1 text-xs text-[var(--color-primary)] hover:underline"
                        @click.stop
                      >
                        <IconOpenInNew class="h-3 w-3" />
                        DOI
                      </a>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="mt-4 flex items-center justify-between border-t border-[var(--surface-container-high)] pt-4">
            <span class="text-sm text-[var(--color-on-surface-variant)]">
              已选择 {{ aiSearchSelected.size }} 篇
            </span>
            <div class="flex gap-2">
              <button
                type="button"
                class="rounded-lg px-4 py-2 text-sm text-[var(--color-on-surface-variant)] hover:bg-[var(--surface-container)]"
                @click="closeAISearchModal"
              >
                关闭
              </button>
              <button
                type="button"
                class="flex items-center gap-2 rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-[var(--color-on-primary)] disabled:opacity-50"
                :disabled="aiSearchImporting || aiSearchSelected.size === 0"
                @click="importSelectedResults"
              >
                <IconLoading v-if="aiSearchImporting" class="h-4 w-4 animate-spin" />
                <IconPlus v-else class="h-4 w-4" />
                导入选中
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </section>
</template>

<style scoped>
.library-sidebar {
  min-width: 16rem;
}

.library-main {
  min-width: 0;
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

dd :deep(p) {
  margin-bottom: 0.5rem;
}

dd :deep(p:last-child) {
  margin-bottom: 0;
}

dd :deep(sup) {
  font-size: 0.75em;
  vertical-align: super;
}

dd :deep(sub) {
  font-size: 0.75em;
  vertical-align: sub;
}

dd :deep(i) {
  font-style: italic;
}

dd :deep(b) {
  font-weight: 600;
}
</style>