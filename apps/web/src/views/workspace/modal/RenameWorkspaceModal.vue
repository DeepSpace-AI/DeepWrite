<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = defineProps<{
  modelValue: boolean
  workspaceName?: string
  submitting?: boolean
  errorMessage?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'submit', payload: { name: string }): void
}>()

const localName = ref('')
const localError = ref('')

const visible = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      localName.value = props.workspaceName || ''
      localError.value = ''
      return
    }

    localName.value = ''
    localError.value = ''
  },
)

watch(
  () => props.workspaceName,
  (name) => {
    if (props.modelValue) {
      localName.value = name || ''
    }
  },
)

function handleClose() {
  visible.value = false
}

function handleSubmit() {
  const name = localName.value.trim()
  if (!name) {
    localError.value = '请输入工作空间名称。'
    return
  }

  if (name.length > 255) {
    localError.value = '工作空间名称不能超过 255 字符。'
    return
  }

  localError.value = ''
  emit('submit', { name })
}
</script>

<template>
  <dialog class="modal" :class="{ 'modal-open': visible }" @close="handleClose">
    <div class="modal-box rounded-sm no-line bg-(--surface-raised) shadow-[0_18px_42px_oklch(0.36_0.008_105/0.12)]">
      <h3 class="text-lg font-semibold text-base-content">重命名工作空间</h3>
      <p class="mt-1 text-sm text-base-content/60">更新工作空间名称，成员会看到新的名称。</p>

      <div class="mt-4 space-y-4">
        <label class="fieldset">
          <legend class="fieldset-legend text-sm">名称</legend>
          <input
            v-model="localName"
            type="text"
            class="input input-bordered w-full rounded-sm"
            placeholder="请输入新的工作空间名称"
            maxlength="255"
          />
        </label>

        <p v-if="localError || errorMessage" class="rounded-sm border border-error/30 bg-error/10 px-3 py-2 text-sm text-error">
          {{ localError || errorMessage }}
        </p>
      </div>

      <div class="modal-action">
        <button type="button" class="btn btn-ghost rounded-sm" :disabled="submitting" @click="handleClose">取消</button>
        <button type="button" class="btn btn-primary rounded-sm" :disabled="submitting" @click="handleSubmit">
          {{ submitting ? '保存中...' : '确认重命名' }}
        </button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button @click="handleClose">close</button>
    </form>
  </dialog>
</template>



