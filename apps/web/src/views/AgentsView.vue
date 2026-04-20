<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import IconSend from '~icons/mdi/send'
import IconAttachFile from '~icons/mdi/paperclip'
import IconAutoAwesome from '~icons/mdi/auto-fix'
import IconSummarize from '~icons/mdi/text-box-outline'
import IconFactCheck from '~icons/material-symbols/fact-check'
import IconLink from '~icons/mdi/link'
import IconRule from '~icons/mdi/ruler'
import IconTranslate from '~icons/mdi/translate'
import IconMenuBook from '~icons/mdi/book-open-page-variant'
import IconPlusCircle from '~icons/mdi/plus-circle'

const { t } = useI18n()

const inputMessage = ref('')

const sessions = [
  {
    id: '1',
    title: t('agents.sessions.digital'),
    preview: t('agents.sessions.digitalPreview'),
    date: 'today',
  },
  {
    id: '2',
    title: t('agents.sessions.neuroplasticity'),
    preview: t('agents.sessions.neuroplasticityPreview'),
    date: 'today',
  },
  {
    id: '3',
    title: t('agents.sessions.socratic'),
    preview: t('agents.sessions.socraticPreview'),
    date: 'lastWeek',
  },
]

const messages = [
  {
    id: '1',
    role: 'assistant',
    content: t('agents.messages.assistant1'),
  },
  {
    id: '2',
    role: 'assistant',
    content: t('agents.messages.assistant2'),
  },
  {
    id: '3',
    role: 'user',
    content: t('agents.messages.user1'),
  },
]

const tools = [
  { id: 'citation', icon: IconLink, label: t('agents.tools.citation') },
  { id: 'logic', icon: IconRule, label: t('agents.tools.logic') },
  { id: 'translate', icon: IconTranslate, label: t('agents.tools.translate') },
  { id: 'abstract', icon: IconMenuBook, label: t('agents.tools.abstract') },
]

const documents = [
  { id: '1', name: 'Borges_Library_Of_Babel.pdf', description: t('agents.docs.borges') },
  { id: '2', name: 'Attention_Economy_Study.md', description: t('agents.docs.attention') },
]
</script>

<template>
  <section class="scholar-agents flex h-[calc(100vh-4rem)]">
    <!-- Left Panel: Session History -->
    <aside class="scholar-sidebar-left w-72 flex-shrink-0 overflow-y-auto p-6">
      <div class="mb-6 flex items-center justify-between">
        <h2 class="text-editorial text-lg text-[var(--color-on-background)]">
          {{ t('agents.history') }}
        </h2>
        <button type="button" class="text-[var(--color-on-surface-variant)] hover:text-[var(--color-primary)]">
          <IconPlusCircle class="h-5 w-5" />
        </button>
      </div>

      <div class="space-y-4">
        <div v-for="session in sessions" :key="session.id" class="group cursor-pointer">
          <p v-if="session.date === 'today'" class="label-sm mb-1 text-[var(--color-on-surface-variant)]">
            {{ t('agents.today') }}
          </p>
          <p v-else-if="session.date === 'lastWeek'" class="label-sm mb-1 mt-4 text-[var(--color-on-surface-variant)]">
            {{ t('agents.lastWeek') }}
          </p>
          <div class="scholar-session rounded-lg p-3 transition-all">
            <h3 class="label-sm mb-1 font-medium">{{ session.title }}</h3>
            <p class="text-xs text-[var(--color-on-surface-variant)] line-clamp-1">{{ session.preview }}</p>
          </div>
        </div>
      </div>
    </aside>

    <!-- Center Panel: Chat Interface -->
    <main class="scholar-chat relative flex flex-1 flex-col overflow-y-auto">
      <div class="scholar-chat-messages flex-1 px-8 pt-8">
        <div class="mx-auto w-full max-w-3xl flex flex-col gap-8 pb-32">
          <!-- AI Messages -->
          <div v-for="msg in messages" :key="msg.id" class="flex flex-col gap-3" :class="msg.role === 'user' ? 'items-end' : ''">
            <div v-if="msg.role === 'assistant'" class="flex items-center gap-2">
              <IconAutoAwesome class="h-4 w-4 text-[var(--color-primary)]" />
              <span class="label-sm font-medium text-[var(--color-primary)]">{{ t('agents.agentLabel') }}</span>
            </div>

            <div
              class="body-lg leading-relaxed"
              :class="msg.role === 'user' ? 'scholar-user-msg max-w-[85%] rounded-xl p-6' : 'text-[var(--color-on-background)]'"
            >
              <p class="whitespace-pre-line">{{ msg.content }}</p>
            </div>

            <!-- Action Chips -->
            <div v-if="msg.role === 'assistant' && msg.id === '2'" class="flex gap-2">
              <button type="button" class="scholar-chip flex items-center gap-2 rounded-full px-3 py-1.5">
                <IconSummarize class="h-4 w-4" />
                <span class="label-sm">{{ t('agents.actions.summarize') }}</span>
              </button>
              <button type="button" class="scholar-chip flex items-center gap-2 rounded-full px-3 py-1.5">
                <IconFactCheck class="h-4 w-4" />
                <span class="label-sm">{{ t('agents.actions.logicCheck') }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Input Interface -->
      <div class="scholar-input-area absolute bottom-0 left-0 right-0 px-8 pb-8 pt-4">
        <div class="scholar-input-container mx-auto max-w-3xl rounded-xl p-4">
          <div class="flex items-end gap-4">
            <div class="flex-1">
              <label class="label-sm ml-1 mb-1 block text-[var(--color-on-surface-variant)]">
                {{ t('agents.queryLabel') }}
              </label>
              <textarea
                v-model="inputMessage"
                class="w-full resize-none bg-transparent p-1 text-[var(--color-on-surface)] outline-none"
                :placeholder="t('agents.queryPlaceholder')"
                rows="1"
              />
            </div>
            <div class="flex gap-2 pb-1">
              <button type="button" class="scholar-attach-btn rounded-lg p-2">
                <IconAttachFile class="h-5 w-5" />
              </button>
              <button type="button" class="scholar-send-btn rounded-lg p-2.5">
                <IconSend class="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Right Panel: Context & Tools -->
    <aside class="scholar-sidebar-right w-80 flex-shrink-0 overflow-y-auto p-6">
      <div class="mb-6">
        <h2 class="label-sm font-bold text-[var(--color-on-background)]">{{ t('agents.agentTitle') }}</h2>
        <p class="label-sm mt-1 text-[var(--color-on-surface-variant)]">{{ t('agents.agentSubtitle') }}</p>
      </div>

      <!-- Tabs -->
      <nav class="scholar-tabs mb-6 flex gap-6 py-2">
        <button type="button" class="scholar-tab label-sm font-semibold text-[var(--color-tertiary)]">
          {{ t('agents.tabs.context') }}
        </button>
        <button type="button" class="scholar-tab label-sm text-[var(--color-on-surface-variant)]">
          {{ t('agents.tabs.history') }}
        </button>
        <button type="button" class="scholar-tab label-sm text-[var(--color-on-surface-variant)]">
          {{ t('agents.tabs.prompts') }}
        </button>
      </nav>

      <!-- Active Documents -->
      <div class="mb-8">
        <h4 class="label-sm mb-4 uppercase tracking-widest text-[var(--color-on-surface-variant)]">
          {{ t('agents.activeDocs') }}
        </h4>
        <div v-for="doc in documents" :key="doc.id" class="scholar-doc-card mb-3 rounded-lg p-4">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-[var(--color-tertiary)]">{{ doc.name }}</span>
          </div>
          <p class="body-sm mt-2 text-[var(--color-on-surface-variant)]">{{ doc.description }}</p>
        </div>
      </div>

      <!-- Agent Toolbox -->
      <div class="scholar-toolbox mb-8 pt-4">
        <h4 class="label-sm mb-4 uppercase tracking-widest text-[var(--color-on-surface-variant)]">
          {{ t('agents.toolbox') }}
        </h4>
        <div class="grid grid-cols-2 gap-2">
          <button
            v-for="tool in tools"
            :key="tool.id"
            type="button"
            class="scholar-tool-btn flex flex-col items-center gap-2 rounded-lg p-3"
          >
            <component :is="tool.icon" class="h-5 w-5 text-[var(--color-tertiary)]" />
            <span class="text-xs">{{ tool.label }}</span>
          </button>
        </div>
      </div>

      <!-- Session Stats -->
      <div class="scholar-stats rounded-xl p-6">
        <p class="label-sm mb-4 uppercase text-[var(--color-tertiary)]">{{ t('agents.sessionPower') }}</p>
        <div class="space-y-3">
          <div class="flex justify-between text-sm">
            <span class="text-[var(--color-on-surface-variant)]">{{ t('agents.sourcesAnalyzed') }}</span>
            <span class="font-bold text-[var(--color-tertiary)]">14</span>
          </div>
          <div class="h-1 w-full rounded-full bg-[var(--color-tertiary-container)]">
            <div class="h-1 w-3/4 rounded-full bg-[var(--color-tertiary)]" />
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-[var(--color-on-surface-variant)]">{{ t('agents.semanticDepth') }}</span>
            <span class="font-bold text-[var(--color-tertiary)]">{{ t('agents.level4') }}</span>
          </div>
        </div>
      </div>
    </aside>
  </section>
</template>

<style scoped>
.scholar-agents {
  background-color: var(--surface);
}

.scholar-sidebar-left {
  background-color: var(--surface-container);
}

.scholar-session {
  background-color: transparent;
}

.scholar-session:hover {
  background-color: var(--surface-container-high);
}

.scholar-chat {
  background-color: var(--surface);
}

.scholar-chat-messages {
  background-color: var(--surface);
}

.scholar-input-area {
  background: linear-gradient(to top, var(--surface) 70%, transparent);
}

.scholar-user-msg {
  background-color: var(--surface-container-lowest);
  border-left: 4px solid var(--color-primary);
  font-style: italic;
  box-shadow: 0 4px 16px oklch(0.28 0.008 105 / 0.04);
}

.scholar-chip {
  background-color: var(--surface-container-high);
  color: var(--color-on-surface-variant);
  transition: background-color 0.2s;
}

.scholar-chip:hover {
  background-color: var(--color-primary-container);
}

.scholar-input-container {
  background-color: var(--surface-container-lowest);
  box-shadow: 0 8px 32px oklch(0.28 0.008 105 / 0.08);
}

.scholar-attach-btn {
  color: var(--color-on-surface-variant);
  transition: background-color 0.2s;
}

.scholar-attach-btn:hover {
  background-color: var(--surface-container-high);
}

.scholar-send-btn {
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-dim));
  color: var(--color-on-primary);
}

.scholar-sidebar-right {
  background-color: var(--surface-container-low);
}

.scholar-tabs {
  border-bottom: 1px solid var(--surface-container-high);
}

.scholar-tab {
  position: relative;
  padding-bottom: 0.5rem;
}

.scholar-tab.text-\[var\(--color-tertiary\)\]::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 2px;
  background-color: var(--color-tertiary);
}

.scholar-doc-card {
  background-color: var(--surface-container-lowest);
}

.scholar-toolbox {
  border-top: 1px solid var(--surface-container-high);
}

.scholar-tool-btn {
  background-color: var(--surface-container);
  color: var(--color-on-surface-variant);
  transition: background-color 0.2s;
}

.scholar-tool-btn:hover {
  background-color: var(--color-secondary-container);
}

.scholar-stats {
  background-color: color-mix(in oklch, var(--color-secondary) 8%, transparent);
  border: 1px solid color-mix(in oklch, var(--color-secondary) 15%, transparent);
}
</style>
