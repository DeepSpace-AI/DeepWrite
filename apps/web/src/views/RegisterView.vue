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
    <div class="mb-6 flex flex-col items-start justify-between gap-4 border-b border-base-300 pb-4 sm:flex-row sm:items-center">
      <div>
        <div class="text-xs font-mono uppercase tracking-[0.2em] text-base-content/38">Workspace Enrollment</div>
        <h2 class="mt-2 text-2xl font-semibold text-base-content">{{ t('auth.register.heading') }}</h2>
      </div>
      <div class="tabs tabs-box w-full rounded-sm border border-base-300 bg-base-200 p-1 sm:w-auto">
        <RouterLink :to="{ name: 'login' }" class="tab rounded-sm">{{ t('auth.tabLogin') }}</RouterLink>
        <RouterLink :to="{ name: 'register' }" class="tab tab-active rounded-sm">{{ t('auth.tabRegister') }}</RouterLink>
      </div>
    </div>

    <form class="grid gap-5" @submit.prevent="onSubmit">
      <label class="fieldset">
        <legend class="fieldset-legend text-sm">{{ t('auth.register.nameLabel') }}</legend>
        <input v-model="displayName" type="text" class="input input-bordered w-full rounded-sm" :placeholder="t('auth.register.namePlaceholder')" autocomplete="name" />
      </label>

      <label class="fieldset">
        <legend class="fieldset-legend text-sm">{{ t('auth.register.emailLabel') }}</legend>
        <input v-model="email" type="email" class="input input-bordered w-full rounded-sm" placeholder="name@workspace.com" autocomplete="email" />
      </label>

      <div class="grid gap-5 md:grid-cols-2">
        <label class="fieldset">
         <legend class="fieldset-legend text-sm">{{ t('auth.register.passwordLabel') }}</legend>
         <input v-model="password" type="password" class="input input-bordered w-full rounded-sm" :placeholder="t('auth.register.passwordPlaceholder')" autocomplete="new-password" />
        </label>
        <label class="fieldset">
         <legend class="fieldset-legend text-sm">{{ t('auth.register.confirmLabel') }}</legend>
         <input v-model="confirmPassword" type="password" class="input input-bordered w-full rounded-sm" :placeholder="t('auth.register.confirmPlaceholder')" autocomplete="new-password" />
        </label>
      </div>

      <label class="label cursor-pointer items-start justify-start gap-3 rounded-sm border border-base-300 bg-base-200/40 px-4 py-3">
        <input v-model="agreed" type="checkbox" class="checkbox checkbox-sm rounded-xs" />
        <span class="label-text whitespace-normal text-sm leading-7 text-base-content/62">
         {{ t('auth.register.agreeTerms') }}
        </span>
      </label>

      <p v-if="errorMessage" class="rounded-sm border border-error/30 bg-error/8 px-3 py-2 text-sm text-error">
        {{ errorMessage }}
      </p>
      <p v-if="successMessage" class="rounded-sm border border-success/30 bg-success/8 px-3 py-2 text-sm text-success">
        {{ successMessage }}
      </p>

      <button type="submit" class="btn btn-primary w-full rounded-sm" :disabled="isSubmitting">
        {{ isSubmitting ? t('auth.register.submitting') : t('auth.register.submit') }}
      </button>
    </form>

    <p class="mt-5 text-center text-sm text-base-content/50">
      {{ t('auth.register.hasAccount') }}
      <RouterLink :to="{ name: 'login' }" class="link link-hover text-primary">{{ t('auth.register.backToLogin') }}</RouterLink>
    </p>
  </AuthShell>
</template>
