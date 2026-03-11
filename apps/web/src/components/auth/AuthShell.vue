<script setup lang="ts">
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
    <div class="pointer-events-none absolute inset-x-0 top-0 h-px bg-linear-to-r from-transparent via-primary/30 to-transparent" />
    <div class="pointer-events-none absolute -top-32 right-12 size-80 rounded-full bg-primary/6 blur-3xl" />
    <div class="pointer-events-none absolute bottom-0 left-0 size-72 rounded-full bg-secondary/8 blur-3xl" />

    <div class="relative mx-auto grid max-w-6xl gap-10 px-4 lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)] lg:items-start">
      <section class="max-w-xl min-w-0 pt-4 lg:pt-8">
        <div class="border-y border-base-300/80 py-3 text-[11px] font-mono uppercase tracking-[0.22em] text-base-content/45">
          {{ eyebrow }}
        </div>
        <div class="mt-8 space-y-5">
          <div class="text-xs font-mono uppercase tracking-[0.24em] text-primary/70">权限访问</div>
          <h1 class="heading-serif text-4xl font-bold leading-[1.08] text-base-content lg:text-5xl">
            {{ title }}
          </h1>
          <p class="max-w-lg text-base leading-8 text-base-content/64">
            {{ description }}
          </p>
          <blockquote class="max-w-md border-l-2 border-secondary/60 pl-4 text-sm italic leading-7 text-base-content/58">
            {{ quote }}
          </blockquote>
        </div>

        <div class="mt-8 grid gap-3 text-sm text-base-content/58">
          <div class="publication-note rounded-sm border border-base-300 bg-base-100/80 p-4">
            <div class="text-[11px] font-mono uppercase tracking-[0.2em] text-base-content/38">{{ noteTitle }}</div>
            <p class="mt-2 leading-7">{{ noteBody }}</p>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div class="rounded-sm border border-base-300 bg-base-100/80 p-4">
              <div class="text-[11px] font-mono uppercase tracking-[0.2em] text-base-content/38">Session</div>
              <p class="mt-2 leading-7">短会话登录与设备记忆，便于持续写作。</p>
            </div>
            <div class="rounded-sm border border-base-300 bg-base-100/80 p-4">
              <div class="text-[11px] font-mono uppercase tracking-[0.2em] text-base-content/38">Recovery</div>
              <p class="mt-2 leading-7">支持邮箱找回与后续密码重置流程接入。</p>
            </div>
          </div>
        </div>
      </section>

      <section class="auth-paper relative min-w-0 overflow-hidden rounded-sm border border-base-300 bg-base-100 shadow-2xl">
        <div class="border-b border-base-300 bg-base-200 px-6 py-3 text-[11px] font-mono uppercase tracking-[0.22em] text-base-content/40">
          Authorization Dossier
        </div>
        <div class="border-b border-base-300/70 px-6 py-2 text-[10px] font-mono uppercase tracking-[0.22em] text-base-content/35">
          Verified access / publication workspace / secure recovery
        </div>
        <div class="p-6 lg:p-8">
          <slot />
        </div>
      </section>
    </div>
  </main>
</template>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Playfair+Display:ital,wght@0,600;0,700;1,600&display=swap');

.heading-serif {
  font-family: 'Playfair Display', Georgia, serif;
}

.dot-grid {
  background-color: var(--color-base-100);
  background-image: radial-gradient(
    circle,
    color-mix(in srgb, var(--color-base-content) 7%, transparent) 1.5px,
    transparent 1.5px
  );
  background-size: 28px 28px;
}

.auth-paper {
  box-shadow:
    0 30px 70px color-mix(in srgb, var(--color-neutral) 12%, transparent),
    18px 18px 0 color-mix(in srgb, var(--color-base-300) 50%, transparent);
}

.publication-note {
  backdrop-filter: blur(8px);
}
</style>
