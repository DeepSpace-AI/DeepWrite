<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ApiError } from '@/api/http'
import {
  fetchCurrentAdminProfile,
  fetchRuntimeStatuses,
  getCachedSystemSettings,
  saveSystemSettings,
  type RuntimeServiceStatus,
  type SystemSettings,
} from '@/api/system'

const settings = reactive<SystemSettings>(getCachedSystemSettings())
const runtimeStatuses = ref<RuntimeServiceStatus[]>([
  { key: 'gateway', name: 'Gateway', status: 'unknown', detail: '未检测', endpoint: '-' },
  { key: 'web', name: 'Web', status: 'unknown', detail: '未检测', endpoint: '-' },
  { key: 'ai', name: 'AI', status: 'unknown', detail: '未检测', endpoint: '-' },
  { key: 'worker', name: 'Worker', status: 'unknown', detail: '未检测', endpoint: '-' },
])
const adminIdentity = ref<{ id: string; email: string; role: string } | null>(null)
const isLoading = ref(false)
const isSaving = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

function statusBadgeClass(status: RuntimeServiceStatus['status']) {
  if (status === 'online') return 'badge-success'
  if (status === 'offline') return 'badge-error'
  return 'badge-ghost'
}

async function loadRuntimeInfo() {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const [statuses, me] = await Promise.all([
      fetchRuntimeStatuses(),
      fetchCurrentAdminProfile(),
    ])
    runtimeStatuses.value = statuses
    adminIdentity.value = me
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '读取运行状态失败'
  } finally {
    isLoading.value = false
  }
}

async function saveSettings() {
  isSaving.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    saveSystemSettings({
      appName: settings.appName,
      appEnv: settings.appEnv,
      defaultProvider: settings.defaultProvider,
      strictAdminMode: settings.strictAdminMode,
      maintenanceMode: settings.maintenanceMode,
    })
    successMessage.value = '系统配置已保存（本地持久化）'
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '保存失败'
  } finally {
    isSaving.value = false
  }
}

onMounted(() => {
  loadRuntimeInfo()
})
</script>

<template>
  <section class="space-y-4 rounded-sm border border-base-300 bg-base-100 p-5 shadow-sm">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h2 class="text-xl font-semibold text-base-content">系统配置</h2>
        <p class="mt-1 text-sm text-base-content/65">管理后台策略与网关运行状态。</p>
      </div>
      <button class="btn btn-outline btn-sm rounded-sm" :disabled="isLoading" @click="loadRuntimeInfo">刷新状态</button>
    </div>

    <p v-if="errorMessage" class="rounded-sm border border-error/30 bg-error/10 px-3 py-2 text-sm text-error">{{ errorMessage }}</p>
    <p v-if="successMessage" class="rounded-sm border border-success/30 bg-success/10 px-3 py-2 text-sm text-success">{{ successMessage }}</p>

    <div class="grid gap-4 lg:grid-cols-2">
      <section class="rounded-sm border border-base-300 p-4">
        <h3 class="font-semibold">运行状态</h3>
        <div v-if="isLoading" class="mt-3 text-sm text-base-content/70">
          <span class="loading loading-spinner loading-sm" />
        </div>
        <div v-else class="mt-3 grid gap-3 sm:grid-cols-2">
          <div v-for="item in runtimeStatuses" :key="item.key" class="rounded-sm border border-base-300 bg-base-200/40 p-3">
            <div class="flex items-center justify-between gap-2">
              <p class="text-sm font-semibold">{{ item.name }}</p>
              <span class="badge badge-sm" :class="statusBadgeClass(item.status)">
                {{ item.status }}
              </span>
            </div>
            <p class="mt-2 text-xs text-base-content/65">{{ item.detail }}</p>
            <p class="mt-1 truncate font-mono text-[11px] text-base-content/55">{{ item.endpoint }}</p>
          </div>
        </div>
        <dl class="mt-4 space-y-2 text-sm">
          <div class="flex justify-between gap-3 rounded-sm border border-base-300 bg-base-200/30 px-2 py-1.5">
            <dt class="text-base-content/60">Admin ID</dt>
            <dd class="font-mono">{{ adminIdentity?.id || '-' }}</dd>
          </div>
          <div class="flex justify-between gap-3 rounded-sm border border-base-300 bg-base-200/30 px-2 py-1.5">
            <dt class="text-base-content/60">Admin Email</dt>
            <dd class="font-mono">{{ adminIdentity?.email || '-' }}</dd>
          </div>
          <div class="flex justify-between gap-3 rounded-sm border border-base-300 bg-base-200/30 px-2 py-1.5">
            <dt class="text-base-content/60">Role</dt>
            <dd class="font-mono">{{ adminIdentity?.role || '-' }}</dd>
          </div>
        </dl>
      </section>

      <section class="rounded-sm border border-base-300 p-4">
        <h3 class="font-semibold">配置面板</h3>
        <fieldset class="fieldset mt-3 rounded-sm border border-base-300 bg-base-200/40 p-4">
          <legend class="fieldset-legend">系统设置</legend>

          <label class="label">应用名称</label>
          <input v-model="settings.appName" type="text" class="input input-bordered w-full" placeholder="DeepWrite Admin" />
          <p class="label">用于后台系统标题展示</p>

          <label class="label mt-2">运行环境</label>
          <select v-model="settings.appEnv" class="select select-bordered w-full">
            <option value="development">development</option>
            <option value="staging">staging</option>
            <option value="production">production</option>
          </select>
          <p class="label">影响后台展示的环境标识</p>

          <label class="label mt-2">默认 Provider</label>
          <input v-model="settings.defaultProvider" type="text" class="input input-bordered w-full" placeholder="openai-compatible" />
          <p class="label">新配置项的默认 provider 建议值</p>

          <fieldset class="fieldset mt-3 rounded-sm border border-base-300 bg-base-100 p-3">
            <legend class="fieldset-legend">运行策略</legend>
            <label class="label cursor-pointer justify-start gap-2">
              <input v-model="settings.strictAdminMode" type="checkbox" class="toggle toggle-sm" />
              严格管理员模式
            </label>
            <label class="label cursor-pointer justify-start gap-2">
              <input v-model="settings.maintenanceMode" type="checkbox" class="toggle toggle-sm" />
              维护模式
            </label>
          </fieldset>

          <div class="mt-4">
            <button class="btn btn-primary rounded-sm" :disabled="isSaving" @click="saveSettings">
              <span v-if="isSaving" class="loading loading-spinner loading-xs" />
              <span>{{ isSaving ? '保存中...' : '保存配置' }}</span>
            </button>
          </div>
        </fieldset>
      </section>
    </div>
  </section>
</template>
