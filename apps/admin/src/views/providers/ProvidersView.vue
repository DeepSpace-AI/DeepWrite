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
  <section class="space-y-4 rounded-sm border border-base-300 bg-base-100 p-5 shadow-sm">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h2 class="text-xl font-semibold text-base-content">AI Provider 管理</h2>
        <p class="mt-1 text-sm text-base-content/65">统一管理厂商接入配置、状态与模型配置入口。</p>
      </div>
      <button class="btn btn-primary rounded-sm" @click="openVendorModal">新增 Provider 厂商</button>
    </div>

    <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      <div class="rounded-sm border border-base-300 bg-base-200/40 p-3">
        <p class="text-xs uppercase tracking-wide text-base-content/55">总厂商</p>
        <p class="mt-1 text-2xl font-semibold">{{ vendorStats.total }}</p>
      </div>
      <div class="rounded-sm border border-success/30 bg-success/10 p-3">
        <p class="text-xs uppercase tracking-wide text-success/80">启用中</p>
        <p class="mt-1 text-2xl font-semibold text-success">{{ vendorStats.enabled }}</p>
      </div>
      <div class="rounded-sm border border-warning/30 bg-warning/10 p-3">
        <p class="text-xs uppercase tracking-wide text-warning/80">停用中</p>
        <p class="mt-1 text-2xl font-semibold text-warning">{{ vendorStats.disabled }}</p>
      </div>
      <div class="rounded-sm border border-info/30 bg-info/10 p-3">
        <p class="text-xs uppercase tracking-wide text-info/80">当前筛选</p>
        <p class="mt-1 text-2xl font-semibold text-info">{{ vendorStats.filtered }}</p>
      </div>
    </div>

    <p v-if="errorMessage" class="rounded-sm border border-error/30 bg-error/10 px-3 py-2 text-sm text-error">{{ errorMessage }}</p>
    <p v-if="successMessage" class="rounded-sm border border-success/30 bg-success/10 px-3 py-2 text-sm text-success">{{ successMessage }}</p>

    <form class="flex flex-wrap items-center gap-2" autocomplete="off" @submit.prevent>
      <input
        v-model="searchKeyword"
        class="input input-bordered input-sm w-full rounded-sm sm:max-w-xs"
        placeholder="搜索 provider/base_url/organization"
        autocomplete="off"
        autocapitalize="off"
        autocorrect="off"
        spellcheck="false"
      />
      <select v-model="enabledFilter" class="select select-bordered select-sm rounded-sm">
        <option value="all">全部</option>
        <option value="enabled">仅启用</option>
        <option value="disabled">仅停用</option>
      </select>
      <button class="btn btn-outline btn-sm rounded-sm" :disabled="isLoading" @click="loadVendors">刷新</button>
      <button v-if="hasActiveFilters" class="btn btn-ghost btn-sm rounded-sm" @click="clearFilters">清空筛选</button>
    </form>

    <div class="rounded-sm border border-base-300">
      <div v-if="isLoading" class="py-8 text-center"><span class="loading loading-spinner loading-md" /></div>
      <div v-else class="overflow-x-auto">
        <table class="table table-zebra">
          <thead>
            <tr>
              <th>厂商</th>
              <th>Base URL</th>
              <th>Organization</th>
              <th>API Key</th>
              <th>状态</th>
              <th class="text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="!sortedFilteredVendors.length">
              <td colspan="6" class="py-10 text-center text-sm text-base-content/60">
                暂无匹配厂商，请调整筛选条件或新增厂商。
              </td>
            </tr>
            <tr v-for="item in sortedFilteredVendors" :key="item.id" class="hover">
              <td class="font-medium">{{ item.name }}</td>
              <td class="font-mono text-xs">{{ item.base_url }}</td>
              <td class="text-sm">{{ item.organization || '-' }}</td>
              <td class="font-mono text-xs">{{ item.api_key_masked || '-' }}</td>
              <td><span class="badge badge-outline">{{ item.enabled ? '启用' : '停用' }}</span></td>
              <td class="text-right">
                <div class="flex justify-end gap-2">
                  <button
                    class="btn btn-xs rounded-sm"
                    :class="item.enabled ? 'btn-warning' : 'btn-success'"
                    :disabled="togglingVendorId === item.id"
                    @click="toggleVendorStatus(item)"
                  >
                    <span v-if="togglingVendorId === item.id" class="loading loading-spinner loading-xs" />
                    <span>{{ item.enabled ? '停用' : '启用' }}</span>
                  </button>
                  <button class="btn btn-primary btn-xs rounded-sm" @click="router.push({ name: 'provider-detail', params: { providerId: item.id } })">进入详情</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <dialog :open="isVendorModalOpen" class="modal">
      <div class="modal-box max-w-4xl rounded-sm">
        <h3 class="text-lg font-semibold">新增 Provider 厂商</h3>
        <p class="mt-1 text-sm text-base-content/65">先填写核心接入信息，路径字段可按需展开调整。</p>
        <fieldset class="fieldset mt-3 rounded-sm border border-base-300 bg-base-200/35 p-4">
          <legend class="fieldset-legend">厂商接入信息</legend>
          <div class="grid gap-3 sm:grid-cols-2">
            <div>
              <label class="label">Provider <span class="text-error">*</span></label>
              <input v-model="vendorForm.provider" class="input input-bordered w-full" placeholder="openai-compatible" />
            </div>
            <div>
              <label class="label">Organization</label>
              <input v-model="vendorForm.organization" class="input input-bordered w-full" placeholder="org_xxx" />
            </div>
            <div class="sm:col-span-2">
              <label class="label">Base URL <span class="text-error">*</span></label>
              <input v-model="vendorForm.base_url" class="input input-bordered w-full" placeholder="https://api.example.com/v1" />
            </div>
            <div class="sm:col-span-2">
              <div class="label">
                <span>API Key <span class="text-error">*</span></span>
                <button type="button" class="btn btn-ghost btn-xs rounded-sm" @click="showApiKey = !showApiKey">
                  {{ showApiKey ? '隐藏' : '显示' }}
                </button>
              </div>
              <input
                v-model="vendorForm.api_key"
                :type="showApiKey ? 'text' : 'password'"
                class="input input-bordered w-full"
                placeholder="sk-..."
                autocomplete="off"
              />
            </div>
            <div class="sm:col-span-2">
              <button type="button" class="btn btn-outline btn-sm rounded-sm" @click="showAdvancedPaths = !showAdvancedPaths">
                {{ showAdvancedPaths ? '收起高级路径配置' : '展开高级路径配置' }}
              </button>
            </div>
            <template v-if="showAdvancedPaths">
              <div>
                <label class="label">Chat Completions Path</label>
                <input v-model="vendorForm.chat_completions_path" class="input input-bordered w-full" placeholder="/chat/completions" />
              </div>
              <div>
                <label class="label">Responses Path</label>
                <input v-model="vendorForm.chat_responses_path" class="input input-bordered w-full" placeholder="/responses" />
              </div>
              <div>
                <label class="label">Embeddings Path</label>
                <input v-model="vendorForm.embeddings_path" class="input input-bordered w-full" placeholder="/embeddings" />
              </div>
              <div>
                <label class="label">Rerank Path</label>
                <input v-model="vendorForm.rerank_path" class="input input-bordered w-full" placeholder="/rerank" />
              </div>
              <div>
                <label class="label">Audio Speech Path</label>
                <input v-model="vendorForm.audio_speech_path" class="input input-bordered w-full" placeholder="/audio/speech" />
              </div>
              <div>
                <label class="label">Audio Transcriptions Path</label>
                <input v-model="vendorForm.audio_transcriptions_path" class="input input-bordered w-full" placeholder="/audio/transcriptions" />
              </div>
            </template>
            <div>
              <label class="label">Models Path</label>
              <input v-model="vendorForm.models_path" class="input input-bordered w-full" placeholder="/models" />
            </div>
            <label class="label cursor-pointer justify-start gap-2 mt-6">
              <input v-model="vendorForm.enabled" type="checkbox" class="toggle toggle-sm" />
              Enabled
            </label>
          </div>
        </fieldset>
        <div class="modal-action">
          <button class="btn btn-ghost rounded-sm" @click="closeVendorModal">取消</button>
          <button class="btn btn-primary rounded-sm" :disabled="isSubmitting || !canSubmitVendor" @click="submitVendor">
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
