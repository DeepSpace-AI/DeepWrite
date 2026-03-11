<script setup lang="ts">
import { computed, onBeforeMount } from 'vue'
import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'
import { useI18n } from 'vue-i18n'

const authStore = useAuthStore()
const userStore = useUserStore()
const { t } = useI18n()

const { isAuthenticated } = storeToRefs(authStore)
const { user } = storeToRefs(userStore)

const showWorkbench = computed(() => isAuthenticated.value && !!user.value?.id)

// 刷新/重新进入时兜底初始化，确保导航栏状态正确
onBeforeMount(() => {
  authStore.initializeFromStorage()
  userStore.initializeFromStorage()
})
</script>

<template>
  <div class="min-h-screen bg-base-100">
    <header class="sticky top-0 z-50 bg-base-100/90 backdrop-blur-sm border-b border-base-300/50">
      <div class="navbar mx-auto max-w-6xl px-5 lg:px-6">
        <div class="navbar-start">
          <RouterLink to="/" class="btn btn-ghost text-lg font-bold tracking-tight px-2">
            DeepWrite
          </RouterLink>
        </div>
        <div class="navbar-center hidden md:flex">
          <ul class="menu menu-horizontal text-sm px-1">
            <li><RouterLink :to="{ name: 'home', hash: '#demo' }">{{ t('nav.demo') }}</RouterLink></li>
            <li><RouterLink :to="{ name: 'home', hash: '#integrations' }">{{ t('nav.integrations') }}</RouterLink></li>
            <li><RouterLink :to="{ name: 'home', hash: '#pricing' }">{{ t('nav.pricing') }}</RouterLink></li>
          </ul>
        </div>
        <div class="navbar-end gap-2">
          <RouterLink :to="{ name: 'dashboard' }" class="btn btn-ghost btn-sm" v-if="showWorkbench"
            >{{ t('nav.workbench') }}</RouterLink
          >
          <template v-else>
            <RouterLink :to="{ name: 'login' }" class="btn btn-ghost btn-sm">{{ t('nav.login') }}</RouterLink>
            <RouterLink :to="{ name: 'register' }" class="btn btn-primary btn-sm"
              >{{ t('nav.startFree') }}</RouterLink
            >
          </template>
        </div>
      </div>
    </header>
    <RouterView />
  </div>
</template>
