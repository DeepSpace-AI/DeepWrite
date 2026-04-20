<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ApiError } from '@/api/http'
import {
  listPublicAgents,
  moderateAgent,
  type Agent,
} from '@/api/agentAdmin'
import IconRefresh from '~icons/mdi/refresh'
import IconCheck from '~icons/mdi/check'
import IconClose from '~icons/mdi/close'
import IconEye from '~icons/mdi/eye'
import IconEarth from '~icons/mdi/earth'
import IconRobot from '~icons/mdi/robot'

const isLoading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const agents = ref<Agent[]>([])

async function loadAgents() {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const res = await listPublicAgents()
    agents.value = res.items || []
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '加载失败'
  } finally {
    isLoading.value = false
  }
}

async function handleModerate(agent: Agent, enabled: boolean) {
  const action = enabled ? '通过' : '下架'
  if (!confirm(`确定要${action} Agent "${agent.name}" 吗？`)) {
    return
  }
  try {
    await moderateAgent(agent.id, enabled)
    successMessage.value = enabled ? '已通过审核' : '已下架'
    await loadAgents()
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '操作失败'
  }
}

onMounted(() => loadAgents())
</script>

<template>
  <section class="space-y-6">
    <header class="flex items-center justify-between">
      <div>
        <h1 class="text-display text-2xl text-[var(--text-primary)]">公开 Agent 审核</h1>
        <p class="text-secondary mt-1 text-sm">审核和管理用户公开分享的 Agent</p>
      </div>
      <button class="btn-ghost-tech flex items-center gap-1" @click="loadAgents">
        <IconRefresh :class="{ 'animate-spin': isLoading }" class="h-4 w-4" />
        刷新
      </button>
    </header>

    <div v-if="errorMessage" class="tech-card text-sm text-[var(--accent-error)]">
      {{ errorMessage }}
    </div>
    <div v-if="successMessage" class="tech-card bg-[var(--accent-success)]/10 text-sm text-[var(--accent-success)]">
      {{ successMessage }}
    </div>

    <div class="tech-card overflow-hidden">
      <div v-if="isLoading" class="flex justify-center py-12">
        <span class="loading loading-spinner text-[var(--accent-primary)]" />
      </div>

      <table v-else-if="agents.length > 0" class="w-full text-sm">
        <thead class="border-b bg-[var(--bg-elevated)]">
          <tr>
            <th class="px-4 py-3 text-left text-label">名称</th>
            <th class="px-4 py-3 text-left text-label">创建者</th>
            <th class="px-4 py-3 text-left text-label">描述</th>
            <th class="px-4 py-3 text-center text-label">评分</th>
            <th class="px-4 py-3 text-center text-label">使用次数</th>
            <th class="px-4 py-3 text-center text-label">状态</th>
            <th class="px-4 py-3 text-right text-label">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="agent in agents"
            :key="agent.id"
            class="border-b hover:bg-[var(--bg-elevated)]"
          >
            <td class="px-4 py-3">
              <div class="flex items-center gap-2">
                <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-[var(--accent-secondary)]/10">
                  <IconEarth class="h-4 w-4 text-[var(--accent-secondary)]" />
                </div>
                <span class="font-medium text-[var(--text-primary)]">{{ agent.name }}</span>
              </div>
            </td>
            <td class="px-4 py-3 text-[var(--text-secondary)]">
              {{ agent.owner_id ? agent.owner_id.substring(0, 8) + '...' : '-' }}
            </td>
            <td class="max-w-xs truncate px-4 py-3 text-[var(--text-secondary)]">
              {{ agent.description || '-' }}
            </td>
            <td class="px-4 py-3 text-center">
              <span class="font-medium text-[var(--accent-warning)]">{{ agent.rating.toFixed(1) }}</span>
            </td>
            <td class="px-4 py-3 text-center text-[var(--text-secondary)]">
              {{ agent.usage_count }}
            </td>
            <td class="px-4 py-3 text-center">
              <span
                class="rounded-full px-2 py-0.5 text-xs"
                :class="agent.enabled
                  ? 'bg-[var(--accent-success)]/10 text-[var(--accent-success)]'
                  : 'bg-[var(--accent-error)]/10 text-[var(--accent-error)]'"
              >
                {{ agent.enabled ? '已通过' : '已下架' }}
              </span>
            </td>
            <td class="px-4 py-3">
              <div class="flex justify-end gap-1">
                <button
                  class="rounded p-1.5 text-[var(--text-secondary)] hover:bg-[var(--bg-card)] hover:text-[var(--accent-primary)]"
                  title="查看详情"
                >
                  <IconEye class="h-4 w-4" />
                </button>
                <button
                  v-if="!agent.enabled"
                  class="rounded p-1.5 text-[var(--text-secondary)] hover:bg-[var(--bg-card)] hover:text-[var(--accent-success)]"
                  title="通过审核"
                  @click="handleModerate(agent, true)"
                >
                  <IconCheck class="h-4 w-4" />
                </button>
                <button
                  v-else
                  class="rounded p-1.5 text-[var(--text-secondary)] hover:bg-[var(--bg-card)] hover:text-[var(--accent-error)]"
                  title="下架"
                  @click="handleModerate(agent, false)"
                >
                  <IconClose class="h-4 w-4" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-else class="flex flex-col items-center justify-center py-16 text-[var(--text-muted)]">
        <IconRobot class="mb-4 h-12 w-12 opacity-50" />
        <p>暂无公开 Agent</p>
      </div>
    </div>
  </section>
</template>