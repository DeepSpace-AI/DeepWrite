import { createApp } from 'vue'
import { watch } from 'vue'
import { createPinia } from 'pinia'
import './style.css'

import App from './App.vue'
import router from './router'
import { fetchCurrentUserProfile } from './api/auth'
import { useAuthStore } from './stores/auth'
import { useUserStore } from './stores/user'
import { i18n, getCachedLocale, setAppLocale } from '../locales/i18n'

const app = createApp(App)

const pinia = createPinia()
app.use(pinia)
app.use(router)
app.use(i18n)

// 应用挂载前，从存储恢复用户和令牌信息
const authStore = useAuthStore(pinia)
const userStore = useUserStore(pinia)
authStore.initializeFromStorage()
userStore.initializeFromStorage()

// 根据用户缓存的语言偏好设置初始 locale
const savedLang = userStore.user?.language
if (savedLang === 'en' || savedLang === 'zh-CN') {
	setAppLocale(savedLang)
} else {
	const cachedLocale = getCachedLocale()
	if (cachedLocale) {
		setAppLocale(cachedLocale)
	}
}

// 用户信息变化时，同步语言到 i18n 和本地缓存
watch(
	() => userStore.user?.language,
	(nextLang) => {
		if (nextLang === 'en' || nextLang === 'zh-CN') {
			setAppLocale(nextLang)
		}
	},
	{ immediate: true },
)

if (authStore.accessToken) {
	fetchCurrentUserProfile().catch(() => {
		// 失败时继续使用缓存中的用户信息，不阻断应用启动
	})
}

app.mount('#app')
