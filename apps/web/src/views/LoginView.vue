<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ApiError } from '@/api/http'
import { login, saveSession } from '@/api/auth'
import AuthShell from '@/components/auth/AuthShell.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
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
    errorMessage.value = t('auth.login.emptyFieldsError')
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
    errorMessage.value = error instanceof ApiError ? error.message : t('auth.login.failedError')
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <AuthShell
     :eyebrow="t('auth.login.eyebrow')"
     :title="t('auth.login.title')"
     :description="t('auth.login.description')"
     :note-title="t('auth.login.noteTitle')"
     :note-body="t('auth.login.noteBody')"
     :quote="t('auth.login.quote')"
  >
    <div class="mb-6 flex items-center justify-between gap-4 border-b border-base-300 pb-4">
      <div>
        <div class="text-xs font-mono uppercase tracking-[0.2em] text-base-content/38">Member Access</div>
        <h2 class="mt-2 text-2xl font-semibold text-base-content">{{ t('auth.login.heading') }}</h2>
      </div>
      <div class="tabs tabs-box rounded-sm border border-base-300 bg-base-200 p-1">
        <RouterLink :to="{ name: 'login' }" class="tab tab-active rounded-sm">{{ t('auth.tabLogin') }}</RouterLink>
        <RouterLink :to="{ name: 'register' }" class="tab rounded-sm">{{ t('auth.tabRegister') }}</RouterLink>
      </div>
    </div>

    <form class="space-y-5" @submit.prevent="onSubmit">
      <label class="fieldset">
        <legend class="fieldset-legend text-sm">{{ t('auth.login.emailLabel') }}</legend>
        <input v-model="email" type="email" class="input input-bordered w-full rounded-sm" placeholder="name@workspace.com" autocomplete="email" />
      </label>

      <label class="fieldset">
        <legend class="fieldset-legend text-sm">{{ t('auth.login.passwordLabel') }}</legend>
        <input v-model="password" type="password" class="input input-bordered w-full rounded-sm" :placeholder="t('auth.login.passwordPlaceholder')" autocomplete="current-password" />
      </label>

      <div class="flex items-center justify-between gap-4 text-sm">
        <label class="label cursor-pointer justify-start gap-3 p-0">
          <input v-model="rememberDevice" type="checkbox" class="checkbox checkbox-sm rounded-xs" />
         <span class="label-text text-base-content/60">{{ t('auth.login.rememberDevice') }}</span>
        </label>
        <RouterLink :to="{ name: 'forgot-password' }" class="link link-hover text-primary">
         {{ t('auth.login.forgotPassword') }}
        </RouterLink>
      </div>

      <p v-if="errorMessage" class="rounded-sm border border-error/30 bg-error/8 px-3 py-2 text-sm text-error">
        {{ errorMessage }}
      </p>

      <button type="submit" class="btn btn-primary w-full rounded-sm" :disabled="isSubmitting">
        {{ isSubmitting ? t('auth.login.submitting') : t('auth.login.submit') }}
      </button>
    </form>

    <div class="divider my-6 text-xs font-mono uppercase tracking-[0.18em] text-base-content/35">{{ t('auth.login.or') }}</div>

    <div class="space-y-3">
      <button type="button" class="btn btn-outline w-full rounded-sm">{{ t('auth.login.magicLink') }}</button>
      <p class="text-center text-sm text-base-content/50">
        {{ t('auth.login.noAccount') }}
        <RouterLink :to="{ name: 'register' }" class="link link-hover text-primary">{{ t('auth.login.createAccount') }}</RouterLink>
      </p>
    </div>
  </AuthShell>
</template>
