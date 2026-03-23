<script setup lang="ts">
import { ref } from 'vue'
import { ApiError } from '@/api/http'
import { forgotPassword } from '@/api/auth'
import AuthShell from '@/components/auth/AuthShell.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const email = ref('')
const isSubmitting = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

async function onSubmit() {
  if (isSubmitting.value) return

  const mailbox = email.value.trim().toLowerCase()
  if (!mailbox) {
    errorMessage.value = t('auth.forgotPassword.emptyEmailError')
    return
  }

  isSubmitting.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    successMessage.value = await forgotPassword({ email: mailbox })
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : t('auth.forgotPassword.failedError')
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <AuthShell
      :eyebrow="t('auth.forgotPassword.eyebrow')"
      :title="t('auth.forgotPassword.title')"
      :description="t('auth.forgotPassword.description')"
      :note-title="t('auth.forgotPassword.noteTitle')"
      :note-body="t('auth.forgotPassword.noteBody')"
      :quote="t('auth.forgotPassword.quote')"
  >
    <div class="mb-6 rounded-md bg-[var(--surface-overlay)] p-4">
      <div class="text-xs font-mono uppercase tracking-[0.2em] text-pretty-muted">Credential Recovery</div>
        <h2 class="mt-2 text-2xl font-semibold text-pretty">{{ t('auth.forgotPassword.heading') }}</h2>
    </div>

    <form class="space-y-5" @submit.prevent="onSubmit">
      <label class="fieldset">
          <legend class="fieldset-legend text-sm">{{ t('auth.forgotPassword.emailLabel') }}</legend>
        <input v-model="email" type="email" class="glass-input input input-bordered w-full rounded-md" placeholder="name@workspace.com" autocomplete="email" />
      </label>

      <div class="rounded-md bg-[var(--glow-secondary)] p-4 text-sm leading-7 text-pretty-secondary">
          {{ t('auth.forgotPassword.infoText') }}
      </div>

      <p v-if="errorMessage" class="rounded-md bg-error/8 px-3 py-2 text-sm text-error">
        {{ errorMessage }}
      </p>
      <p v-if="successMessage" class="rounded-md bg-success/8 px-3 py-2 text-sm text-success">
        {{ successMessage }}
      </p>

      <button type="submit" class="btn-primary-vellum w-full rounded-md px-4 py-2" :disabled="isSubmitting">
        {{ isSubmitting ? t('auth.forgotPassword.submitting') : t('auth.forgotPassword.submit') }}
      </button>
    </form>

    <div class="mt-6 flex flex-wrap items-center justify-between gap-3 text-sm">
      <RouterLink :to="{ name: 'login' }" class="link link-hover text-primary">{{ t('auth.forgotPassword.backToLogin') }}</RouterLink>
      <RouterLink :to="{ name: 'register' }" class="link link-hover text-base-content/58">{{ t('auth.forgotPassword.createAccount') }}</RouterLink>
    </div>
  </AuthShell>
</template>
