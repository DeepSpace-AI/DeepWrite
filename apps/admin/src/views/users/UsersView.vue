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
  <section class="space-y-4 rounded-sm border border-base-300 bg-base-100 p-5 shadow-sm">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h2 class="text-xl font-semibold text-base-content">用户管理</h2>
        <p class="mt-1 text-sm text-base-content/65">仅展示用户列表信息。</p>
      </div>
      <button class="btn btn-outline btn-sm rounded-sm" :disabled="isLoading" @click="loadUsers">刷新</button>
    </div>

    <div class="flex flex-wrap gap-2">
      <input v-model="keyword" class="input input-bordered input-sm rounded-sm" placeholder="按邮箱/姓名搜索" />
      <select v-model="statusFilter" class="select select-bordered select-sm rounded-sm">
        <option value="all">全部状态</option>
        <option value="active">active</option>
        <option value="inactive">非 active</option>
      </select>
    </div>

    <p v-if="errorMessage" class="rounded-sm border border-error/30 bg-error/10 px-3 py-2 text-sm text-error">{{ errorMessage }}</p>
    <div v-if="isLoading" class="py-8 text-center text-base-content/70">
      <span class="loading loading-spinner loading-md" />
    </div>

    <div v-else class="overflow-x-auto">
      <table class="table table-zebra">
        <thead>
          <tr>
            <th>用户</th>
            <th>邮箱</th>
            <th>状态</th>
            <th>用户ID</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in filteredUsers" :key="user.userId">
            <td class="font-medium">{{ user.displayName }}</td>
            <td class="text-sm text-base-content/70">{{ user.email || '-' }}</td>
            <td><span class="badge badge-outline">{{ user.status }}</span></td>
            <td class="font-mono text-xs">{{ user.userId }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>
