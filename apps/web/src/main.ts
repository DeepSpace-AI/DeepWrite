import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './style.css'

import App from './App.vue'
import router from './router'
import { fetchCurrentUserProfile } from './api/auth'
import { useAuthStore } from './stores/auth'
import { useUserStore } from './stores/user'

const app = createApp(App)

const pinia = createPinia()
app.use(pinia)
app.use(router)

// 应用挂载前，从存储恢复用户和令牌信息
const authStore = useAuthStore(pinia)
const userStore = useUserStore(pinia)
authStore.initializeFromStorage()
userStore.initializeFromStorage()

if (authStore.accessToken) {
	fetchCurrentUserProfile().catch(() => {
		// 失败时继续使用缓存中的用户信息，不阻断应用启动
	})
}

app.mount('#app')
