<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <section class="h-full min-h-0 rounded-sm border border-base-300 bg-base-100 shadow-sm flex flex-col">
    <div class="border-b border-base-300 px-4 py-3">
      <h3 class="text-base font-semibold text-base-content">{{ t('workspace.detail.navAnalysis') }}</h3>
    </div>

    <div class="min-h-0 flex-1 space-y-3 overflow-y-auto p-4">
      <div class="rounded-sm border border-dashed border-base-300 bg-base-200/30 px-4 py-12 text-center">
        <p class="text-sm font-semibold text-base-content">{{ t('workspace.detail.navAnalysis') }}</p>
        <p class="mt-2 text-sm text-base-content/60">{{ t('workspace.detail.analysisPlaceholder') }}</p>
      </div>
    </div>
  </section>
</template>
