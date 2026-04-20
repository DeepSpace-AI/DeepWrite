<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { ApiError } from '@/api/http'
import { updateCurrentUserProfile, uploadUserAvatar } from '@/api/auth'
import { useUserStore } from '@/stores/user'
import { useI18n } from 'vue-i18n'
import IconAccountCircle from '~icons/mdi/account-circle'
import IconUpload from '~icons/mdi/upload'

const userStore = useUserStore()
const { user } = storeToRefs(userStore)
const { t } = useI18n()

const displayName = ref('')
const avatarUrl = ref('')
const bio = ref('')
const language = ref('zh-CN')
const timezone = ref('UTC')
const isSubmitting = ref(false)
const isUploadingAvatar = ref(false)
const uploadAvatarError = ref('')
const avatarFileInput = ref<HTMLInputElement | null>(null)
const errorMessage = ref('')
const successMessage = ref('')

const languageOptions = [
  { label: '简体中文', value: 'zh-CN' },
  { label: 'English', value: 'en' },
]

const timezoneOptions = [
  { label: 'UTC', value: 'UTC' },
  { label: 'Asia/Shanghai', value: 'Asia/Shanghai' },
  { label: 'Asia/Tokyo', value: 'Asia/Tokyo' },
  { label: 'Europe/Berlin', value: 'Europe/Berlin' },
  { label: 'America/New_York', value: 'America/New_York' },
]

const email = computed(() => user.value?.email || '')

const userInitial = computed(() => {
  const name = user.value?.displayName || user.value?.email || ''
  return name ? name.slice(0, 1).toUpperCase() : 'D'
})

watch(
  user,
  (nextUser) => {
    displayName.value = nextUser?.displayName || ''
    avatarUrl.value = nextUser?.avatarUrl || ''
    bio.value = nextUser?.bio || ''
    language.value = nextUser?.language || 'zh-CN'
    timezone.value = nextUser?.timezone || 'UTC'
  },
  { immediate: true },
)

async function onAvatarFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  input.value = ''

  if (file.size > 5 * 1024 * 1024) {
    uploadAvatarError.value = t('profile.avatarSizeError')
    return
  }

  isUploadingAvatar.value = true
  uploadAvatarError.value = ''

  try {
    const url = await uploadUserAvatar(file)
    avatarUrl.value = url
  } catch (error) {
    uploadAvatarError.value = error instanceof ApiError ? error.message : t('profile.avatarUploadError')
  } finally {
    isUploadingAvatar.value = false
  }
}

async function onSubmit() {
  if (isSubmitting.value) return

  const name = displayName.value.trim()
  if (!name) {
    errorMessage.value = t('profile.nameRequired')
    return
  }
  if (name.length > 30) {
    errorMessage.value = t('profile.nameTooLong')
    return
  }
  if (bio.value.trim().length > 255) {
    errorMessage.value = t('profile.bioTooLong')
    return
  }

  isSubmitting.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    const updated = await updateCurrentUserProfile({
      displayName: name,
      avatarUrl: avatarUrl.value.trim(),
      bio: bio.value.trim(),
      language: language.value,
      timezone: timezone.value,
    })

    userStore.setUser(updated)
    successMessage.value = t('profile.success')
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : t('profile.error')
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <section class="scholar-profile p-8">
    <div class="mx-auto max-w-2xl">
      <!-- Header -->
      <header class="mb-8">
        <p class="label-sm text-[var(--color-on-surface-variant)] uppercase tracking-widest">
          Profile / Account
        </p>
        <h1 class="text-editorial mt-3 text-4xl font-light text-[var(--color-on-background)]">
          {{ t('profile.title') }}
        </h1>
        <p class="body-md mt-3 text-[var(--color-on-surface-variant)]">
          {{ t('profile.subtitle') }}
        </p>
      </header>

      <!-- Form Card -->
      <div class="scholar-form-card rounded-xl p-8">
        <form class="space-y-8" @submit.prevent="onSubmit">
          <!-- Avatar Section -->
          <div class="scholar-form-section">
            <label class="label-sm mb-4 block text-[var(--color-on-surface-variant)]">
              {{ t('profile.avatar') }}
            </label>
            <div class="flex items-center gap-6">
              <div class="scholar-avatar-container relative shrink-0">
                <img
                  v-if="avatarUrl"
                  :src="avatarUrl"
                  alt="avatar"
                  class="h-20 w-20 rounded-xl object-cover"
                />
                <div v-else class="scholar-avatar-placeholder flex h-20 w-20 items-center justify-center rounded-xl">
                  <span class="text-2xl font-medium">{{ userInitial }}</span>
                </div>
                <div
                  v-if="isUploadingAvatar"
                  class="absolute inset-0 flex items-center justify-center rounded-xl bg-[var(--surface-base)]/70"
                >
                  <span class="loading loading-spinner loading-md" />
                </div>
              </div>
              <div class="flex flex-col gap-2">
                <button
                  type="button"
                  class="scholar-upload-btn flex items-center gap-2 rounded-lg px-4 py-2.5"
                  :disabled="isUploadingAvatar"
                  @click="avatarFileInput?.click()"
                >
                  <IconUpload class="h-4 w-4" />
                  {{ isUploadingAvatar ? t('profile.uploading') : t('profile.uploadAvatar') }}
                </button>
                <p class="text-xs text-[var(--color-on-surface-muted)]">{{ t('profile.avatarHint') }}</p>
              </div>
            </div>
            <input
              ref="avatarFileInput"
              type="file"
              accept="image/*"
              class="hidden"
              @change="onAvatarFileChange"
            />
            <p v-if="uploadAvatarError" class="mt-2 text-sm text-[var(--color-error)]">{{ uploadAvatarError }}</p>
          </div>

          <!-- Name & Email -->
          <div class="grid gap-6 md:grid-cols-2">
            <div class="scholar-form-section">
              <label class="label-sm mb-2 block text-[var(--color-on-surface-variant)]">
                {{ t('profile.displayName') }}
              </label>
              <input
                v-model="displayName"
                type="text"
                class="scholar-input w-full rounded-lg px-4 py-3"
                maxlength="30"
                :placeholder="t('profile.displayNamePlaceholder')"
              />
            </div>

            <div class="scholar-form-section">
              <label class="label-sm mb-2 block text-[var(--color-on-surface-variant)]">
                {{ t('profile.email') }}
              </label>
              <input
                :value="email"
                type="email"
                class="scholar-input w-full rounded-lg px-4 py-3 opacity-60"
                disabled
              />
            </div>
          </div>

          <!-- Avatar URL -->
          <div class="scholar-form-section">
            <label class="label-sm mb-2 block text-[var(--color-on-surface-variant)]">
              {{ t('profile.avatarUrl') }}
            </label>
            <input
              v-model="avatarUrl"
              type="url"
              class="scholar-input w-full rounded-lg px-4 py-3"
              placeholder="https://example.com/avatar.jpg"
            />
            <p class="mt-2 text-xs text-[var(--color-on-surface-muted)]">{{ t('profile.avatarUrlHint') }}</p>
          </div>

          <!-- Bio -->
          <div class="scholar-form-section">
            <label class="label-sm mb-2 block text-[var(--color-on-surface-variant)]">
              {{ t('profile.bio') }}
            </label>
            <textarea
              v-model="bio"
              class="scholar-input min-h-28 w-full rounded-lg px-4 py-3"
              maxlength="255"
              :placeholder="t('profile.bioPlaceholder')"
            />
            <p class="mt-2 text-xs text-[var(--color-on-surface-muted)]">{{ bio.trim().length }} / 255</p>
          </div>

          <!-- Language & Timezone -->
          <div class="grid gap-6 md:grid-cols-2">
            <div class="scholar-form-section">
              <label class="label-sm mb-2 block text-[var(--color-on-surface-variant)]">
                {{ t('profile.language') }}
              </label>
              <select v-model="language" class="scholar-select w-full rounded-lg px-4 py-3">
                <option v-for="item in languageOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
            </div>

            <div class="scholar-form-section">
              <label class="label-sm mb-2 block text-[var(--color-on-surface-variant)]">
                {{ t('profile.timezone') }}
              </label>
              <select v-model="timezone" class="scholar-select w-full rounded-lg px-4 py-3">
                <option v-for="item in timezoneOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
            </div>
          </div>

          <!-- Messages -->
          <div v-if="errorMessage" class="scholar-error rounded-lg p-4">
            {{ errorMessage }}
          </div>

          <div v-if="successMessage" class="scholar-success rounded-lg p-4">
            {{ successMessage }}
          </div>

          <!-- Submit -->
          <div class="flex justify-end">
            <button type="submit" class="scholar-submit-btn rounded-lg px-6 py-3" :disabled="isSubmitting">
              {{ isSubmitting ? t('profile.saving') : t('profile.save') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </section>
</template>

<style scoped>
.scholar-form-card {
  background-color: var(--surface-container-lowest);
  box-shadow: 0 8px 32px oklch(0.28 0.008 105 / 0.08);
}

[data-theme="vellum-dark"] .scholar-form-card {
  box-shadow: 0 8px 32px oklch(0.15 0.02 75 / 0.20);
}

.scholar-avatar-container {
  background-color: var(--surface-container-low);
  border-radius: 0.75rem;
}

.scholar-avatar-placeholder {
  background-color: var(--surface-container);
  color: var(--color-on-surface);
}

.scholar-upload-btn {
  background-color: var(--surface-container-low);
  color: var(--color-on-surface);
  font-weight: 500;
  transition: background-color 0.2s;
}

.scholar-upload-btn:hover {
  background-color: var(--surface-container);
}

.scholar-upload-btn:disabled {
  opacity: 0.5;
}

.scholar-input {
  background-color: var(--surface-container-low);
  color: var(--color-on-surface);
  border: none;
  outline: none;
  transition: background-color 0.2s;
}

.scholar-input:focus {
  background-color: var(--surface-container);
}

.scholar-select {
  background-color: var(--surface-container-low);
  color: var(--color-on-surface);
  border: none;
  outline: none;
  cursor: pointer;
}

.scholar-select:focus {
  background-color: var(--surface-container);
}

.scholar-error {
  background-color: var(--color-error-container);
  color: var(--color-on-error-container);
}

.scholar-success {
  background-color: var(--color-tertiary-container);
  color: var(--color-on-tertiary-container);
}

.scholar-submit-btn {
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-dim));
  color: var(--color-on-primary);
  font-weight: 500;
  transition: opacity 0.2s;
}

.scholar-submit-btn:hover {
  opacity: 0.88;
}

.scholar-submit-btn:disabled {
  opacity: 0.5;
}
</style>