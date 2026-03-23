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
  <main class="relative min-h-screen overflow-hidden bg-dot-grid px-4 py-10" data-theme="forest">
    <div class="glow-blob -left-32 top-0 h-[32rem] w-[32rem] bg-[var(--glow-primary)] opacity-50" />
    <div class="glow-blob -right-24 bottom-0 h-[28rem] w-[28rem] bg-[var(--glow-secondary)] opacity-40" />
    <div class="glow-blob left-1/2 top-1/2 h-80 w-80 -translate-x-1/2 -translate-y-1/2 rounded-full bg-[var(--glow-primary)] opacity-20" />

    <section class="glass-card relative mx-auto max-w-md rounded-2xl p-8">
      <div class="mb-8">
        <p class="text-xs font-medium uppercase tracking-[0.2em] text-pretty-muted">DeepWrite Admin</p>
        <h1 class="mt-3 text-2xl font-semibold text-pretty">管理员登录</h1>
        <p class="mt-2 text-sm text-pretty-secondary">仅限管理员账号访问后台管理系统</p>
      </div>

      <form class="space-y-5" @submit.prevent="onSubmit">
        <label class="form-control">
          <span class="label-text text-sm text-pretty-secondary">邮箱</span>
          <input 
            v-model="email" 
            type="email" 
            required 
            class="glass-input input input-bordered mt-1.5 w-full rounded-xl" 
            placeholder="admin@deepwrite.io" 
          />
        </label>
        <label class="form-control">
          <span class="label-text text-sm text-pretty-secondary">密码</span>
          <input 
            v-model="password" 
            type="password" 
            required 
            class="glass-input input input-bordered mt-1.5 w-full rounded-xl" 
          />
        </label>
        <p v-if="errorMessage" class="rounded-xl border border-error/20 bg-error/10 px-4 py-3 text-sm text-error">
          {{ errorMessage }}
        </p>
        <button 
          type="submit" 
          class="btn w-full rounded-xl bg-[var(--glow-primary)] text-pretty transition-all hover:bg-[var(--glow-primary)] hover:opacity-90" 
          :disabled="isSubmitting"
        >
          <span v-if="isSubmitting" class="loading loading-spinner loading-xs" />
          <span>{{ isSubmitting ? '登录中...' : '登录后台' }}</span>
        </button>
      </form>
    </section>
  </main>
</template>
