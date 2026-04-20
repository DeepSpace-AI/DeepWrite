<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import IconSearch from '~icons/mdi/magnify'
import IconStar from '~icons/mdi/star'
import IconCheckCircle from '~icons/mdi/check-circle'
import IconTranslate from '~icons/mdi/translate'
import IconAutoFix from '~icons/mdi/auto-fix'
import IconFunction from '~icons/mdi/function'
import IconFormatQuote from '~icons/ic/round-format-quote'
import IconRule from '~icons/mdi/ruler'
import IconHub from '~icons/mdi/hub'

const { t } = useI18n()

const searchKeyword = ref('')
const selectedCategory = ref('all')

const categories = [
  { id: 'all', label: t('skills.categories.all') },
  { id: 'writing', label: t('skills.categories.writing') },
  { id: 'research', label: t('skills.categories.research') },
  { id: 'data', label: t('skills.categories.data') },
  { id: 'publishing', label: t('skills.categories.publishing') },
]

const skills = [
  {
    id: '1',
    name: 'Translation Pro',
    description: t('skills.items.translation.description'),
    category: 'research',
    rating: 4.9,
    icon: IconTranslate,
    active: false,
  },
  {
    id: '2',
    name: 'Paper Polishing AI',
    description: t('skills.items.polishing.description'),
    category: 'writing',
    rating: 5.0,
    icon: IconAutoFix,
    active: true,
  },
  {
    id: '3',
    name: 'LaTeX Converter',
    description: t('skills.items.latex.description'),
    category: 'publishing',
    rating: 4.7,
    icon: IconFunction,
    active: false,
  },
  {
    id: '4',
    name: 'BibTeX Formatter',
    description: t('skills.items.bibtex.description'),
    category: 'data',
    rating: 4.8,
    icon: IconFormatQuote,
    active: false,
  },
  {
    id: '5',
    name: 'Academic Style Checker',
    description: t('skills.items.style.description'),
    category: 'writing',
    rating: 4.6,
    icon: IconRule,
    active: false,
  },
  {
    id: '6',
    name: 'Semantic Explorer',
    description: t('skills.items.semantic.description'),
    category: 'research',
    rating: 5.0,
    icon: IconHub,
    active: false,
  },
]

const filteredSkills = ref(skills)

function filterByCategory(categoryId: string) {
  selectedCategory.value = categoryId
  if (categoryId === 'all') {
    filteredSkills.value = skills
  } else {
    filteredSkills.value = skills.filter(s => s.category === categoryId)
  }
}

function getCategoryColor(category: string) {
  switch (category) {
    case 'research':
      return 'bg-[var(--color-tertiary-container)] text-[var(--color-on-tertiary-container)]'
    case 'writing':
      return 'bg-[var(--color-primary-container)] text-[var(--color-on-primary-container)]'
    case 'publishing':
      return 'bg-[var(--surface-container)] text-[var(--color-on-surface)]'
    case 'data':
      return 'bg-[var(--color-secondary-container)] text-[var(--color-on-secondary-container)]'
    default:
      return 'bg-[var(--surface-container-high)] text-[var(--color-on-surface-variant)]'
  }
}
</script>

<template>
  <section class="scholar-skills">
    <!-- Hero Section -->
    <header class="scholar-hero px-8 pt-10 pb-8">
      <div class="mx-auto max-w-6xl">
        <h1 class="text-editorial text-5xl font-light leading-tight text-[var(--color-on-background)]">
          {{ t('skills.title') }}
        </h1>
        <p class="body-lg mt-4 max-w-xl text-[var(--color-on-surface-variant)]">
          {{ t('skills.subtitle') }}
        </p>
      </div>
    </header>

    <!-- Toolbar Section -->
    <div class="scholar-toolbar px-8 py-5">
      <div class="mx-auto flex max-w-6xl flex-wrap items-center gap-4">
        <div class="scholar-search relative flex-1 max-w-md">
          <IconSearch class="absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-[var(--color-on-surface-variant)] opacity-50" />
          <input
            v-model="searchKeyword"
            type="text"
            class="w-full rounded-full py-2.5 pl-11 pr-4"
            :placeholder="t('skills.searchPlaceholder')"
          />
        </div>

        <div class="flex gap-2">
          <button
            v-for="cat in categories"
            :key="cat.id"
            type="button"
            class="scholar-category-btn rounded-full px-5 py-2"
            :class="selectedCategory === cat.id ? 'active' : ''"
            @click="filterByCategory(cat.id)"
          >
            {{ cat.label }}
          </button>
        </div>
      </div>
    </div>

    <!-- Skills Grid Container -->
    <div class="scholar-grid-container px-8 py-8">
      <div class="mx-auto max-w-6xl grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        <article
          v-for="skill in filteredSkills"
          :key="skill.id"
          class="scholar-skill-card group rounded-xl p-6 transition-all"
        >
          <!-- Header -->
          <div class="mb-5 flex items-start justify-between">
            <div class="flex h-12 w-12 items-center justify-center rounded-xl" :class="getCategoryColor(skill.category)">
              <component :is="skill.icon" class="h-6 w-6" />
            </div>
            <div class="flex items-center gap-1">
              <IconStar class="h-4 w-4 text-amber-500" />
              <span class="label-sm font-medium">{{ skill.rating }}</span>
            </div>
          </div>

          <!-- Content -->
          <h3 class="text-editorial mb-3 text-xl font-medium text-[var(--color-on-background)]">
            {{ skill.name }}
          </h3>
          <p class="body-sm mb-6 text-[var(--color-on-surface-variant)]">
            {{ skill.description }}
          </p>

          <!-- Footer -->
          <div class="flex items-center justify-between">
            <span class="scholar-category-tag label-sm rounded-full px-3 py-1 uppercase" :class="getCategoryColor(skill.category)">
              {{ t(`skills.categories.${skill.category}`) }}
            </span>
            <button
              v-if="skill.active"
              type="button"
              class="scholar-active-btn label-sm flex items-center gap-1 font-semibold text-[var(--color-tertiary)]"
            >
              {{ t('skills.active') }}
              <IconCheckCircle class="h-4 w-4" />
            </button>
            <button
              v-else
              type="button"
              class="scholar-activate-btn label-sm rounded-lg px-5 py-2 font-medium"
            >
              {{ t('skills.activate') }}
            </button>
          </div>
        </article>
      </div>
    </div>

    <!-- Promo Banner -->
    <div class="scholar-promo-container px-8 py-8">
      <div class="scholar-promo mx-auto max-w-6xl rounded-xl p-10">
        <div class="flex flex-col gap-8 md:flex-row md:items-center md:justify-between">
          <div>
            <span class="label-sm mb-3 block uppercase tracking-widest text-[var(--color-tertiary)]">
              {{ t('skills.promo.label') }}
            </span>
            <h3 class="text-editorial mb-4 text-3xl leading-tight text-[var(--color-on-background)]">
              {{ t('skills.promo.title') }}
            </h3>
            <p class="body-md max-w-lg text-[var(--color-on-surface-variant)]">
              {{ t('skills.promo.description') }}
            </p>
          </div>
          <button
            type="button"
            class="scholar-promo-btn rounded-lg px-6 py-3"
          >
            {{ t('skills.promo.cta') }}
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.scholar-skills {
  background-color: var(--surface);
  min-height: calc(100vh - 4rem);
}

.scholar-hero {
  background-color: var(--surface-container-low);
}

.scholar-toolbar {
  background-color: var(--surface-container);
}

.scholar-search input {
  background-color: var(--surface-container-lowest);
  color: var(--color-on-surface);
  border: none;
  outline: none;
  font-size: 0.875rem;
}

.scholar-search input:focus {
  background-color: var(--surface-container-lowest);
}

.scholar-category-btn {
  background-color: var(--surface-container-low);
  color: var(--color-on-surface-variant);
  font-size: 0.75rem;
  font-weight: 500;
  white-space: nowrap;
  transition: background-color 0.2s, color 0.2s;
}

.scholar-category-btn:hover {
  background-color: var(--surface-container-high);
}

.scholar-category-btn.active {
  background-color: var(--color-primary);
  color: var(--color-on-primary);
}

.scholar-grid-container {
  background-color: var(--surface-container);
}

.scholar-skill-card {
  background-color: var(--surface-container-lowest);
  box-shadow: 0 4px 24px oklch(0.28 0.008 105 / 0.06);
}

.scholar-skill-card:hover {
  box-shadow: 0 8px 32px oklch(0.28 0.008 105 / 0.10);
  transform: translateY(-4px);
}

[data-theme="vellum-dark"] .scholar-skill-card {
  box-shadow: 0 4px 24px oklch(0.15 0.02 75 / 0.15);
}

[data-theme="vellum-dark"] .scholar-skill-card:hover {
  box-shadow: 0 8px 32px oklch(0.15 0.02 75 / 0.20);
}

.scholar-category-tag {
  font-size: 0.65rem;
}

.scholar-activate-btn {
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-dim));
  color: var(--color-on-primary);
  transition: opacity 0.2s;
}

.scholar-activate-btn:hover {
  opacity: 0.88;
}

.scholar-promo-container {
  background-color: var(--surface);
}

.scholar-promo {
  background-color: var(--surface-container-low);
}

.scholar-promo-btn {
  background-color: var(--color-on-background);
  color: var(--surface-base);
  font-weight: 500;
  transition: opacity 0.2s;
}

.scholar-promo-btn:hover {
  opacity: 0.88;
}
</style>
