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
    <div class="mb-6 flex items-center justify-between gap-4 pb-4">
      <div>
        <p class="label-sm">Member Access</p>
        <h2 class="text-editorial mt-2 text-2xl font-semibold">{{ t('auth.login.heading') }}</h2>
      </div>
      <div class="flex rounded-xl bg-[var(--surface-overlay)] p-1">
        <RouterLink :to="{ name: 'login' }" class="px-4 py-2 rounded-lg text-sm bg-[var(--color-primary)] text-white">{{ t('auth.tabLogin') }}</RouterLink>
        <RouterLink :to="{ name: 'register' }" class="px-4 py-2 rounded-lg text-sm text-pretty-secondary hover:bg-[var(--surface-sunken)]">{{ t('auth.tabRegister') }}</RouterLink>
      </div>
    </div>

    <form class="space-y-5" @submit.prevent="onSubmit">
      <div class="space-y-2">
        <label class="label-sm text-pretty-secondary">{{ t('auth.login.emailLabel') }}</label>
        <input v-model="email" type="email" class="glass-input input input-bordered w-full rounded-xl" placeholder="name@workspace.com" autocomplete="email" />
      </div>

      <div class="space-y-2">
        <label class="label-sm text-pretty-secondary">{{ t('auth.login.passwordLabel') }}</label>
        <input v-model="password" type="password" class="glass-input input input-bordered w-full rounded-xl" :placeholder="t('auth.login.passwordPlaceholder')" autocomplete="current-password" />
      </div>

      <div class="flex items-center justify-between gap-4 text-sm">
        <label class="flex items-center gap-3 cursor-pointer">
          <input v-model="rememberDevice" type="checkbox" class="checkbox checkbox-sm rounded" />
          <span class="text-pretty-secondary">{{ t('auth.login.rememberDevice') }}</span>
        </label>
        <RouterLink :to="{ name: 'forgot-password' }" class="text-pretty-secondary hover:text-[var(--color-primary)]">
         {{ t('auth.login.forgotPassword') }}
        </RouterLink>
      </div>

      <p v-if="errorMessage" class="rounded-xl bg-red-500/10 px-3 py-2 text-sm text-red-500">
        {{ errorMessage }}
      </p>

      <button type="submit" class="btn-primary-vellum w-full rounded-xl py-3" :disabled="isSubmitting">
        {{ isSubmitting ? t('auth.login.submitting') : t('auth.login.submit') }}
      </button>
    </form>

    <div class="my-6 text-center">
      <p class="label-sm text-pretty-muted">{{ t('auth.login.or') }}</p>
    </div>

    <div class="space-y-3">
      <button type="button" class="btn-tertiary w-full rounded-xl py-3">{{ t('auth.login.magicLink') }}</button>
      <p class="text-center text-sm text-pretty-muted">
        {{ t('auth.login.noAccount') }}
        <RouterLink :to="{ name: 'register' }" class="text-pretty-secondary hover:text-[var(--color-primary)]">{{ t('auth.login.createAccount') }}</RouterLink>
      </p>
    </div>
  </AuthShell>
</template>
