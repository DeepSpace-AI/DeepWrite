<script setup lang="ts">
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const { user } = storeToRefs(userStore)

const userName = computed(() => user.value?.displayName || '写作者')
const userMeta = computed(() => {
  if (!user.value) return '未登录'
  return `${user.value.role} · ${user.value.language || 'en'} · ${user.value.timezone || 'UTC'}`
})
</script>

<template>
  <section class="space-y-5">
    <section class="rounded-sm border border-base-300 bg-base-100 p-5 shadow-sm">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p class="text-[11px] font-mono uppercase tracking-[0.2em] text-base-content/45">Workbench / Overview</p>
          <h2 class="heading-serif mt-2 text-3xl font-bold text-base-content">欢迎回来，{{ userName }}</h2>
          <p class="mt-2 max-w-2xl text-sm leading-7 text-base-content/62">在同一工作台内管理协作空间、研究文献、技能能力与 Agent 助手，让编辑流程和知识沉淀保持连续。</p>
        </div>
        <div class="rounded-sm border border-secondary/30 bg-secondary/10 px-4 py-3 text-xs text-secondary-content/90">
          <p class="font-medium">{{ user?.email || 'guest@deepwrite.local' }}</p>
          <p class="mt-1 opacity-80">{{ userMeta }}</p>
        </div>
      </div>
    </section>

    <section class="grid gap-5 md:grid-cols-2">
      <article class="rounded-sm border border-base-300 bg-base-100 p-5 shadow-sm">
        <div class="mb-3 flex items-center justify-between">
          <h3 class="text-lg font-semibold text-base-content">工作空间</h3>
          <span class="badge badge-outline rounded-sm">6 个活跃项目</span>
        </div>
        <ul class="space-y-2 text-sm text-base-content/70">
          <li class="flex items-center justify-between rounded-sm bg-base-200/70 px-3 py-2"><span>产品路线图 2026</span><span class="text-xs">3 人在线</span></li>
          <li class="flex items-center justify-between rounded-sm bg-base-200/70 px-3 py-2"><span>研究方法白皮书</span><span class="text-xs">待评审</span></li>
          <li class="flex items-center justify-between rounded-sm bg-base-200/70 px-3 py-2"><span>发布稿件 Issue 04</span><span class="text-xs">今日更新</span></li>
        </ul>
      </article>

      <article class="rounded-sm border border-base-300 bg-base-100 p-5 shadow-sm">
        <div class="mb-3 flex items-center justify-between">
          <h3 class="text-lg font-semibold text-base-content">文献库</h3>
          <span class="badge badge-outline rounded-sm">128 条条目</span>
        </div>
        <ul class="space-y-2 text-sm text-base-content/70">
          <li class="rounded-sm border border-base-300/70 px-3 py-2">可检索引文与注释摘要，支持导出到写作上下文。</li>
          <li class="rounded-sm border border-base-300/70 px-3 py-2">按主题分类：LLM、协作系统、知识管理、科研写作。</li>
          <li class="rounded-sm border border-base-300/70 px-3 py-2">最近同步：15 分钟前。</li>
        </ul>
      </article>

      <article class="rounded-sm border border-base-300 bg-base-100 p-5 shadow-sm">
        <div class="mb-3 flex items-center justify-between">
          <h3 class="text-lg font-semibold text-base-content">技能市场</h3>
          <span class="badge badge-outline rounded-sm">9 项已启用</span>
        </div>
        <div class="grid grid-cols-2 gap-2 text-xs">
          <div class="rounded-sm bg-primary/10 px-3 py-2 text-primary">学术改写</div>
          <div class="rounded-sm bg-secondary/10 px-3 py-2 text-secondary">文献综述</div>
          <div class="rounded-sm bg-accent/10 px-3 py-2 text-accent">结构提纲</div>
          <div class="rounded-sm bg-primary/10 px-3 py-2 text-primary">术语统一</div>
        </div>
      </article>

      <article class="rounded-sm border border-base-300 bg-base-100 p-5 shadow-sm">
        <div class="mb-3 flex items-center justify-between">
          <h3 class="text-lg font-semibold text-base-content">Agent 助手</h3>
          <span class="badge badge-outline rounded-sm">2 个任务执行中</span>
        </div>
        <div class="space-y-2 text-sm text-base-content/70">
          <p class="rounded-sm bg-base-200/70 px-3 py-2">Agent 01 正在整理引文格式并修复参考文献顺序。</p>
          <p class="rounded-sm bg-base-200/70 px-3 py-2">Agent 02 正在检查章节逻辑并提出结构优化建议。</p>
        </div>
      </article>
    </section>
  </section>
</template>
