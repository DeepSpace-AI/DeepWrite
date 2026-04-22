import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type {
  Reference,
  CollectionTree,
  ListReferencesParams,
  ListReferencesResponse,
  CreateReferenceInput,
  UpdateReferenceInput,
} from '@/types/reference'
import * as api from '@/api/reference'

export const useReferenceStore = defineStore('reference', () => {
  const references = ref<Reference[]>([])
  const totalReferences = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(20)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  const selectedReference = ref<Reference | null>(null)
  const collections = ref<CollectionTree[]>([])
  const selectedCollectionId = ref<string | null>(null)

  const searchQuery = ref('')
  const filterType = ref<string>('')
  const filterYearFrom = ref<number | undefined>()
  const filterYearTo = ref<number | undefined>()
  const filterStarred = ref<boolean | undefined>()
  const filterStatus = ref<string>('')

  const hasMore = computed(() => {
    return references.value.length < totalReferences.value
  })

  async function fetchReferences(params: Partial<ListReferencesParams> = {}) {
    isLoading.value = true
    error.value = null

    try {
      const response: ListReferencesResponse = await api.listReferences({
        q: searchQuery.value || undefined,
        type: filterType.value || undefined,
        year_from: filterYearFrom.value,
        year_to: filterYearTo.value,
        collection_id: selectedCollectionId.value || undefined,
        starred: filterStarred.value,
        status: filterStatus.value || undefined,
        sort_by: 'created_at',
        sort_order: 'desc',
        page: params.page || currentPage.value,
        page_size: pageSize.value,
        ...params,
      })

      references.value = response.items
      totalReferences.value = response.total
      currentPage.value = response.page
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch references'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function fetchMoreReferences() {
    if (!hasMore.value || isLoading.value) return

    isLoading.value = true
    error.value = null

    try {
      const response = await api.listReferences({
        q: searchQuery.value || undefined,
        type: filterType.value || undefined,
        year_from: filterYearFrom.value,
        year_to: filterYearTo.value,
        collection_id: selectedCollectionId.value || undefined,
        starred: filterStarred.value,
        status: filterStatus.value || undefined,
        sort_by: 'created_at',
        sort_order: 'desc',
        page: currentPage.value + 1,
        page_size: pageSize.value,
      })

      references.value = [...references.value, ...response.items]
      totalReferences.value = response.total
      currentPage.value = response.page
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch more references'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function fetchReference(id: string) {
    isLoading.value = true
    error.value = null

    try {
      const ref = await api.getReference(id)
      selectedReference.value = ref
      return ref
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch reference'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function createReference(input: CreateReferenceInput) {
    isLoading.value = true
    error.value = null

    try {
      const ref = await api.createReference(input)
      references.value.unshift(ref)
      totalReferences.value++
      return ref
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to create reference'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function updateReference(id: string, input: UpdateReferenceInput) {
    isLoading.value = true
    error.value = null

    try {
      const ref = await api.updateReference(id, input)
      const index = references.value.findIndex(r => r.id === id)
      if (index !== -1) {
        references.value[index] = ref
      }
      if (selectedReference.value?.id === id) {
        selectedReference.value = ref
      }
      return ref
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to update reference'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function deleteReference(id: string) {
    isLoading.value = true
    error.value = null

    try {
      await api.deleteReference(id)
      references.value = references.value.filter(r => r.id !== id)
      totalReferences.value--
      if (selectedReference.value?.id === id) {
        selectedReference.value = null
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to delete reference'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function toggleStar(id: string) {
    const ref = references.value.find(r => r.id === id)
    if (!ref) return

    return updateReference(id, { starred: !ref.starred })
  }

  async function fetchCollections(workspaceId?: string) {
    try {
      collections.value = await api.listCollections(workspaceId)
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch collections'
      throw e
    }
  }

  async function createCollection(name: string, parentId?: string, workspaceId?: string) {
    try {
      const col = await api.createCollection({
        name,
        parent_id: parentId,
        workspace_id: workspaceId,
      })
      await fetchCollections(workspaceId)
      return col
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to create collection'
      throw e
    }
  }

  async function deleteCollection(id: string) {
    try {
      await api.deleteCollection(id)
      await fetchCollections()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to delete collection'
      throw e
    }
  }

  async function addReferenceToCollection(collectionId: string, referenceId: string) {
    try {
      await api.addReferenceToCollection(collectionId, referenceId)
      await fetchCollections()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to add reference to collection'
      throw e
    }
  }

  async function removeReferenceFromCollection(collectionId: string, referenceId: string) {
    try {
      await api.removeReferenceFromCollection(collectionId, referenceId)
      await fetchCollections()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to remove reference from collection'
      throw e
    }
  }

  function setSearchQuery(query: string) {
    searchQuery.value = query
  }

  function setFilters(filters: {
    type?: string
    yearFrom?: number
    yearTo?: number
    starred?: boolean
    status?: string
  }) {
    filterType.value = filters.type || ''
    filterYearFrom.value = filters.yearFrom
    filterYearTo.value = filters.yearTo
    filterStarred.value = filters.starred
    filterStatus.value = filters.status || ''
  }

  function selectCollection(collectionId: string | null) {
    selectedCollectionId.value = collectionId
  }

  function clearFilters() {
    searchQuery.value = ''
    filterType.value = ''
    filterYearFrom.value = undefined
    filterYearTo.value = undefined
    filterStarred.value = undefined
    filterStatus.value = ''
    selectedCollectionId.value = null
  }

  function reset() {
    references.value = []
    totalReferences.value = 0
    currentPage.value = 1
    selectedReference.value = null
    collections.value = []
    selectedCollectionId.value = null
    searchQuery.value = ''
    filterType.value = ''
    filterYearFrom.value = undefined
    filterYearTo.value = undefined
    filterStarred.value = undefined
    filterStatus.value = ''
    error.value = null
  }

  const isExtracting = ref(false)
  const extractionTaskId = ref<string | null>(null)
  const extractionResult = ref<api.PdfExtractResult | null>(null)

  async function extractPdfMetadata(referenceId: string, wait = true) {
    isExtracting.value = true
    error.value = null
    extractionResult.value = null

    try {
      const response = await api.extractPdfFromReference(referenceId, { wait })

      if (response.task_id && !wait) {
        extractionTaskId.value = response.task_id
        return response
      }

      if (response.extracted) {
        extractionResult.value = response.extracted
        return response.extracted
      }

      if (response.error) {
        error.value = response.error
        throw new Error(response.error)
      }

      return response
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to extract PDF metadata'
      throw e
    } finally {
      isExtracting.value = false
    }
  }

  async function waitForExtraction(taskId: string) {
    isExtracting.value = true
    try {
      const status = await api.waitForExtractionTask(taskId)
      if (status.result) {
        extractionResult.value = status.result
      }
      if (status.error) {
        error.value = status.error
      }
      return status
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to wait for extraction'
      throw e
    } finally {
      isExtracting.value = false
    }
  }

  async function createReferenceFromPdf(referenceId: string) {
    isExtracting.value = true
    try {
      const result = await extractPdfMetadata(referenceId, true)
      if (result && 'reference' in result && result.reference) {
        const newRef = await createReference({
          title: result.reference.title,
          authors: result.reference.authors,
          year: result.reference.year,
          source: result.reference.source,
          doi: result.reference.doi,
          type: result.reference.type as unknown,
          abstract: result.reference.abstract,
          volume: result.reference.volume,
          issue: result.reference.issue,
          pages: result.reference.pages,
          publisher: result.reference.publisher,
          url: result.reference.url,
          file_id: referenceId,
        })
        return newRef
      }
      return null
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to create reference from PDF'
      throw e
    } finally {
      isExtracting.value = false
    }
  }

  return {
    references,
    totalReferences,
    currentPage,
    pageSize,
    isLoading,
    error,
    selectedReference,
    collections,
    selectedCollectionId,
    searchQuery,
    filterType,
    filterYearFrom,
    filterYearTo,
    filterStarred,
    filterStatus,
    hasMore,
    fetchReferences,
    fetchMoreReferences,
    fetchReference,
    createReference,
    updateReference,
    deleteReference,
    toggleStar,
    fetchCollections,
    createCollection,
    deleteCollection,
    addReferenceToCollection,
    removeReferenceFromCollection,
    setSearchQuery,
    setFilters,
    selectCollection,
    clearFilters,
    reset,
    isExtracting,
    extractionTaskId,
    extractionResult,
    extractPdfMetadata,
    waitForExtraction,
    createReferenceFromPdf,
  }
})