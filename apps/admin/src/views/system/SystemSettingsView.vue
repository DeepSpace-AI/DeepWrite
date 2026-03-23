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
  <section class="space-y-4">
    <div class="glass-card rounded-2xl p-5">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 class="text-xl font-semibold text-pretty">系统配置</h2>
          <p class="mt-1 text-sm text-pretty-secondary">管理后台策略与网关运行状态</p>
        </div>
        <button class="btn btn-ghost btn-sm rounded-xl" :disabled="isLoading" @click="loadRuntimeInfo">刷新状态</button>
      </div>
    </div>

    <div v-if="errorMessage" class="glass-card rounded-xl p-4 text-sm text-error border border-error/20 bg-error/5">{{ errorMessage }}</div>
    <div v-if="successMessage" class="glass-card rounded-xl p-4 text-sm text-emerald-600 border border-emerald-500/20 bg-emerald-500/5">{{ successMessage }}</div>

    <div class="grid gap-4 lg:grid-cols-2">
      <section class="glass-card rounded-2xl p-5">
        <h3 class="font-semibold text-pretty">运行状态</h3>
        <div v-if="isLoading" class="mt-4 text-sm text-pretty-muted">
          <span class="loading loading-spinner loading-sm" />
        </div>
        <div v-else class="mt-4 grid gap-3 sm:grid-cols-2">
          <div v-for="item in runtimeStatuses" :key="item.key" class="glass-card rounded-xl p-3">
            <div class="flex items-center justify-between gap-2">
              <p class="text-sm font-semibold text-pretty">{{ item.name }}</p>
              <span class="badge rounded-xl badge-sm text-xs px-3 py-2" :class="item.status === 'online' ? 'bg-emerald-500/15 text-emerald-600 border-0' : item.status === 'offline' ? 'bg-error/15 text-error border-0' : 'bg-[var(--glow-primary)] text-pretty-muted border-0'">
                {{ item.status }}
              </span>
            </div>
            <p class="mt-2 text-xs text-pretty-muted">{{ item.detail }}</p>
            <p class="mt-1 truncate font-mono text-[11px] text-pretty-muted">{{ item.endpoint }}</p>
          </div>
        </div>
        <dl class="mt-5 space-y-2 text-sm">
          <div class="flex justify-between gap-3 rounded-xl border border-[var(--border-subtle)] bg-[var(--bg-base)]/50 px-3 py-2.5">
            <dt class="text-pretty-muted">Admin ID</dt>
            <dd class="font-mono text-pretty">{{ adminIdentity?.id || '-' }}</dd>
          </div>
          <div class="flex justify-between gap-3 rounded-xl border border-[var(--border-subtle)] bg-[var(--bg-base)]/50 px-3 py-2.5">
            <dt class="text-pretty-muted">Admin Email</dt>
            <dd class="font-mono text-pretty">{{ adminIdentity?.email || '-' }}</dd>
          </div>
          <div class="flex justify-between gap-3 rounded-xl border border-[var(--border-subtle)] bg-[var(--bg-base)]/50 px-3 py-2.5">
            <dt class="text-pretty-muted">Role</dt>
            <dd class="font-mono text-pretty">{{ adminIdentity?.role || '-' }}</dd>
          </div>
        </dl>
      </section>

      <section class="glass-card rounded-2xl p-5">
        <h3 class="font-semibold text-pretty">配置面板</h3>
        <fieldset class="fieldset mt-4 rounded-xl border border-[var(--border-subtle)] bg-[var(--bg-base)]/50 p-4">
          <legend class="px-2 text-pretty-secondary">系统设置</legend>

          <label class="label"><span class="label-text text-pretty-secondary">应用名称</span></label>
          <input v-model="settings.appName" type="text" class="glass-input input input-bordered w-full rounded-xl" placeholder="DeepWrite Admin" />
          <p class="label"><span class="label-text-alt text-pretty-muted">用于后台系统标题展示</span></p>

          <label class="label mt-3"><span class="label-text text-pretty-secondary">运行环境</span></label>
          <select v-model="settings.appEnv" class="glass-input select select-bordered w-full rounded-xl">
            <option value="development">development</option>
            <option value="staging">staging</option>
            <option value="production">production</option>
          </select>
          <p class="label"><span class="label-text-alt text-pretty-muted">影响后台展示的环境标识</span></p>

          <label class="label mt-3"><span class="label-text text-pretty-secondary">默认 Provider</span></label>
          <input v-model="settings.defaultProvider" type="text" class="glass-input input input-bordered w-full rounded-xl" placeholder="openai-compatible" />
          <p class="label"><span class="label-text-alt text-pretty-muted">新配置项的默认 provider 建议值</span></p>

          <fieldset class="fieldset mt-4 rounded-xl border border-[var(--border-subtle)] bg-[var(--bg-base)]/30 p-3">
            <legend class="px-2 text-pretty-secondary">运行策略</legend>
            <label class="label cursor-pointer justify-start gap-2">
              <input v-model="settings.strictAdminMode" type="checkbox" class="toggle toggle-sm" />
              <span class="label-text text-pretty-secondary">严格管理员模式</span>
            </label>
            <label class="label cursor-pointer justify-start gap-2 mt-2">
              <input v-model="settings.maintenanceMode" type="checkbox" class="toggle toggle-sm" />
              <span class="label-text text-pretty-secondary">维护模式</span>
            </label>
          </fieldset>

          <div class="mt-5">
            <button class="btn rounded-xl bg-[var(--glow-primary)] text-pretty hover:opacity-90" :disabled="isSaving" @click="saveSettings">
              <span v-if="isSaving" class="loading loading-spinner loading-xs" />
              <span>{{ isSaving ? '保存中...' : '保存配置' }}</span>
            </button>
          </div>
        </fieldset>
      </section>
    </div>
  </section>
</template>
