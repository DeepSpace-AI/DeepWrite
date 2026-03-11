import {createI18n} from "vue-i18n";


import zhCN from './zh'
import en from './en'
import { cacheManager, CACHE_TTL } from '../src/utils/cache'
import type { User } from '../src/stores/user'

export type MessageSchema = typeof zhCN
export type SupportedLocale = 'zh-CN' | 'en'

const CACHE_KEY_LOCALE = 'deepwrite_locale'
const CACHE_KEY_USER = 'deepwrite_user'
const LOCALE_TTL = CACHE_TTL.ONE_DAY

function isSupportedLocale(value: unknown): value is SupportedLocale {
	return value === 'zh-CN' || value === 'en'
}

export function getCachedLocale(): SupportedLocale | null {
	const rawLocale = cacheManager.get<string>(CACHE_KEY_LOCALE)
	if (isSupportedLocale(rawLocale)) return rawLocale

	const cachedUser = cacheManager.get<User>(CACHE_KEY_USER)
	const userLocale = cachedUser?.language
	return isSupportedLocale(userLocale) ? userLocale : null
}

export function setAppLocale(locale: string): SupportedLocale {
	const nextLocale: SupportedLocale = isSupportedLocale(locale) ? locale : 'zh-CN'
	i18n.global.locale.value = nextLocale
	cacheManager.set(CACHE_KEY_LOCALE, nextLocale, LOCALE_TTL)

	return nextLocale
}

export const i18n = createI18n({
	legacy: false,
	locale: getCachedLocale() ?? 'zh-CN',
	fallbackLocale: 'en',
	messages: {
		'zh-CN': zhCN,
		en,
	},
})


