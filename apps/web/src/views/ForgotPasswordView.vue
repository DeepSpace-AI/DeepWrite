<script setup lang="ts">
import { ref } from 'vue'
import { ApiError } from '@/api/http'
import { forgotPassword } from '@/api/auth'
import AuthShell from '@/components/auth/AuthShell.vue'

const email = ref('')
const isSubmitting = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

async function onSubmit() {
  if (isSubmitting.value) return

  const mailbox = email.value.trim().toLowerCase()
  if (!mailbox) {
    errorMessage.value = '请输入注册邮箱。'
    return
  }

  isSubmitting.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    successMessage.value = await forgotPassword({ email: mailbox })
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '发送失败，请稍后重试。'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <AuthShell
    eyebrow="Issue 04 / Access Ledger / Recovery"
    title="恢复你的访问权限"
    description="输入注册邮箱后，系统将发送重置密码所需的恢复链接。版式延续现有期刊式视觉，使权限流程也保持统一而可信。"
    note-title="恢复说明"
    note-body="找回流程通常涉及邮件发送、令牌校验与密码重置页。当前页面先完成入口表单与交互结构的落地。"
    quote="恢复权限不是异常路径，它是账号体系可信度的一部分。"
  >
    <div class="mb-6 border-b border-base-300 pb-4">
      <div class="text-xs font-mono uppercase tracking-[0.2em] text-base-content/38">Credential Recovery</div>
      <h2 class="mt-2 text-2xl font-semibold text-base-content">忘记密码</h2>
    </div>

    <form class="space-y-5" @submit.prevent="onSubmit">
      <label class="fieldset">
        <legend class="fieldset-legend text-sm">注册邮箱</legend>
        <input v-model="email" type="email" class="input input-bordered w-full rounded-sm" placeholder="name@workspace.com" autocomplete="email" />
      </label>

      <div class="rounded-sm border border-secondary/25 bg-secondary/8 p-4 text-sm leading-7 text-base-content/58">
        我们会向你的邮箱发送一封恢复邮件，链接通常在 30 分钟内有效。若未收到，请检查垃圾邮件箱或联系管理员。
      </div>

      <p v-if="errorMessage" class="rounded-sm border border-error/30 bg-error/8 px-3 py-2 text-sm text-error">
        {{ errorMessage }}
      </p>
      <p v-if="successMessage" class="rounded-sm border border-success/30 bg-success/8 px-3 py-2 text-sm text-success">
        {{ successMessage }}
      </p>

      <button type="submit" class="btn btn-primary w-full rounded-sm" :disabled="isSubmitting">
        {{ isSubmitting ? '发送中...' : '发送恢复链接' }}
      </button>
    </form>

    <div class="mt-6 flex flex-wrap items-center justify-between gap-3 text-sm">
      <RouterLink :to="{ name: 'login' }" class="link link-hover text-primary">返回登录</RouterLink>
      <RouterLink :to="{ name: 'register' }" class="link link-hover text-base-content/58">创建新账号</RouterLink>
    </div>
  </AuthShell>
</template>
