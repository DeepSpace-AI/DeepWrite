<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getSharedSession, type SessionShare, type Session, type Message } from '@/api/agent'
import { ApiError } from '@/api/http'
import { ChatMessageList } from '@/components/chat'
import type { ChatMessage } from '@/types/chat'
import IconLoading from '~icons/mdi/loading'
import IconAlert from '~icons/mdi/alert-circle'
import IconHome from '~icons/mdi/home'

const route = useRoute()
const router = useRouter()

const isLoading = ref(true)
const error = ref('')
const share = ref<SessionShare | null>(null)
const session = ref<Session | null>(null)
const rawMessages = ref<Message[]>([])

const messages = computed<ChatMessage[]>(() => {
  return rawMessages.value.map(m => ({
    id: m.id,
    role: m.role as 'user' | 'assistant',
    parts: [{ type: 'text' as const, text: m.content }],
    status: 'done' as const,
    createdAt: new Date(m.created_at),
  }))
})

onMounted(async () => {
  const token = route.params.token as string
  if (!token) {
    error.value = 'Invalid share link'
    isLoading.value = false
    return
  }

  try {
    const result = await getSharedSession(token)
    share.value = result.share
    session.value = result.session
    rawMessages.value = result.messages || []
  } catch (e) {
    if (e instanceof ApiError) {
      if (e.statusCode === 404) {
        error.value = 'Share link not found or has expired'
      } else {
        error.value = e.message || 'Failed to load shared session'
      }
    } else {
      error.value = 'Failed to load shared session'
    }
  } finally {
    isLoading.value = false
  }
})

function goToHome() {
  router.push('/')
}
</script>

<template>
  <div class="share-view">
    <header class="share-header">
      <div class="header-content">
        <h1 class="logo">DeepWrite</h1>
        <button class="home-btn" @click="goToHome">
          <IconHome class="h-4 w-4" />
          <span>Home</span>
        </button>
      </div>
    </header>

    <main class="share-main">
      <div v-if="isLoading" class="loading-state">
        <IconLoading class="loading-icon" />
        <p>Loading shared conversation...</p>
      </div>

      <div v-else-if="error" class="error-state">
        <IconAlert class="error-icon" />
        <h2>Unable to Load</h2>
        <p>{{ error }}</p>
        <button class="retry-btn" @click="goToHome">Go to Home</button>
      </div>

      <template v-else>
        <div class="session-info">
          <h2 class="session-title">{{ session?.title || 'Shared Conversation' }}</h2>
          <div v-if="share?.expires_at" class="session-meta">
            <span>Expires: {{ new Date(share.expires_at).toLocaleString() }}</span>
          </div>
        </div>

        <ChatMessageList
          :messages="messages"
          :agent-name="'Assistant'"
        />
      </template>
    </main>
  </div>
</template>

<style scoped>
.share-view {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--surface);
}

.share-header {
  border-bottom: 1px solid var(--color-surface-container-high);
  background: var(--surface-container);
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 56rem;
  margin: 0 auto;
  padding: 0.75rem 1rem;
}

.logo {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--color-on-background);
  margin: 0;
}

.home-btn {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 0.75rem;
  border: none;
  border-radius: 0.5rem;
  background: var(--surface-container-high);
  color: var(--color-on-surface-variant);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.home-btn:hover {
  background: var(--color-primary-container);
  color: var(--color-primary);
}

.share-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  max-width: 56rem;
  width: 100%;
  margin: 0 auto;
  padding: 0 1rem;
}

.loading-state,
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  text-align: center;
  padding: 2rem;
}

.loading-icon {
  width: 2.5rem;
  height: 2.5rem;
  color: var(--color-primary);
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-icon {
  width: 3rem;
  height: 3rem;
  color: var(--color-error);
  margin-bottom: 1rem;
}

.error-state h2 {
  margin: 0 0 0.5rem;
  font-size: 1.25rem;
  color: var(--color-on-background);
}

.error-state p {
  margin: 0 0 1.5rem;
  color: var(--color-on-surface-variant);
}

.retry-btn {
  padding: 0.625rem 1.25rem;
  border: none;
  border-radius: 0.5rem;
  background: var(--color-primary);
  color: var(--color-on-primary);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: filter 0.15s ease;
}

.retry-btn:hover {
  filter: brightness(1.1);
}

.session-info {
  padding: 1.5rem 0;
  border-bottom: 1px solid var(--color-surface-container-high);
}

.session-title {
  margin: 0 0 0.5rem;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-on-background);
}

.session-meta {
  font-size: 0.8125rem;
  color: var(--color-on-surface-variant);
}
</style>