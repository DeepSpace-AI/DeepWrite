import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import defaultLayout from '@/layouts/defaultLayout.vue'
import workbenchLayout from '@/layouts/workbenchLayout.vue'
import workspaceDetailLayout from '@/layouts/workspaceDetailLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'

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
      }
    ],
  },
  {
    path: '/',
    name: 'workbenchLayout',
    component: workbenchLayout,
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('@/views/DashboardView.vue'),
      },
      {
        path: 'workspaces',
        name: 'workspace-list',
        component: () => import('@/views/workspace/WorkspaceListView.vue'),
      },
      {
        path: 'workspaces/:id',
        redirect: (to) => ({ name: 'workspace-detail', params: { id: to.params.id }, query: to.query }),
      },
      {
        path: 'profile',
        name: 'profile',
        component: () => import('@/views/ProfileView.vue'),
      },
      {
        path: 'library',
        name: 'library',
        component: () => import('@/views/LibraryView.vue'),
      },
      {
        path: 'skills',
        name: 'skills',
        component: () => import('@/views/SkillsView.vue'),
      },
      {
        path: 'agents',
        name: 'agents',
        component: () => import('@/views/AgentsView.vue'),
      },
      {
        path: 'notifications',
        name: 'notifications',
        component: () => import('@/views/NotificationsView.vue'),
      },
    ],
  },
  {
    path: '/workspace/:id',
    component: workspaceDetailLayout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'workspace-detail',
        component: () => import('@/views/workspace/WorkspaceDetailView.vue'),
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach((to) => {
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth)
  if (!requiresAuth) {
    return true
  }

  const authStore = useAuthStore()

  if (!authStore.hasValidAccessToken()) {
    authStore.clearTokens()
    const userStore = useUserStore()
    userStore.clearUser()

    return {
      name: 'login',
      query: { redirect: to.fullPath },
    }
  }

  return true
})

export default router
