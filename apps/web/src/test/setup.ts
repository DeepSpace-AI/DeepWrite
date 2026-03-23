import { afterEach, vi } from 'vitest'
import '@vue/test-utils'

const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => { store[key] = value },
    removeItem: (key: string) => { delete store[key] },
    clear: () => { store = {} },
    get length() { return Object.keys(store).length },
    key: (i: number) => Object.keys(store)[i] || null,
  }
})()

Object.defineProperty(global, 'localStorage', { value: localStorageMock })

Object.defineProperty(global, 'navigator', {
  value: { userAgent: 'test-agent', platform: 'test-platform' },
  configurable: true,
})

Object.defineProperty(global, 'window', {
  value: { location: { host: 'localhost' } },
  configurable: true,
})

global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}))

afterEach(() => {
  localStorageMock.clear()
})

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
    locale: { value: 'en' },
  }),
}))
