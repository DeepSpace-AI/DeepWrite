<script setup lang="ts">
import { computed, ref, watch } from 'vue'

interface CreateWorkspaceForm {
  name: string
  description: string
  isPublic: boolean
}

const props = defineProps<{
  modelValue: boolean
  submitting?: boolean
  errorMessage?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'submit', payload: { name: string; description: string; public: boolean }): void
}>()

const form = ref<CreateWorkspaceForm>({
  name: '',
  description: '',
  isPublic: false,
})
const localError = ref('')

const visible = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      localError.value = ''
      return
    }

    form.value = { name: '', description: '', isPublic: false }
    localError.value = ''
  },
)

function handleClose() {
  visible.value = false
}

function handleSubmit() {
  const name = form.value.name.trim()
  if (!name) {
    localError.value = '请输入工作空间名称。'
    return
  }

  if (name.length > 255) {
    localError.value = '工作空间名称不能超过 255 字符。'
    return
  }

  if (form.value.description.length > 1000) {
    localError.value = '描述不能超过 1000 字符。'
    return
  }

  localError.value = ''
  emit('submit', {
    name,
    description: form.value.description.trim(),
    public: form.value.isPublic,
  })
}
</script>

<template>
  <dialog class="modal" :class="{ 'modal-open': visible }" @close="handleClose">
    <div class="modal-box rounded-sm border border-base-300">
      <h3 class="text-lg font-semibold text-base-content">新增工作空间</h3>
      <p class="mt-1 text-sm text-base-content/60">创建一个新的工作空间，并将你设为所有者。</p>

      <div class="mt-4 space-y-4">
        <label class="fieldset">
          <legend class="fieldset-legend text-sm">名称</legend>
          <input
            v-model="form.name"
            type="text"
            class="input input-bordered w-full rounded-sm"
            placeholder="例如：技术写作中台"
            maxlength="255"
          />
        </label>

        <label class="fieldset">
          <legend class="fieldset-legend text-sm">描述</legend>
          <textarea
            v-model="form.description"
            class="textarea textarea-bordered min-h-24 w-full rounded-sm"
            placeholder="描述该工作空间的目标、成员和内容范围"
            maxlength="1000"
          />
        </label>

        <label class="label cursor-pointer justify-start gap-3 rounded-sm border border-base-300/70 bg-base-200/60 px-3 py-2">
          <input v-model="form.isPublic" type="checkbox" class="checkbox checkbox-sm rounded-xs" />
          <span class="label-text text-sm text-base-content/70">设为公开工作空间</span>
        </label>

        <p v-if="localError || errorMessage" class="rounded-sm border border-error/30 bg-error/10 px-3 py-2 text-sm text-error">
          {{ localError || errorMessage }}
        </p>
      </div>

      <div class="modal-action">
        <button type="button" class="btn btn-ghost rounded-sm" :disabled="submitting" @click="handleClose">取消</button>
        <button type="button" class="btn btn-primary rounded-sm" :disabled="submitting" @click="handleSubmit">
          {{ submitting ? '创建中...' : '确认创建' }}
        </button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button @click="handleClose">close</button>
    </form>
  </dialog>
</template>
