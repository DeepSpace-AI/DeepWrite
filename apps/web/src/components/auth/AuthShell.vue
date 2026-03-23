<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps<{
  eyebrow: string
  title: string
  description: string
  noteTitle: string
  noteBody: string
  quote: string
}>()
</script>

<template>
  <main class="dot-grid relative overflow-hidden py-14 lg:py-20">
    <div class="glow-blob -top-32 right-12 size-80 bg-[var(--glow-primary)] opacity-30" />
    <div class="glow-blob bottom-0 left-0 size-72 bg-[var(--glow-secondary)] opacity-25" />

    <div class="relative mx-auto grid max-w-6xl gap-10 px-4 lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)] lg:items-start">
      <section class="max-w-xl min-w-0 pt-4 lg:pt-8">
        <p class="label-sm py-3">
          {{ eyebrow }}
        </p>
        <div class="mt-8 space-y-5">
          <p class="label-sm">{{ t('auth.accessLabel') }}</p>
          <h1 class="text-editorial text-4xl font-bold leading-[1.08] lg:text-5xl">
            {{ title }}
          </h1>
          <p class="body-lg text-pretty-secondary">
            {{ description }}
          </p>
          <blockquote class="paper-quote body-md">
            {{ quote }}
          </blockquote>
        </div>

        <div class="mt-8 grid gap-4 text-sm">
          <div class="paper-card rounded-md p-5">
            <p class="label-sm mb-2">{{ noteTitle }}</p>
            <p class="body-md text-pretty-secondary">{{ noteBody }}</p>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="paper-card rounded-md p-5">
              <p class="label-sm mb-2">Session</p>
              <p class="body-md text-pretty-secondary">{{ t('auth.sessionText') }}</p>
            </div>
            <div class="paper-card rounded-md p-5">
              <p class="label-sm mb-2">Recovery</p>
              <p class="body-md text-pretty-secondary">{{ t('auth.recoveryText') }}</p>
            </div>
          </div>
        </div>
      </section>

      <section class="paper-card relative min-w-0 overflow-hidden rounded-lg">
        <div class="bg-[var(--surface-raised)] px-6 py-3">
          <p class="label-sm">Authorization Dossier</p>
        </div>
        <div class="bg-[var(--surface-base)]/50 px-6 py-2">
          <p class="label-sm opacity-60">Verified access / publication workspace / secure recovery</p>
        </div>
        <div class="p-6 lg:p-8">
          <slot />
        </div>
      </section>
    </div>
  </main>
</template>

<style scoped>
.dot-grid {
  background-color: var(--surface-base);
  background-image: radial-gradient(
    oklch(0.70 0.005 100 / 0.45) 1px,
    transparent 1px
  );
  background-size: 28px 28px;
}
</style>
