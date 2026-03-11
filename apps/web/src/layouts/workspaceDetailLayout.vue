<script setup lang="ts">
</script>

<template>
  <section class="dot-grid relative h-dvh overflow-hidden bg-base-100">
    <div class="pointer-events-none absolute -left-24 top-0 h-72 w-72 rounded-full bg-secondary/10 blur-3xl" />
    <div class="pointer-events-none absolute -right-20 bottom-0 h-80 w-80 rounded-full bg-primary/10 blur-3xl" />

    <div class="relative mx-auto h-full px-4 py-4 lg:px-5">
      <div class="h-full">
        <RouterView />
      </div>
    </div>
  </section>
</template>
