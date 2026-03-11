import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import defaultLayout from '@/layouts/defaultLayout.vue'
import workbenchLayout from '@/layouts/workbenchLayout.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'defaultLayout',
    component: defaultLayout,
    children: [
      {
        path: '',
        name: 'home',
        component: () => import('@/views/HomeView.vue'),
      },
      {
        path: 'login',
        name: 'login',
        component: () => import('@/views/LoginView.vue'),
      },
      {
        path: 'register',
        name: 'register',
        component: () => import('@/views/RegisterView.vue'),
      },
      {
        path: 'forgot-password',
        name: 'forgot-password',
        component: () => import('@/views/ForgotPasswordView.vue'),
      },
      {
        path: '/documents/:id',
        name: 'document-collab',
        component: () => import('@/views/DocumentCollabView.vue'),
      },
    ],
  },
  {
    path: '/dashboard',
    name: 'workbenchLayout',
    component: workbenchLayout,
    children: [
      {
        path: '',
        name: 'dashboard',
        component: () => import('@/views/DashboardView.vue'),
      },
      {
        path: 'workspaces',
        name: 'workspace-list',
        component: () => import('@/views/workspace/WorkspaceListView.vue'),
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router
