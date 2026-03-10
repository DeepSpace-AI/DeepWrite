<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const documentId = ref('')

function openDocument() {
  const id = documentId.value.trim()
  if (!id) return
  void router.push({ name: 'document-collab', params: { id } })
}
</script>

<template>
  <main class="mx-auto flex min-h-screen max-w-xl items-center p-4">
    <section class="w-full rounded-xl border border-base-300 bg-base-100 p-6 shadow-sm">
      <h1 class="mb-2 text-2xl font-semibold">DeepWrite 协作文档</h1>
      <p class="mb-4 text-base-content/70">输入文档 ID 进入实时协作页面。</p>

      <label class="form-control">
        <span class="label-text">文档 ID</span>
        <input v-model="documentId" class="input input-bordered" placeholder="例如：c4d1f2e0-xxxx-xxxx-xxxx-xxxxxxxxxxxx" />
      </label>

      <button class="btn btn-primary mt-4 w-full" @click="openDocument">进入协作页</button>
    </section>
  </main>
</template>
