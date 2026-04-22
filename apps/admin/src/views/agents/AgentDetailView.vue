<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ApiError } from '@/api/http'
import {
  getAgent,
  createAgent,
  updateAgent,
} from '@/api/agentAdmin'
import { listAIProviders, type AIProviderModel } from '@/api/aiProvider'
import IconSave from '~icons/mdi/content-save'
import IconArrowLeft from '~icons/mdi/arrow-left'

const route = useRoute()
const router = useRouter()

const isNew = computed(() => route.name === 'agent-create')

const isLoading = ref(false)
const isSaving = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

const agentId = computed(() => route.params.id as string || '')

const availableModels = ref<AIProviderModel[]>([])

const form = ref({
  name: '',
  description: '',
  category: 'research',
  icon_url: '',
  system_prompt: '',
  default_model: '',
  skills: '[]',
  tools: '[]',
  temperature: 0.7,
  max_tokens: 4096,
})

const categoryOptions = [
  { value: 'research', label: '科研助手' },
  { value: 'writing', label: '写作助手' },
  { value: 'data', label: '数据处理' },
  { value: 'publishing', label: '出版辅助' },
  { value: 'general', label: '通用助手' },
]

async function loadAgent() {
  if (isNew.value) return
  
  isLoading.value = true
  try {
    const agent = await getAgent(agentId.value)
    form.value = {
      name: agent.name,
      description: agent.description || '',
      category: agent.category || 'research',
      icon_url: agent.icon_url || '',
      system_prompt: agent.system_prompt || '',
      default_model: agent.default_model || '',
      skills: typeof agent.skills === 'string' ? agent.skills : JSON.stringify(agent.skills || []),
      tools: typeof agent.tools === 'string' ? agent.tools : JSON.stringify(agent.tools || []),
      temperature: agent.temperature || 0.7,
      max_tokens: agent.max_tokens || 4096,
    }
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '加载失败'
  } finally {
    isLoading.value = false
  }
}

async function loadModels() {
  try {
    const res = await listAIProviders({ enabled: true })
    availableModels.value = res.items || []
  } catch (error) {
    console.error('Failed to load models:', error)
  }
}

async function handleSave() {
  if (!form.value.name.trim()) {
    errorMessage.value = '名称不能为空'
    return
  }
  if (!form.value.system_prompt.trim()) {
    errorMessage.value = 'System Prompt 不能为空'
    return
  }

  isSaving.value = true
  errorMessage.value = ''
  
  try {
    const input = {
      name: form.value.name,
      description: form.value.description,
      category: form.value.category,
      icon_url: form.value.icon_url,
      system_prompt: form.value.system_prompt,
      default_model: form.value.default_model,
      skills: form.value.skills,
      tools: form.value.tools,
      temperature: form.value.temperature,
      max_tokens: form.value.max_tokens,
    }

    if (isNew.value) {
      const agent = await createAgent(input)
      successMessage.value = '创建成功'
      router.replace(`/agents/${agent.id}`)
    } else {
      await updateAgent(agentId.value, input)
      successMessage.value = '保存成功'
    }
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '保存失败'
  } finally {
    isSaving.value = false
  }
}

onMounted(() => {
  loadAgent()
  loadModels()
})
</script>

<template>
  <section class="space-y-6">
    <header class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <button
          class="rounded p-2 text-[var(--text-secondary)] hover:bg-[var(--bg-card)]"
          @click="router.push('/agents')"
        >
          <IconArrowLeft class="h-5 w-5" />
        </button>
        <div>
          <h1 class="text-display text-2xl text-[var(--text-primary)]">
            {{ isNew ? '创建官方 Agent' : '编辑 Agent' }}
          </h1>
          <p class="text-secondary mt-1 text-sm">
            {{ isNew ? '创建新的官方 Agent 助手' : '修改 Agent 配置' }}
          </p>
        </div>
      </div>
      <button
        class="btn-tech flex items-center gap-1"
        :disabled="isSaving"
        @click="handleSave"
      >
        <IconSave class="h-4 w-4" />
        {{ isSaving ? '保存中...' : '保存' }}
      </button>
    </header>

    <div v-if="errorMessage" class="tech-card text-sm text-[var(--accent-error)]">
      {{ errorMessage }}
    </div>
    <div v-if="successMessage" class="tech-card bg-[var(--accent-success)]/10 text-sm text-[var(--accent-success)]">
      {{ successMessage }}
    </div>

    <div v-if="isLoading" class="flex justify-center py-12">
      <span class="loading loading-spinner text-[var(--accent-primary)]" />
    </div>

    <div v-else class="grid gap-6 lg:grid-cols-3">
      <div class="lg:col-span-2 space-y-6">
        <div class="tech-card p-6">
          <h2 class="text-heading mb-4 text-lg">基本信息</h2>
          
          <div class="space-y-4">
            <div>
              <label class="text-label mb-1 block">名称 *</label>
              <input
                v-model="form.name"
                type="text"
                class="tech-input w-full"
                placeholder="输入 Agent 名称"
              >
            </div>

            <div>
              <label class="text-label mb-1 block">分类</label>
              <select v-model="form.category" class="tech-input w-full">
                <option v-for="cat in categoryOptions" :key="cat.value" :value="cat.value">
                  {{ cat.label }}
                </option>
              </select>
            </div>

            <div>
              <label class="text-label mb-1 block">描述</label>
              <textarea
                v-model="form.description"
                class="tech-input w-full"
                rows="3"
                placeholder="描述这个 Agent 的功能和用途"
              />
            </div>

            <div>
              <label class="text-label mb-1 block">图标 URL</label>
              <input
                v-model="form.icon_url"
                type="text"
                class="tech-input w-full"
                placeholder="https://example.com/icon.png"
              >
            </div>
          </div>
        </div>

        <div class="tech-card p-6">
          <h2 class="text-heading mb-4 text-lg">System Prompt *</h2>
          <textarea
            v-model="form.system_prompt"
            class="tech-input font-mono text-sm"
            rows="12"
            placeholder="定义 Agent 的角色、行为和能力..."
          />
        </div>
      </div>

      <div class="space-y-6">
        <div class="tech-card p-6">
          <h2 class="text-heading mb-4 text-lg">模型配置</h2>
          
          <div class="space-y-4">
            <div>
              <label class="text-label mb-1 block">默认模型</label>
              <select v-model="form.default_model" class="tech-input w-full">
                <option value="">选择模型</option>
                <option v-for="model in availableModels" :key="model.id" :value="model.model">
                  {{ model.model }}
                </option>
              </select>
            </div>

            <div>
              <label class="text-label mb-1 block">Temperature</label>
              <input
                v-model.number="form.temperature"
                type="number"
                step="0.1"
                min="0"
                max="2"
                class="tech-input w-full"
              >
            </div>

            <div>
              <label class="text-label mb-1 block">Max Tokens</label>
              <input
                v-model.number="form.max_tokens"
                type="number"
                min="1"
                max="128000"
                class="tech-input w-full"
              >
            </div>
          </div>
        </div>

        <div class="tech-card p-6">
          <h2 class="text-heading mb-4 text-lg">Skills (JSON)</h2>
          <textarea
            v-model="form.skills"
            class="tech-input font-mono text-sm"
            rows="4"
            placeholder='[{"id": "skill-1", "name": "技能名称"}]'
          />
        </div>

        <div class="tech-card p-6">
          <h2 class="text-heading mb-4 text-lg">Tools (JSON)</h2>
          <textarea
            v-model="form.tools"
            class="tech-input font-mono text-sm"
            rows="4"
            placeholder='[{"id": "tool-1", "name": "工具名称", "enabled": true}]'
          />
        </div>
      </div>
    </div>
  </section>
</template>