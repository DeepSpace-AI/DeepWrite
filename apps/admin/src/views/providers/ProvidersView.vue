<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { createVendor, listVendors, updateVendorEnabled, type AIProviderVendor, type AIProviderVendorInput } from '@/api/aiProvider'
import { ApiError } from '@/api/http'

const router = useRouter()
const isLoading = ref(false)
const isSubmitting = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const searchKeyword = ref('')
const enabledFilter = ref<'all' | 'enabled' | 'disabled'>('all')
const vendors = ref<AIProviderVendor[]>([])
const isVendorModalOpen = ref(false)
const showAdvancedPaths = ref(false)
const showApiKey = ref(false)
const togglingVendorId = ref('')

const vendorForm = reactive<AIProviderVendorInput>({
  provider: '',
  base_url: '',
  api_key: '',
  organization: '',
  chat_completions_path: '/chat/completions',
  chat_responses_path: '/responses',
  embeddings_path: '/embeddings',
  rerank_path: '/rerank',
  audio_speech_path: '/audio/speech',
  audio_transcriptions_path: '/audio/transcriptions',
  models_path: '/models',
  enabled: true,
})

const filteredVendors = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()
  return vendors.value.filter((item) => {
    const passesEnabled = enabledFilter.value === 'all'
      || (enabledFilter.value === 'enabled' ? item.enabled : !item.enabled)
    const passesKeyword = !keyword
      || item.name?.toLowerCase().includes(keyword)
      || item.base_url?.toLowerCase().includes(keyword)
      || item.organization?.toLowerCase().includes(keyword)
    return passesEnabled && passesKeyword
  })
})

const sortedFilteredVendors = computed(() => {
  return [...filteredVendors.value].sort((a, b) => {
    return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
  })
})

const vendorStats = computed(() => {
  const total = vendors.value.length
  const enabled = vendors.value.filter(item => item.enabled).length
  const disabled = total - enabled
  return {
    total,
    enabled,
    disabled,
    filtered: sortedFilteredVendors.value.length,
  }
})

const hasActiveFilters = computed(() => {
  return Boolean(searchKeyword.value.trim()) || enabledFilter.value !== 'all'
})

const canSubmitVendor = computed(() => {
  return Boolean(
    vendorForm.provider.trim()
    && vendorForm.base_url.trim()
    && vendorForm.api_key.trim(),
  )
})

function resetVendorForm() {
  vendorForm.provider = ''
  vendorForm.base_url = ''
  vendorForm.api_key = ''
  vendorForm.organization = ''
  vendorForm.chat_completions_path = '/chat/completions'
  vendorForm.chat_responses_path = '/responses'
  vendorForm.embeddings_path = '/embeddings'
  vendorForm.rerank_path = '/rerank'
  vendorForm.audio_speech_path = '/audio/speech'
  vendorForm.audio_transcriptions_path = '/audio/transcriptions'
  vendorForm.models_path = '/models'
  vendorForm.enabled = true
  showAdvancedPaths.value = false
  showApiKey.value = false
}

function openVendorModal() {
  resetVendorForm()
  errorMessage.value = ''
  successMessage.value = ''
  isVendorModalOpen.value = true
}

function closeVendorModal() {
  isVendorModalOpen.value = false
  resetVendorForm()
}

function clearFilters() {
  searchKeyword.value = ''
  enabledFilter.value = 'all'
}

async function loadVendors() {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const res = await listVendors()
    vendors.value = res.items
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '加载厂商失败'
  } finally {
    isLoading.value = false
  }
}

async function submitVendor() {
  if (!canSubmitVendor.value) {
    errorMessage.value = '请填写 Provider、Base URL、API Key 后再提交'
    return
  }
  isSubmitting.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const provider = await createVendor({
      ...vendorForm,
      provider: vendorForm.provider.trim(),
      base_url: vendorForm.base_url.trim(),
      api_key: vendorForm.api_key.trim(),
      organization: vendorForm.organization?.trim(),
    })
    successMessage.value = '厂商已落库，请在详情页拉取模型列表'
    closeVendorModal()
    await loadVendors()
    await router.push({ name: 'provider-detail', params: { providerId: provider.id } })
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '创建厂商失败'
  } finally {
    isSubmitting.value = false
  }
}

async function toggleVendorStatus(item: AIProviderVendor) {
  togglingVendorId.value = item.id
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const nextEnabled = !item.enabled
    await updateVendorEnabled(item.id, nextEnabled)
    successMessage.value = nextEnabled
      ? `已启用厂商 ${item.name}`
      : `已停用厂商 ${item.name}`
    await loadVendors()
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '更新厂商状态失败'
  } finally {
    togglingVendorId.value = ''
  }
}

onMounted(() => {
  loadVendors()
})
</script>

<template>
  <section class="space-y-4">
    <div class="glass-card rounded-2xl p-5">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 class="text-xl font-semibold text-pretty">AI Provider 管理</h2>
          <p class="mt-1 text-sm text-pretty-secondary">统一管理厂商接入配置、状态与模型配置入口</p>
        </div>
        <button class="btn rounded-xl bg-[var(--glow-primary)] text-pretty hover:opacity-90" @click="openVendorModal">新增 Provider</button>
      </div>
    </div>

    <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      <div class="glass-card rounded-xl p-3">
        <p class="text-[10px] uppercase tracking-wider text-pretty-muted">总厂商</p>
        <p class="mt-1 text-2xl font-semibold text-pretty">{{ vendorStats.total }}</p>
      </div>
      <div class="glass-card rounded-xl p-3 border border-emerald-500/20">
        <p class="text-[10px] uppercase tracking-wider text-emerald-600/80">启用中</p>
        <p class="mt-1 text-2xl font-semibold text-emerald-600">{{ vendorStats.enabled }}</p>
      </div>
      <div class="glass-card rounded-xl p-3 border border-amber-500/20">
        <p class="text-[10px] uppercase tracking-wider text-amber-600/80">停用中</p>
        <p class="mt-1 text-2xl font-semibold text-amber-600">{{ vendorStats.disabled }}</p>
      </div>
      <div class="glass-card rounded-xl p-3 border border-blue-500/20">
        <p class="text-[10px] uppercase tracking-wider text-blue-600/80">当前筛选</p>
        <p class="mt-1 text-2xl font-semibold text-blue-600">{{ vendorStats.filtered }}</p>
      </div>
    </div>

    <div v-if="errorMessage" class="glass-card rounded-xl p-4 text-sm text-error border border-error/20 bg-error/5">{{ errorMessage }}</div>
    <div v-if="successMessage" class="glass-card rounded-xl p-4 text-sm text-emerald-600 border border-emerald-500/20 bg-emerald-500/5">{{ successMessage }}</div>

    <div class="glass-card rounded-2xl p-5">
      <form class="flex flex-wrap items-center gap-2" autocomplete="off" @submit.prevent>
        <input
          v-model="searchKeyword"
          class="glass-input input input-bordered input-sm w-full rounded-xl sm:max-w-xs"
          placeholder="搜索 provider/base_url/organization"
          autocomplete="off"
          autocapitalize="off"
          autocorrect="off"
          spellcheck="false"
        />
        <select v-model="enabledFilter" class="glass-input select select-bordered select-sm rounded-xl">
          <option value="all">全部</option>
          <option value="enabled">仅启用</option>
          <option value="disabled">仅停用</option>
        </select>
        <button class="btn btn-ghost btn-sm rounded-xl" :disabled="isLoading" @click="loadVendors">刷新</button>
        <button v-if="hasActiveFilters" class="btn btn-ghost btn-sm rounded-xl" @click="clearFilters">清空筛选</button>
      </form>
    </div>

    <div class="glass-card rounded-2xl overflow-hidden">
      <div v-if="isLoading" class="py-12 text-center"><span class="loading loading-spinner loading-md text-pretty-muted" /></div>
      <div v-else class="overflow-x-auto">
        <table class="table">
          <thead>
            <tr class="border-b border-[var(--border-subtle)]">
              <th class="text-pretty-secondary font-medium">厂商</th>
              <th class="text-pretty-secondary font-medium">Base URL</th>
              <th class="text-pretty-secondary font-medium">Organization</th>
              <th class="text-pretty-secondary font-medium">API Key</th>
              <th class="text-pretty-secondary font-medium">状态</th>
              <th class="text-right text-pretty-secondary font-medium">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="!sortedFilteredVendors.length">
              <td colspan="6" class="py-12 text-center text-sm text-pretty-muted">
                暂无匹配厂商，请调整筛选条件或新增厂商
              </td>
            </tr>
            <tr v-for="item in sortedFilteredVendors" :key="item.id" class="border-b border-[var(--border-subtle)] hover:bg-[var(--bg-elevated)]/50 transition-colors">
              <td class="font-medium text-pretty">{{ item.name }}</td>
              <td class="font-mono text-xs text-pretty-secondary">{{ item.base_url }}</td>
              <td class="text-sm text-pretty-secondary">{{ item.organization || '-' }}</td>
              <td class="font-mono text-xs text-pretty-muted">{{ item.api_key_masked || '-' }}</td>
              <td><span class="badge rounded-xl" :class="item.enabled ? 'bg-emerald-500/15 text-emerald-600 border-0' : 'bg-[var(--glow-primary)] text-pretty-muted border-0'">{{ item.enabled ? '启用' : '停用' }}</span></td>
              <td class="text-right">
                <div class="flex justify-end gap-2">
                  <button
                    class="btn btn-xs rounded-xl"
                    :class="item.enabled ? 'bg-amber-500/15 text-amber-600 border-0 hover:bg-amber-500/25' : 'bg-emerald-500/15 text-emerald-600 border-0 hover:bg-emerald-500/25'"
                    :disabled="togglingVendorId === item.id"
                    @click="toggleVendorStatus(item)"
                  >
                    <span v-if="togglingVendorId === item.id" class="loading loading-spinner loading-xs" />
                    <span>{{ item.enabled ? '停用' : '启用' }}</span>
                  </button>
                  <button class="btn btn-ghost btn-xs rounded-xl text-pretty-secondary hover:bg-[var(--glow-primary)] hover:text-pretty" @click="router.push({ name: 'provider-detail', params: { providerId: item.id } })">详情</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <dialog :open="isVendorModalOpen" class="modal">
      <div class="modal-box max-w-4xl rounded-2xl bg-[var(--bg-elevated)]">
        <h3 class="text-lg font-semibold text-pretty">新增 Provider 厂商</h3>
        <p class="mt-1 text-sm text-pretty-secondary">先填写核心接入信息，路径字段可按需展开调整</p>
        <fieldset class="fieldset mt-4 rounded-xl border border-[var(--border-subtle)] bg-[var(--bg-base)] p-4">
          <legend class="fieldset-legend text-pretty-secondary">厂商接入信息</legend>
          <div class="grid gap-4 sm:grid-cols-2">
            <div>
              <label class="label"><span class="label-text text-pretty-secondary">Provider <span class="text-error">*</span></span></label>
              <input v-model="vendorForm.provider" class="glass-input input input-bordered w-full rounded-xl" placeholder="openai-compatible" />
            </div>
            <div>
              <label class="label"><span class="label-text text-pretty-secondary">Organization</span></label>
              <input v-model="vendorForm.organization" class="glass-input input input-bordered w-full rounded-xl" placeholder="org_xxx" />
            </div>
            <div class="sm:col-span-2">
              <label class="label"><span class="label-text text-pretty-secondary">Base URL <span class="text-error">*</span></span></label>
              <input v-model="vendorForm.base_url" class="glass-input input input-bordered w-full rounded-xl" placeholder="https://api.example.com/v1" />
            </div>
            <div class="sm:col-span-2">
              <div class="label">
                <span class="label-text text-pretty-secondary">API Key <span class="text-error">*</span></span>
                <button type="button" class="btn btn-ghost btn-xs rounded-xl" @click="showApiKey = !showApiKey">
                  {{ showApiKey ? '隐藏' : '显示' }}
                </button>
              </div>
              <input
                v-model="vendorForm.api_key"
                :type="showApiKey ? 'text' : 'password'"
                class="glass-input input input-bordered w-full rounded-xl"
                placeholder="sk-..."
                autocomplete="off"
              />
            </div>
            <div class="sm:col-span-2">
              <button type="button" class="btn btn-ghost btn-sm rounded-xl" @click="showAdvancedPaths = !showAdvancedPaths">
                {{ showAdvancedPaths ? '收起高级路径配置' : '展开高级路径配置' }}
              </button>
            </div>
            <template v-if="showAdvancedPaths">
              <div>
                <label class="label"><span class="label-text text-pretty-secondary">Chat Completions Path</span></label>
                <input v-model="vendorForm.chat_completions_path" class="glass-input input input-bordered w-full rounded-xl" placeholder="/chat/completions" />
              </div>
              <div>
                <label class="label"><span class="label-text text-pretty-secondary">Responses Path</span></label>
                <input v-model="vendorForm.chat_responses_path" class="glass-input input input-bordered w-full rounded-xl" placeholder="/responses" />
              </div>
              <div>
                <label class="label"><span class="label-text text-pretty-secondary">Embeddings Path</span></label>
                <input v-model="vendorForm.embeddings_path" class="glass-input input input-bordered w-full rounded-xl" placeholder="/embeddings" />
              </div>
              <div>
                <label class="label"><span class="label-text text-pretty-secondary">Rerank Path</span></label>
                <input v-model="vendorForm.rerank_path" class="glass-input input input-bordered w-full rounded-xl" placeholder="/rerank" />
              </div>
              <div>
                <label class="label"><span class="label-text text-pretty-secondary">Audio Speech Path</span></label>
                <input v-model="vendorForm.audio_speech_path" class="glass-input input input-bordered w-full rounded-xl" placeholder="/audio/speech" />
              </div>
              <div>
                <label class="label"><span class="label-text text-pretty-secondary">Audio Transcriptions Path</span></label>
                <input v-model="vendorForm.audio_transcriptions_path" class="glass-input input input-bordered w-full rounded-xl" placeholder="/audio/transcriptions" />
              </div>
            </template>
            <div>
              <label class="label"><span class="label-text text-pretty-secondary">Models Path</span></label>
              <input v-model="vendorForm.models_path" class="glass-input input input-bordered w-full rounded-xl" placeholder="/models" />
            </div>
            <label class="label cursor-pointer justify-start gap-2 mt-2">
              <input v-model="vendorForm.enabled" type="checkbox" class="toggle toggle-sm" />
              <span class="label-text text-pretty-secondary">Enabled</span>
            </label>
          </div>
        </fieldset>
        <div class="modal-action">
          <button class="btn btn-ghost rounded-xl" @click="closeVendorModal">取消</button>
          <button class="btn rounded-xl bg-[var(--glow-primary)] text-pretty hover:opacity-90" :disabled="isSubmitting || !canSubmitVendor" @click="submitVendor">
            <span v-if="isSubmitting" class="loading loading-spinner loading-xs" />
            <span>{{ isSubmitting ? '提交中...' : '创建厂商并落库' }}</span>
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button @click="closeVendorModal">close</button>
      </form>
    </dialog>
  </section>
</template>
