<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ApiError } from '@/api/http'
import { fetchUsers, updateUser, deleteUser, resetUserPassword, type UserProfile } from '@/api/admin/user'
import IconRefresh from '~icons/mdi/refresh'
import IconAccountOff from '~icons/mdi/account-off'
import IconAccountCheck from '~icons/mdi/account-check'
import IconKeyVariant from '~icons/mdi/key-variant'
import IconDelete from '~icons/mdi/delete'
import IconDotsVertical from '~icons/mdi/dots-vertical'

const isLoading = ref(false)
const isUpdating = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const keyword = ref('')
const statusFilter = ref<string>('all')
const page = ref(1)
const pageSize = ref(50)
const actionMenuId = ref<string | null>(null)
const resetPasswordResult = ref<{ email: string; password: string } | null>(null)
const showResetPasswordModal = ref(false)

const users = ref<UserProfile[]>([])
const total = ref(0)

const filteredUsers = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  return users.value.filter((item) => {
    const matchKeyword = !q ||
      item.email.toLowerCase().includes(q) ||
      (item.display_name || '').toLowerCase().includes(q)
    const matchStatus = statusFilter.value === 'all' || item.status === statusFilter.value
    return matchKeyword && matchStatus
  })
})

const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

const statusOptions = [
  { value: 'all', label: '全部' },
  { value: 'active', label: '正常' },
  { value: 'inactive', label: '禁用' },
]

function formatDate(dateStr?: string) {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

async function loadUsers() {
  isLoading.value = true
  errorMessage.value = ''
  successMessage.value = ''
  actionMenuId.value = null
  try {
    const params = {
      limit: pageSize.value,
      offset: (page.value - 1) * pageSize.value,
      keyword: keyword.value.trim() || undefined,
      status: statusFilter.value !== 'all' ? statusFilter.value : undefined,
    }
    const data = await fetchUsers(params)
    users.value = data.users || []
    total.value = data.total || 0
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '加载用户失败'
    users.value = []
    total.value = 0
  } finally {
    isLoading.value = false
  }
}

async function toggleUserStatus(user: UserProfile) {
  const newStatus = user.status === 'active' ? 'inactive' : 'active'
  if (!confirm(`确定要${newStatus === 'active' ? '启用' : '禁用'}用户 ${user.email} 吗？`)) {
    return
  }

  isUpdating.value = true
  errorMessage.value = ''
  successMessage.value = ''
  actionMenuId.value = null
  try {
    await updateUser(user.id, { status: newStatus })
    user.status = newStatus
    successMessage.value = `用户 ${user.email} 已${newStatus === 'active' ? '启用' : '禁用'}`
    setTimeout(() => { successMessage.value = '' }, 3000)
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '更新失败'
  } finally {
    isUpdating.value = false
  }
}

async function handleResetPassword(user: UserProfile) {
  if (!confirm(`确定要重置用户 ${user.email} 的密码吗？`)) {
    actionMenuId.value = null
    return
  }

  isUpdating.value = true
  errorMessage.value = ''
  successMessage.value = ''
  actionMenuId.value = null
  try {
    const result = await resetUserPassword(user.id)
    resetPasswordResult.value = { email: result.email, password: result.new_password }
    showResetPasswordModal.value = true
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '重置密码失败'
  } finally {
    isUpdating.value = false
  }
}

async function handleDeleteUser(user: UserProfile) {
  if (!confirm(`确定要删除用户 ${user.email} 吗？此操作不可恢复！`)) {
    actionMenuId.value = null
    return
  }

  if (!confirm(`再次确认：删除用户 ${user.email} 的所有数据将被永久删除！`)) {
    actionMenuId.value = null
    return
  }

  isUpdating.value = true
  errorMessage.value = ''
  successMessage.value = ''
  actionMenuId.value = null
  try {
    await deleteUser(user.id)
    successMessage.value = `用户 ${user.email} 已删除`
    setTimeout(() => { successMessage.value = '' }, 3000)
    await loadUsers()
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '删除失败'
  } finally {
    isUpdating.value = false
  }
}

function toggleActionMenu(userId: string) {
  actionMenuId.value = actionMenuId.value === userId ? null : userId
}

function closeActionMenu() {
  actionMenuId.value = null
}

function handleSearch() {
  page.value = 1
  loadUsers()
}

function handlePageChange(newPage: number) {
  if (newPage < 1 || newPage > totalPages.value) return
  page.value = newPage
  loadUsers()
}

onMounted(() => {
  loadUsers()
})
</script>

<template>
  <section class="space-y-4 rounded-sm border border-base-300 bg-base-100 p-5 shadow-sm">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h2 class="text-xl font-semibold text-base-content">用户管理</h2>
        <p class="mt-1 text-sm text-base-content/65">直接管理平台用户，支持启用/禁用、密码重置等操作</p>
      </div>
      <button class="btn btn-outline btn-sm rounded-sm" :disabled="isLoading" @click="loadUsers">
        <IconRefresh class="h-4 w-4" />
        刷新
      </button>
    </div>

    <div class="flex flex-wrap items-center gap-3">
      <input
        v-model="keyword"
        class="input input-bordered input-sm rounded-sm w-64"
        placeholder="搜索邮箱/名称"
        @keyup.enter="handleSearch"
      />
      <select v-model="statusFilter" class="select select-bordered select-sm rounded-sm" @change="handleSearch">
        <option v-for="opt in statusOptions" :key="opt.value" :value="opt.value">
          {{ opt.label }}
        </option>
      </select>
      <button class="btn btn-primary btn-sm rounded-sm" @click="handleSearch">搜索</button>
    </div>

    <div v-if="errorMessage" class="alert alert-error rounded-sm text-sm py-2">{{ errorMessage }}</div>
    <div v-if="successMessage" class="alert alert-success rounded-sm text-sm py-2">{{ successMessage }}</div>

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
            <th>角色</th>
            <th>注册时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in filteredUsers" :key="user.id" class="relative">
            <td>
              <div class="flex items-center gap-3">
                <div class="avatar placeholder">
                  <div class="w-8 rounded-full bg-neutral text-neutral-content text-sm">
                    <span v-if="user.avatar_url">
                      <img :src="user.avatar_url" :alt="user.display_name" />
                    </span>
                    <span v-else>{{ (user.display_name || user.email || '?').slice(0, 1).toUpperCase() }}</span>
                  </div>
                </div>
                <span class="font-medium">{{ user.display_name || '-' }}</span>
              </div>
            </td>
            <td class="text-sm">{{ user.email }}</td>
            <td>
              <span
                class="badge badge-sm"
                :class="user.status === 'active' ? 'badge-success' : 'badge-error'"
              >
                {{ user.status === 'active' ? '正常' : '禁用' }}
              </span>
            </td>
            <td>
              <span class="badge badge-outline badge-sm">{{ user.role || 'user' }}</span>
            </td>
            <td class="text-sm text-base-content/70 whitespace-nowrap">{{ formatDate(user.created_at) }}</td>
            <td class="relative">
              <div class="dropdown dropdown-end">
                <button
                  tabindex="0"
                  class="btn btn-ghost btn-xs rounded-sm"
                  :disabled="isUpdating"
                  @click="toggleActionMenu(user.id)"
                >
                  <IconDotsVertical class="h-4 w-4" />
                </button>
                <ul
                  v-if="actionMenuId === user.id"
                  tabindex="0"
                  class="menu dropdown-content dropdown-left z-10 mt-1 w-40 rounded-sm border border-base-300 bg-base-100 p-2 shadow-lg"
                >
                  <li>
                    <button
                      class="flex items-center gap-2 text-sm"
                      :disabled="isUpdating || user.role === 'admin'"
                      @click="toggleUserStatus(user)"
                    >
                      <component
                        :is="user.status === 'active' ? IconAccountOff : IconAccountCheck"
                        class="h-4 w-4"
                      />
                      {{ user.status === 'active' ? '禁用' : '启用' }}
                    </button>
                  </li>
                  <li>
                    <button
                      class="flex items-center gap-2 text-sm"
                      :disabled="isUpdating"
                      @click="handleResetPassword(user)"
                    >
                      <IconKeyVariant class="h-4 w-4" />
                      重置密码
                    </button>
                  </li>
                  <li class="border-t border-base-300">
                    <button
                      class="flex items-center gap-2 text-error"
                      :disabled="isUpdating || user.role === 'admin'"
                      @click="handleDeleteUser(user)"
                    >
                      <IconDelete class="h-4 w-4" />
                      删除用户
                    </button>
                  </li>
                </ul>
              </div>
              <div
                v-if="actionMenuId === user.id"
                class="fixed inset-0 z-0"
                @click="closeActionMenu"
              />
            </td>
          </tr>
          <tr v-if="filteredUsers.length === 0">
            <td colspan="6" class="text-center py-8 text-base-content/50">暂无数据</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="total > pageSize" class="flex items-center justify-between pt-2">
      <span class="text-sm text-base-content/60">共 {{ total }} 条，第 {{ page }} / {{ totalPages }} 页</span>
      <div class="join">
        <button
          class="join-item btn btn-sm rounded-sm"
          :disabled="page <= 1"
          @click="handlePageChange(page - 1)"
        >
          ‹
        </button>
        <button class="join-item btn btn-sm rounded-sm btn-active">{{ page }}</button>
        <button
          class="join-item btn btn-sm rounded-sm"
          :disabled="page >= totalPages"
          @click="handlePageChange(page + 1)"
        >
          ›
        </button>
      </div>
    </div>
  </section>

  <dialog :open="showResetPasswordModal" class="modal">
    <div class="modal-box rounded-sm">
      <h3 class="font-bold text-lg">密码重置成功</h3>
      <div class="py-4 space-y-3">
        <p class="text-sm">用户 <strong>{{ resetPasswordResult?.email }}</strong> 的密码已重置为：</p>
        <div class="alert alert-warning rounded-sm">
          <code class="text-lg font-mono break-all">{{ resetPasswordResult?.password }}</code>
        </div>
        <p class="text-xs text-base-content/70">请将此密码安全地发送给用户，并建议用户在登录后立即修改密码。</p>
      </div>
      <div class="modal-action">
        <button class="btn btn-primary btn-sm rounded-sm" @click="showResetPasswordModal = false">关闭</button>
      </div>
    </div>
    <div class="modal-backdrop bg-black/30" @click="showResetPasswordModal = false" />
  </dialog>
</template>