<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ApiError } from '@/api/http'
import { login, saveSession } from '@/api/auth'
import AuthShell from '@/components/auth/AuthShell.vue'

const router = useRouter()
const route = useRoute()

const email = ref('')
const password = ref('')
const rememberDevice = ref(true)
const isSubmitting = ref(false)
const errorMessage = ref('')

async function onSubmit() {
  if (isSubmitting.value) return

  const trimmedEmail = email.value.trim().toLowerCase()
  if (!trimmedEmail || !password.value.trim()) {
    errorMessage.value = '请填写邮箱和密码。'
    return
  }

  isSubmitting.value = true
  errorMessage.value = ''

  try {
    const data = await login({
      email: trimmedEmail,
      password: password.value,
    })

    await saveSession(data)
    if (!rememberDevice.value) {
      sessionStorage.setItem('deepwrite_access_token', data.access_token)
    }

    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
    if (redirect) {
      await router.push(redirect)
    } else {
      await router.push({ name: 'home' })
    }
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '登录失败，请稍后重试。'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <AuthShell
    eyebrow="Issue 04 / Access Ledger / Sign In"
    title="继续你的写作进程"
    description="登录 DeepWrite，回到你的文稿、协作会话与版本记录。界面延续与首页一致的出版式结构，减少切换场景带来的认知断层。"
    note-title="访问说明"
    note-body="登录后可恢复最近工作区、协作状态与编辑偏好。后续可直接接入后端鉴权接口。"
    quote="权限页面不应只是表单，它应该是进入工作状态前的最后一页整备。"
  >
    <div class="mb-6 flex items-center justify-between gap-4 border-b border-base-300 pb-4">
      <div>
        <div class="text-xs font-mono uppercase tracking-[0.2em] text-base-content/38">Member Access</div>
        <h2 class="mt-2 text-2xl font-semibold text-base-content">登录</h2>
      </div>
      <div class="tabs tabs-box rounded-sm border border-base-300 bg-base-200 p-1">
        <RouterLink :to="{ name: 'login' }" class="tab tab-active rounded-sm">登录</RouterLink>
        <RouterLink :to="{ name: 'register' }" class="tab rounded-sm">注册</RouterLink>
      </div>
    </div>

    <form class="space-y-5" @submit.prevent="onSubmit">
      <label class="fieldset">
        <legend class="fieldset-legend text-sm">邮箱地址</legend>
        <input v-model="email" type="email" class="input input-bordered w-full rounded-sm" placeholder="name@workspace.com" autocomplete="email" />
      </label>

      <label class="fieldset">
        <legend class="fieldset-legend text-sm">密码</legend>
        <input v-model="password" type="password" class="input input-bordered w-full rounded-sm" placeholder="输入你的密码" autocomplete="current-password" />
      </label>

      <div class="flex items-center justify-between gap-4 text-sm">
        <label class="label cursor-pointer justify-start gap-3 p-0">
          <input v-model="rememberDevice" type="checkbox" class="checkbox checkbox-sm rounded-xs" />
          <span class="label-text text-base-content/60">记住当前设备</span>
        </label>
        <RouterLink :to="{ name: 'forgot-password' }" class="link link-hover text-primary">
          忘记密码
        </RouterLink>
      </div>

      <p v-if="errorMessage" class="rounded-sm border border-error/30 bg-error/8 px-3 py-2 text-sm text-error">
        {{ errorMessage }}
      </p>

      <button type="submit" class="btn btn-primary w-full rounded-sm" :disabled="isSubmitting">
        {{ isSubmitting ? '登录中...' : '登录并继续' }}
      </button>
    </form>

    <div class="divider my-6 text-xs font-mono uppercase tracking-[0.18em] text-base-content/35">或</div>

    <div class="space-y-3">
      <button type="button" class="btn btn-outline w-full rounded-sm">使用工作区邮箱链接登录</button>
      <p class="text-center text-sm text-base-content/50">
        还没有账号？
        <RouterLink :to="{ name: 'register' }" class="link link-hover text-primary">创建账号</RouterLink>
      </p>
    </div>
  </AuthShell>
</template>
