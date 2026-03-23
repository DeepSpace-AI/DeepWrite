<script setup lang="ts">
</script>

<template>
  <section class="dot-grid relative h-dvh overflow-hidden">
    <div class="glow-blob -left-24 top-0 h-72 w-72 bg-[var(--glow-secondary)] opacity-40" />
    <div class="glow-blob -right-20 bottom-0 h-80 w-80 bg-[var(--glow-primary)] opacity-30" />
    <div class="glow-blob left-1/3 top-1/4 h-48 w-48 -translate-x-1/2 rounded-full bg-[var(--glow-secondary)] opacity-20" />

    <div class="relative mx-auto h-full px-4 py-4 lg:px-5">
      <div class="paper-panel h-full rounded-lg p-2">
        <RouterView />
      </div>
    </div>
  </section>
</template>
