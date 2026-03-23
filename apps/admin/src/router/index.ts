import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
  },
  {
    path: '/',
    component: () => import('@/layouts/AdminLayout.vue'),
    meta: { requiresAuth: true, requiresAdmin: true },
    children: [
      {
        path: '',
        name: 'dashboard',
        component: () => import('@/views/DashboardView.vue'),
      },
      {
        path: 'providers',
        name: 'providers',
        component: () => import('@/views/providers/ProvidersView.vue'),
      },
      {
        path: 'providers/:providerId',
        name: 'provider-detail',
        component: () => import('@/views/providers/ProviderDetailView.vue'),
      },
      {
        path: 'users',
        name: 'users',
        component: () => import('@/views/users/UsersView.vue'),
      },
      {
        path: 'users-new',
        name: 'users-new',
        component: () => import('@/views/admin/UserListView.vue'),
      },
      {
        path: 'workspaces',
        name: 'workspaces',
        component: () => import('@/views/workspaces/WorkspacesView.vue'),
      },
      {
        path: 'system',
        name: 'system',
        component: () => import('@/views/system/SystemSettingsView.vue'),
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach((to) => {
  const authStore = useAuthStore()
  const userStore = useUserStore()

  if (to.matched.some(record => record.meta.requiresAuth)) {
    if (!authStore.hasValidAccessToken()) {
      authStore.clearTokens()
      userStore.clearUser()
      return {
        name: 'login',
        query: { redirect: to.fullPath },
      }
    }
  }

  if (to.matched.some(record => record.meta.requiresAdmin)) {
    const role = userStore.user?.role?.toLowerCase() || ''
    if (role && role !== 'admin') {
      authStore.clearTokens()
      userStore.clearUser()
      return { name: 'login' }
    }
  }

  if (to.name === 'login' && authStore.hasValidAccessToken()) {
    const role = userStore.user?.role?.toLowerCase()
    if (role === 'admin') {
      return { name: 'dashboard' }
    }
  }

  return true
})

export default router
