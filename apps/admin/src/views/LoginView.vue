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
  <main class="dot-grid min-h-screen bg-base-100 px-4 py-10">
    <section class="mx-auto max-w-md rounded-sm border border-base-300 bg-base-100 p-6 shadow-sm">
      <div class="mb-6">
        <p class="text-xs font-mono uppercase tracking-[0.2em] text-base-content/50">DeepWrite Admin</p>
        <h1 class="mt-3 text-2xl font-semibold text-base-content">管理员登录</h1>
        <p class="mt-2 text-sm text-base-content/70">仅 role=admin 账号可访问后台管理系统</p>
      </div>

      <form class="space-y-4" @submit.prevent="onSubmit">
        <label class="form-control">
          <span class="label-text text-sm">邮箱</span>
          <input v-model="email" type="email" required class="input input-bordered rounded-sm" placeholder="admin@deepwrite.io" />
        </label>
        <label class="form-control">
          <span class="label-text text-sm">密码</span>
          <input v-model="password" type="password" required class="input input-bordered rounded-sm" />
        </label>
        <p v-if="errorMessage" class="rounded-sm border border-error/30 bg-error/10 px-3 py-2 text-sm text-error">{{ errorMessage }}</p>
        <button type="submit" class="btn btn-primary w-full rounded-sm" :disabled="isSubmitting">
          <span v-if="isSubmitting" class="loading loading-spinner loading-xs" />
          <span>{{ isSubmitting ? '登录中...' : '登录后台' }}</span>
        </button>
      </form>
    </section>
  </main>
</template>
