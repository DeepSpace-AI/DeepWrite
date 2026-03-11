<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ApiError } from '@/api/http'
import { register } from '@/api/auth'
import AuthShell from '@/components/auth/AuthShell.vue'

const router = useRouter()

const displayName = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const agreed = ref(false)
const isSubmitting = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

async function onSubmit() {
  if (isSubmitting.value) return

  const name = displayName.value.trim()
  const mailbox = email.value.trim().toLowerCase()
  const pass = password.value
  const confirm = confirmPassword.value

  if (!name || !mailbox || !pass) {
    errorMessage.value = '请填写姓名、邮箱和密码。'
    return
  }
  if (pass.length < 6) {
    errorMessage.value = '密码长度至少为 6 位。'
    return
  }
  if (pass !== confirm) {
    errorMessage.value = '两次输入的密码不一致。'
    return
  }
  if (!agreed.value) {
    errorMessage.value = '请先同意服务条款与隐私声明。'
    return
  }

  isSubmitting.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    await register({
      displayName: name,
      email: mailbox,
      password: pass,
    })
    successMessage.value = '注册成功，正在跳转到登录页面。'
    await router.push({ name: 'login', query: { email: mailbox } })
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '注册失败，请稍后重试。'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <AuthShell
    eyebrow="Issue 04 / Access Ledger / Registration"
    title="建立你的写作工作区"
    description="创建账号后，你可以开始建立个人或团队写作流程，并逐步接入协作权限。页面语言保持克制，强调长期写作系统，而不是一次性试用入口。"
    note-title="注册说明"
    note-body="注册仅保留姓名、邮箱与密码三个核心字段，降低首次进入门槛，后续再补充工作区与团队信息。"
    quote="从账号创建开始，就应该让用户感受到这是一套可长期使用的写作基础设施。"
  >
    <div class="mb-6 flex flex-col items-start justify-between gap-4 border-b border-base-300 pb-4 sm:flex-row sm:items-center">
      <div>
        <div class="text-xs font-mono uppercase tracking-[0.2em] text-base-content/38">Workspace Enrollment</div>
        <h2 class="mt-2 text-2xl font-semibold text-base-content">注册</h2>
      </div>
      <div class="tabs tabs-box w-full rounded-sm border border-base-300 bg-base-200 p-1 sm:w-auto">
        <RouterLink :to="{ name: 'login' }" class="tab rounded-sm">登录</RouterLink>
        <RouterLink :to="{ name: 'register' }" class="tab tab-active rounded-sm">注册</RouterLink>
      </div>
    </div>

    <form class="grid gap-5" @submit.prevent="onSubmit">
      <label class="fieldset">
        <legend class="fieldset-legend text-sm">姓名</legend>
        <input v-model="displayName" type="text" class="input input-bordered w-full rounded-sm" placeholder="例如：Lin Chen" autocomplete="name" />
      </label>

      <label class="fieldset">
        <legend class="fieldset-legend text-sm">邮箱地址</legend>
        <input v-model="email" type="email" class="input input-bordered w-full rounded-sm" placeholder="name@workspace.com" autocomplete="email" />
      </label>

      <div class="grid gap-5 md:grid-cols-2">
        <label class="fieldset">
          <legend class="fieldset-legend text-sm">密码</legend>
          <input v-model="password" type="password" class="input input-bordered w-full rounded-sm" placeholder="至少 6 位" autocomplete="new-password" />
        </label>
        <label class="fieldset">
          <legend class="fieldset-legend text-sm">确认密码</legend>
          <input v-model="confirmPassword" type="password" class="input input-bordered w-full rounded-sm" placeholder="再次输入密码" autocomplete="new-password" />
        </label>
      </div>

      <label class="label cursor-pointer items-start justify-start gap-3 rounded-sm border border-base-300 bg-base-200/40 px-4 py-3">
        <input v-model="agreed" type="checkbox" class="checkbox checkbox-sm rounded-xs" />
        <span class="label-text whitespace-normal text-sm leading-7 text-base-content/62">
          我已阅读并同意服务条款与隐私声明，接受以邮箱方式接收账户相关通知。
        </span>
      </label>

      <p v-if="errorMessage" class="rounded-sm border border-error/30 bg-error/8 px-3 py-2 text-sm text-error">
        {{ errorMessage }}
      </p>
      <p v-if="successMessage" class="rounded-sm border border-success/30 bg-success/8 px-3 py-2 text-sm text-success">
        {{ successMessage }}
      </p>

      <button type="submit" class="btn btn-primary w-full rounded-sm" :disabled="isSubmitting">
        {{ isSubmitting ? '创建中...' : '创建账号' }}
      </button>
    </form>

    <p class="mt-5 text-center text-sm text-base-content/50">
      已有账号？
      <RouterLink :to="{ name: 'login' }" class="link link-hover text-primary">返回登录</RouterLink>
    </p>
  </AuthShell>
</template>
