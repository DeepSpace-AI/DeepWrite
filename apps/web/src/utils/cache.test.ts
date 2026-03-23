import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { cacheManager, CACHE_TTL } from '@/utils/cache'

const TEST_KEY = 'test-cache-key'
const TEST_VALUE = { name: 'test', value: 123 }

describe('CacheManager', () => {
  beforeEach(() => {
    cacheManager.clear()
  })

  afterEach(() => {
    cacheManager.clear()
  })

  describe('set and get', () => {
    it('should store and retrieve a string value', () => {
      cacheManager.set('string-key', 'hello')
      expect(cacheManager.get<string>('string-key')).toBe('hello')
    })

    it('should store and retrieve an object value', () => {
      cacheManager.set('object-key', TEST_VALUE)
      expect(cacheManager.get<typeof TEST_VALUE>('object-key')).toEqual(TEST_VALUE)
    })

    it('should store and retrieve array values', () => {
      const arr = [1, 2, 3]
      cacheManager.set('array-key', arr)
      expect(cacheManager.get<number[]>('array-key')).toEqual(arr)
    })

    it('should return null for non-existent key', () => {
      expect(cacheManager.get('non-existent')).toBeNull()
    })
  })

  describe('expiration', () => {
    it('should expire data after TTL', async () => {
      cacheManager.set(TEST_KEY, TEST_VALUE, 100)
      expect(cacheManager.get<typeof TEST_VALUE>(TEST_KEY)).toEqual(TEST_VALUE)

      await new Promise((resolve) => setTimeout(resolve, 150))
      expect(cacheManager.get<typeof TEST_VALUE>(TEST_KEY)).toBeNull()
    })

    it('should not expire if no TTL provided', () => {
      cacheManager.set(TEST_KEY, TEST_VALUE)
      expect(cacheManager.get<typeof TEST_VALUE>(TEST_KEY)).toEqual(TEST_VALUE)
    })
  })

  describe('remove', () => {
    it('should remove single key', () => {
      cacheManager.set(TEST_KEY, TEST_VALUE)
      cacheManager.remove(TEST_KEY)
      expect(cacheManager.get(TEST_KEY)).toBeNull()
    })

    it('should remove multiple keys', () => {
      cacheManager.set('key1', 'value1')
      cacheManager.set('key2', 'value2')
      cacheManager.removeMultiple(['key1', 'key2'])
      expect(cacheManager.get('key1')).toBeNull()
      expect(cacheManager.get('key2')).toBeNull()
    })
  })

  describe('clear', () => {
    it('should clear all cached data', () => {
      cacheManager.set('key1', 'value1')
      cacheManager.set('key2', 'value2')
      cacheManager.clear()
      expect(cacheManager.get('key1')).toBeNull()
      expect(cacheManager.get('key2')).toBeNull()
    })
  })

  describe('CACHE_TTL constants', () => {
    it('should have correct FIVE_MINUTES value', () => {
      expect(CACHE_TTL.FIVE_MINUTES).toBe(5 * 60 * 1000)
    })

    it('should have correct ONE_HOUR value', () => {
      expect(CACHE_TTL.ONE_HOUR).toBe(60 * 60 * 1000)
    })

    it('should have correct ONE_DAY value', () => {
      expect(CACHE_TTL.ONE_DAY).toBe(24 * 60 * 60 * 1000)
    })

    it('should have all expected TTL constants', () => {
      expect(CACHE_TTL.FIVE_MINUTES).toBeDefined()
      expect(CACHE_TTL.FIFTEEN_MINUTES).toBeDefined()
      expect(CACHE_TTL.ONE_HOUR).toBeDefined()
      expect(CACHE_TTL.SIX_HOURS).toBeDefined()
      expect(CACHE_TTL.ONE_DAY).toBeDefined()
      expect(CACHE_TTL.SEVEN_DAYS).toBeDefined()
    })
  })
})
