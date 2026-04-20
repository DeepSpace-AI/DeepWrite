<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ApiError } from '@/api/http'
import { login, saveSession } from '@/api/auth'

const route = useRoute()
const router = useRouter()

const email = ref('')
const password = ref('')
const isSubmitting = ref(false)
const errorMessage = ref('')

async function onSubmit() {
  isSubmitting.value = true
  errorMessage.value = ''
  try {
    const payload = await login({
      email: email.value.trim(),
      password: password.value,
    })
    await saveSession(payload)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/providers'
    await router.push(redirect)
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '登录失败，请稍后重试。'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <main class="flex min-h-screen items-center justify-center bg-[var(--bg-base)] px-4 py-10">
    <div class="w-full max-w-md">
      <div class="tech-card p-8">
        <div class="mb-8 text-center">
          <p class="text-label">DeepWrite</p>
          <h1 class="text-display mt-3 text-2xl">管理员登录</h1>
          <p class="text-muted mt-2 text-sm">仅限管理员账号访问后台系统</p>
        </div>

        <form class="space-y-5" @submit.prevent="onSubmit">
          <div>
            <label class="text-label mb-2 block">邮箱</label>
            <input 
              v-model="email" 
              type="email" 
              required 
              class="tech-input w-full px-4 py-3 text-sm" 
              placeholder="admin@deepwrite.io" 
            />
          </div>
          <div>
            <label class="text-label mb-2 block">密码</label>
            <input 
              v-model="password" 
              type="password" 
              required 
              class="tech-input w-full px-4 py-3 text-sm" 
            />
          </div>
          
          <div v-if="errorMessage" class="rounded-xl border border-[var(--accent-primary)] bg-[var(--surface-elevated)] px-4 py-3 text-sm text-[var(--accent-primary)]">
            {{ errorMessage }}
          </div>
          
          <button 
            type="submit" 
            class="btn-tech w-full py-3 text-sm"
            :disabled="isSubmitting"
          >
            <span v-if="isSubmitting" class="loading loading-spinner loading-xs" />
            <span>{{ isSubmitting ? '登录中...' : '登录后台' }}</span>
          </button>
        </form>
      </div>

      <p class="mt-6 text-center text-xs text-[var(--text-muted)]">
        DeepWrite Admin &copy; {{ new Date().getFullYear() }}
      </p>
    </div>
  </main>
</template>