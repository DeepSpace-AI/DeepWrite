<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ApiError } from '@/api/http'
import { register } from '@/api/auth'
import AuthShell from '@/components/auth/AuthShell.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
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
    errorMessage.value = t('auth.register.emptyFieldsError')
    return
  }
  if (pass.length < 6) {
    errorMessage.value = t('auth.register.shortPasswordError')
    return
  }
  if (pass !== confirm) {
    errorMessage.value = t('auth.register.passwordMismatchError')
    return
  }
  if (!agreed.value) {
    errorMessage.value = t('auth.register.agreeRequiredError')
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
      successMessage.value = t('auth.register.successMessage')
    await router.push({ name: 'login', query: { email: mailbox } })
  } catch (error) {
      errorMessage.value = error instanceof ApiError ? error.message : t('auth.register.failedError')
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <AuthShell
     :eyebrow="t('auth.register.eyebrow')"
     :title="t('auth.register.title')"
     :description="t('auth.register.description')"
     :note-title="t('auth.register.noteTitle')"
     :note-body="t('auth.register.noteBody')"
     :quote="t('auth.register.quote')"
  >
    <div class="mb-6 flex flex-col items-start justify-between gap-4 pb-4 sm:flex-row sm:items-center">
      <div>
        <p class="label-sm">Workspace Enrollment</p>
        <h2 class="text-editorial mt-2 text-2xl font-semibold">{{ t('auth.register.heading') }}</h2>
      </div>
      <div class="flex rounded-xl bg-[var(--surface-overlay)] p-1">
        <RouterLink :to="{ name: 'login' }" class="px-4 py-2 rounded-lg text-sm text-pretty-secondary hover:bg-[var(--surface-sunken)]">{{ t('auth.tabLogin') }}</RouterLink>
        <RouterLink :to="{ name: 'register' }" class="px-4 py-2 rounded-lg text-sm bg-[var(--color-primary)] text-white">{{ t('auth.tabRegister') }}</RouterLink>
      </div>
    </div>

    <form class="grid gap-5" @submit.prevent="onSubmit">
      <div class="space-y-2">
        <label class="label-sm text-pretty-secondary">{{ t('auth.register.nameLabel') }}</label>
        <input v-model="displayName" type="text" class="glass-input input input-bordered w-full rounded-xl" :placeholder="t('auth.register.namePlaceholder')" autocomplete="name" />
      </div>

      <div class="space-y-2">
        <label class="label-sm text-pretty-secondary">{{ t('auth.register.emailLabel') }}</label>
        <input v-model="email" type="email" class="glass-input input input-bordered w-full rounded-xl" placeholder="name@workspace.com" autocomplete="email" />
      </div>

      <div class="grid gap-5 md:grid-cols-2">
        <div class="space-y-2">
          <label class="label-sm text-pretty-secondary">{{ t('auth.register.passwordLabel') }}</label>
          <input v-model="password" type="password" class="glass-input input input-bordered w-full rounded-xl" :placeholder="t('auth.register.passwordPlaceholder')" autocomplete="new-password" />
        </div>
        <div class="space-y-2">
          <label class="label-sm text-pretty-secondary">{{ t('auth.register.confirmLabel') }}</label>
          <input v-model="confirmPassword" type="password" class="glass-input input input-bordered w-full rounded-xl" :placeholder="t('auth.register.confirmPlaceholder')" autocomplete="new-password" />
        </div>
      </div>

      <label class="flex items-start gap-3 rounded-xl bg-[var(--surface-overlay)] px-4 py-3 cursor-pointer">
        <input v-model="agreed" type="checkbox" class="checkbox checkbox-sm rounded mt-0.5" />
        <span class="body-md text-pretty-secondary">
         {{ t('auth.register.agreeTerms') }}
        </span>
      </label>

      <p v-if="errorMessage" class="rounded-xl bg-red-500/10 px-3 py-2 text-sm text-red-500">
        {{ errorMessage }}
      </p>
      <p v-if="successMessage" class="rounded-xl bg-emerald-500/10 px-3 py-2 text-sm text-emerald-600">
        {{ successMessage }}
      </p>

      <button type="submit" class="btn-primary-vellum w-full rounded-xl py-3" :disabled="isSubmitting">
        {{ isSubmitting ? t('auth.register.submitting') : t('auth.register.submit') }}
      </button>
    </form>

    <p class="mt-5 text-center text-sm text-pretty-muted">
      {{ t('auth.register.hasAccount') }}
      <RouterLink :to="{ name: 'login' }" class="text-pretty-secondary hover:text-[var(--color-primary)]">{{ t('auth.register.backToLogin') }}</RouterLink>
    </p>
  </AuthShell>
</template>
