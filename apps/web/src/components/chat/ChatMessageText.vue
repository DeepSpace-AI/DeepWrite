<script setup lang="ts">
import { computed } from 'vue'
import { marked } from 'marked'
import hljs from 'highlight.js'

const props = defineProps<{
  text: string
  isStreaming?: boolean
}>()

const renderer = new marked.Renderer()

renderer.code = function({ text, lang }: { text: string; lang?: string }) {
  let highlighted: string
  if (lang && hljs.getLanguage(lang)) {
    try {
      highlighted = hljs.highlight(text, { language: lang }).value
    } catch {
      highlighted = hljs.highlightAuto(text).value
    }
  } else {
    highlighted = hljs.highlightAuto(text).value
  }
  return `<pre><code class="hljs language-${lang || 'auto'}">${highlighted}</code></pre>`
}

marked.setOptions({
  renderer,
  breaks: true,
  gfm: true,
})

const renderedContent = computed(() => {
  return marked.parse(props.text) as string
})
</script>

<template>
  <div class="text-part markdown-content" :class="{ streaming: isStreaming }">
    <div v-html="renderedContent" />
    <span v-if="isStreaming" class="cursor" />
  </div>
</template>

<style scoped>
.text-part {
  position: relative;
}

.streaming {
  display: inline;
}

.cursor {
  display: inline-block;
  width: 2px;
  height: 1.2em;
  margin-left: 2px;
  background: var(--color-primary);
  animation: blink 1s infinite;
  vertical-align: text-bottom;
}

@keyframes blink {
  0%, 50% {
    opacity: 1;
  }
  51%, 100% {
    opacity: 0;
  }
}
</style>

<style>
.markdown-content {
  line-height: 1.6;
}

.markdown-content h1,
.markdown-content h2,
.markdown-content h3,
.markdown-content h4,
.markdown-content h5,
.markdown-content h6 {
  margin-top: 1.25em;
  margin-bottom: 0.5em;
  font-weight: 600;
  line-height: 1.3;
}

.markdown-content h1 { font-size: 1.5em; }
.markdown-content h2 { font-size: 1.35em; }
.markdown-content h3 { font-size: 1.2em; }
.markdown-content h4 { font-size: 1.1em; }
.markdown-content h5 { font-size: 1em; }
.markdown-content h6 { font-size: 0.9em; color: var(--color-on-surface-variant); }

.markdown-content p {
  margin: 0.75em 0;
}

.markdown-content p:first-child {
  margin-top: 0;
}

.markdown-content p:last-child {
  margin-bottom: 0;
}

.markdown-content a {
  color: var(--color-primary);
  text-decoration: none;
}

.markdown-content a:hover {
  text-decoration: underline;
}

.markdown-content strong {
  font-weight: 600;
}

.markdown-content em {
  font-style: italic;
}

.markdown-content ul,
.markdown-content ol {
  margin: 0.75em 0;
  padding-left: 1.5em;
}

.markdown-content li {
  margin: 0.25em 0;
}

.markdown-content li > ul,
.markdown-content li > ol {
  margin: 0.125em 0;
}

.markdown-content blockquote {
  margin: 0.75em 0;
  padding: 0.5em 1em;
  border-left: 3px solid var(--color-primary);
  background: var(--surface-container-low);
  color: var(--color-on-surface-variant);
}

.markdown-content blockquote p {
  margin: 0.25em 0;
}

.markdown-content code {
  font-family: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 0.875em;
  padding: 0.125em 0.375em;
  background: var(--surface-container);
  border-radius: 0.25rem;
}

.markdown-content pre {
  margin: 0.75em 0;
  padding: 0.75rem 1rem;
  background: var(--surface-container);
  border-radius: 0.5rem;
  overflow-x: auto;
}

.markdown-content pre code {
  padding: 0;
  background: transparent;
  font-size: 0.8125rem;
  line-height: 1.5;
}

.markdown-content hr {
  margin: 1.5em 0;
  border: none;
  border-top: 1px solid var(--color-surface-container-high);
}

.markdown-content table {
  margin: 0.75em 0;
  border-collapse: collapse;
  width: 100%;
}

.markdown-content th,
.markdown-content td {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--color-surface-container-high);
  text-align: left;
}

.markdown-content th {
  background: var(--surface-container-low);
  font-weight: 600;
}

.markdown-content img {
  max-width: 100%;
  height: auto;
  border-radius: 0.5rem;
}

.hljs {
  color: var(--color-on-background);
  background: transparent;
}

.hljs-keyword,
.hljs-selector-tag,
.hljs-built_in,
.hljs-name,
.hljs-tag {
  color: #c678dd;
}

.hljs-string,
.hljs-title,
.hljs-section,
.hljs-attribute,
.hljs-literal,
.hljs-template-tag,
.hljs-template-variable,
.hljs-type {
  color: #98c379;
}

.hljs-comment,
.hljs-deletion {
  color: #5c6370;
  font-style: italic;
}

.hljs-number,
.hljs-regexp,
.hljs-addition {
  color: #d19a66;
}

.hljs-function {
  color: #61afef;
}

.hljs-variable,
.hljs-params {
  color: #e06c75;
}

.hljs-selector-class,
.hljs-selector-id {
  color: #e5c07b;
}
</style>