<script setup lang="ts">
import IconAlert from '~icons/mdi/alert-circle-outline'

defineProps<{
  message: string
  code?: string
}>()
</script>

<template>
  <div class="error-part">
    <IconAlert class="error-icon" />
    <span class="error-message">{{ message }}</span>
    <span v-if="code" class="error-code">({{ code }})</span>
  </div>
</template>

<style scoped>
.error-part {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  padding: 0.75rem;
  border-radius: 0.5rem;
  background: oklch(0.25 0.02 25 / 0.1);
  border: 1px solid oklch(0.5 0.15 25 / 0.3);
}

.error-icon {
  flex-shrink: 0;
  width: 1.25rem;
  height: 1.25rem;
  color: oklch(0.6 0.15 25);
  margin-top: 0.125rem;
}

.error-message {
  font-size: 0.875rem;
  color: oklch(0.7 0.1 25);
  line-height: 1.5;
}

.error-code {
  font-size: 0.75rem;
  color: oklch(0.5 0.08 25);
  margin-left: 0.25rem;
}
</style>