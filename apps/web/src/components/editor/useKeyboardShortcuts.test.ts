import { describe, it, expect } from 'vitest'
import { formatShortcut, getShortcutDisplay } from '@/components/editor/useKeyboardShortcuts'
import type { KeyboardShortcut } from '@/components/editor/useKeyboardShortcuts'

describe('useKeyboardShortcuts', () => {
  describe('formatShortcut', () => {
    it('should format simple key shortcut', () => {
      const shortcut: KeyboardShortcut = {
        key: 'b',
        action: vi.fn(),
        description: '加粗',
      }
      const result = formatShortcut(shortcut)
      expect(result).toContain('B')
    })

    it('should format shortcut with modifiers', () => {
      const shortcut: KeyboardShortcut = {
        key: 'b',
        modifiers: ['ctrl'],
        action: vi.fn(),
        description: '加粗',
      }
      const result = formatShortcut(shortcut)
      expect(result).toContain('B')
    })

    it('should format special keys', () => {
      const shortcut: KeyboardShortcut = {
        key: 'Enter',
        action: vi.fn(),
        description: '确认',
      }
      const result = formatShortcut(shortcut)
      expect(result).toContain('Enter')
    })

    it('should format arrow keys', () => {
      const shortcut: KeyboardShortcut = {
        key: 'ArrowLeft',
        modifiers: ['ctrl', 'alt'],
        action: vi.fn(),
        description: '减少缩进',
      }
      const result = formatShortcut(shortcut)
      expect(result).toContain('←')
    })
  })

  describe('getShortcutDisplay', () => {
    it('should include description', () => {
      const shortcut: KeyboardShortcut = {
        key: 'b',
        action: vi.fn(),
        description: '加粗',
      }
      const result = getShortcutDisplay(shortcut)
      expect(result).toContain('加粗')
    })

    it('should use mac description on Mac', () => {
      const shortcut: KeyboardShortcut = {
        key: 'b',
        modifiers: ['meta'],
        action: vi.fn(),
        description: '加粗',
        macDescription: '⌘B',
      }
      Object.defineProperty(navigator, 'platform', { value: 'Mac', configurable: true })
      const result = getShortcutDisplay(shortcut)
      expect(result).toContain('⌘B')
    })
  })
})
