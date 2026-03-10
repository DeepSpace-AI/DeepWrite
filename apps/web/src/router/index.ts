import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import DocumentCollabView from '@/views/DocumentCollabView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/documents/:id',
      name: 'document-collab',
      component: DocumentCollabView,
    },
  ],
})

export default router
