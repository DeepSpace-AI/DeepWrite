<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ApiError } from '@/api/http'
import { fetchAdminStats } from '@/api/admin/dashboard'
import IconUserGroup from '~icons/mdi/account-group'
import IconFolderMultiple from '~icons/mdi/folder-multiple'
import IconFileDocument from '~icons/mdi/file-document'
import IconAttachment from '~icons/mdi/paperclip'
import IconStar from '~icons/mdi/star'
import IconTrendingUp from '~icons/mdi/trending-up'
import IconDatabase from '~icons/mdi/database'
import IconRefresh from '~icons/mdi/refresh'

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

const metricCards = computed(() => [
  { key: 'users', label: '总用户数', value: stats.value.total_users, icon: IconUserGroup, glow: 'bg-[var(--glow-primary)]' },
  { key: 'workspaces', label: '总工作区', value: stats.value.total_workspaces, icon: IconFolderMultiple, glow: 'bg-[var(--glow-secondary)]' },
  { key: 'documents', label: '总文档', value: stats.value.total_documents, icon: IconFileDocument, glow: 'bg-[var(--glow-primary)]' },
  { key: 'files', label: '总文件', value: stats.value.total_files, icon: IconAttachment, glow: 'bg-[var(--glow-secondary)]' },
  { key: 'active', label: '活跃工作区', value: stats.value.active_workspaces, icon: IconTrendingUp, glow: 'bg-[var(--glow-primary)]' },
  { key: 'online', label: '今日活跃', value: stats.value.active_users_today, icon: IconStar, glow: 'bg-[var(--glow-secondary)]' },
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
    <div class="glass-card rounded-2xl p-6">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p class="text-[11px] font-medium uppercase tracking-[0.2em] text-pretty-muted">Admin Dashboard</p>
          <h2 class="mt-2 text-3xl font-bold text-pretty">系统概览</h2>
          <p class="mt-2 text-sm text-pretty-secondary">DeepWrite 平台运行状态一览</p>
        </div>
        <button class="btn btn-ghost btn-sm rounded-xl" :disabled="isLoading" @click="loadStats">
          <IconRefresh class="h-4 w-4" />
          刷新
        </button>
      </div>
    </div>

    <div v-if="errorMessage" class="glass-card rounded-xl p-4 text-sm text-error border border-error/20 bg-error/5">{{ errorMessage }}</div>

    <div v-if="isLoading && !stats.total_users" class="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
      <div v-for="i in 6" :key="i" class="skeleton h-32 rounded-2xl" />
    </div>

    <div v-else class="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
      <article
        v-for="item in metricCards"
        :key="item.key"
        class="glass-card group relative overflow-hidden rounded-2xl p-5 transition-all duration-300 hover:-translate-y-1 hover:shadow-lg"
      >
        <div class="absolute -right-4 -top-4 h-20 w-20 rounded-full opacity-20 blur-xl transition-all group-hover:opacity-30" :class="item.glow" />
        <div class="flex items-start justify-between">
          <div>
            <p class="text-[11px] uppercase tracking-[0.15em] text-pretty-muted">{{ item.label }}</p>
            <p class="mt-2 text-3xl font-bold text-pretty">{{ formatNumber(item.value) }}</p>
          </div>
          <span class="inline-flex h-11 w-11 items-center justify-center rounded-xl" :class="item.glow">
            <component :is="item.icon" class="h-5 w-5 text-pretty" />
          </span>
        </div>
      </article>

      <article class="glass-card group relative overflow-hidden rounded-2xl p-5 transition-all duration-300 hover:-translate-y-1 hover:shadow-lg">
        <div class="absolute -right-4 -top-4 h-20 w-20 rounded-full bg-yellow-500/20 opacity-20 blur-xl transition-all group-hover:opacity-30" />
        <div class="flex items-start justify-between">
          <div>
            <p class="text-[11px] uppercase tracking-[0.15em] text-pretty-muted">存储使用</p>
            <p class="mt-2 text-3xl font-bold text-pretty">{{ formatBytes(stats.storage_used_bytes) }}</p>
          </div>
          <span class="inline-flex h-11 w-11 items-center justify-center rounded-xl bg-yellow-500/20">
            <IconDatabase class="h-5 w-5 text-pretty" />
          </span>
        </div>
      </article>
    </div>

    <div class="grid gap-5 lg:grid-cols-2">
      <article class="glass-card rounded-2xl p-5">
        <h3 class="text-lg font-semibold text-pretty">快捷操作</h3>
        <div class="mt-4 grid grid-cols-2 gap-3">
          <RouterLink :to="{ name: 'users' }" class="btn btn-outline btn-sm justify-start rounded-xl border-[var(--border-subtle)] text-pretty-secondary hover:bg-[var(--glow-primary)] hover:text-pretty hover:border-[var(--border-subtle)]">
            <IconUserGroup class="h-4 w-4" />
            用户管理
          </RouterLink>
          <RouterLink :to="{ name: 'workspaces' }" class="btn btn-outline btn-sm justify-start rounded-xl border-[var(--border-subtle)] text-pretty-secondary hover:bg-[var(--glow-primary)] hover:text-pretty hover:border-[var(--border-subtle)]">
            <IconFolderMultiple class="h-4 w-4" />
            工作区列表
          </RouterLink>
          <RouterLink :to="{ name: 'providers' }" class="btn btn-outline btn-sm justify-start rounded-xl border-[var(--border-subtle)] text-pretty-secondary hover:bg-[var(--glow-primary)] hover:text-pretty hover:border-[var(--border-subtle)]">
            <IconStar class="h-4 w-4" />
            AI 配置
          </RouterLink>
          <RouterLink :to="{ name: 'system' }" class="btn btn-outline btn-sm justify-start rounded-xl border-[var(--border-subtle)] text-pretty-secondary hover:bg-[var(--glow-primary)] hover:text-pretty hover:border-[var(--border-subtle)]">
            <IconDatabase class="h-4 w-4" />
            系统设置
          </RouterLink>
        </div>
      </article>

      <article class="glass-card rounded-2xl p-5">
        <h3 class="text-lg font-semibold text-pretty">系统状态</h3>
        <div class="mt-4 space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-sm text-pretty-secondary">Gateway 服务</span>
            <span class="badge rounded-xl bg-emerald-500/15 text-emerald-600 border-0 text-xs px-3 py-2">运行中</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-sm text-pretty-secondary">数据库</span>
            <span class="badge rounded-xl bg-emerald-500/15 text-emerald-600 border-0 text-xs px-3 py-2">正常</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-sm text-pretty-secondary">缓存服务</span>
            <span class="badge rounded-xl bg-emerald-500/15 text-emerald-600 border-0 text-xs px-3 py-2">正常</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-sm text-pretty-secondary">AI 服务</span>
            <span class="badge rounded-xl bg-amber-500/15 text-amber-600 border-0 text-xs px-3 py-2">待配置</span>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>
