<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ApiError } from '@/api/http'
import { createWorkspace, listWorkspaces, type Workspace } from '@/api/workspace'
import { useUserStore } from '@/stores/user'
import CreateWorkspaceModal from '@/views/workspace/modal/CreateWorkspaceModal.vue'

type WorkspaceRole = 'owner' | 'admin' | 'editor' | 'viewer'
type ViewMode = 'card' | 'table'

interface WorkspaceItem {
  id: string
  name: string
  role: WorkspaceRole
  members: number
  docs: number
  updatedAt: string
  status: 'active' | 'archived'
}

const userStore = useUserStore()

const rawWorkspaces = ref<Workspace[]>([])
const isLoading = ref(false)
const fetchError = ref('')

const viewMode = ref<ViewMode>('card')
const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(6)

const isCreateModalOpen = ref(false)
const isSubmittingCreate = ref(false)
const createErrorMessage = ref('')

const roleLabelMap: Record<WorkspaceRole, string> = {
  owner: '所有者',
  admin: '管理员',
  editor: '编辑',
  viewer: '访客',
}

function normalizeRole(value: string): WorkspaceRole {
  if (value === 'owner' || value === 'admin' || value === 'editor' || value === 'viewer') {
    return value
  }
  return 'viewer'
}

function formatUpdatedAt(value: string): string {
  if (!value) return '未知'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
  return d.toLocaleString('zh-CN', { hour12: false })
}

const workspaceItems = computed<WorkspaceItem[]>(() => {
  const currentUserId = userStore.user?.id || ''

  return rawWorkspaces.value.map((item) => {
    const currentMember = item.members?.find((m) => m.user_id === currentUserId)
    const role = normalizeRole(currentMember?.role || (item.owner_id === currentUserId ? 'owner' : 'viewer'))

    return {
      id: item.id,
      name: item.name,
      role,
      members: item.members?.length || 1,
      docs: 0,
      updatedAt: formatUpdatedAt(item.updated_at),
      status: item.status === 'archived' ? 'archived' : 'active',
    }
  })
})

const filteredWorkspaces = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (!keyword) return workspaceItems.value

  return workspaceItems.value.filter((item) => {
    const roleLabel = roleLabelMap[item.role]
    return (
      item.name.toLowerCase().includes(keyword)
      || roleLabel.toLowerCase().includes(keyword)
      || item.status.toLowerCase().includes(keyword)
    )
  })
})

const totalPages = computed(() => {
  const pages = Math.ceil(filteredWorkspaces.value.length / pageSize.value)
  return pages > 0 ? pages : 1
})

const pagedWorkspaces = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return filteredWorkspaces.value.slice(start, start + pageSize.value)
})

const pageNumbers = computed(() => Array.from({ length: totalPages.value }, (_, i) => i + 1))

async function loadWorkspaces() {
  isLoading.value = true
  fetchError.value = ''

  try {
    rawWorkspaces.value = await listWorkspaces(100, 0)
  } catch (error) {
    fetchError.value = error instanceof ApiError ? error.message : '加载工作空间失败，请稍后重试。'
  } finally {
    isLoading.value = false
  }
}

function switchView(mode: ViewMode) {
  viewMode.value = mode
}

function applySearch() {
  currentPage.value = 1
}

function goToPage(page: number) {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
}

function prevPage() {
  goToPage(currentPage.value - 1)
}

function nextPage() {
  goToPage(currentPage.value + 1)
}

function openCreateModal() {
  createErrorMessage.value = ''
  isCreateModalOpen.value = true
}

async function handleCreateWorkspace(payload: { name: string; description: string; public: boolean }) {
  if (isSubmittingCreate.value) return

  isSubmittingCreate.value = true
  createErrorMessage.value = ''

  try {
    const created = await createWorkspace(payload)
    rawWorkspaces.value = [created, ...rawWorkspaces.value]
    isCreateModalOpen.value = false
    currentPage.value = 1
  } catch (error) {
    createErrorMessage.value = error instanceof ApiError ? error.message : '创建工作空间失败，请稍后重试。'
  } finally {
    isSubmittingCreate.value = false
  }
}

onMounted(() => {
  loadWorkspaces()
})
</script>

<template>
  <section class="space-y-5">
    <section class="rounded-sm border border-base-300 bg-base-100 p-5 shadow-sm">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p class="text-[11px] font-mono uppercase tracking-[0.2em] text-base-content/45">Workspace / Directory</p>
          <h2 class="heading-serif mt-2 text-3xl font-bold text-base-content">工作空间</h2>
          <p class="mt-2 max-w-2xl text-sm leading-7 text-base-content/62">你的项目入口。统一查看成员、文档规模、角色权限与最近活动。</p>
        </div>
        <button type="button" class="btn btn-primary rounded-sm" @click="openCreateModal">新建工作空间</button>
      </div>

      <div class="mt-5 flex flex-wrap items-center gap-3">
        <label class="input input-bordered flex w-full max-w-md items-center gap-2 rounded-sm">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" class="h-4 w-4 text-base-content/50">
            <path d="m21 21-4.3-4.3" />
            <circle cx="11" cy="11" r="6.5" />
          </svg>
          <input
            v-model="searchKeyword"
            type="text"
            class="grow"
            placeholder="搜索工作空间名称、角色或状态"
            @input="applySearch"
          />
        </label>

        <div class="tabs tabs-box rounded-sm border border-base-300 bg-base-200 p-1">
          <button type="button" class="tab rounded-sm" :class="viewMode === 'card' ? 'tab-active' : ''" @click="switchView('card')">
            卡片
          </button>
          <button type="button" class="tab rounded-sm" :class="viewMode === 'table' ? 'tab-active' : ''" @click="switchView('table')">
            表格
          </button>
        </div>
      </div>

      <div class="mt-3 text-xs text-base-content/50">
        共 {{ filteredWorkspaces.length }} 个结果，第 {{ currentPage }} / {{ totalPages }} 页
      </div>
    </section>

    <p v-if="fetchError" class="rounded-sm border border-error/30 bg-error/10 px-4 py-3 text-sm text-error">
      {{ fetchError }}
    </p>

    <section v-if="isLoading" class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <div v-for="skeleton in 3" :key="`skeleton-${skeleton}`" class="h-36 animate-pulse rounded-sm border border-base-300 bg-base-200/60" />
    </section>

    <section v-else-if="viewMode === 'card'" class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <article
        v-for="item in pagedWorkspaces"
        :key="item.id"
        class="rounded-sm border border-base-300 bg-base-100 p-4 shadow-sm"
      >
        <div class="flex items-center justify-between gap-3">
          <h3 class="truncate text-base font-semibold text-base-content">{{ item.name }}</h3>
          <span
            class="badge rounded-sm"
            :class="item.status === 'active' ? 'badge-success badge-outline' : 'badge-ghost'"
          >
            {{ item.status === 'active' ? '活跃' : '归档' }}
          </span>
        </div>

        <div class="mt-3 grid grid-cols-2 gap-2 text-xs">
          <div class="rounded-sm bg-base-200/70 px-3 py-2 text-base-content/70">成员 {{ item.members }}</div>
          <div class="rounded-sm bg-base-200/70 px-3 py-2 text-base-content/70">文档 {{ item.docs }}</div>
        </div>

        <div class="mt-3 flex items-center justify-between text-xs text-base-content/55">
          <span>{{ roleLabelMap[item.role] }}</span>
          <span>更新于 {{ item.updatedAt }}</span>
        </div>
      </article>

      <p v-if="!pagedWorkspaces.length" class="col-span-full rounded-sm border border-base-300 bg-base-100 px-4 py-10 text-center text-sm text-base-content/60">
        当前没有匹配的工作空间。
      </p>
    </section>

    <section v-else-if="!isLoading" class="overflow-hidden rounded-sm border border-base-300 bg-base-100 shadow-sm">
      <div class="border-b border-base-300 px-4 py-3">
        <h3 class="text-sm font-semibold text-base-content">工作空间列表</h3>
      </div>
      <div class="overflow-x-auto">
        <table class="table table-zebra">
          <thead>
            <tr>
              <th>名称</th>
              <th>角色</th>
              <th>成员</th>
              <th>文档</th>
              <th>状态</th>
              <th>最近更新</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in pagedWorkspaces" :key="`${item.id}-row`">
              <td class="font-medium text-base-content">{{ item.name }}</td>
              <td>{{ roleLabelMap[item.role] }}</td>
              <td>{{ item.members }}</td>
              <td>{{ item.docs }}</td>
              <td>
                <span
                  class="badge rounded-sm"
                  :class="item.status === 'active' ? 'badge-success badge-outline' : 'badge-ghost'"
                >
                  {{ item.status === 'active' ? '活跃' : '归档' }}
                </span>
              </td>
              <td>{{ item.updatedAt }}</td>
            </tr>
            <tr v-if="!pagedWorkspaces.length">
              <td colspan="6" class="py-10 text-center text-sm text-base-content/60">当前没有匹配的工作空间。</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="flex flex-wrap items-center justify-end gap-2">
      <button type="button" class="btn btn-sm rounded-sm" :disabled="currentPage <= 1" @click="prevPage">
        上一页
      </button>

      <button
        v-for="page in pageNumbers"
        :key="`page-${page}`"
        type="button"
        class="btn btn-sm rounded-sm"
        :class="page === currentPage ? 'btn-primary' : 'btn-ghost'"
        @click="goToPage(page)"
      >
        {{ page }}
      </button>

      <button type="button" class="btn btn-sm rounded-sm" :disabled="currentPage >= totalPages" @click="nextPage">
        下一页
      </button>
    </section>

    <CreateWorkspaceModal
      v-model="isCreateModalOpen"
      :submitting="isSubmittingCreate"
      :error-message="createErrorMessage"
      @submit="handleCreateWorkspace"
    />
  </section>
</template>
