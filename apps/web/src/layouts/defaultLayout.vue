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

onBeforeMount(() => {
  authStore.initializeFromStorage()
  userStore.initializeFromStorage()
})
</script>

<template>
  <div class="min-h-screen bg-dot-grid px-4 pb-6">
    <header class="paper-card-static sticky top-4 z-50 mx-auto mt-4 max-w-6xl rounded-lg">
      <nav class="mx-auto flex max-w-6xl items-center justify-between px-5 py-4 lg:px-6">
        <div class="flex items-center gap-8">
          <RouterLink to="/" class="text-editorial text-xl font-semibold tracking-tight">
            DeepWrite
          </RouterLink>
          <ul class="hidden md:flex items-center gap-1">
            <li>
              <RouterLink
                :to="{ name: 'home', hash: '#demo' }"
                class="label-md rounded-md px-4 py-2 hover:bg-[var(--glow-primary)] transition-colors"
              >
                {{ t('nav.demo') }}
              </RouterLink>
            </li>
            <li>
              <RouterLink
                :to="{ name: 'home', hash: '#integrations' }"
                class="label-md rounded-md px-4 py-2 hover:bg-[var(--glow-primary)] transition-colors"
              >
                {{ t('nav.integrations') }}
              </RouterLink>
            </li>
            <li>
              <RouterLink
                :to="{ name: 'home', hash: '#pricing' }"
                class="label-md rounded-md px-4 py-2 hover:bg-[var(--glow-primary)] transition-colors"
              >
                {{ t('nav.pricing') }}
              </RouterLink>
            </li>
          </ul>
        </div>
        <div class="flex items-center gap-3">
          <RouterLink
            v-if="showWorkbench"
            :to="{ name: 'dashboard' }"
            class="label-md rounded-md px-4 py-2 hover:bg-[var(--glow-primary)] transition-colors"
          >
            {{ t('nav.workbench') }}
          </RouterLink>
          <template v-else>
            <RouterLink
              :to="{ name: 'login' }"
              class="label-md rounded-md px-4 py-2 hover:bg-[var(--glow-primary)] transition-colors"
            >
              {{ t('nav.login') }}
            </RouterLink>
            <RouterLink
              :to="{ name: 'register' }"
              class="btn-primary-vellum rounded-md px-5 py-2 text-sm"
            >
              {{ t('nav.startFree') }}
            </RouterLink>
          </template>
        </div>
      </nav>
    </header>
    <RouterView />
  </div>
</template>
