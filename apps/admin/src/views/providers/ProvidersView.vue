<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { createVendor, listVendors, updateVendorEnabled, type AIProviderVendor, type AIProviderVendorInput } from '@/api/aiProvider'
import { ApiError } from '@/api/http'
import IconAdd from '~icons/mdi/plus'
import IconSearch from '~icons/mdi/magnify'
import IconRefresh from '~icons/mdi/refresh'
import IconEye from '~icons/mdi/eye'
import IconEyeOff from '~icons/mdi/eye-off'
import IconChevronDown from '~icons/mdi/chevron-down'
import IconChevronUp from '~icons/mdi/chevron-up'
import IconCheck from '~icons/mdi/check'
import IconClose from '~icons/mdi/close'

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
    return passesEnabled && passesKeyword
  })
})

const sortedFilteredVendors = computed(() => {
  return [...filteredVendors.value].sort((a, b) => 
    new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
  )
})

const vendorStats = computed(() => {
  const total = vendors.value.length
  const enabled = vendors.value.filter(item => item.enabled).length
  return { total, enabled, disabled: total - enabled }
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
    errorMessage.value = '请填写 Provider、Base URL、API Key'
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
    successMessage.value = '厂商已创建'
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
    successMessage.value = nextEnabled ? `已启用 ${item.name}` : `已停用 ${item.name}`
    await loadVendors()
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '更新状态失败'
  } finally {
    togglingVendorId.value = ''
  }
}

onMounted(() => {
  loadVendors()
})
</script>

<template>
  <section class="space-y-6">
    <header class="flex items-center justify-between">
      <div>
        <h1 class="text-display text-2xl">AI Provider</h1>
        <p class="text-secondary mt-1 text-sm">管理厂商接入配置与模型</p>
      </div>
      <button class="btn-tech flex items-center gap-2 px-4 py-2 text-sm" @click="openVendorModal">
        <IconAdd class="h-4 w-4" />
        新增 Provider
      </button>
    </header>

    <div v-if="errorMessage" class="tech-card px-4 py-3 text-sm text-[var(--accent-error)]">
      {{ errorMessage }}
    </div>
    <div v-if="successMessage" class="tech-card border-[var(--accent-success)] px-4 py-3 text-sm text-[var(--accent-success)]">
      {{ successMessage }}
    </div>

    <div class="grid gap-3 sm:grid-cols-3">
      <div class="tech-card p-4">
        <p class="text-muted text-xs">总厂商</p>
        <p class="text-display mt-1 text-2xl">{{ vendorStats.total }}</p>
      </div>
      <div class="tech-card p-4">
        <p class="text-muted text-xs">启用中</p>
        <p class="text-display mt-1 text-2xl text-[var(--accent-success)]">{{ vendorStats.enabled }}</p>
      </div>
      <div class="tech-card p-4">
        <p class="text-muted text-xs">已停用</p>
        <p class="text-display mt-1 text-2xl text-[var(--text-muted)]">{{ vendorStats.disabled }}</p>
      </div>
    </div>

    <div class="tech-card p-4">
      <div class="flex flex-wrap items-center gap-3">
        <div class="relative min-w-0 flex-1">
          <IconSearch class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--text-muted)]" />
          <input
            v-model="searchKeyword"
            class="tech-input w-full py-2 pl-9 pr-3 text-sm"
            placeholder="搜索厂商..."
            autocomplete="off"
          />
        </div>
        <select v-model="enabledFilter" class="tech-input px-3 py-2 text-sm">
          <option value="all">全部</option>
          <option value="enabled">启用</option>
          <option value="disabled">停用</option>
        </select>
        <button class="btn-ghost-tech flex items-center gap-2 px-3 py-2 text-sm" :disabled="isLoading" @click="loadVendors">
          <IconRefresh class="h-4 w-4" :class="{ 'animate-spin': isLoading }" />
          <span class="hidden sm:inline">刷新</span>
        </button>
      </div>
    </div>

    <div class="tech-card overflow-hidden">
      <div v-if="isLoading" class="flex items-center justify-center py-12">
        <span class="loading loading-spinner loading-md text-[var(--text-muted)]" />
      </div>
      <div v-else-if="!sortedFilteredVendors.length" class="flex flex-col items-center justify-center py-12">
        <p class="text-secondary text-sm">暂无厂商数据</p>
        <button class="btn-ghost-tech mt-3 px-3 py-1.5 text-sm" @click="openVendorModal">新增 Provider</button>
      </div>
      <table v-else class="tech-table">
        <thead>
          <tr>
            <th>厂商</th>
            <th>Base URL</th>
            <th>API Key</th>
            <th>状态</th>
            <th class="text-right">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in sortedFilteredVendors" :key="item.id">
            <td>
              <p class="font-medium">{{ item.name }}</p>
              <p v-if="item.organization" class="text-xs text-[var(--text-muted)]">{{ item.organization }}</p>
            </td>
            <td>
              <code class="text-mono text-xs">{{ item.base_url }}</code>
            </td>
            <td>
              <code class="text-mono text-xs text-[var(--text-muted)]">{{ item.api_key_masked || '-' }}</code>
            </td>
            <td>
              <span 
                class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium"
                :class="item.enabled ? 'badge-success' : 'badge-neutral'"
              >
                {{ item.enabled ? '启用' : '停用' }}
              </span>
            </td>
            <td class="text-right">
              <div class="flex items-center justify-end gap-1">
                <button
                  class="inline-flex items-center rounded px-2 py-1 text-xs font-medium transition-colors"
                  :class="item.enabled 
                    ? 'text-[var(--text-muted)] hover:bg-[var(--surface-hover)] hover:text-[var(--text-primary)]' 
                    : 'text-[var(--accent-primary)] hover:bg-[var(--surface-hover)]'"
                  :disabled="togglingVendorId === item.id"
                  @click="toggleVendorStatus(item)"
                >
                  <span v-if="togglingVendorId === item.id" class="loading loading-spinner loading-xs" />
                  <span>{{ item.enabled ? '停用' : '启用' }}</span>
                </button>
                <button 
                  class="inline-flex items-center rounded px-2 py-1 text-xs font-medium text-[var(--text-muted)] transition-colors hover:bg-[var(--surface-hover)] hover:text-[var(--text-primary)]"
                  @click="router.push({ name: 'provider-detail', params: { providerId: item.id } })"
                >
                  详情
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <dialog :open="isVendorModalOpen" class="modal">
      <div class="tech-modal modal-box max-w-xl">
        <div class="flex items-center justify-between border-b border-[var(--border-subtle)] px-5 py-4">
          <h3 class="text-heading text-base">新增 Provider</h3>
          <button class="flex h-7 w-7 items-center justify-center rounded text-[var(--text-muted)] hover:bg-[var(--surface-hover)]" @click="closeVendorModal">
            <IconClose class="h-5 w-5" />
          </button>
        </div>

        <form class="space-y-4 p-5" @submit.prevent="submitVendor">
          <div class="grid gap-4 sm:grid-cols-2">
            <div>
              <label class="text-label mb-1.5 block">Provider <span class="text-[var(--accent-error)]">*</span></label>
              <input v-model="vendorForm.provider" class="tech-input w-full px-3 py-2 text-sm" placeholder="openai-compatible" />
            </div>
            <div>
              <label class="text-label mb-1.5 block">Organization</label>
              <input v-model="vendorForm.organization" class="tech-input w-full px-3 py-2 text-sm" placeholder="org_xxx" />
            </div>
          </div>

          <div>
            <label class="text-label mb-1.5 block">Base URL <span class="text-[var(--accent-error)]">*</span></label>
            <input v-model="vendorForm.base_url" class="tech-input w-full px-3 py-2 text-sm" placeholder="https://api.example.com/v1" />
          </div>

          <div>
            <div class="flex items-center justify-between">
              <label class="text-label mb-1.5 block">API Key <span class="text-[var(--accent-error)]">*</span></label>
              <button type="button" class="flex items-center gap-1 text-xs text-[var(--text-muted)] hover:text-[var(--text-primary)]" @click="showApiKey = !showApiKey">
                <component :is="showApiKey ? IconEyeOff : IconEye" class="h-4 w-4" />
                {{ showApiKey ? '隐藏' : '显示' }}
              </button>
            </div>
            <input
              v-model="vendorForm.api_key"
              :type="showApiKey ? 'text' : 'password'"
              class="tech-input w-full px-3 py-2 text-sm"
              placeholder="sk-..."
              autocomplete="off"
            />
          </div>

          <div>
            <button type="button" class="flex items-center gap-1.5 text-sm text-[var(--text-muted)] hover:text-[var(--text-primary)]" @click="showAdvancedPaths = !showAdvancedPaths">
              <component :is="showAdvancedPaths ? IconChevronUp : IconChevronDown" class="h-4 w-4" />
              高级配置
            </button>
          </div>

          <div v-if="showAdvancedPaths" class="grid gap-4 sm:grid-cols-2">
            <div>
              <label class="text-label mb-1.5 block">Chat Path</label>
              <input v-model="vendorForm.chat_completions_path" class="tech-input w-full px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="text-label mb-1.5 block">Embeddings Path</label>
              <input v-model="vendorForm.embeddings_path" class="tech-input w-full px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="text-label mb-1.5 block">Models Path</label>
              <input v-model="vendorForm.models_path" class="tech-input w-full px-3 py-2 text-sm" />
            </div>
          </div>

          <label class="flex cursor-pointer items-center gap-2">
            <input v-model="vendorForm.enabled" type="checkbox" class="toggle toggle-sm" />
            <span class="text-sm">启用此厂商</span>
          </label>
        </form>

        <div class="flex justify-end gap-2 border-t border-[var(--border-subtle)] px-5 py-3">
          <button class="btn-ghost-tech px-4 py-2 text-sm" @click="closeVendorModal">取消</button>
          <button 
            class="btn-tech flex items-center gap-2 px-4 py-2 text-sm" 
            :disabled="isSubmitting || !canSubmitVendor"
            @click="submitVendor"
          >
            <span v-if="isSubmitting" class="loading loading-spinner loading-xs" />
            <IconCheck v-else class="h-4 w-4" />
            创建
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop-tech">
        <button @click="closeVendorModal">close</button>
      </form>
    </dialog>
  </section>
</template>