<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { ApiError } from '@/api/http'
import { updateCurrentUserProfile, uploadUserAvatar } from '@/api/auth'
import { useUserStore } from '@/stores/user'
import IconAccountCircle from '~icons/mdi/account-circle'

const userStore = useUserStore()
const { user } = storeToRefs(userStore)

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
    uploadAvatarError.value = '头像文件不能超过 5 MB。'
    return
  }

  isUploadingAvatar.value = true
  uploadAvatarError.value = ''

  try {
    const url = await uploadUserAvatar(file)
    avatarUrl.value = url
  } catch (error) {
    uploadAvatarError.value = error instanceof ApiError ? error.message : '头像上传失败，请稍后重试。'
  } finally {
    isUploadingAvatar.value = false
  }
}

async function onSubmit() {
  if (isSubmitting.value) return

  const name = displayName.value.trim()
  if (!name) {
    errorMessage.value = '显示名不能为空。'
    return
  }
  if (name.length > 30) {
    errorMessage.value = '显示名不能超过 30 个字符。'
    return
  }
  if (bio.value.trim().length > 255) {
    errorMessage.value = '个人简介不能超过 255 个字符。'
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
    successMessage.value = '个人信息已更新。'
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '更新失败，请稍后重试。'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <section class="space-y-5">
    <section class="paper-card rounded-lg p-5">
      <p class="text-[11px] font-mono uppercase tracking-[0.2em] text-pretty-muted">Profile / Account</p>
      <h2 class="heading-serif mt-2 text-3xl font-bold text-pretty">个人信息</h2>
      <p class="mt-2 max-w-2xl text-sm leading-7 text-pretty-secondary">管理你的显示名、头像地址、语言和时区设置。这些设置会同步到后端并在刷新后保持一致。</p>
    </section>

    <section class="paper-card rounded-lg p-5 bg-[var(--surface-raised)]">
      <form class="grid gap-5" @submit.prevent="onSubmit">
        <div class="grid gap-5 md:grid-cols-2">
          <label class="fieldset">
            <legend class="fieldset-legend text-sm text-pretty-secondary">显示名</legend>
            <input v-model="displayName" type="text" class="input input-bordered w-full rounded-xl" maxlength="30" placeholder="请输入显示名" />
          </label>

          <label class="fieldset">
            <legend class="fieldset-legend text-sm text-pretty-secondary">邮箱</legend>
            <input :value="email" type="email" class="input input-bordered w-full rounded-xl" disabled />
          </label>
        </div>

        <div class="fieldset">
          <legend class="fieldset-legend text-sm text-pretty-secondary">头像</legend>
          <div class="flex items-center gap-4">
            <div class="relative shrink-0">
              <img
                v-if="avatarUrl"
                :src="avatarUrl"
                alt="avatar"
                class="h-16 w-16 rounded-md object-cover"
              />
              <div
                v-else
                class="flex h-16 w-16 items-center justify-center rounded-md bg-[var(--surface-overlay)] text-pretty-muted"
              >
                <IconAccountCircle class="h-8 w-8" />
              </div>
              <div
                v-if="isUploadingAvatar"
                class="absolute inset-0 flex items-center justify-center rounded-md bg-[var(--bg-base)]/70"
              >
                <span class="loading loading-spinner loading-sm" />
              </div>
            </div>
            <div class="flex flex-col gap-1">
              <button
                type="button"
                class="btn-tertiary rounded-md px-3 py-1.5 text-sm"
                :disabled="isUploadingAvatar"
                @click="avatarFileInput?.click()"
              >
                {{ isUploadingAvatar ? '上传中...' : '上传头像' }}
              </button>
              <p class="text-xs text-pretty-muted">最大 5 MB，支持 JPG / PNG / GIF / WebP</p>
            </div>
          </div>
          <input
            ref="avatarFileInput"
            type="file"
            accept="image/*"
            class="hidden"
            @change="onAvatarFileChange"
          />
          <p v-if="uploadAvatarError" class="mt-1 text-xs text-error">{{ uploadAvatarError }}</p>
        </div>

        <label class="fieldset">
          <legend class="fieldset-legend text-sm text-pretty-secondary">头像地址</legend>
          <input v-model="avatarUrl" type="url" class="input input-bordered w-full rounded-md" placeholder="https://example.com/avatar.jpg" />
          <div class="fieldset-label text-xs text-pretty-muted">上传头像后自动填入，也可手动输入 URL</div>
        </label>

        <label class="fieldset">
          <legend class="fieldset-legend text-sm text-pretty-secondary">个人简介</legend>
          <textarea v-model="bio" class="textarea textarea-bordered min-h-28 w-full rounded-md" maxlength="255" placeholder="介绍一下你自己" />
          <div class="mt-1 text-xs text-pretty-muted">{{ bio.trim().length }} / 255</div>
        </label>

        <div class="grid gap-5 md:grid-cols-2">
          <label class="fieldset">
            <legend class="fieldset-legend text-sm text-pretty-secondary">语言</legend>
            <select v-model="language" class="select select-bordered w-full rounded-md">
              <option v-for="item in languageOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
            </select>
          </label>

          <label class="fieldset">
            <legend class="fieldset-legend text-sm text-pretty-secondary">时区</legend>
            <select v-model="timezone" class="select select-bordered w-full rounded-md">
              <option v-for="item in timezoneOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
            </select>
          </label>
        </div>

        <p v-if="errorMessage" class="rounded-md bg-error/8 px-3 py-2 text-sm text-error">
          {{ errorMessage }}
        </p>

        <p v-if="successMessage" class="rounded-md bg-emerald-500/8 px-3 py-2 text-sm text-emerald-600">
          {{ successMessage }}
        </p>

        <div class="flex justify-end">
          <button type="submit" class="btn-primary-vellum rounded-md px-4 py-2" :disabled="isSubmitting">
            {{ isSubmitting ? '保存中...' : '保存更改' }}
          </button>
        </div>
      </form>
    </section>
  </section>
</template>
