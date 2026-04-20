<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ApiError } from '@/api/http'
import { fetchAdminStats } from '@/api/admin/dashboard'
import IconUsers from '~icons/mdi/account-group'
import IconWorkspaces from '~icons/mdi/folder-multiple'
import IconDocuments from '~icons/mdi/file-document-multiple'
import IconFiles from '~icons/mdi/file'
import IconTrending from '~icons/mdi/trending-up'
import IconActive from '~icons/mdi/account-check'
import IconStorage from '~icons/mdi/database'
import IconRefresh from '~icons/mdi/refresh'
import IconChevronRight from '~icons/mdi/chevron-right'

const isLoading = ref(false)
const errorMessage = ref('')

interface Stats {
  total_users: number
  total_workspaces: number
  total_documents: number
  total_files: number
  active_workspaces: number
  active_users_today: number
  storage_used_bytes: number
}

const stats = ref<Stats>({
  total_users: 0,
  total_workspaces: 0,
  total_documents: 0,
  total_files: 0,
  active_workspaces: 0,
  active_users_today: 0,
  storage_used_bytes: 0,
})

const metrics = computed(() => [
  { key: 'users', label: '用户', value: stats.value.total_users, icon: IconUsers },
  { key: 'workspaces', label: '工作区', value: stats.value.total_workspaces, icon: IconWorkspaces },
  { key: 'documents', label: '文档', value: stats.value.total_documents, icon: IconDocuments },
  { key: 'files', label: '文件', value: stats.value.total_files, icon: IconFiles },
])

const activityMetrics = computed(() => [
  { key: 'active', label: '活跃工作区', value: stats.value.active_workspaces, icon: IconTrending },
  { key: 'online', label: '今日活跃用户', value: stats.value.active_users_today, icon: IconActive },
])

function formatBytes(bytes: number) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

function formatNumber(num: number) {
  return new Intl.NumberFormat().format(num)
}

async function loadStats() {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const data = await fetchAdminStats()
    stats.value = {
      total_users: data.total_users || 0,
      total_workspaces: data.total_workspaces || 0,
      total_documents: data.total_documents || 0,
      total_files: data.total_files || 0,
      active_workspaces: data.active_workspaces || 0,
      active_users_today: data.active_users_today || 0,
      storage_used_bytes: data.storage_used_bytes || 0,
    }
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '加载统计数据失败'
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  loadStats()
})
</script>

<template>
  <section class="space-y-6">
    <header class="flex items-center justify-between">
      <div>
        <h1 class="text-display text-2xl">仪表盘</h1>
        <p class="text-secondary mt-1 text-sm">平台运行状态与数据概览</p>
      </div>
      <button 
        class="btn-ghost-tech flex items-center gap-2 px-3 py-2 text-sm"
        :disabled="isLoading"
        @click="loadStats"
      >
        <IconRefresh class="h-4 w-4" :class="{ 'animate-spin': isLoading }" />
        <span>刷新</span>
      </button>
    </header>

    <div v-if="errorMessage" class="tech-card px-4 py-3 text-sm text-[var(--accent-error)]">
      {{ errorMessage }}
    </div>

    <div v-if="isLoading && !stats.total_users" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <div v-for="i in 4" :key="i" class="skeleton-tech h-24" />
    </div>

    <div v-else class="animate-stagger grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <div
        v-for="item in metrics"
        :key="item.key"
        class="tech-card p-4"
      >
        <div class="flex items-center gap-3">
          <span class="flex h-10 w-10 items-center justify-center rounded-lg bg-[var(--surface-elevated)]">
            <component :is="item.icon" class="h-5 w-5 text-[var(--accent-primary)]" />
          </span>
          <div>
            <p class="text-muted text-xs">{{ item.label }}</p>
            <p class="text-display text-xl">{{ formatNumber(item.value) }}</p>
          </div>
        </div>
      </div>
    </div>

    <div class="grid gap-6 lg:grid-cols-3">
      <div class="lg:col-span-2 space-y-4">
        <div class="tech-card p-4">
          <h2 class="text-heading text-base">活跃数据</h2>
          <div class="mt-4 grid gap-3 sm:grid-cols-2">
            <div 
              v-for="item in activityMetrics"
              :key="item.key"
              class="flex items-center gap-3 rounded-lg border border-[var(--border-subtle)] p-3"
            >
              <span class="flex h-9 w-9 items-center justify-center rounded-lg border border-[var(--border-subtle)]">
                <component :is="item.icon" class="h-4 w-4 text-[var(--text-secondary)]" />
              </span>
              <div>
                <p class="text-muted text-xs">{{ item.label }}</p>
                <p class="text-heading text-lg">{{ formatNumber(item.value) }}</p>
              </div>
            </div>
          </div>
        </div>

        <div class="tech-card p-4">
          <h2 class="text-heading text-base">存储使用</h2>
          <div class="mt-4 flex items-center gap-4">
            <span class="flex h-12 w-12 items-center justify-center rounded-lg border border-[var(--border-subtle)]">
              <IconStorage class="h-6 w-6 text-[var(--text-secondary)]" />
            </span>
            <div>
              <p class="text-display text-xl">{{ formatBytes(stats.storage_used_bytes) }}</p>
              <p class="text-muted text-xs">总存储占用</p>
            </div>
          </div>
        </div>
      </div>

      <div class="space-y-4">
        <div class="tech-card p-4">
          <h2 class="text-heading text-base">快捷操作</h2>
          <nav class="mt-3 space-y-1">
            <RouterLink 
              :to="{ name: 'users' }" 
              class="flex items-center justify-between rounded-lg px-3 py-2.5 text-[var(--text-secondary)] transition-colors hover:bg-[var(--surface-hover)] hover:text-[var(--text-primary)]"
            >
              <div class="flex items-center gap-3">
                <IconUsers class="h-4 w-4" />
                <span class="text-sm">用户管理</span>
              </div>
              <IconChevronRight class="h-4 w-4 text-[var(--text-muted)]" />
            </RouterLink>
            <RouterLink 
              :to="{ name: 'workspaces' }" 
              class="flex items-center justify-between rounded-lg px-3 py-2.5 text-[var(--text-secondary)] transition-colors hover:bg-[var(--surface-hover)] hover:text-[var(--text-primary)]"
            >
              <div class="flex items-center gap-3">
                <IconWorkspaces class="h-4 w-4" />
                <span class="text-sm">工作区列表</span>
              </div>
              <IconChevronRight class="h-4 w-4 text-[var(--text-muted)]" />
            </RouterLink>
            <RouterLink 
              :to="{ name: 'providers' }" 
              class="flex items-center justify-between rounded-lg px-3 py-2.5 text-[var(--text-secondary)] transition-colors hover:bg-[var(--surface-hover)] hover:text-[var(--text-primary)]"
            >
              <div class="flex items-center gap-3">
                <IconTrending class="h-4 w-4" />
                <span class="text-sm">AI Provider</span>
              </div>
              <IconChevronRight class="h-4 w-4 text-[var(--text-muted)]" />
            </RouterLink>
            <RouterLink 
              :to="{ name: 'system' }" 
              class="flex items-center justify-between rounded-lg px-3 py-2.5 text-[var(--text-secondary)] transition-colors hover:bg-[var(--surface-hover)] hover:text-[var(--text-primary)]"
            >
              <div class="flex items-center gap-3">
                <IconStorage class="h-4 w-4" />
                <span class="text-sm">系统设置</span>
              </div>
              <IconChevronRight class="h-4 w-4 text-[var(--text-muted)]" />
            </RouterLink>
          </nav>
        </div>

        <div class="tech-card p-4">
          <h2 class="text-heading text-base">系统状态</h2>
          <div class="mt-3 space-y-2">
            <div class="flex items-center justify-between py-2">
              <span class="text-secondary text-sm">Gateway</span>
              <span class="badge-success px-2 py-0.5 text-xs">运行中</span>
            </div>
            <div class="flex items-center justify-between py-2">
              <span class="text-secondary text-sm">数据库</span>
              <span class="badge-success px-2 py-0.5 text-xs">正常</span>
            </div>
            <div class="flex items-center justify-between py-2">
              <span class="text-secondary text-sm">缓存</span>
              <span class="badge-success px-2 py-0.5 text-xs">正常</span>
            </div>
            <div class="flex items-center justify-between py-2">
              <span class="text-secondary text-sm">AI 服务</span>
              <span class="badge-warning px-2 py-0.5 text-xs">待配置</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>