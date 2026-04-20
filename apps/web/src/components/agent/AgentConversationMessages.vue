<script setup lang="ts">
import type { Agent } from '@/api/agent'
import type { ChatMessage } from '@/types/chat'
import { getTextFromMessage } from '@/types/chat'
import IconAutoAwesome from '~icons/mdi/auto-fix'
import IconUser from '~icons/mdi/account'

const props = defineProps<{
  messages: ChatMessage[]
  isLoading?: boolean
  agent?: Agent | null
}>()
</script>

<template>
  <div class="messages-area flex-1 overflow-y-auto px-8 py-6">
    <div class="mx-auto max-w-3xl space-y-6">
      <div v-if="messages.length === 0 && !isLoading" class="flex flex-col items-center py-16 text-center">
        <div class="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-[var(--color-primary-container)]">
          <IconAutoAwesome class="h-8 w-8 text-[var(--color-primary)]" />
        </div>
        <h3 class="text-editorial mb-2 text-xl text-[var(--color-on-background)]">
          {{ agent?.name || 'Agent' }}
        </h3>
        <p class="body-md max-w-md text-[var(--color-on-surface-variant)]">
          {{ agent?.description || 'Start a conversation with your AI assistant.' }}
        </p>
      </div>

      <div
        v-for="message in messages"
        :key="message.id"
        class="message-item"
        :class="message.role"
      >
        <div v-if="message.role === 'assistant'" class="mb-2 flex items-center gap-2">
          <IconAutoAwesome class="h-4 w-4 text-[var(--color-primary)]" />
          <span class="label-sm font-medium text-[var(--color-primary)]">{{ agent?.name || 'Agent' }}</span>
        </div>

        <div
          class="message-content body-lg leading-relaxed"
          :class="{
            'user-message ml-auto max-w-[85%] rounded-xl p-4': message.role === 'user',
            'assistant-message': message.role === 'assistant'
          }"
        >
          <p class="whitespace-pre-wrap">{{ getTextFromMessage(message) }}</p>
        </div>

        <div v-if="message.role === 'user'" class="mt-1 flex justify-end">
          <IconUser class="h-4 w-4 text-[var(--color-on-surface-variant)]" />
        </div>
      </div>

      <div v-if="isLoading" class="message-item assistant">
        <div class="mb-2 flex items-center gap-2">
          <IconAutoAwesome class="h-4 w-4 text-[var(--color-primary)]" />
          <span class="label-sm font-medium text-[var(--color-primary)]">{{ agent?.name || 'Agent' }}</span>
        </div>
        <div class="message-content assistant-message">
          <div class="flex gap-1">
            <span class="h-2 w-2 animate-bounce rounded-full bg-[var(--color-primary)]" style="animation-delay: 0ms" />
            <span class="h-2 w-2 animate-bounce rounded-full bg-[var(--color-primary)]" style="animation-delay: 150ms" />
            <span class="h-2 w-2 animate-bounce rounded-full bg-[var(--color-primary)]" style="animation-delay: 300ms" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.user-message {
  background-color: var(--surface-container-lowest);
  border-left: 4px solid var(--color-primary);
  font-style: italic;
  box-shadow: 0 4px 16px oklch(0.28 0.008 105 / 0.04);
}

.assistant-message {
  color: var(--color-on-background);
}
</style>