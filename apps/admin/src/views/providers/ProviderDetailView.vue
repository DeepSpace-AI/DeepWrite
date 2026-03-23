<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  deleteAIProvider,
  createModelByVendor,
  discoverVendorModels,
  getVendorDetail,
  updateAIProvider,
  type AIProviderModel,
  type AIProviderVendor,
  type CreateModelByVendorInput,
  type DiscoveredProviderModel,
} from '@/api/aiProvider'
import { ApiError } from '@/api/http'

interface DiscoverSelection extends Required<CreateModelByVendorInput> {
  selected: boolean
}

const route = useRoute()
const router = useRouter()

const isLoading = ref(false)
const isDiscovering = ref(false)
const isSavingModels = ref(false)
const isSavingManualModel = ref(false)
const isSavingPersistedModel = ref(false)
const togglingPersistedModelName = ref('')
const deletingPersistedModelName = ref('')
const errorMessage = ref('')
const successMessage = ref('')
const detailProvider = ref<AIProviderVendor | null>(null)
const persistedModels = ref<AIProviderModel[]>([])
const discoveredModels = ref<DiscoveredProviderModel[]>([])
const discoverState = ref<Record<string, DiscoverSelection>>({})
const isModelCompareModalOpen = ref(false)
const isManualModelModalOpen = ref(false)
const isPersistedEditModalOpen = ref(false)
const modelSearchKeyword = ref('')
const modelPage = ref(1)
const modelPageSize = ref(20)
const persistSearchKeyword = ref('')
const editingPersistedModelName = ref('')
const manualModelForm = reactive<Required<CreateModelByVendorInput>>({
  model: '',
  request_model: '',
  supports_chat_completions: true,
  supports_chat_responses: true,
  supports_embeddings: true,
  supports_rerank: false,
  supports_audio_speech: false,
  supports_audio_transcriptions: false,
  supports_models: true,
  enabled: true,
})
const persistedModelForm = reactive<Required<CreateModelByVendorInput>>({
  model: '',
  request_model: '',
  supports_chat_completions: true,
  supports_chat_responses: true,
  supports_embeddings: true,
  supports_rerank: false,
  supports_audio_speech: false,
  supports_audio_transcriptions: false,
  supports_models: true,
  enabled: true,
})

const providerId = computed(() => String(route.params.providerId || '').trim())
const selectedDiscoverCount = computed(() => Object.values(discoverState.value).filter(item => item.selected).length)
const persistedModelMap = computed(() => {
  const map = new Map<string, AIProviderModel>()
  persistedModels.value.forEach((item) => map.set(item.model, item))
  return map
})
const filteredDiscoveredModels = computed(() => {
  const keyword = modelSearchKeyword.value.trim().toLowerCase()
  if (!keyword) return discoveredModels.value
  return discoveredModels.value.filter(item =>
    item.model.toLowerCase().includes(keyword)
    || item.name?.toLowerCase().includes(keyword)
    || item.owned_by?.toLowerCase().includes(keyword),
  )
})
const totalModelPages = computed(() => {
  const total = filteredDiscoveredModels.value.length
  return Math.max(1, Math.ceil(total / modelPageSize.value))
})
const pagedDiscoveredModels = computed(() => {
  const start = (modelPage.value - 1) * modelPageSize.value
  return filteredDiscoveredModels.value.slice(start, start + modelPageSize.value)
})
const selectedOnPageCount = computed(() => {
  return pagedDiscoveredModels.value.filter(item => discoverState.value[item.model]?.selected).length
})
const canSaveSelected = computed(() => selectedDiscoverCount.value > 0 && !isSavingModels.value)
const persistedEnabledCount = computed(() => persistedModels.value.filter(item => item.enabled).length)
const filteredPersistedModels = computed(() => {
  const keyword = persistSearchKeyword.value.trim().toLowerCase()
  if (!keyword) return persistedModels.value
  return persistedModels.value.filter(item => {
    return item.model.toLowerCase().includes(keyword)
      || item.request_model.toLowerCase().includes(keyword)
      || item.provider.toLowerCase().includes(keyword)
  })
})

function resetManualModelForm() {
  manualModelForm.model = ''
  manualModelForm.request_model = ''
  manualModelForm.supports_chat_completions = true
  manualModelForm.supports_chat_responses = true
  manualModelForm.supports_embeddings = true
  manualModelForm.supports_rerank = false
  manualModelForm.supports_audio_speech = false
  manualModelForm.supports_audio_transcriptions = false
  manualModelForm.supports_models = true
  manualModelForm.enabled = true
}

function resetPersistedModelForm() {
  persistedModelForm.model = ''
  persistedModelForm.request_model = ''
  persistedModelForm.supports_chat_completions = true
  persistedModelForm.supports_chat_responses = true
  persistedModelForm.supports_embeddings = true
  persistedModelForm.supports_rerank = false
  persistedModelForm.supports_audio_speech = false
  persistedModelForm.supports_audio_transcriptions = false
  persistedModelForm.supports_models = true
  persistedModelForm.enabled = true
  editingPersistedModelName.value = ''
}

function closeManualModelModal() {
  isManualModelModalOpen.value = false
  resetManualModelForm()
}

function closeModelCompareModal() {
  isModelCompareModalOpen.value = false
}

function openPersistedEditModal(row: AIProviderModel) {
  editingPersistedModelName.value = row.model
  persistedModelForm.model = row.model
  persistedModelForm.request_model = row.request_model
  persistedModelForm.supports_chat_completions = row.supports_chat_completions
  persistedModelForm.supports_chat_responses = row.supports_chat_responses
  persistedModelForm.supports_embeddings = row.supports_embeddings
  persistedModelForm.supports_rerank = row.supports_rerank
  persistedModelForm.supports_audio_speech = row.supports_audio_speech
  persistedModelForm.supports_audio_transcriptions = row.supports_audio_transcriptions
  persistedModelForm.supports_models = row.supports_models
  persistedModelForm.enabled = row.enabled
  isPersistedEditModalOpen.value = true
}

function closePersistedEditModal() {
  isPersistedEditModalOpen.value = false
  resetPersistedModelForm()
}

function toggleSelectCurrentPage(selected: boolean) {
  pagedDiscoveredModels.value.forEach((item) => {
    const state = discoverState.value[item.model]
    if (state) {
      state.selected = selected
    }
  })
}

function selectOnlyNotPersisted() {
  Object.entries(discoverState.value).forEach(([model, state]) => {
    state.selected = !persistedModelMap.value.has(model)
  })
}

function applyCapabilityPreset(preset: 'chat' | 'embedding' | 'all') {
  const selectedItems = Object.values(discoverState.value).filter(item => item.selected)
  if (!selectedItems.length) return
  selectedItems.forEach((item) => {
    if (preset === 'chat') {
      item.supports_chat_completions = true
      item.supports_chat_responses = true
      item.supports_embeddings = false
      item.supports_rerank = false
      item.supports_audio_speech = false
      item.supports_audio_transcriptions = false
      item.supports_models = true
      item.enabled = true
      return
    }
    if (preset === 'embedding') {
      item.supports_chat_completions = false
      item.supports_chat_responses = false
      item.supports_embeddings = true
      item.supports_rerank = false
      item.supports_audio_speech = false
      item.supports_audio_transcriptions = false
      item.supports_models = true
      item.enabled = true
      return
    }
    item.supports_chat_completions = true
    item.supports_chat_responses = true
    item.supports_embeddings = true
    item.supports_rerank = true
    item.supports_audio_speech = true
    item.supports_audio_transcriptions = true
    item.supports_models = true
    item.enabled = true
  })
}

async function loadVendorDetail() {
  if (!providerId.value) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    const detail = await getVendorDetail(providerId.value)
    detailProvider.value = detail.provider
    persistedModels.value = detail.models
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '加载厂商详情失败'
  } finally {
    isLoading.value = false
  }
}

async function discoverModels() {
  if (!providerId.value) return
  isDiscovering.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const res = await discoverVendorModels(providerId.value)
    discoveredModels.value = res.items
    const next: Record<string, DiscoverSelection> = {}
    discoveredModels.value.forEach((item) => {
      next[item.model] = {
        selected: false,
        model: item.model,
        request_model: item.model,
        supports_chat_completions: true,
        supports_chat_responses: true,
        supports_embeddings: true,
        supports_rerank: false,
        supports_audio_speech: false,
        supports_audio_transcriptions: false,
        supports_models: true,
        enabled: true,
      }
    })
    discoverState.value = next
    modelPage.value = 1
    successMessage.value = `已拉取 ${res.items.length} 个模型，请选择并配置后落库`
    isModelCompareModalOpen.value = true
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '拉取模型失败'
  } finally {
    isDiscovering.value = false
  }
}

function capabilityLabelsFromSelection(selection: DiscoverSelection) {
  const labels: string[] = []
  if (selection.supports_chat_completions) labels.push('chat')
  if (selection.supports_chat_responses) labels.push('responses')
  if (selection.supports_embeddings) labels.push('embeddings')
  if (selection.supports_rerank) labels.push('rerank')
  if (selection.supports_audio_speech) labels.push('speech')
  if (selection.supports_audio_transcriptions) labels.push('transcribe')
  if (selection.supports_models) labels.push('models')
  if (selection.enabled) labels.push('enabled')
  return labels
}

function capabilityLabelsFromPersisted(model: AIProviderModel) {
  const labels: string[] = []
  if (model.supports_chat_completions) labels.push('chat')
  if (model.supports_chat_responses) labels.push('responses')
  if (model.supports_embeddings) labels.push('embeddings')
  if (model.supports_rerank) labels.push('rerank')
  if (model.supports_audio_speech) labels.push('speech')
  if (model.supports_audio_transcriptions) labels.push('transcribe')
  if (model.supports_models) labels.push('models')
  if (model.enabled) labels.push('enabled')
  return labels
}

function compareStatus(modelName: string) {
  const local = persistedModelMap.value.get(modelName)
  if (!local) return '未落库'
  const selection = discoverState.value[modelName]
  if (!selection) return '已落库'
  const same = local.request_model === selection.request_model
    && local.supports_chat_completions === selection.supports_chat_completions
    && local.supports_chat_responses === selection.supports_chat_responses
    && local.supports_embeddings === selection.supports_embeddings
    && local.supports_rerank === selection.supports_rerank
    && local.supports_audio_speech === selection.supports_audio_speech
    && local.supports_audio_transcriptions === selection.supports_audio_transcriptions
    && local.supports_models === selection.supports_models
    && local.enabled === selection.enabled
  return same ? '配置一致' : '存在差异'
}

async function saveSelectedModels() {
  if (!providerId.value) return
  const selectedItems = Object.values(discoverState.value).filter(item => item.selected)
  if (!selectedItems.length) return
  isSavingModels.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const results = await Promise.allSettled(selectedItems.map(item =>
      createModelByVendor(providerId.value, {
        model: item.model,
        request_model: item.request_model.trim(),
        supports_chat_completions: item.supports_chat_completions,
        supports_chat_responses: item.supports_chat_responses,
        supports_embeddings: item.supports_embeddings,
        supports_rerank: item.supports_rerank,
        supports_audio_speech: item.supports_audio_speech,
        supports_audio_transcriptions: item.supports_audio_transcriptions,
        supports_models: item.supports_models,
        enabled: item.enabled,
      }),
    ))
    const successCount = results.filter(item => item.status === 'fulfilled').length
    const failed = results.filter(item => item.status === 'rejected')
    if (successCount > 0) {
      successMessage.value = failed.length
        ? `已落库 ${successCount} 个模型，${failed.length} 个失败`
        : `已落库 ${successCount} 个模型`
    }
    if (!successCount && failed.length) {
      const firstFailed = failed[0]
      const firstReason = firstFailed?.reason
      errorMessage.value = firstReason instanceof ApiError ? firstReason.message : '模型落库失败'
    }
    await loadVendorDetail()
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '模型落库失败'
  } finally {
    isSavingModels.value = false
  }
}

async function saveManualModel() {
  if (!providerId.value) return
  const model = manualModelForm.model.trim()
  if (!model) {
    errorMessage.value = '请先填写模型名称'
    return
  }
  isSavingManualModel.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    await createModelByVendor(providerId.value, {
      model,
      request_model: (manualModelForm.request_model || model).trim(),
      supports_chat_completions: manualModelForm.supports_chat_completions,
      supports_chat_responses: manualModelForm.supports_chat_responses,
      supports_embeddings: manualModelForm.supports_embeddings,
      supports_rerank: manualModelForm.supports_rerank,
      supports_audio_speech: manualModelForm.supports_audio_speech,
      supports_audio_transcriptions: manualModelForm.supports_audio_transcriptions,
      supports_models: manualModelForm.supports_models,
      enabled: manualModelForm.enabled,
    })
    successMessage.value = `模型 ${model} 已手动落库`
    closeManualModelModal()
    await loadVendorDetail()
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '手动新增模型失败'
  } finally {
    isSavingManualModel.value = false
  }
}

async function savePersistedModelEdit() {
  const model = editingPersistedModelName.value.trim()
  if (!model) return
  isSavingPersistedModel.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    await updateAIProvider(model, {
      request_model: persistedModelForm.request_model.trim() || model,
      supports_chat_completions: persistedModelForm.supports_chat_completions,
      supports_chat_responses: persistedModelForm.supports_chat_responses,
      supports_embeddings: persistedModelForm.supports_embeddings,
      supports_rerank: persistedModelForm.supports_rerank,
      supports_audio_speech: persistedModelForm.supports_audio_speech,
      supports_audio_transcriptions: persistedModelForm.supports_audio_transcriptions,
      supports_models: persistedModelForm.supports_models,
      enabled: persistedModelForm.enabled,
    })
    successMessage.value = `模型 ${model} 配置已更新`
    closePersistedEditModal()
    await loadVendorDetail()
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '更新模型配置失败'
  } finally {
    isSavingPersistedModel.value = false
  }
}

async function togglePersistedModelEnabled(row: AIProviderModel) {
  togglingPersistedModelName.value = row.model
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const nextEnabled = !row.enabled
    await updateAIProvider(row.model, { enabled: nextEnabled })
    successMessage.value = nextEnabled
      ? `模型 ${row.model} 已启用`
      : `模型 ${row.model} 已停用`
    await loadVendorDetail()
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '更新模型状态失败'
  } finally {
    togglingPersistedModelName.value = ''
  }
}

async function deletePersistedModel(row: AIProviderModel) {
  if (!window.confirm(`确认删除模型 ${row.model} 吗？`)) {
    return
  }
  deletingPersistedModelName.value = row.model
  errorMessage.value = ''
  successMessage.value = ''
  try {
    await deleteAIProvider(row.model)
    successMessage.value = `模型 ${row.model} 已删除`
    await loadVendorDetail()
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '删除模型失败'
  } finally {
    deletingPersistedModelName.value = ''
  }
}

watch(providerId, () => {
  discoverState.value = {}
  discoveredModels.value = []
  modelSearchKeyword.value = ''
  persistSearchKeyword.value = ''
  modelPage.value = 1
  isModelCompareModalOpen.value = false
  isManualModelModalOpen.value = false
  isPersistedEditModalOpen.value = false
  resetManualModelForm()
  resetPersistedModelForm()
  loadVendorDetail()
})

watch(modelSearchKeyword, () => {
  modelPage.value = 1
})

watch([filteredDiscoveredModels, modelPageSize], () => {
  if (modelPage.value > totalModelPages.value) {
    modelPage.value = totalModelPages.value
  }
})

onMounted(() => {
  loadVendorDetail()
})
</script>

<template>
  <section class="space-y-4">
    <div class="glass-card rounded-2xl p-5">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 class="text-xl font-semibold text-pretty">Provider 详情</h2>
          <p class="mt-1 text-sm text-pretty-secondary">在详情页拉取厂商模型列表，选择并配置后落库</p>
        </div>
        <button class="btn btn-ghost btn-sm rounded-xl" @click="router.push({ name: 'providers' })">返回列表</button>
      </div>
    </div>

    <div v-if="errorMessage" class="glass-card rounded-xl p-4 text-sm text-error border border-error/20 bg-error/5">{{ errorMessage }}</div>
    <div v-if="successMessage" class="glass-card rounded-xl p-4 text-sm text-emerald-600 border border-emerald-500/20 bg-emerald-500/5">{{ successMessage }}</div>

    <fieldset class="glass-card rounded-2xl p-5">
      <legend class="px-2 text-pretty-secondary font-medium">厂商详情</legend>
      <div v-if="isLoading" class="py-8 text-center"><span class="loading loading-spinner loading-md text-pretty-muted" /></div>
      <div v-else-if="detailProvider">
        <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
          <div class="glass-card rounded-xl p-3">
            <p class="text-[10px] uppercase tracking-wider text-pretty-muted">Provider</p>
            <p class="mt-1 text-sm font-semibold text-pretty">{{ detailProvider.name }}</p>
          </div>
          <div class="glass-card rounded-xl p-3">
            <p class="text-[10px] uppercase tracking-wider text-pretty-muted">已落库模型</p>
            <p class="mt-1 text-2xl font-semibold text-pretty">{{ persistedModels.length }}</p>
          </div>
          <div class="glass-card rounded-xl p-3 border border-emerald-500/20">
            <p class="text-[10px] uppercase tracking-wider text-emerald-600/80">启用模型</p>
            <p class="mt-1 text-2xl font-semibold text-emerald-600">{{ persistedEnabledCount }}</p>
          </div>
          <div class="glass-card rounded-xl p-3 border border-blue-500/20">
            <p class="text-[10px] uppercase tracking-wider text-blue-600/80">待选择模型</p>
            <p class="mt-1 text-2xl font-semibold text-blue-600">{{ selectedDiscoverCount }}</p>
          </div>
        </div>
        <div class="mt-4 space-y-1">
          <p class="text-sm text-pretty-secondary"><span class="font-medium text-pretty">Base URL：</span><span class="font-mono text-xs">{{ detailProvider.base_url }}</span></p>
          <p class="text-sm text-pretty-secondary"><span class="font-medium text-pretty">Models Path：</span><span class="font-mono text-xs">{{ detailProvider.models_path }}</span></p>
        </div>
        <div class="mt-4 flex gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl" :disabled="isDiscovering" @click="discoverModels">
            <span v-if="isDiscovering" class="loading loading-spinner loading-xs" />
            <span>{{ isDiscovering ? '拉取中...' : '拉取厂商模型列表' }}</span>
          </button>
        </div>
      </div>
      <p v-else class="text-sm text-pretty-muted">未找到该厂商信息</p>
    </fieldset>

    <fieldset class="glass-card rounded-2xl p-5">
      <legend class="px-2 text-pretty-secondary font-medium">模型选择与配置</legend>
      <div class="mb-4 flex flex-wrap items-center justify-between gap-2">
        <p class="text-sm text-pretty-secondary">已拉取 {{ discoveredModels.length }} 个模型，已选择 {{ selectedDiscoverCount }} 个</p>
        <div class="flex flex-wrap items-center gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl" @click="isManualModelModalOpen = true">手动新增模型</button>
          <button class="btn btn-ghost btn-sm rounded-xl" :disabled="!discoveredModels.length" @click="isModelCompareModalOpen = true">打开模型对比弹窗</button>
          <button class="btn rounded-xl bg-[var(--glow-primary)] text-pretty hover:opacity-90 btn-sm" :disabled="!canSaveSelected" @click="saveSelectedModels">
            <span v-if="isSavingModels" class="loading loading-spinner loading-xs" />
            <span>{{ isSavingModels ? '落库中...' : '落库选中模型' }}</span>
          </button>
        </div>
      </div>
      <div v-if="!discoveredModels.length" class="text-sm text-pretty-muted py-4">请先在上方点击"拉取厂商模型列表"</div>
    </fieldset>

    <dialog :open="isManualModelModalOpen" class="modal">
      <div class="modal-box w-11/12 max-w-3xl rounded-2xl bg-[var(--bg-elevated)]">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2" @click="closeManualModelModal">✕</button>
        </form>
        <h3 class="text-lg font-semibold text-pretty">手动新增模型</h3>

        <div class="mt-6 grid gap-5">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="form-control w-full">
              <label class="label"><span class="label-text font-medium text-pretty">Model ID</span><span class="label-text-alt text-error">*</span></label>
              <input v-model="manualModelForm.model" type="text" placeholder="e.g. gpt-4o-mini" class="glass-input input input-bordered w-full rounded-xl" :class="{'input-error': !manualModelForm.model}" />
            </div>
            <div class="form-control w-full">
              <label class="label"><span class="label-text font-medium text-pretty">Request Model</span><span class="label-text-alt text-pretty-muted">Optional</span></label>
              <input v-model="manualModelForm.request_model" type="text" placeholder="Defaults to Model ID if empty" class="glass-input input input-bordered w-full rounded-xl" />
            </div>
          </div>

          <div class="glass-card rounded-xl p-4">
            <h4 class="text-sm font-semibold text-pretty-secondary mb-3 uppercase tracking-wider">Capabilities & Status</h4>
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-x-4 gap-y-3">
              <div class="form-control">
                <label class="label cursor-pointer justify-start gap-3 p-0">
                  <input v-model="manualModelForm.supports_chat_completions" type="checkbox" class="checkbox checkbox-sm checkbox-primary" />
                  <span class="label-text text-sm text-pretty-secondary">Chat</span>
                </label>
              </div>
              <div class="form-control">
                <label class="label cursor-pointer justify-start gap-3 p-0">
                  <input v-model="manualModelForm.supports_chat_responses" type="checkbox" class="checkbox checkbox-sm checkbox-primary" />
                  <span class="label-text text-sm text-pretty-secondary">Responses</span>
                </label>
              </div>
              <div class="form-control">
                <label class="label cursor-pointer justify-start gap-3 p-0">
                  <input v-model="manualModelForm.supports_embeddings" type="checkbox" class="checkbox checkbox-sm checkbox-primary" />
                  <span class="label-text text-sm text-pretty-secondary">Embeddings</span>
                </label>
              </div>
              <div class="form-control">
                <label class="label cursor-pointer justify-start gap-3 p-0">
                  <input v-model="manualModelForm.supports_rerank" type="checkbox" class="checkbox checkbox-sm checkbox-primary" />
                  <span class="label-text text-sm text-pretty-secondary">Rerank</span>
                </label>
              </div>
              <div class="form-control">
                <label class="label cursor-pointer justify-start gap-3 p-0">
                  <input v-model="manualModelForm.supports_audio_speech" type="checkbox" class="checkbox checkbox-sm checkbox-primary" />
                  <span class="label-text text-sm text-pretty-secondary">Speech</span>
                </label>
              </div>
              <div class="form-control">
                <label class="label cursor-pointer justify-start gap-3 p-0">
                  <input v-model="manualModelForm.supports_audio_transcriptions" type="checkbox" class="checkbox checkbox-sm checkbox-primary" />
                  <span class="label-text text-sm text-pretty-secondary">Transcribe</span>
                </label>
              </div>
              <div class="form-control">
                <label class="label cursor-pointer justify-start gap-3 p-0">
                  <input v-model="manualModelForm.supports_models" type="checkbox" class="checkbox checkbox-sm checkbox-primary" />
                  <span class="label-text text-sm text-pretty-secondary">Models</span>
                </label>
              </div>
              <div class="divider col-span-full my-0 opacity-10"></div>
              <div class="form-control col-span-full sm:col-span-2">
                <label class="label cursor-pointer justify-start gap-3 p-0">
                  <input v-model="manualModelForm.enabled" type="checkbox" class="toggle toggle-success toggle-sm" />
                  <span class="label-text font-medium text-pretty">Enable Model</span>
                </label>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-action mt-8">
          <button class="btn btn-ghost rounded-xl" @click="closeManualModelModal">取消</button>
          <button class="btn rounded-xl bg-[var(--glow-primary)] text-pretty hover:opacity-90" :disabled="isSavingManualModel || !manualModelForm.model.trim()" @click="saveManualModel">
            <span v-if="isSavingManualModel" class="loading loading-spinner loading-xs" />
            <span>{{ isSavingManualModel ? '正在保存...' : '确认新增' }}</span>
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button @click="closeManualModelModal">close</button>
      </form>
    </dialog>

    <dialog :open="isModelCompareModalOpen" class="modal">
      <div class="modal-box flex h-[86vh] w-11/12 max-w-384 flex-col rounded-2xl bg-[var(--bg-elevated)] p-0">
        <div class="border-b border-[var(--border-subtle)] px-6 pb-4 pt-5">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <h3 class="text-lg font-semibold text-pretty">厂商模型对比与落库</h3>
            <div class="flex flex-wrap items-center gap-2">
              <button class="btn btn-ghost btn-sm rounded-xl" @click="toggleSelectCurrentPage(true)">选择本页</button>
              <button class="btn btn-ghost btn-sm rounded-xl" @click="toggleSelectCurrentPage(false)">取消本页</button>
              <button class="btn btn-ghost btn-sm rounded-xl" @click="selectOnlyNotPersisted">仅选未落库</button>
              <button class="btn btn-ghost btn-sm rounded-xl" :disabled="!selectedDiscoverCount" @click="applyCapabilityPreset('chat')">预设 Chat</button>
              <button class="btn btn-ghost btn-sm rounded-xl" :disabled="!selectedDiscoverCount" @click="applyCapabilityPreset('embedding')">预设 Embedding</button>
              <button class="btn btn-ghost btn-sm rounded-xl" :disabled="!selectedDiscoverCount" @click="applyCapabilityPreset('all')">预设全功能</button>
              <button class="btn rounded-xl bg-[var(--glow-primary)] text-pretty hover:opacity-90 btn-sm" :disabled="!canSaveSelected" @click="saveSelectedModels">
                <span v-if="isSavingModels" class="loading loading-spinner loading-xs" />
                <span>{{ isSavingModels ? '落库中...' : '落库选中模型' }}</span>
              </button>
            </div>
          </div>
          <p class="mt-2 text-sm text-pretty-muted">
            总计 {{ discoveredModels.length }} 个模型，筛选后 {{ filteredDiscoveredModels.length }} 个，当前页已选 {{ selectedOnPageCount }} 个，合计已选 {{ selectedDiscoverCount }} 个
          </p>
        </div>
        <div class="px-6 py-4">
          <fieldset class="glass-card rounded-xl p-3">
            <legend class="px-2 text-pretty-secondary text-sm">搜索</legend>
            <div class="max-w-md">
              <input
                v-model="modelSearchKeyword"
                class="glass-input input input-bordered input-sm w-full rounded-xl"
                placeholder="搜索 model/name/owned_by"
                autocomplete="off"
                autocapitalize="off"
                autocorrect="off"
                spellcheck="false"
              />
              <p class="label"><span class="label-text-alt text-pretty-muted">支持模型标识、名称、归属方检索</span></p>
            </div>
          </fieldset>
        </div>
        <div class="min-h-0 flex-1 px-6 pb-4">
          <div class="h-full overflow-auto rounded-xl border border-[var(--border-subtle)]">
            <table class="table table-sm">
              <thead>
                <tr class="border-b border-[var(--border-subtle)] bg-[var(--bg-base)]/50">
                  <th class="text-pretty-secondary font-medium">选择</th>
                  <th class="text-pretty-secondary font-medium">Model 名称</th>
                  <th class="text-pretty-secondary font-medium">功能</th>
                  <th class="text-pretty-secondary font-medium">本地落库</th>
                  <th class="text-pretty-secondary font-medium">落库对比</th>
                  <th class="text-pretty-secondary font-medium">Request Model</th>
                  <th class="text-pretty-secondary font-medium">待落库配置</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="!pagedDiscoveredModels.length">
                  <td colspan="7" class="py-12 text-center text-sm text-pretty-muted">当前筛选下无模型数据</td>
                </tr>
                <tr v-for="item in pagedDiscoveredModels" :key="item.model" class="border-b border-[var(--border-subtle)] hover:bg-[var(--bg-elevated)]/30 transition-colors">
                  <td><input v-model="discoverState[item.model]!.selected" type="checkbox" class="checkbox checkbox-sm" /></td>
                  <td>
                    <p class="font-mono text-xs text-pretty">{{ item.model }}</p>
                    <p class="text-[11px] text-pretty-muted">{{ item.name || '-' }} / {{ item.owned_by || '-' }}</p>
                  </td>
                  <td class="text-[11px] text-pretty-secondary">{{ item.name || 'unknown' }}</td>
                  <td>
                    <span v-if="persistedModelMap.has(item.model)" class="badge rounded-xl bg-emerald-500/15 text-emerald-600 border-0 text-xs">已落库</span>
                    <span v-else class="badge rounded-xl bg-[var(--glow-primary)] text-pretty-muted border-0 text-xs">未落库</span>
                  </td>
                  <td>
                    <span
                      class="badge rounded-xl text-xs"
                      :class="compareStatus(item.model) === '配置一致'
                        ? 'bg-emerald-500/15 text-emerald-600 border-0'
                        : compareStatus(item.model) === '存在差异'
                          ? 'bg-amber-500/15 text-amber-600 border-0'
                          : 'bg-[var(--glow-primary)] text-pretty-muted border-0'"
                    >
                      {{ compareStatus(item.model) }}
                    </span>
                    <p v-if="persistedModelMap.get(item.model)" class="mt-1 text-[11px] text-pretty-muted">
                      本地: {{ capabilityLabelsFromPersisted(persistedModelMap.get(item.model)!).join(', ') || '-' }}
                    </p>
                  </td>
                  <td><input v-model="discoverState[item.model]!.request_model" class="glass-input input input-bordered input-xs w-44 rounded-xl" /></td>
                  <td>
                    <div class="grid grid-cols-2 gap-1 text-xs">
                      <label class="label cursor-pointer justify-start gap-1 p-0"><input v-model="discoverState[item.model]!.supports_chat_completions" type="checkbox" class="toggle toggle-xs" />chat</label>
                      <label class="label cursor-pointer justify-start gap-1 p-0"><input v-model="discoverState[item.model]!.supports_chat_responses" type="checkbox" class="toggle toggle-xs" />responses</label>
                      <label class="label cursor-pointer justify-start gap-1 p-0"><input v-model="discoverState[item.model]!.supports_embeddings" type="checkbox" class="toggle toggle-xs" />embeddings</label>
                      <label class="label cursor-pointer justify-start gap-1 p-0"><input v-model="discoverState[item.model]!.supports_rerank" type="checkbox" class="toggle toggle-xs" />rerank</label>
                      <label class="label cursor-pointer justify-start gap-1 p-0"><input v-model="discoverState[item.model]!.supports_audio_speech" type="checkbox" class="toggle toggle-xs" />speech</label>
                      <label class="label cursor-pointer justify-start gap-1 p-0"><input v-model="discoverState[item.model]!.supports_audio_transcriptions" type="checkbox" class="toggle toggle-xs" />transcribe</label>
                      <label class="label cursor-pointer justify-start gap-1 p-0"><input v-model="discoverState[item.model]!.supports_models" type="checkbox" class="toggle toggle-xs" />models</label>
                      <label class="label cursor-pointer justify-start gap-1 p-0"><input v-model="discoverState[item.model]!.enabled" type="checkbox" class="toggle toggle-xs" />enabled</label>
                    </div>
                    <p class="mt-1 text-[11px] text-pretty-muted">
                      目标: {{ capabilityLabelsFromSelection(discoverState[item.model]!).join(', ') || '-' }}
                    </p>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
        <div class="border-t border-[var(--border-subtle)] px-6 py-4">
          <div class="flex flex-wrap items-end justify-between gap-3">
            <div class="flex items-end gap-2">
              <div class="w-24">
                <label class="label"><span class="label-text text-pretty-secondary text-xs">每页</span></label>
                <select v-model.number="modelPageSize" class="glass-input select select-bordered select-sm w-full rounded-xl">
                  <option :value="10">10</option>
                  <option :value="20">20</option>
                  <option :value="50">50</option>
                  <option :value="100">100</option>
                </select>
              </div>
              <div>
                <label class="label"><span class="label-text text-pretty-secondary text-xs">页码</span></label>
                <div class="glass-input input input-bordered input-sm flex h-9 items-center px-3 text-xs rounded-xl text-pretty">
                  {{ modelPage }} / {{ totalModelPages }}
                </div>
              </div>
              <button class="btn btn-ghost btn-sm rounded-xl" :disabled="modelPage <= 1" @click="modelPage--">上一页</button>
              <button class="btn btn-ghost btn-sm rounded-xl" :disabled="modelPage >= totalModelPages" @click="modelPage++">下一页</button>
            </div>
            <div class="modal-action m-0">
              <button class="btn btn-ghost rounded-xl" @click="closeModelCompareModal">关闭</button>
              <button class="btn rounded-xl bg-[var(--glow-primary)] text-pretty hover:opacity-90" :disabled="!canSaveSelected" @click="saveSelectedModels">
                <span v-if="isSavingModels" class="loading loading-spinner loading-xs" />
                <span>{{ isSavingModels ? '落库中...' : '落库选中模型' }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button @click="closeModelCompareModal">close</button>
      </form>
    </dialog>

    <fieldset class="glass-card rounded-2xl p-5">
      <legend class="px-2 text-pretty-secondary font-medium">已落库模型</legend>
      <div class="mb-4 max-w-sm">
        <input
          v-model="persistSearchKeyword"
          class="glass-input input input-bordered input-sm w-full rounded-xl"
          placeholder="搜索已落库模型 model/request_model"
          autocomplete="off"
          autocapitalize="off"
          autocorrect="off"
          spellcheck="false"
        />
      </div>
      <div class="overflow-x-auto rounded-xl border border-[var(--border-subtle)]">
        <table class="table">
          <thead>
            <tr class="border-b border-[var(--border-subtle)] bg-[var(--bg-base)]/50">
              <th class="text-pretty-secondary font-medium">Model</th>
              <th class="text-pretty-secondary font-medium">Request Model</th>
              <th class="text-pretty-secondary font-medium">Enabled</th>
              <th class="text-pretty-secondary font-medium">更新时间</th>
              <th class="text-right text-pretty-secondary font-medium">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="!filteredPersistedModels.length">
              <td colspan="5" class="py-10 text-center text-sm text-pretty-muted">暂无已落库模型</td>
            </tr>
            <tr v-for="row in filteredPersistedModels" :key="row.id" class="border-b border-[var(--border-subtle)] hover:bg-[var(--bg-elevated)]/30 transition-colors">
              <td class="font-mono text-xs text-pretty">{{ row.model }}</td>
              <td class="font-mono text-xs text-pretty-secondary">{{ row.request_model }}</td>
              <td><span class="badge rounded-xl" :class="row.enabled ? 'bg-emerald-500/15 text-emerald-600 border-0' : 'bg-[var(--glow-primary)] text-pretty-muted border-0'">{{ row.enabled ? '启用' : '停用' }}</span></td>
              <td class="text-xs text-pretty-muted">{{ row.updated_at }}</td>
              <td>
                <div class="flex justify-end gap-2">
                  <button
                    class="btn btn-xs rounded-xl"
                    :class="row.enabled ? 'bg-amber-500/15 text-amber-600 border-0 hover:bg-amber-500/25' : 'bg-emerald-500/15 text-emerald-600 border-0 hover:bg-emerald-500/25'"
                    :disabled="togglingPersistedModelName === row.model"
                    @click="togglePersistedModelEnabled(row)"
                  >
                    <span v-if="togglingPersistedModelName === row.model" class="loading loading-spinner loading-xs" />
                    <span>{{ row.enabled ? '停用' : '启用' }}</span>
                  </button>
                  <button class="btn btn-ghost btn-xs rounded-xl text-pretty-secondary hover:bg-[var(--glow-primary)] hover:text-pretty" @click="openPersistedEditModal(row)">编辑</button>
                  <button
                    class="btn btn-error btn-xs rounded-xl border-0"
                    :disabled="deletingPersistedModelName === row.model"
                    @click="deletePersistedModel(row)"
                  >
                    <span v-if="deletingPersistedModelName === row.model" class="loading loading-spinner loading-xs" />
                    <span>删除</span>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </fieldset>

    <dialog :open="isPersistedEditModalOpen" class="modal">
      <div class="modal-box w-11/12 max-w-3xl rounded-2xl bg-[var(--bg-elevated)]">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2" @click="closePersistedEditModal">✕</button>
        </form>
        <h3 class="text-lg font-semibold text-pretty">编辑已落库模型</h3>

        <div class="mt-5 space-y-4">
          <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div>
              <label class="label"><span class="label-text font-medium text-pretty">Model</span></label>
              <input v-model="persistedModelForm.model" type="text" class="glass-input input input-bordered w-full rounded-xl" disabled />
            </div>
            <div>
              <label class="label"><span class="label-text font-medium text-pretty">Request Model</span></label>
              <input v-model="persistedModelForm.request_model" type="text" class="glass-input input input-bordered w-full rounded-xl" placeholder="默认同 model" />
            </div>
          </div>

          <div class="glass-card rounded-xl p-4">
            <h4 class="mb-3 text-sm font-semibold text-pretty-secondary uppercase tracking-wider">Capabilities & Status</h4>
            <div class="grid grid-cols-2 gap-x-4 gap-y-3 sm:grid-cols-4">
              <label class="label cursor-pointer justify-start gap-3 p-0"><input v-model="persistedModelForm.supports_chat_completions" type="checkbox" class="checkbox checkbox-sm checkbox-primary" /> <span class="label-text text-sm text-pretty-secondary">Chat</span></label>
              <label class="label cursor-pointer justify-start gap-3 p-0"><input v-model="persistedModelForm.supports_chat_responses" type="checkbox" class="checkbox checkbox-sm checkbox-primary" /> <span class="label-text text-sm text-pretty-secondary">Responses</span></label>
              <label class="label cursor-pointer justify-start gap-3 p-0"><input v-model="persistedModelForm.supports_embeddings" type="checkbox" class="checkbox checkbox-sm checkbox-primary" /> <span class="label-text text-sm text-pretty-secondary">Embeddings</span></label>
              <label class="label cursor-pointer justify-start gap-3 p-0"><input v-model="persistedModelForm.supports_rerank" type="checkbox" class="checkbox checkbox-sm checkbox-primary" /> <span class="label-text text-sm text-pretty-secondary">Rerank</span></label>
              <label class="label cursor-pointer justify-start gap-3 p-0"><input v-model="persistedModelForm.supports_audio_speech" type="checkbox" class="checkbox checkbox-sm checkbox-primary" /> <span class="label-text text-sm text-pretty-secondary">Speech</span></label>
              <label class="label cursor-pointer justify-start gap-3 p-0"><input v-model="persistedModelForm.supports_audio_transcriptions" type="checkbox" class="checkbox checkbox-sm checkbox-primary" /> <span class="label-text text-sm text-pretty-secondary">Transcribe</span></label>
              <label class="label cursor-pointer justify-start gap-3 p-0"><input v-model="persistedModelForm.supports_models" type="checkbox" class="checkbox checkbox-sm checkbox-primary" /> <span class="label-text text-sm text-pretty-secondary">Models</span></label>
              <label class="label cursor-pointer justify-start gap-3 p-0"><input v-model="persistedModelForm.enabled" type="checkbox" class="toggle toggle-success toggle-sm" /> <span class="label-text text-sm font-medium text-pretty">Enable</span></label>
            </div>
          </div>
        </div>

        <div class="modal-action mt-8">
          <button class="btn btn-ghost rounded-xl" @click="closePersistedEditModal">取消</button>
          <button class="btn rounded-xl bg-[var(--glow-primary)] text-pretty hover:opacity-90" :disabled="isSavingPersistedModel" @click="savePersistedModelEdit">
            <span v-if="isSavingPersistedModel" class="loading loading-spinner loading-xs" />
            <span>{{ isSavingPersistedModel ? '保存中...' : '保存修改' }}</span>
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button @click="closePersistedEditModal">close</button>
      </form>
    </dialog>
  </section>
</template>
