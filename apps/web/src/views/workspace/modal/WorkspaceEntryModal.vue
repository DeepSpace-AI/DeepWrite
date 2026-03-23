<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = withDefaults(defineProps<{
  modelValue: boolean
  title: string
  subtitle?: string
  nameLabel: string
  namePlaceholder?: string
  descriptionLabel?: string
  descriptionPlaceholder?: string
  submitText: string
  submitting?: boolean
  errorMessage?: string
  initialName?: string
  initialDescription?: string
  showDescription?: boolean
}>(), {
  subtitle: '',
  namePlaceholder: '',
  descriptionLabel: '',
  descriptionPlaceholder: '',
  submitting: false,
  errorMessage: '',
  initialName: '',
  initialDescription: '',
  showDescription: false,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'submit', payload: { name: string; description: string }): void
}>()

const localName = ref('')
const localDescription = ref('')
const localError = ref('')

const visible = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      localName.value = props.initialName
      localDescription.value = props.initialDescription
      localError.value = ''
      return
    }

    localName.value = ''
    localDescription.value = ''
    localError.value = ''
  },
)

watch(
  () => [props.initialName, props.initialDescription],
  ([name, description]) => {
    if (!props.modelValue) return
    localName.value = name || ''
    localDescription.value = description || ''
  },
)

function handleClose() {
  visible.value = false
}

function handleSubmit() {
  const name = localName.value.trim()
  if (!name) {
    localError.value = '请输入名称。'
    return
  }
  if (name.length > 255) {
    localError.value = '名称不能超过 255 字符。'
    return
  }
  if (localDescription.value.length > 1000) {
    localError.value = '描述不能超过 1000 字符。'
    return
  }

  localError.value = ''
  emit('submit', {
    name,
    description: localDescription.value.trim(),
  })
}
</script>

<template>
  <dialog class="modal" :class="{ 'modal-open': visible }" @close="handleClose">
    <div class="modal-box rounded-sm no-line bg-(--surface-raised) shadow-[0_18px_42px_oklch(0.36_0.008_105/0.12)]">
      <h3 class="text-lg font-semibold text-base-content">{{ title }}</h3>
      <p v-if="subtitle" class="mt-1 text-sm text-base-content/60">{{ subtitle }}</p>

      <div class="mt-4 space-y-4">
        <label class="fieldset">
          <legend class="fieldset-legend text-sm">{{ nameLabel }}</legend>
          <input
            v-model="localName"
            type="text"
            class="input input-bordered w-full rounded-sm"
            :placeholder="namePlaceholder"
            maxlength="255"
          />
        </label>

        <label v-if="showDescription" class="fieldset">
          <legend class="fieldset-legend text-sm">{{ descriptionLabel }}</legend>
          <textarea
            v-model="localDescription"
            class="textarea textarea-bordered min-h-24 w-full rounded-sm"
            :placeholder="descriptionPlaceholder"
            maxlength="1000"
          />
        </label>

        <p v-if="localError || errorMessage" class="rounded-sm border border-error/30 bg-error/10 px-3 py-2 text-sm text-error">
          {{ localError || errorMessage }}
        </p>
      </div>

      <div class="modal-action">
        <button type="button" class="btn btn-ghost rounded-sm" :disabled="submitting" @click="handleClose">取消</button>
        <button type="button" class="btn btn-primary rounded-sm" :disabled="submitting" @click="handleSubmit">
          {{ submitting ? '处理中...' : submitText }}
        </button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button @click="handleClose">close</button>
    </form>
  </dialog>
</template>



