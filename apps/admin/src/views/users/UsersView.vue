<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ApiError } from '@/api/http'
import {
  listWorkspaces,
  listWorkspaceMembers,
} from '@/api/workspace'

interface ManagedUser {
  userId: string
  email: string
  displayName: string
  status: string
}

const isLoading = ref(false)
const errorMessage = ref('')
const keyword = ref('')
const statusFilter = ref<'all' | 'active' | 'inactive'>('all')
const users = ref<ManagedUser[]>([])

const filteredUsers = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  return users.value.filter((item) => {
    const hitKeyword = !q || item.email.toLowerCase().includes(q) || item.displayName.toLowerCase().includes(q)
    const hitStatus = statusFilter.value === 'all' || (statusFilter.value === 'active' ? item.status === 'active' : item.status !== 'active')
    return hitKeyword && hitStatus
  })
})

async function loadUsers() {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const workspaces = await listWorkspaces(200, 0)
    const memberResults = await Promise.all(
      workspaces.map(async (workspace) => ({
        workspace,
        members: await listWorkspaceMembers(workspace.id),
      })),
    )

    const map = new Map<string, ManagedUser>()
    memberResults.forEach(({ members }) => {
      members.forEach((member) => {
        if (!member.user_id) return
        const existing = map.get(member.user_id)
        if (existing) {
          if (!existing.email && member.email) {
            existing.email = member.email
          }
          if (!existing.displayName && (member.display_name || member.email)) {
            existing.displayName = member.display_name || member.email || existing.userId
          }
          if (existing.status !== 'active' && member.status === 'active') {
            existing.status = 'active'
          }
          return
        }
        map.set(member.user_id, {
          userId: member.user_id,
          email: member.email || '',
          displayName: member.display_name || member.email || member.user_id,
          status: member.status || 'active',
        })
      })
    })

    users.value = Array.from(map.values()).sort((a, b) => a.displayName.localeCompare(b.displayName, 'zh-CN'))
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '加载用户失败'
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  loadUsers()
})
</script>

<template>
  <section class="space-y-4">
    <div class="glass-card rounded-2xl p-5">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 class="text-xl font-semibold text-pretty">用户管理</h2>
          <p class="mt-1 text-sm text-pretty-secondary">平台所有成员检索与状态治理</p>
        </div>
        <button class="btn btn-ghost btn-sm rounded-xl" :disabled="isLoading" @click="loadUsers">刷新</button>
      </div>

      <div class="mt-4 flex flex-wrap gap-2">
        <input v-model="keyword" class="glass-input input input-bordered input-sm rounded-xl w-full max-w-xs" placeholder="按邮箱/姓名搜索" />
        <select v-model="statusFilter" class="glass-input select select-bordered select-sm rounded-xl">
          <option value="all">全部状态</option>
          <option value="active">active</option>
          <option value="inactive">非 active</option>
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
              <th class="text-pretty-secondary font-medium">用户</th>
              <th class="text-pretty-secondary font-medium">邮箱</th>
              <th class="text-pretty-secondary font-medium">状态</th>
              <th class="text-pretty-secondary font-medium">用户ID</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in filteredUsers" :key="user.userId" class="border-b border-[var(--border-subtle)] hover:bg-[var(--bg-elevated)]/50 transition-colors">
              <td class="font-medium text-pretty">{{ user.displayName }}</td>
              <td class="text-sm text-pretty-secondary">{{ user.email || '-' }}</td>
              <td><span class="badge rounded-xl bg-[var(--glow-primary)] text-pretty border-0 text-xs px-3 py-2">{{ user.status }}</span></td>
              <td class="font-mono text-xs text-pretty-muted">{{ user.userId }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>
