<script setup lang="ts">
import { ref, computed } from 'vue'
import { DButton, DModal } from '@/components/base'
import IconShare from '~icons/mdi/share-variant'
import IconCopy from '~icons/mdi/content-copy'
import IconCheck from '~icons/mdi/check'
import type { SessionShare } from '@/api/agent'

const props = defineProps<{
  share: SessionShare | null
  isCreating: boolean
  error: string | null
}>()

const emit = defineEmits<{
  create: []
  close: []
  copy: [token: string]
}>()

const copied = ref(false)

const shareUrl = computed(() => {
  if (!props.share) return ''
  return `${window.location.origin}/share/${props.share.share_token}`
})

async function handleCopy() {
  if (!props.share) return
  
  try {
    await navigator.clipboard.writeText(shareUrl.value)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
    emit('copy', props.share.share_token)
  } catch {
    // Failed to copy
  }
}
</script>

<template>
  <DModal :visible="!!share || isCreating" :title="share ? '分享对话' : '创建分享'" @close="emit('close')">
    <div class="share-dialog">
      <div v-if="isCreating" class="loading-state">
        <div class="spinner" />
        <p>正在创建分享链接...</p>
      </div>

      <div v-else-if="error" class="error-state">
        <p>{{ error }}</p>
        <DButton @click="emit('create')">重试</DButton>
      </div>

      <div v-else-if="share" class="share-content">
        <p class="share-info">
          分享链接已创建，任何人都可以查看此对话。
        </p>

        <div class="share-url">
          <input
            type="text"
            :value="shareUrl"
            readonly
            class="url-input"
          />
          <DButton @click="handleCopy">
            <IconCopy v-if="!copied" class="h-4 w-4" />
            <IconCheck v-else class="h-4 w-4" />
            {{ copied ? '已复制' : '复制' }}
          </DButton>
        </div>

        <div v-if="share.expires_at" class="share-expiry">
          <span class="expiry-label">过期时间：</span>
          <span class="expiry-value">{{ new Date(share.expires_at).toLocaleString() }}</span>
        </div>

        <div class="share-stats">
          <span class="stats-label">浏览次数：</span>
          <span class="stats-value">{{ share.view_count }}</span>
        </div>
      </div>

      <div v-else class="empty-state">
        <p>创建一个分享链接，让其他人可以查看此对话。</p>
        <DButton @click="emit('create')">
          <IconShare class="h-4 w-4" />
          创建分享链接
        </DButton>
      </div>
    </div>
  </DModal>
</template>

<style scoped>
.share-dialog {
  min-width: 24rem;
}

.loading-state,
.error-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  text-align: center;
}

.spinner {
  width: 2rem;
  height: 2rem;
  border: 2px solid var(--color-surface-container-high);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.share-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.share-info {
  margin: 0;
  font-size: 0.875rem;
  color: var(--color-on-surface-variant);
}

.share-url {
  display: flex;
  gap: 0.5rem;
}

.url-input {
  flex: 1;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--color-surface-container-high);
  border-radius: 0.5rem;
  background: var(--surface-container-low);
  font-size: 0.8125rem;
  color: var(--color-on-background);
  font-family: monospace;
}

.share-expiry,
.share-stats {
  display: flex;
  gap: 0.5rem;
  font-size: 0.8125rem;
}

.expiry-label,
.stats-label {
  color: var(--color-on-surface-variant);
}

.expiry-value,
.stats-value {
  color: var(--color-on-background);
}
</style>