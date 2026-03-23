<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <section class="paper-panel h-full min-h-0 flex flex-col">
    <div class="bg-(--surface-overlay) px-4 py-3">
      <h3 class="text-base font-semibold text-base-content">{{ t('workspace.detail.navLiterature') }}</h3>
    </div>

    <div class="min-h-0 flex-1 space-y-3 overflow-y-auto p-4">
      <div class="paper-panel-embedded rounded-sm px-4 py-12 text-center">
        <p class="text-sm font-semibold text-base-content">{{ t('workspace.detail.navLiterature') }}</p>
        <p class="mt-2 text-sm text-base-content/60">{{ t('workspace.detail.literaturePlaceholder') }}</p>
      </div>
    </div>
  </section>
</template>



