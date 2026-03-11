<script setup lang="ts">
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/stores/user'
import { useI18n } from 'vue-i18n'

const userStore = useUserStore()
const { user } = storeToRefs(userStore)
const { t } = useI18n()

const userName = computed(() => user.value?.displayName || t('common.writer'))
const userMeta = computed(() => {
  if (!user.value) return t('dashboard.notLoggedIn')
  return `${user.value.role} · ${user.value.language || 'en'} · ${user.value.timezone || 'UTC'}`
})
</script>

<template>
  <section class="space-y-5">
    <section class="rounded-sm border border-base-300 bg-base-100 p-5 shadow-sm">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p class="text-[11px] font-mono uppercase tracking-[0.2em] text-base-content/45">Workbench / Overview</p>
           <h2 class="heading-serif mt-2 text-3xl font-bold text-base-content">{{ t('dashboard.welcomeBack', { name: userName }) }}</h2>
           <p class="mt-2 max-w-2xl text-sm leading-7 text-base-content/62">{{ t('dashboard.subtitle') }}</p>
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
           <h3 class="text-lg font-semibold text-base-content">{{ t('dashboard.workspacesTitle') }}</h3>
           <span class="badge badge-outline rounded-sm">{{ t('dashboard.workspacesCount') }}</span>
        </div>
        <ul class="space-y-2 text-sm text-base-content/70">
           <li class="flex items-center justify-between rounded-sm bg-base-200/70 px-3 py-2"><span>{{ t('dashboard.workspaceItems.item1') }}</span><span class="text-xs">{{ t('dashboard.workspaceItems.item1Meta') }}</span></li>
           <li class="flex items-center justify-between rounded-sm bg-base-200/70 px-3 py-2"><span>{{ t('dashboard.workspaceItems.item2') }}</span><span class="text-xs">{{ t('dashboard.workspaceItems.item2Meta') }}</span></li>
           <li class="flex items-center justify-between rounded-sm bg-base-200/70 px-3 py-2"><span>{{ t('dashboard.workspaceItems.item3') }}</span><span class="text-xs">{{ t('dashboard.workspaceItems.item3Meta') }}</span></li>
        </ul>
      </article>

      <article class="rounded-sm border border-base-300 bg-base-100 p-5 shadow-sm">
        <div class="mb-3 flex items-center justify-between">
           <h3 class="text-lg font-semibold text-base-content">{{ t('dashboard.libraryTitle') }}</h3>
           <span class="badge badge-outline rounded-sm">{{ t('dashboard.libraryCount') }}</span>
        </div>
        <ul class="space-y-2 text-sm text-base-content/70">
           <li class="rounded-sm border border-base-300/70 px-3 py-2">{{ t('dashboard.libraryItems.item1') }}</li>
           <li class="rounded-sm border border-base-300/70 px-3 py-2">{{ t('dashboard.libraryItems.item2') }}</li>
           <li class="rounded-sm border border-base-300/70 px-3 py-2">{{ t('dashboard.libraryItems.item3') }}</li>
        </ul>
      </article>

      <article class="rounded-sm border border-base-300 bg-base-100 p-5 shadow-sm">
        <div class="mb-3 flex items-center justify-between">
           <h3 class="text-lg font-semibold text-base-content">{{ t('dashboard.skillsTitle') }}</h3>
           <span class="badge badge-outline rounded-sm">{{ t('dashboard.skillsEnabled') }}</span>
        </div>
        <div class="grid grid-cols-2 gap-2 text-xs">
           <div class="rounded-sm bg-primary/10 px-3 py-2 text-primary">{{ t('dashboard.skillItems.rewrite') }}</div>
           <div class="rounded-sm bg-secondary/10 px-3 py-2 text-secondary">{{ t('dashboard.skillItems.review') }}</div>
           <div class="rounded-sm bg-accent/10 px-3 py-2 text-accent">{{ t('dashboard.skillItems.outline') }}</div>
           <div class="rounded-sm bg-primary/10 px-3 py-2 text-primary">{{ t('dashboard.skillItems.term') }}</div>
        </div>
      </article>

      <article class="rounded-sm border border-base-300 bg-base-100 p-5 shadow-sm">
        <div class="mb-3 flex items-center justify-between">
           <h3 class="text-lg font-semibold text-base-content">{{ t('dashboard.agentTitle') }}</h3>
           <span class="badge badge-outline rounded-sm">{{ t('dashboard.agentCount') }}</span>
        </div>
        <div class="space-y-2 text-sm text-base-content/70">
           <p class="rounded-sm bg-base-200/70 px-3 py-2">{{ t('dashboard.agentItems.task1') }}</p>
           <p class="rounded-sm bg-base-200/70 px-3 py-2">{{ t('dashboard.agentItems.task2') }}</p>
        </div>
      </article>
    </section>
  </section>
</template>
