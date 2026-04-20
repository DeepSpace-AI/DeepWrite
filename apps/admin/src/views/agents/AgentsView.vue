<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ApiError } from '@/api/http'
import {
  listOfficialAgents,
  deleteAgent,
  updateAgentStatus,
  type Agent,
} from '@/api/agentAdmin'
import IconPlus from '~icons/mdi/plus'
import IconRefresh from '~icons/mdi/refresh'
import IconEdit from '~icons/mdi/pencil'
import IconDelete from '~icons/mdi/delete'
import IconToggleOn from '~icons/mdi/toggle-switch'
import IconToggleOff from '~icons/mdi/toggle-switch-off'
import IconRobot from '~icons/mdi/robot'
import IconResearch from '~icons/mdi/beaker'
import IconWrite from '~icons/mdi/pencil-outline'
import IconData from '~icons/mdi/database'
import IconPublish from '~icons/mdi/book-open-page-variant'

const router = useRouter()

const isLoading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const agents = ref<Agent[]>([])

const categoryOptions = [
  { value: 'all', label: '全部分类' },
  { value: 'research', label: '科研助手' },
  { value: 'writing', label: '写作助手' },
  { value: 'data', label: '数据处理' },
  { value: 'publishing', label: '出版辅助' },
]

const selectedCategory = ref('all')

const filteredAgents = computed(() => {
  if (selectedCategory.value === 'all') {
    return agents.value
  }
  return agents.value.filter(a => a.category === selectedCategory.value)
})

async function loadAgents() {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const res = await listOfficialAgents()
    agents.value = res.items || []
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '加载失败'
  } finally {
    isLoading.value = false
  }
}

async function handleToggleStatus(agent: Agent) {
  try {
    await updateAgentStatus(agent.id, !agent.enabled)
    successMessage.value = agent.enabled ? '已禁用 Agent' : '已启用 Agent'
    await loadAgents()
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '操作失败'
  }
}

async function handleDelete(agent: Agent) {
  if (!confirm(`确定要删除 Agent "${agent.name}" 吗？此操作不可撤销。`)) {
    return
  }
  try {
    await deleteAgent(agent.id)
    successMessage.value = '删除成功'
    await loadAgents()
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '删除失败'
  }
}

function getCategoryIcon(category: string) {
  switch (category) {
    case 'research':
      return IconResearch
    case 'writing':
      return IconWrite
    case 'data':
      return IconData
    case 'publishing':
      return IconPublish
    default:
      return IconRobot
  }
}

function getCategoryLabel(category: string) {
  const found = categoryOptions.find(c => c.value === category)
  return found ? found.label : category
}

onMounted(() => loadAgents())
</script>

<template>
  <section class="space-y-6">
    <header class="flex items-center justify-between">
      <div>
        <h1 class="text-display text-2xl text-[var(--text-primary)]">官方 Agent 管理</h1>
        <p class="text-secondary mt-1 text-sm">管理平台预置的官方 Agent 助手</p>
      </div>
      <div class="flex gap-2">
        <button class="btn-ghost-tech flex items-center gap-1" @click="loadAgents">
          <IconRefresh :class="{ 'animate-spin': isLoading }" class="h-4 w-4" />
          刷新
        </button>
        <button class="btn-tech flex items-center gap-1" @click="router.push('/agents/create')">
          <IconPlus class="h-4 w-4" />
          新建 Agent
        </button>
      </div>
    </header>

    <div v-if="errorMessage" class="tech-card text-sm text-[var(--accent-error)]">
      {{ errorMessage }}
    </div>
    <div v-if="successMessage" class="tech-card bg-[var(--accent-success)]/10 text-sm text-[var(--accent-success)]">
      {{ successMessage }}
    </div>

    <div class="tech-card p-4">
      <div class="flex items-center gap-4">
        <span class="text-label">分类筛选</span>
        <div class="flex gap-2">
          <button
            v-for="cat in categoryOptions"
            :key="cat.value"
            class="rounded-full px-4 py-1.5 text-sm transition-colors"
            :class="selectedCategory === cat.value
              ? 'bg-[var(--accent-primary)] text-white'
              : 'bg-[var(--bg-elevated)] text-[var(--text-secondary)] hover:bg-[var(--bg-card)]'"
            @click="selectedCategory = cat.value"
          >
            {{ cat.label }}
          </button>
        </div>
      </div>
    </div>

    <div class="tech-card overflow-hidden">
      <div v-if="isLoading" class="flex justify-center py-12">
        <span class="loading loading-spinner text-[var(--accent-primary)]" />
      </div>

      <table v-else-if="filteredAgents.length > 0" class="w-full text-sm">
        <thead class="border-b bg-[var(--bg-elevated)]">
          <tr>
            <th class="px-4 py-3 text-left text-label">名称</th>
            <th class="px-4 py-3 text-left text-label">分类</th>
            <th class="px-4 py-3 text-left text-label">描述</th>
            <th class="px-4 py-3 text-center text-label">使用次数</th>
            <th class="px-4 py-3 text-center text-label">评分</th>
            <th class="px-4 py-3 text-center text-label">状态</th>
            <th class="px-4 py-3 text-right text-label">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="agent in filteredAgents"
            :key="agent.id"
            class="border-b hover:bg-[var(--bg-elevated)]"
          >
            <td class="px-4 py-3">
              <div class="flex items-center gap-2">
                <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-[var(--accent-primary)]/10">
                  <component :is="getCategoryIcon(agent.category)" class="h-4 w-4 text-[var(--accent-primary)]" />
                </div>
                <span class="font-medium text-[var(--text-primary)]">{{ agent.name }}</span>
              </div>
            </td>
            <td class="px-4 py-3">
              <span class="rounded-full bg-[var(--bg-elevated)] px-2 py-0.5 text-xs">
                {{ getCategoryLabel(agent.category) }}
              </span>
            </td>
            <td class="max-w-xs truncate px-4 py-3 text-[var(--text-secondary)]">
              {{ agent.description || '-' }}
            </td>
            <td class="px-4 py-3 text-center text-[var(--text-secondary)]">
              {{ agent.usage_count }}
            </td>
            <td class="px-4 py-3 text-center">
              <span class="font-medium text-[var(--accent-warning)]">{{ agent.rating.toFixed(1) }}</span>
            </td>
            <td class="px-4 py-3 text-center">
              <span
                class="rounded-full px-2 py-0.5 text-xs"
                :class="agent.enabled
                  ? 'bg-[var(--accent-success)]/10 text-[var(--accent-success)]'
                  : 'bg-[var(--accent-error)]/10 text-[var(--accent-error)]'"
              >
                {{ agent.enabled ? '启用' : '禁用' }}
              </span>
            </td>
            <td class="px-4 py-3">
              <div class="flex justify-end gap-1">
                <button
                  class="rounded p-1.5 text-[var(--text-secondary)] hover:bg-[var(--bg-card)] hover:text-[var(--accent-primary)]"
                  title="编辑"
                  @click="router.push(`/agents/${agent.id}`)"
                >
                  <IconEdit class="h-4 w-4" />
                </button>
                <button
                  class="rounded p-1.5 text-[var(--text-secondary)] hover:bg-[var(--bg-card)]"
                  :class="agent.enabled ? 'hover:text-[var(--accent-error)]' : 'hover:text-[var(--accent-success)]'"
                  :title="agent.enabled ? '禁用' : '启用'"
                  @click="handleToggleStatus(agent)"
                >
                  <IconToggleOn v-if="agent.enabled" class="h-4 w-4" />
                  <IconToggleOff v-else class="h-4 w-4" />
                </button>
                <button
                  class="rounded p-1.5 text-[var(--text-secondary)] hover:bg-[var(--bg-card)] hover:text-[var(--accent-error)]"
                  title="删除"
                  @click="handleDelete(agent)"
                >
                  <IconDelete class="h-4 w-4" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-else class="flex flex-col items-center justify-center py-16 text-[var(--text-muted)]">
        <IconRobot class="mb-4 h-12 w-12 opacity-50" />
        <p>暂无官方 Agent</p>
        <button class="btn-ghost-tech mt-4" @click="router.push('/agents/create')">
          创建第一个 Agent
        </button>
      </div>
    </div>
  </section>
</template>