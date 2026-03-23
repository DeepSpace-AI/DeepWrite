<script setup lang="ts">
import { computed, nextTick, onUnmounted, ref, shallowRef, watch, type ComponentPublicInstance } from 'vue'
import { useI18n } from 'vue-i18n'
import { GlobalWorkerOptions, getDocument, type PDFDocumentProxy, type PDFPageProxy, type RenderTask } from 'pdfjs-dist'
import { TextLayerBuilder } from 'pdfjs-dist/web/pdf_viewer'
import pdfWorkerSrc from 'pdfjs-dist/build/pdf.worker?url'
import 'pdfjs-dist/web/pdf_viewer.css'
import IconClose from '~icons/mdi/close'
import IconChevronLeft from '~icons/mdi/chevron-left'
import IconChevronRight from '~icons/mdi/chevron-right'
import IconMagnifyPlus from '~icons/mdi/magnify-plus'
import IconMagnifyMinus from '~icons/mdi/magnify-minus'
import IconAspectRatio from '~icons/mdi/aspect-ratio'
import IconPrinterOutline from '~icons/mdi/printer-outline'
import IconContentCopy from '~icons/mdi/content-copy'
import IconFormatColorHighlight from '~icons/mdi/format-color-highlight'
import IconNoteOutline from '~icons/mdi/note-outline'
import IconDraw from '~icons/mdi/draw'
import IconPalette from '~icons/mdi/palette'
import IconDeleteOutline from '~icons/mdi/delete-outline'
import IconCheck from '~icons/mdi/check'
import IconCloseCircleOutline from '~icons/mdi/close-circle-outline'
import type { PDFAnnotation, WorkspaceFile, DrawingPath } from '@/views/workspace/types'

GlobalWorkerOptions.workerSrc = pdfWorkerSrc

const props = defineProps<{
  file: WorkspaceFile
  annotations: PDFAnnotation[]
  previewUrl?: string
  workspaceId: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'annotation-create', annotation: Partial<PDFAnnotation>): void
  (e: 'annotation-update', annotationId: string, updates: Partial<PDFAnnotation>): void
  (e: 'annotation-delete', annotationId: string): void
}>()

const { t } = useI18n()

const HIGHLIGHT_COLORS = [
  { name: 'yellow', hex: '#FFEB3B' },
  { name: 'green', hex: '#4CAF50' },
  { name: 'blue', hex: '#2196F3' },
  { name: 'orange', hex: '#FF9800' },
  { name: 'pink', hex: '#E91E63' },
  { name: 'purple', hex: '#9C27B0' },
]

const DRAWING_COLORS = [
  { name: 'black', hex: '#000000' },
  { name: 'red', hex: '#F44336' },
  { name: 'blue', hex: '#2196F3' },
  { name: 'green', hex: '#4CAF50' },
  { name: 'yellow', hex: '#FFEB3B' },
]

type AnnotationTool = 'none' | 'highlight' | 'note' | 'drawing'
type AnnotationFilter = 'all' | 'highlight' | 'note' | 'drawing'
type ReadingMode = 'paged' | 'continuous'

const currentPage = ref(1)
const totalPages = ref(0)
const scale = ref(1.0)
const readingMode = ref<ReadingMode>('paged')
const isLoadingPdf = ref(false)
const isRenderingPage = ref(false)
const isRenderingContinuous = ref(false)
const renderErrorMessage = ref('')
const copyStatusMessage = ref('')

const pdfDocument = shallowRef<PDFDocumentProxy | null>(null)
const currentPdfPage = shallowRef<PDFPageProxy | null>(null)
const currentRenderTask = shallowRef<RenderTask | null>(null)
const currentTextLayerBuilder = shallowRef<TextLayerBuilder | null>(null)

const pageBaseSize = ref({ width: 0, height: 0 })
const pageDisplaySize = ref({ width: 0, height: 0 })

const activeTool = ref<AnnotationTool>('none')
const selectedColor = ref(HIGHLIGHT_COLORS[0]?.hex ?? '#FFEB3B')
const selectedDrawingColor = ref(DRAWING_COLORS[0]?.hex ?? '#000000')
const strokeWidth = ref(2)
const annotationFilter = ref<AnnotationFilter>('all')
const editingAnnotationId = ref<string | null>(null)
const editingContent = ref('')

const containerRef = ref<HTMLElement | null>(null)
const pageLayerRef = ref<HTMLElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)
const textLayerRef = ref<HTMLElement | null>(null)
const noteInputPosition = ref<{ x: number; y: number } | null>(null)
const noteInputValue = ref('')
const isDrawing = ref(false)
const currentPath = ref<{ x: number; y: number }[]>([])
const drawingStartPos = ref<{ x: number; y: number } | null>(null)
const lastWheelFlipAt = ref(0)
const autoFitEnabled = ref(false)
const isSyncingCurrentPageFromScroll = ref(false)

const continuousPageSizes = ref<Record<number, { width: number; height: number }>>({})
const continuousRenderVersion = ref(0)
const continuousPageRefs = new Map<number, HTMLElement>()
const continuousCanvasRefs = new Map<number, HTMLCanvasElement>()
const continuousTextLayerRefs = new Map<number, HTMLElement>()
const continuousRenderTasks = new Map<number, RenderTask>()
const continuousTextLayerBuilders = new Map<number, TextLayerBuilder>()

const zoomLevels = [0.5, 0.75, 1.0, 1.25, 1.5, 2.0]

const zoomOptions = computed(() => {
  const epsilon = 0.001
  const hasCurrent = zoomLevels.some(level => Math.abs(level - scale.value) < epsilon)
  if (hasCurrent) return zoomLevels
  return [...zoomLevels, Number(scale.value.toFixed(2))].sort((a, b) => a - b)
})

const filteredAnnotations = computed(() => {
  if (annotationFilter.value === 'all') {
    return props.annotations
  }
  return props.annotations.filter(a => a.type === annotationFilter.value)
})

const annotationsByPage = computed(() => {
  const byPage = new Map<number, PDFAnnotation[]>()
  for (const ann of filteredAnnotations.value) {
    if (!byPage.has(ann.page)) {
      byPage.set(ann.page, [])
    }
    byPage.get(ann.page)!.push(ann)
  }
  return byPage
})

const pageAnnotations = computed(() => {
  return annotationsByPage.value.get(currentPage.value) || []
})

const annotationFilterOptions: AnnotationFilter[] = ['all', 'highlight', 'note', 'drawing']

function annotationFilterLabel(filter: AnnotationFilter) {
  if (filter === 'all') return t('workspace.pdf.filterAll')
  if (filter === 'highlight') return t('workspace.pdf.filterHighlight')
  if (filter === 'note') return t('workspace.pdf.filterNote')
  return t('workspace.pdf.filterDrawing')
}

function onDocumentLoad({ numPages }: { numPages: number }) {
  totalPages.value = numPages
  if (currentPage.value > numPages) {
    currentPage.value = numPages
  }
}

function clearCopyStatus() {
  if (!copyStatusMessage.value) return
  window.setTimeout(() => {
    copyStatusMessage.value = ''
  }, 1200)
}

async function copySelectedText() {
  try {
    const selected = window.getSelection()?.toString()?.trim() || ''
    if (selected) {
      await navigator.clipboard.writeText(selected)
      copyStatusMessage.value = t('workspace.pdf.copySelectionSuccess')
      clearCopyStatus()
      return
    }

    if (!currentPdfPage.value) {
      copyStatusMessage.value = t('workspace.pdf.copyNoText')
      clearCopyStatus()
      return
    }

    const textContent = await currentPdfPage.value.getTextContent()
    const fullText = textContent.items
      .map(item => ('str' in item ? item.str : ''))
      .join(' ')
      .trim()

    if (!fullText) {
      copyStatusMessage.value = t('workspace.pdf.copyNoText')
      clearCopyStatus()
      return
    }

    await navigator.clipboard.writeText(fullText)
    copyStatusMessage.value = t('workspace.pdf.copyPageSuccess')
    clearCopyStatus()
  } catch {
    copyStatusMessage.value = t('workspace.pdf.copyFailed')
    clearCopyStatus()
  }
}

function printDocument() {
  if (!props.previewUrl) return
  const iframe = document.createElement('iframe')
  iframe.style.position = 'fixed'
  iframe.style.width = '0'
  iframe.style.height = '0'
  iframe.style.opacity = '0'
  iframe.style.pointerEvents = 'none'
  iframe.src = props.previewUrl
  iframe.onload = () => {
    const printWindow = iframe.contentWindow
    if (printWindow) {
      printWindow.focus()
      printWindow.print()
    }
    window.setTimeout(() => {
      iframe.remove()
    }, 1500)
  }
  document.body.appendChild(iframe)
}

function goToPage(page: number) {
  if (page < 1) page = 1
  if (page > totalPages.value) page = totalPages.value
  currentPage.value = page

  if (readingMode.value === 'continuous') {
    const targetPage = continuousPageRefs.get(page)
    targetPage?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}

function previousPage() {
  goToPage(currentPage.value - 1)
}

function nextPage() {
  goToPage(currentPage.value + 1)
}

function updateScale(value: number, enableAutoFit = false) {
  const safeScale = Math.max(0.25, Math.min(4, Number(value.toFixed(3))))
  scale.value = safeScale
  autoFitEnabled.value = enableAutoFit
}

function getNextZoomLevel(current: number) {
  return zoomLevels.find(level => level > current + 0.001)
}

function getPreviousZoomLevel(current: number) {
  for (let i = zoomLevels.length - 1; i >= 0; i -= 1) {
    if ((zoomLevels[i] ?? 0) < current - 0.001) {
      return zoomLevels[i]
    }
  }
  return undefined
}

async function getCurrentPageBaseSize() {
  if (!pdfDocument.value) return null
  try {
    const page = await pdfDocument.value.getPage(currentPage.value)
    const viewport = page.getViewport({ scale: 1 })
    return {
      width: viewport.width,
      height: viewport.height,
    }
  } catch {
    return null
  }
}

async function fitToPage() {
  if (!containerRef.value) return

  const baseSize = await getCurrentPageBaseSize()
  if (!baseSize) return

  const horizontalPadding = 32
  const verticalPadding = 32
  const availableWidth = Math.max(containerRef.value.clientWidth - horizontalPadding, 1)
  const availableHeight = Math.max(containerRef.value.clientHeight - verticalPadding, 1)

  const fitWidth = availableWidth / baseSize.width
  const fitHeight = availableHeight / baseSize.height

  const nextScale = readingMode.value === 'continuous'
    ? fitWidth
    : Math.min(fitWidth, fitHeight)

  updateScale(nextScale, true)
}

function handleScaleSelectChange(value: number) {
  updateScale(value, false)
}

function setReadingMode(mode: ReadingMode) {
  if (readingMode.value === mode) return
  readingMode.value = mode
  if (mode === 'continuous') {
    setTool('none')
  }
}

function setContinuousPageRef(pageNumber: number) {
  return (el: Element | ComponentPublicInstance | null) => {
    if (el instanceof HTMLElement) {
      continuousPageRefs.set(pageNumber, el)
      return
    }
    continuousPageRefs.delete(pageNumber)
  }
}

function setContinuousCanvasRef(pageNumber: number) {
  return (el: Element | ComponentPublicInstance | null) => {
    if (el instanceof HTMLCanvasElement) {
      continuousCanvasRefs.set(pageNumber, el)
      return
    }
    continuousCanvasRefs.delete(pageNumber)
  }
}

function setContinuousTextLayerRef(pageNumber: number) {
  return (el: Element | ComponentPublicInstance | null) => {
    if (el instanceof HTMLElement) {
      continuousTextLayerRefs.set(pageNumber, el)
      return
    }
    continuousTextLayerRefs.delete(pageNumber)
  }
}

function getContinuousPageSize(pageNumber: number) {
  const saved = continuousPageSizes.value[pageNumber]
  if (saved) return saved
  const width = Math.max(pageDisplaySize.value.width, 640)
  const height = Math.max(pageDisplaySize.value.height, 900)
  return { width, height }
}

function cancelContinuousRender() {
  for (const task of continuousRenderTasks.values()) {
    task.cancel()
  }
  continuousRenderTasks.clear()

  for (const builder of continuousTextLayerBuilders.values()) {
    builder.cancel()
  }
  continuousTextLayerBuilders.clear()
}

async function renderContinuousPage(pageNumber: number, renderVersion: number) {
  if (!pdfDocument.value || readingMode.value !== 'continuous') return
  if (renderVersion !== continuousRenderVersion.value) return

  const pageDiv = continuousPageRefs.get(pageNumber)
  const canvas = continuousCanvasRefs.get(pageNumber)
  const textLayerElement = continuousTextLayerRefs.get(pageNumber)
  if (!pageDiv || !canvas || !textLayerElement) return

  const page = await pdfDocument.value.getPage(pageNumber)
  if (renderVersion !== continuousRenderVersion.value || readingMode.value !== 'continuous') return

  const viewport = page.getViewport({ scale: scale.value })
  continuousPageSizes.value = {
    ...continuousPageSizes.value,
    [pageNumber]: {
      width: viewport.width,
      height: viewport.height,
    },
  }

  const context = canvas.getContext('2d')
  if (!context) return

  const dpr = window.devicePixelRatio || 1
  canvas.width = Math.floor(viewport.width * dpr)
  canvas.height = Math.floor(viewport.height * dpr)
  canvas.style.width = `${viewport.width}px`
  canvas.style.height = `${viewport.height}px`

  const renderTask = page.render({
    canvasContext: context,
    viewport,
    transform: dpr === 1 ? undefined : [dpr, 0, 0, dpr, 0, 0],
  })
  continuousRenderTasks.set(pageNumber, renderTask)
  try {
    await renderTask.promise
  } catch (error) {
    if (error instanceof Error && error.name === 'RenderingCancelledException') {
      return
    }
    throw error
  }

  if (renderVersion !== continuousRenderVersion.value || readingMode.value !== 'continuous') return

  textLayerElement.innerHTML = ''
  textLayerElement.style.width = `${viewport.width}px`
  textLayerElement.style.height = `${viewport.height}px`

  const textContent = await page.getTextContent()
  const textLayerBuilder = new TextLayerBuilder({
    highlighter: null,
    accessibilityManager: null,
    isOffscreenCanvasSupported: true,
  })
  textLayerElement.appendChild(textLayerBuilder.div)
  textLayerBuilder.setTextContentSource(textContent)
  continuousTextLayerBuilders.set(pageNumber, textLayerBuilder)
  try {
    await textLayerBuilder.render(viewport)
  } catch (error) {
    if (error instanceof Error && error.name === 'RenderingCancelledException') {
      return
    }
    throw error
  }
}

async function renderContinuousPages() {
  if (!pdfDocument.value || totalPages.value < 1) return
  cancelContinuousRender()
  continuousRenderVersion.value += 1
  const version = continuousRenderVersion.value
  continuousPageSizes.value = {}

  isRenderingContinuous.value = true
  await nextTick()

  try {
    for (let pageNumber = 1; pageNumber <= totalPages.value; pageNumber += 1) {
      if (version !== continuousRenderVersion.value || readingMode.value !== 'continuous') {
        break
      }
      await renderContinuousPage(pageNumber, version)
    }
  } catch (error) {
    if (!(error instanceof Error && error.name === 'RenderingCancelledException')) {
      renderErrorMessage.value = t('workspace.pdf.renderFailed')
    }
  } finally {
    if (version === continuousRenderVersion.value) {
      isRenderingContinuous.value = false
    }
  }
}

function handleReaderWheel(event: WheelEvent) {
  if (readingMode.value !== 'paged') return
  if (totalPages.value <= 1) return

  const now = Date.now()
  if (now - lastWheelFlipAt.value < 220) {
    return
  }

  const deltaY = event.deltaY
  if (Math.abs(deltaY) < 18) return

  if (deltaY > 0 && currentPage.value < totalPages.value) {
    event.preventDefault()
    lastWheelFlipAt.value = now
    nextPage()
    return
  }

  if (deltaY < 0 && currentPage.value > 1) {
    event.preventDefault()
    lastWheelFlipAt.value = now
    previousPage()
  }
}

function syncCurrentPageFromContinuousScroll() {
  if (readingMode.value !== 'continuous' || !containerRef.value || totalPages.value < 1) return

  const containerRect = containerRef.value.getBoundingClientRect()
  const probeY = containerRect.top + containerRect.height * 0.35

  let nearestPage = currentPage.value
  let nearestDistance = Number.POSITIVE_INFINITY

  for (let pageNumber = 1; pageNumber <= totalPages.value; pageNumber += 1) {
    const pageElement = continuousPageRefs.get(pageNumber)
    if (!pageElement) continue

    const rect = pageElement.getBoundingClientRect()
    const center = rect.top + rect.height / 2
    const distance = Math.abs(center - probeY)
    if (distance < nearestDistance) {
      nearestDistance = distance
      nearestPage = pageNumber
    }
  }

  if (nearestPage !== currentPage.value) {
    isSyncingCurrentPageFromScroll.value = true
    currentPage.value = nearestPage
    window.setTimeout(() => {
      isSyncingCurrentPageFromScroll.value = false
    }, 0)
  }
}

function handleContainerScroll() {
  if (readingMode.value !== 'continuous') return
  syncCurrentPageFromContinuousScroll()
}

function zoomIn() {
  const next = getNextZoomLevel(scale.value)
  if (next) {
    updateScale(next, false)
  }
}

function zoomOut() {
  const previous = getPreviousZoomLevel(scale.value)
  if (previous) {
    updateScale(previous, false)
  }
}

function setTool(tool: AnnotationTool) {
  activeTool.value = activeTool.value === tool ? 'none' : tool
  noteInputPosition.value = null
  isDrawing.value = false
  currentPath.value = []
}

function getRelativePosition(e: MouseEvent): { x: number; y: number } | null {
  if (!pageLayerRef.value) return null

  const rect = pageLayerRef.value.getBoundingClientRect()
  const x = e.clientX - rect.left
  const y = e.clientY - rect.top
  if (x < 0 || y < 0 || x > rect.width || y > rect.height) {
    return null
  }

  return {
    x,
    y,
  }
}

function isNormalizedRect(annotation: PDFAnnotation): boolean {
  return (
    annotation.rect_x >= 0
    && annotation.rect_x <= 1
    && annotation.rect_y >= 0
    && annotation.rect_y <= 1
    && annotation.rect_width >= 0
    && annotation.rect_width <= 1
    && annotation.rect_height >= 0
    && annotation.rect_height <= 1
  )
}

function isNormalizedPoint(point: { x: number; y: number }): boolean {
  return point.x >= 0 && point.x <= 1 && point.y >= 0 && point.y <= 1
}

function getScaleRatio() {
  if (!pageBaseSize.value.width || !pageBaseSize.value.height) {
    return { x: 1, y: 1 }
  }
  return {
    x: pageDisplaySize.value.width / pageBaseSize.value.width,
    y: pageDisplaySize.value.height / pageBaseSize.value.height,
  }
}

function getDisplayRect(annotation: PDFAnnotation) {
  if (isNormalizedRect(annotation)) {
    return {
      x: annotation.rect_x * pageDisplaySize.value.width,
      y: annotation.rect_y * pageDisplaySize.value.height,
      width: annotation.rect_width * pageDisplaySize.value.width,
      height: annotation.rect_height * pageDisplaySize.value.height,
    }
  }

  const ratio = getScaleRatio()
  return {
    x: annotation.rect_x * ratio.x,
    y: annotation.rect_y * ratio.y,
    width: annotation.rect_width * ratio.x,
    height: annotation.rect_height * ratio.y,
  }
}

function getNormalizedRectFromPoints(start: { x: number; y: number }, end: { x: number; y: number }) {
  const width = pageDisplaySize.value.width || 1
  const height = pageDisplaySize.value.height || 1
  const minX = Math.min(start.x, end.x)
  const minY = Math.min(start.y, end.y)
  return {
    x: minX / width,
    y: minY / height,
    width: Math.abs(end.x - start.x) / width,
    height: Math.abs(end.y - start.y) / height,
  }
}

function normalizePoint(point: { x: number; y: number }) {
  const width = pageDisplaySize.value.width || 1
  const height = pageDisplaySize.value.height || 1
  return {
    x: point.x / width,
    y: point.y / height,
  }
}

function parseDrawingPaths(annotation: PDFAnnotation): DrawingPath[] {
  const rawPaths = (annotation as PDFAnnotation & { paths?: DrawingPath[] | string }).paths
  if (!rawPaths) return []
  if (Array.isArray(rawPaths)) return rawPaths
  if (typeof rawPaths === 'string') {
    try {
      const parsed = JSON.parse(rawPaths)
      return Array.isArray(parsed) ? parsed : []
    } catch {
      return []
    }
  }
  return []
}

function toDisplayPoint(point: { x: number; y: number }) {
  if (isNormalizedPoint(point)) {
    return {
      x: point.x * pageDisplaySize.value.width,
      y: point.y * pageDisplaySize.value.height,
    }
  }
  const ratio = getScaleRatio()
  return {
    x: point.x * ratio.x,
    y: point.y * ratio.y,
  }
}

function getDrawingPathData(path: DrawingPath) {
  if (!path.points?.length) return ''
  return path.points
    .map((point, index) => {
      const displayPoint = toDisplayPoint(point)
      return `${index === 0 ? 'M' : 'L'} ${displayPoint.x} ${displayPoint.y}`
    })
    .join(' ')
}

function drawingPathStrokeWidth(path: DrawingPath): number {
  const ratio = getScaleRatio()
  const factor = (ratio.x + ratio.y) / 2
  return Math.max((path.width || 2) * factor, 1)
}

function handleContentClick(e: MouseEvent) {
  if (activeTool.value === 'none') return

  const pos = getRelativePosition(e)
  if (!pos) return

  if (activeTool.value === 'note') {
    noteInputPosition.value = pos
    noteInputValue.value = ''
  }
}

function handleContentMouseDown(e: MouseEvent) {
  if (activeTool.value !== 'drawing' && activeTool.value !== 'highlight') return

  const pos = getRelativePosition(e)
  if (!pos) return

  if (activeTool.value === 'drawing') {
    isDrawing.value = true
    currentPath.value = [pos]
    drawingStartPos.value = pos
  } else if (activeTool.value === 'highlight') {
    isDrawing.value = true
    currentPath.value = [pos]
    drawingStartPos.value = pos
  }
}

function handleContentMouseMove(e: MouseEvent) {
  if (!isDrawing.value) return

  const pos = getRelativePosition(e)
  if (!pos) return

  currentPath.value.push(pos)
}

function handleContentMouseUp(e: MouseEvent) {
  if (!isDrawing.value) return

  const pos = getRelativePosition(e)
  if (!pos) return

  if (activeTool.value === 'drawing' && currentPath.value.length > 1) {
    emit('annotation-create', {
      type: 'drawing',
      page: currentPage.value,
      rect_x: 0,
      rect_y: 0,
      rect_width: 0,
      rect_height: 0,
      paths: [{
        points: currentPath.value.map(point => normalizePoint(point)),
        color: selectedDrawingColor.value,
        width: strokeWidth.value,
      }],
      color: selectedDrawingColor.value,
    })
  } else if (activeTool.value === 'highlight' && currentPath.value.length > 1 && drawingStartPos.value) {
    const normalized = getNormalizedRectFromPoints(drawingStartPos.value, pos)
    const width = normalized.width * pageDisplaySize.value.width
    const height = normalized.height * pageDisplaySize.value.height

    if (width > 5 && height > 5) {
      emit('annotation-create', {
        type: 'highlight',
        page: currentPage.value,
        rect_x: normalized.x,
        rect_y: normalized.y,
        rect_width: normalized.width,
        rect_height: normalized.height,
        color: selectedColor.value,
      })
    }
  }

  isDrawing.value = false
  currentPath.value = []
  drawingStartPos.value = null
}

function handleNoteSubmit() {
  if (!noteInputPosition.value || !noteInputValue.value.trim()) {
    noteInputPosition.value = null
    return
  }

  const notePosition = normalizePoint(noteInputPosition.value)
  const noteWidth = 24 / (pageDisplaySize.value.width || 1)
  const noteHeight = 24 / (pageDisplaySize.value.height || 1)

  emit('annotation-create', {
    type: 'note',
    page: currentPage.value,
    rect_x: notePosition.x,
    rect_y: notePosition.y,
    rect_width: noteWidth,
    rect_height: noteHeight,
    content: noteInputValue.value.trim(),
    color: '#FF9800',
  })

  noteInputPosition.value = null
  noteInputValue.value = ''
}

function handleAnnotationClick(annotation: PDFAnnotation) {
  if (editingAnnotationId.value === annotation.id) return

  goToPage(annotation.page)

  if (annotation.type === 'note') {
    editingAnnotationId.value = annotation.id
    editingContent.value = annotation.content || ''
  }
}

function handleAnnotationDelete(annotationId: string) {
  emit('annotation-delete', annotationId)
  editingAnnotationId.value = null
}

function saveAnnotationEdit(annotation: PDFAnnotation) {
  if (!editingContent.value.trim()) return

  emit('annotation-update', annotation.id, {
    content: editingContent.value.trim(),
  })
  editingAnnotationId.value = null
  editingContent.value = ''
}

function cancelAnnotationEdit() {
  editingAnnotationId.value = null
  editingContent.value = ''
}

function getSvgPath(): string {
  if (currentPath.value.length < 2) return ''
  return currentPath.value.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`).join(' ')
}

function getHighlightRect(): { x: number; y: number; width: number; height: number } | null {
  if (!drawingStartPos.value || currentPath.value.length < 2) return null
  const last = currentPath.value[currentPath.value.length - 1]
  if (!last) return null
  const start = drawingStartPos.value
  return {
    x: Math.min(start.x, last.x),
    y: Math.min(start.y, last.y),
    width: Math.abs(last.x - start.x),
    height: Math.abs(last.y - start.y),
  }
}

watch(() => props.previewUrl, () => {
  currentPage.value = 1
})

async function renderCurrentPage() {
  if (!pdfDocument.value || !canvasRef.value || !textLayerRef.value) return
  if (currentPage.value < 1 || currentPage.value > totalPages.value) return

  currentRenderTask.value?.cancel()
  currentTextLayerBuilder.value?.cancel()
  currentTextLayerBuilder.value = null
  isRenderingPage.value = true
  renderErrorMessage.value = ''

  try {
    const page = await pdfDocument.value.getPage(currentPage.value)
    currentPdfPage.value = page

    const baseViewport = page.getViewport({ scale: 1 })
    pageBaseSize.value = {
      width: baseViewport.width,
      height: baseViewport.height,
    }

    const viewport = page.getViewport({ scale: scale.value })
    pageDisplaySize.value = {
      width: viewport.width,
      height: viewport.height,
    }

    const canvas = canvasRef.value
    const context = canvas.getContext('2d')
    if (!context) return

    const dpr = window.devicePixelRatio || 1
    canvas.width = Math.floor(viewport.width * dpr)
    canvas.height = Math.floor(viewport.height * dpr)
    canvas.style.width = `${viewport.width}px`
    canvas.style.height = `${viewport.height}px`

    currentRenderTask.value = page.render({
      canvasContext: context,
      viewport,
      transform: dpr === 1 ? undefined : [dpr, 0, 0, dpr, 0, 0],
    })

    await currentRenderTask.value.promise

    const textLayerElement = textLayerRef.value
    textLayerElement.innerHTML = ''
    textLayerElement.style.width = `${viewport.width}px`
    textLayerElement.style.height = `${viewport.height}px`

    const textContent = await page.getTextContent()
    const textLayerBuilder = new TextLayerBuilder({
      highlighter: null,
      accessibilityManager: null,
      isOffscreenCanvasSupported: true,
    })
    currentTextLayerBuilder.value = textLayerBuilder
    textLayerElement.appendChild(textLayerBuilder.div)
    textLayerBuilder.setTextContentSource(textContent)
    await textLayerBuilder.render(viewport)
  } catch (error) {
    if (!(error instanceof Error && error.name === 'RenderingCancelledException')) {
      renderErrorMessage.value = t('workspace.pdf.renderFailed')
    }
  } finally {
    isRenderingPage.value = false
  }
}

async function loadDocument() {
  currentRenderTask.value?.cancel()
  currentTextLayerBuilder.value?.cancel()
  currentTextLayerBuilder.value = null
  cancelContinuousRender()
  continuousRenderVersion.value += 1
  continuousPageSizes.value = {}
  isRenderingContinuous.value = false
  pdfDocument.value?.destroy()
  pdfDocument.value = null
  currentPdfPage.value = null
  totalPages.value = 0
  renderErrorMessage.value = ''

  if (!props.previewUrl) {
    return
  }

  isLoadingPdf.value = true
  try {
    const doc = await loadPdfFromUrl(props.previewUrl)
    pdfDocument.value = doc
    onDocumentLoad({ numPages: doc.numPages })
    if (readingMode.value === 'continuous') {
      await renderContinuousPages()
    } else {
      await renderCurrentPage()
    }
  } catch (error) {
    renderErrorMessage.value = getLoadErrorMessage(error)
  } finally {
    isLoadingPdf.value = false
  }
}

async function loadPdfFromUrl(url?: string): Promise<PDFDocumentProxy> {
  const normalizedUrl = url?.trim() || ''
  if (!normalizedUrl) {
    throw new Error('empty-url')
  }

  let lastError: unknown = null

  const directLoaders = [
    () => getDocument({
      url: normalizedUrl,
      withCredentials: false,
    }).promise,
    () => getDocument({
      url: normalizedUrl,
      withCredentials: false,
      disableRange: true,
      disableStream: true,
      disableAutoFetch: true,
    }).promise,
  ]

  for (const tryLoad of directLoaders) {
    try {
      return await tryLoad()
    } catch (error) {
      lastError = error
    }
  }

  // Fallback: fetch binary first, then let pdf.js parse ArrayBuffer.
  try {
    const response = await fetch(normalizedUrl, {
      method: 'GET',
      mode: 'cors',
      credentials: 'omit',
    })
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }
    const pdfBytes = await response.arrayBuffer()
    return await getDocument({ data: pdfBytes }).promise
  } catch (error) {
    throw error ?? lastError
  }
}

function getLoadErrorMessage(error: unknown) {
  const rawMessage = error instanceof Error ? error.message : ''
  const message = rawMessage.toLowerCase()

  if (!rawMessage || rawMessage === 'empty-url') {
    return t('workspace.pdf.loadFailedEmpty')
  }

  if (message.includes('cors') || message.includes('access-control') || message.includes('failed to fetch')) {
    return t('workspace.pdf.loadFailedCors')
  }

  if (message.includes('403') || message.includes('401') || message.includes('expired')) {
    return t('workspace.pdf.loadFailedExpired')
  }

  return t('workspace.pdf.loadFailedWithReason', { reason: rawMessage })
}

watch(
  () => props.previewUrl,
  () => {
    void loadDocument()
  },
  { immediate: true },
)

watch(
  () => currentPage.value,
  () => {
    if (!pdfDocument.value) return
    if (autoFitEnabled.value) {
      void fitToPage()
      return
    }
    if (readingMode.value !== 'paged') return
    if (isSyncingCurrentPageFromScroll.value) return
    void renderCurrentPage()
  },
)

watch(
  () => scale.value,
  () => {
    if (!pdfDocument.value) return
    if (readingMode.value === 'continuous') {
      void renderContinuousPages()
      return
    }
    void renderCurrentPage()
  },
)

watch(
  () => readingMode.value,
  async (mode) => {
    if (!pdfDocument.value) return
    if (autoFitEnabled.value) {
      await fitToPage()
    }
    if (mode === 'continuous') {
      await renderContinuousPages()
      await nextTick()
      syncCurrentPageFromContinuousScroll()
      return
    }
    cancelContinuousRender()
    continuousRenderVersion.value += 1
    isRenderingContinuous.value = false
    void renderCurrentPage()
  },
)

function handleWindowResize() {
  if (!autoFitEnabled.value) return
  void fitToPage()
}

watch(
  () => autoFitEnabled.value,
  (enabled) => {
    if (enabled) {
      window.addEventListener('resize', handleWindowResize)
      return
    }
    window.removeEventListener('resize', handleWindowResize)
  },
)

onUnmounted(() => {
  window.removeEventListener('resize', handleWindowResize)
  currentRenderTask.value?.cancel()
  currentTextLayerBuilder.value?.cancel()
  cancelContinuousRender()
  pdfDocument.value?.destroy()
})

defineExpose({})
</script>

<template>
  <section class="paper-panel flex h-full flex-col">
    <header class="flex flex-wrap items-center justify-between gap-3 bg-(--surface-overlay) px-4 py-3">
      <div class="flex items-center gap-3">
        <h3 class="max-w-60 truncate text-base font-semibold text-base-content">{{ file.file_name }}</h3>
        <span class="text-xs text-base-content/50">{{ t('workspace.pdf.pages', { count: totalPages }) }}</span>
        <span v-if="isLoadingPdf || isRenderingPage || isRenderingContinuous" class="text-xs text-info">{{ t('workspace.pdf.rendering') }}</span>
        <span v-if="renderErrorMessage" class="text-xs text-error">{{ renderErrorMessage }}</span>
        <span v-if="copyStatusMessage" class="text-xs text-success">{{ copyStatusMessage }}</span>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <div class="flex items-center gap-1">
          <button type="button" class="btn btn-square btn-xs btn-ghost rounded-sm" :disabled="currentPage <= 1" @click="previousPage">
            <IconChevronLeft class="h-4 w-4" />
          </button>

          <input
            type="number"
            :value="currentPage"
            min="1"
            :max="totalPages"
            class="input input-xs w-14 rounded-sm text-center"
            @change="goToPage(Number(($event.target as HTMLInputElement).value))"
          />
          <span class="text-xs text-base-content/60">/ {{ totalPages }}</span>

          <button type="button" class="btn btn-square btn-xs btn-ghost rounded-sm" :disabled="currentPage >= totalPages" @click="nextPage">
            <IconChevronRight class="h-4 w-4" />
          </button>
        </div>

        <div class="h-4 w-px bg-base-content/15" />

        <div class="flex items-center gap-1">
          <button type="button" class="btn btn-square btn-xs btn-ghost rounded-sm" @click="zoomOut">
            <IconMagnifyMinus class="h-4 w-4" />
          </button>

          <select :value="scale" class="select select-xs w-20 rounded-sm" @change="handleScaleSelectChange(Number(($event.target as HTMLSelectElement).value))">
            <option v-for="level in zoomOptions" :key="level" :value="level">
              {{ Math.round(level * 100) }}%
            </option>
          </select>

          <button type="button" class="btn btn-square btn-xs btn-ghost rounded-sm" @click="zoomIn">
            <IconMagnifyPlus class="h-4 w-4" />
          </button>

          <div class="tooltip tooltip-bottom" :data-tip="t('workspace.pdf.fitPage')">
            <button
              type="button"
              class="btn btn-square btn-xs rounded-sm"
              :class="autoFitEnabled ? 'btn-primary' : 'btn-ghost'"
              @click="fitToPage"
            >
              <IconAspectRatio class="h-4 w-4" />
            </button>
          </div>
        </div>

        <div class="h-4 w-px bg-base-content/15" />

        <div class="join">
          <button
            type="button"
            class="join-item btn btn-xs rounded-sm"
            :class="readingMode === 'paged' ? 'btn-primary' : 'btn-ghost'"
            @click="setReadingMode('paged')"
          >
            {{ t('workspace.pdf.readingModePaged') }}
          </button>
          <button
            type="button"
            class="join-item btn btn-xs rounded-sm"
            :class="readingMode === 'continuous' ? 'btn-primary' : 'btn-ghost'"
            @click="setReadingMode('continuous')"
          >
            {{ t('workspace.pdf.readingModeContinuous') }}
          </button>
        </div>

        <div class="h-4 w-px bg-base-300" />

        <div class="flex items-center gap-1">
          <div class="tooltip tooltip-bottom" :data-tip="t('workspace.pdf.toolHighlight')">
            <button type="button" class="btn btn-square btn-xs rounded-sm" :disabled="readingMode === 'continuous'" :class="activeTool === 'highlight' ? 'btn-primary' : 'btn-ghost'" @click="setTool('highlight')">
              <IconFormatColorHighlight class="h-4 w-4" />
            </button>
          </div>

          <div class="dropdown dropdown-end">
            <button type="button" tabindex="0" class="btn btn-square btn-xs rounded-sm btn-ghost" :disabled="readingMode === 'continuous'" :class="{ 'btn-primary': activeTool === 'highlight' }">
              <IconPalette class="h-4 w-4" />
            </button>
            <ul v-if="activeTool === 'highlight'" tabindex="0" class="dropdown-content z-1 menu p-2 shadow bg-(--surface-raised) rounded-box w-40">
              <li v-for="color in HIGHLIGHT_COLORS" :key="color.hex">
                <a class="flex items-center gap-2" :class="{ 'active': selectedColor === color.hex }" @click="selectedColor = color.hex">
                  <span class="w-4 h-4 rounded" :style="{ backgroundColor: color.hex }" />
                  <span class="capitalize">{{ color.name }}</span>
                </a>
              </li>
            </ul>
          </div>

          <div class="tooltip tooltip-bottom" :data-tip="t('workspace.pdf.toolNote')">
            <button type="button" class="btn btn-square btn-xs rounded-sm" :disabled="readingMode === 'continuous'" :class="activeTool === 'note' ? 'btn-primary' : 'btn-ghost'" @click="setTool('note')">
              <IconNoteOutline class="h-4 w-4" />
            </button>
          </div>

          <div class="tooltip tooltip-bottom" :data-tip="t('workspace.pdf.toolDrawing')">
            <button type="button" class="btn btn-square btn-xs rounded-sm" :disabled="readingMode === 'continuous'" :class="activeTool === 'drawing' ? 'btn-primary' : 'btn-ghost'" @click="setTool('drawing')">
              <IconDraw class="h-4 w-4" />
            </button>
          </div>

          <div v-if="activeTool === 'drawing'" class="flex items-center gap-1 ml-1">
            <input type="color" v-model="selectedDrawingColor" class="w-5 h-5 rounded cursor-pointer" />
            <input type="range" v-model="strokeWidth" min="1" max="10" class="range range-xs w-16" />
          </div>
        </div>

        <div class="h-4 w-px bg-base-content/15" />

        <div class="flex items-center gap-1">
          <div class="tooltip tooltip-bottom" :data-tip="t('workspace.pdf.copyText')">
            <button type="button" class="btn btn-square btn-xs btn-ghost rounded-sm" @click="copySelectedText">
              <IconContentCopy class="h-4 w-4" />
            </button>
          </div>

          <div class="tooltip tooltip-bottom" :data-tip="t('workspace.pdf.printPdf')">
            <button type="button" class="btn btn-square btn-xs btn-ghost rounded-sm" @click="printDocument">
              <IconPrinterOutline class="h-4 w-4" />
            </button>
          </div>
        </div>

        <div class="h-4 w-px bg-base-content/15" />

        <button type="button" class="btn btn-square btn-xs btn-ghost rounded-sm" @click="emit('close')">
          <IconClose class="h-4 w-4" />
        </button>
      </div>
    </header>

    <div class="flex flex-1 overflow-hidden">
      <div ref="containerRef" class="flex-1 overflow-y-auto px-4 py-4 bg-(--surface-overlay)" :class="{ 'cursor-crosshair': readingMode === 'paged' && activeTool !== 'none' }" @wheel="handleReaderWheel" @scroll="handleContainerScroll">
        <div
          v-if="readingMode === 'paged'"
          ref="pageLayerRef"
          class="pdf-page-layer mx-auto relative select-text"
          :style="{ width: `${pageDisplaySize.width}px`, height: `${pageDisplaySize.height}px` }"
          @click="handleContentClick"
          @mousedown="handleContentMouseDown"
          @mousemove="handleContentMouseMove"
          @mouseup="handleContentMouseUp"
        >
          <canvas ref="canvasRef" class="block rounded-sm shadow-sm bg-white" />

          <div
            ref="textLayerRef"
            class="textLayer absolute left-0 top-0"
            :class="{ 'pointer-events-none': activeTool !== 'none' }"
          />

          <svg
            v-if="currentPath.length > 1"
            class="absolute top-0 left-0 pointer-events-none"
            style="width: 100%; height: 100%; overflow: visible;"
          >
            <path
              v-if="activeTool === 'drawing'"
              :d="getSvgPath()"
              fill="none"
              :stroke="selectedDrawingColor"
              :stroke-width="strokeWidth"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
            <rect
              v-if="activeTool === 'highlight' && getHighlightRect()"
              :x="getHighlightRect()!.x"
              :y="getHighlightRect()!.y"
              :width="getHighlightRect()!.width"
              :height="getHighlightRect()!.height"
              :fill="selectedColor + '40'"
              stroke="none"
            />
          </svg>

          <div
            v-for="ann in pageAnnotations"
            :key="ann.id"
            class="absolute pointer-events-auto"
            :style="{
              left: `${getDisplayRect(ann).x}px`,
              top: `${getDisplayRect(ann).y}px`,
              width: ann.type === 'note' ? '24px' : `${getDisplayRect(ann).width}px`,
              height: ann.type === 'note' ? '24px' : `${getDisplayRect(ann).height}px`,
            }"
            @click.stop="handleAnnotationClick(ann)"
          >
            <template v-if="ann.type === 'note'">
              <div class="w-full h-full bg-warning rounded-full flex items-center justify-center text-white font-bold text-sm shadow-md cursor-pointer hover:bg-warning/80">!</div>
            </template>
            <template v-else-if="ann.type === 'highlight'">
              <div class="w-full h-full rounded-sm cursor-pointer hover:opacity-80" :style="{ backgroundColor: (ann.color || '#FFEB3B') + '60' }" />
            </template>
            <template v-else-if="ann.type === 'drawing' && parseDrawingPaths(ann).length > 0">
              <svg class="absolute left-0 top-0 overflow-visible" :style="{ width: `${pageDisplaySize.width}px`, height: `${pageDisplaySize.height}px` }">
                <path
                  v-for="(path, idx) in parseDrawingPaths(ann)"
                  :key="idx"
                  :d="getDrawingPathData(path)"
                  fill="none"
                  :stroke="path.color"
                  :stroke-width="drawingPathStrokeWidth(path)"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
            </template>
          </div>
        </div>

        <div v-else class="mx-auto flex flex-col gap-4">
          <div
            v-for="pageNumber in totalPages"
            :key="pageNumber"
            :ref="setContinuousPageRef(pageNumber)"
            class="relative mx-auto rounded-sm bg-white shadow-sm"
            :style="{
              width: `${getContinuousPageSize(pageNumber).width}px`,
              minHeight: `${getContinuousPageSize(pageNumber).height}px`,
            }"
          >
            <canvas :ref="setContinuousCanvasRef(pageNumber)" class="block rounded-sm bg-white" />
            <div :ref="setContinuousTextLayerRef(pageNumber)" class="textLayer absolute left-0 top-0" />
          </div>
        </div>
      </div>

      <div
        v-if="readingMode === 'paged' && noteInputPosition"
        class="paper-panel absolute z-50 p-3 w-64"
        :style="{ left: `${noteInputPosition.x + 20}px`, top: `${noteInputPosition.y + 20}px` }"
      >
        <textarea
          v-model="noteInputValue"
          class="textarea textarea-sm w-full rounded-sm"
          :placeholder="t('workspace.pdf.notePlaceholder')"
          rows="3"
          autofocus
          @keydown.enter.ctrl="handleNoteSubmit"
          @keydown.esc="noteInputPosition = null"
        />
        <div class="flex justify-end gap-2 mt-2">
          <button type="button" class="btn btn-xs btn-ghost rounded-sm" @click="noteInputPosition = null">
            <IconCloseCircleOutline class="h-4 w-4" />
          </button>
          <button type="button" class="btn btn-xs btn-primary rounded-sm" @click="handleNoteSubmit">
            <IconCheck class="h-4 w-4" />
          </button>
        </div>
      </div>

      <aside class="w-72 bg-(--surface-overlay) flex flex-col overflow-hidden">
        <div class="bg-(--surface-sunken) px-3 py-2">
          <h4 class="text-sm font-semibold text-base-content">{{ t('workspace.pdf.annotations') }}</h4>
          <div class="mt-2 flex gap-1 flex-wrap">
            <button
              v-for="filter in annotationFilterOptions"
              :key="filter"
              type="button"
              class="btn btn-xs rounded-sm"
              :class="annotationFilter === filter ? 'btn-primary' : 'btn-ghost'"
              @click="annotationFilter = filter"
            >
              {{ annotationFilterLabel(filter) }}
            </button>
          </div>
        </div>

        <div class="flex-1 overflow-y-auto px-3 py-2">
          <div v-if="filteredAnnotations.length === 0" class="text-xs text-base-content/50 text-center py-4">
            {{ t('workspace.pdf.noAnnotations') }}
          </div>

          <div v-else class="space-y-2">
            <div
              v-for="ann in filteredAnnotations"
              :key="ann.id"
              class="rounded-sm no-line bg-(--surface-raised) p-2 cursor-pointer hover:bg-(--surface-overlay)"
              :class="{ 'bg-primary/5 border-primary': editingAnnotationId === ann.id }"
              @click="handleAnnotationClick(ann)"
            >
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <span v-if="ann.type === 'highlight'" class="w-3 h-3 rounded-sm" :style="{ backgroundColor: ann.color || '#FFEB3B' }" />
                  <IconNoteOutline v-else-if="ann.type === 'note'" class="h-3 w-3 text-warning" />
                  <IconDraw v-else-if="ann.type === 'drawing'" class="h-3 w-3 text-info" />
                  <span class="text-xs text-base-content/70">{{ t('workspace.pdf.pageWithNumber', { page: ann.page }) }}</span>
                </div>
                <button type="button" class="btn btn-ghost btn-xs btn-square rounded-sm" @click.stop="handleAnnotationDelete(ann.id)">
                  <IconDeleteOutline class="h-3 w-3 text-error" />
                </button>
              </div>

              <div v-if="ann.type === 'note' || ann.type === 'highlight'" class="mt-1 text-xs text-base-content/80 line-clamp-2">
                <template v-if="editingAnnotationId === ann.id">
                  <textarea v-model="editingContent" class="textarea textarea-xs w-full rounded-sm" rows="2" @click.stop />
                  <div class="flex justify-end gap-1 mt-1">
                    <button type="button" class="btn btn-xs btn-ghost rounded-sm" @click.stop="cancelAnnotationEdit">{{ t('workspace.pdf.cancel') }}</button>
                    <button type="button" class="btn btn-xs btn-primary rounded-sm" @click.stop="saveAnnotationEdit(ann)">{{ t('workspace.pdf.save') }}</button>
                  </div>
                </template>
                <template v-else>
                  {{ ann.content || (ann.type === 'highlight' ? t('workspace.pdf.textHighlight') : '') }}
                </template>
              </div>

              <div v-if="ann.type === 'drawing'" class="mt-1 text-xs text-base-content/70">{{ t('workspace.pdf.drawingLabel') }}</div>
            </div>
          </div>
        </div>
      </aside>
    </div>
  </section>
</template>

<style scoped>
.pdf-page-layer {
  line-height: 1;
}
</style>



