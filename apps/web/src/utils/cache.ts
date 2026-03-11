/**
 * Cache Manager - 支持过期时间和开发环境加密
 */

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
    // 前端无法实现“绝对安全”，这里用环境变量 + 设备信息做“部分安全”增强
    const envSecret = (import.meta.env.VITE_CACHE_SECRET as string | undefined)?.trim()
    const host = typeof window !== 'undefined' ? window.location.host : 'unknown-host'
    const ua = typeof navigator !== 'undefined' ? navigator.userAgent : 'unknown-ua'
    return `${envSecret || 'deepwrite-cache-default-secret'}::${host}::${ua.slice(0, 80)}`
  }

  /**
   * 设置缓存值
   * @param key - 缓存键
   * @param value - 缓存值
   * @param ttl - 过期时间（毫秒），如果为 undefined 则永不过期
   */
  set<T>(key: string, value: T, ttl?: number): void {
    const entry: CacheEntry<T> = {
      value,
      expiresAt: ttl ? Date.now() + ttl : undefined,
    }

    const data = JSON.stringify(entry)
    const payload = this.encryptionEnabled ? this.versionPrefix + this.encrypt(data) : data
    localStorage.setItem(key, payload)
  }

  /**
   * 获取缓存值（自动检查过期）
   * @param key - 缓存键
   * @returns 缓存值或 null
   */
  get<T>(key: string): T | null {
    const raw = localStorage.getItem(key)
    if (!raw) return null

    try {
      const data = this.decodePayload(raw)
      const entry = JSON.parse(data) as CacheEntry<T>

      // 检查是否过期
      if (entry.expiresAt && entry.expiresAt < Date.now()) {
        this.remove(key)
        return null
      }

      return entry.value
    } catch {
      // 解析失败或解密失败，移除该项
      this.remove(key)
      return null
    }
  }

  /**
   * 移除单个缓存项
   * @param key - 缓存键
   */
  remove(key: string): void {
    localStorage.removeItem(key)
  }

  /**
   * 移除多个缓存项
   * @param keys - 缓存键数组
   */
  removeMultiple(keys: string[]): void {
    keys.forEach(key => localStorage.removeItem(key))
  }

  /**
   * 清空所有缓存
   */
  clear(): void {
    localStorage.clear()
  }

  /**
   * AES 加密
   * @param data - 原始字符串数据
   * @returns 加密后的字符串
   */
  private encrypt(data: string): string {
    return CryptoJS.AES.encrypt(data, this.getSecretKey()).toString()
  }

  /**
   * AES 解密
   * @param encrypted - 加密的字符串
   * @returns 原始字符串数据
   */
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

    // 兼容旧版本 Base64（仅开发环境曾启用）
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

/**
 * 默认缓存管理器实例
 */
export const cacheManager = new CacheManager()

/**
 * 常用的 TTL 常量（毫秒）
 */
export const CACHE_TTL = {
  /** 5 分钟 */
  FIVE_MINUTES: 5 * 60 * 1000,
  /** 15 分钟 */
  FIFTEEN_MINUTES: 15 * 60 * 1000,
  /** 1 小时 */
  ONE_HOUR: 60 * 60 * 1000,
  /** 6 小时 */
  SIX_HOURS: 6 * 60 * 60 * 1000,
  /** 24 小时 */
  ONE_DAY: 24 * 60 * 60 * 1000,
  /** 7 天 */
  SEVEN_DAYS: 7 * 24 * 60 * 60 * 1000,
}
