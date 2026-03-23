<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ApiError } from '@/api/http'
import { listWorkspaces, listWorkspaceMembers, type Workspace } from '@/api/workspace'
import IconRefresh from '~icons/mdi/refresh'

interface WorkspaceDetail extends Workspace {
  member_count: number
  document_count: number
  file_count: number
  owner_email: string
}

const isLoading = ref(false)
const errorMessage = ref('')
const keyword = ref('')
const statusFilter = ref<'all' | 'active' | 'archived'>('all')
const workspaces = ref<WorkspaceDetail[]>([])

const filteredWorkspaces = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  return workspaces.value.filter((item) => {
    const hitKeyword = !q || item.name.toLowerCase().includes(q) || item.owner_email.toLowerCase().includes(q)
    const hitStatus = statusFilter.value === 'all' || item.status === statusFilter.value
    return hitKeyword && hitStatus
  })
})

async function loadWorkspaces() {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const basicList = await listWorkspaces(200, 0)
    const details = await Promise.all(
      basicList.map(async (ws) => {
        try {
          const members = await listWorkspaceMembers(ws.id)
          const owner = members.find((m) => m.is_workspace_owner)
          return {
            ...ws,
            member_count: members.length,
            document_count: 0,
            file_count: 0,
            owner_email: owner?.email || owner?.display_name || '-',
          } as WorkspaceDetail
        } catch {
          return {
            ...ws,
            member_count: 0,
            document_count: 0,
            file_count: 0,
            owner_email: '-',
          } as WorkspaceDetail
        }
      }),
    )
    workspaces.value = details
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '加载工作区失败'
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  loadWorkspaces()
})
</script>

<template>
  <section class="space-y-4">
    <div class="glass-card rounded-2xl p-5">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 class="text-xl font-semibold text-pretty">工作区管理</h2>
          <p class="mt-1 text-sm text-pretty-secondary">平台所有工作区列表与基本信息</p>
        </div>
        <button class="btn btn-ghost btn-sm rounded-xl" :disabled="isLoading" @click="loadWorkspaces">
          <IconRefresh class="h-4 w-4" />
          刷新
        </button>
      </div>

      <div class="mt-4 flex flex-wrap gap-2">
        <input v-model="keyword" class="glass-input input input-bordered input-sm rounded-xl w-full max-w-xs" placeholder="搜索工作区名称/所有者" />
        <select v-model="statusFilter" class="glass-input select select-bordered select-sm rounded-xl">
          <option value="all">全部状态</option>
          <option value="active">active</option>
          <option value="archived">archived</option>
        </select>
      </div>
    </div>

    <div v-if="errorMessage" class="glass-card rounded-xl p-4 text-sm text-error border border-error/20 bg-error/5">{{ errorMessage }}</div>
    <div v-if="isLoading" class="glass-card rounded-xl p-8 text-center">
      <span class="loading loading-spinner loading-md text-pretty-muted" />
    </div>

    <div v-else class="glass-card rounded-2xl overflow-hidden">
      <div class="overflow-x-auto">
        <table class="table">
          <thead>
            <tr class="border-b border-[var(--border-subtle)]">
              <th class="text-pretty-secondary font-medium">工作区</th>
              <th class="text-pretty-secondary font-medium">状态</th>
              <th class="text-pretty-secondary font-medium">成员</th>
              <th class="text-pretty-secondary font-medium">所有者</th>
              <th class="text-pretty-secondary font-medium">ID</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="ws in filteredWorkspaces" :key="ws.id" class="border-b border-[var(--border-subtle)] hover:bg-[var(--bg-elevated)]/50 transition-colors">
              <td class="font-medium text-pretty">{{ ws.name }}</td>
              <td>
                <span class="badge rounded-xl" :class="ws.status === 'active' ? 'bg-emerald-500/15 text-emerald-600 border-0' : 'bg-[var(--glow-primary)] text-pretty-muted border-0'">
                  {{ ws.status }}
                </span>
              </td>
              <td class="text-sm text-pretty-secondary">{{ ws.member_count }}</td>
              <td class="text-sm text-pretty-secondary">{{ ws.owner_email }}</td>
              <td class="font-mono text-xs text-pretty-muted">{{ ws.id }}</td>
            </tr>
            <tr v-if="filteredWorkspaces.length === 0">
              <td colspan="5" class="py-12 text-center text-pretty-muted">暂无数据</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="border-t border-[var(--border-subtle)] px-5 py-3">
        <p class="text-sm text-pretty-muted">共 {{ filteredWorkspaces.length }} 个工作区</p>
      </div>
    </div>
  </section>
</template>
