import CryptoJS from 'crypto-js'

interface CacheEntry<T> {
  value: T
  expiresAt?: number
}

export class CacheManager {
  private isDev = import.meta.env.DEV
  private encryptionEnabled = true
  private versionPrefix = 'enc:v2:'

  private getSecretKey(): string {
    const envSecret = (import.meta.env.VITE_CACHE_SECRET as string | undefined)?.trim()
    const host = typeof window !== 'undefined' ? window.location.host : 'unknown-host'
    const ua = typeof navigator !== 'undefined' ? navigator.userAgent : 'unknown-ua'
    return `${envSecret || 'deepwrite-cache-default-secret'}::${host}::${ua.slice(0, 80)}`
  }

  set<T>(key: string, value: T, ttl?: number): void {
    const entry: CacheEntry<T> = {
      value,
      expiresAt: ttl ? Date.now() + ttl : undefined,
    }

    const data = JSON.stringify(entry)
    const payload = this.encryptionEnabled ? this.versionPrefix + this.encrypt(data) : data
    localStorage.setItem(key, payload)
  }

  get<T>(key: string): T | null {
    const raw = localStorage.getItem(key)
    if (!raw) return null

    try {
      const data = this.decodePayload(raw)
      const entry = JSON.parse(data) as CacheEntry<T>

      if (entry.expiresAt && entry.expiresAt < Date.now()) {
        this.remove(key)
        return null
      }

      return entry.value
    } catch {
      this.remove(key)
      return null
    }
  }

  remove(key: string): void {
    localStorage.removeItem(key)
  }

  removeMultiple(keys: string[]): void {
    keys.forEach(key => localStorage.removeItem(key))
  }

  clear(): void {
    localStorage.clear()
  }

  private encrypt(data: string): string {
    return CryptoJS.AES.encrypt(data, this.getSecretKey()).toString()
  }

  private decrypt(encrypted: string): string {
    const bytes = CryptoJS.AES.decrypt(encrypted, this.getSecretKey())
    const plainText = bytes.toString(CryptoJS.enc.Utf8)
    if (!plainText) {
      throw new Error('decrypt failed')
    }
    return plainText
  }

  private decodePayload(raw: string): string {
    if (raw.startsWith(this.versionPrefix)) {
      return this.decrypt(raw.slice(this.versionPrefix.length))
    }

    if (this.isDev) {
      try {
        return atob(raw)
      } catch {
        return raw
      }
    }

    return raw
  }
}

export const cacheManager = new CacheManager()

export const CACHE_TTL = {
  FIVE_MINUTES: 5 * 60 * 1000,
  FIFTEEN_MINUTES: 15 * 60 * 1000,
  ONE_HOUR: 60 * 60 * 1000,
  SIX_HOURS: 6 * 60 * 60 * 1000,
  ONE_DAY: 24 * 60 * 60 * 1000,
  SEVEN_DAYS: 7 * 24 * 60 * 60 * 1000,
}
