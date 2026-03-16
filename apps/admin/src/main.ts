import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './style.css'

import App from './App.vue'
import router from './router'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'
import { fetchCurrentUserProfile } from '@/api/auth'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

const authStore = useAuthStore(pinia)
const userStore = useUserStore(pinia)
authStore.initializeFromStorage()
userStore.initializeFromStorage()

if (authStore.accessToken) {
  fetchCurrentUserProfile().catch(() => {})
}

app.mount('#app')
