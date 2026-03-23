import type { Editor } from '@tiptap/vue-3'

export interface KeyboardShortcut {
  key: string
  modifiers?: ('ctrl' | 'meta' | 'alt' | 'shift')[]
  action: (editor: Editor) => void
  description: string
  macDescription?: string
}

export function buildKeyboardShortcuts(editor: Editor | null): KeyboardShortcut[] {
  if (!editor) return []

  const isMac = typeof navigator !== 'undefined' && /Mac|iPod|iPhone|iPad/.test(navigator.platform)
  const mod = isMac ? 'meta' : 'ctrl'

  return [
    {
      key: 'b',
      modifiers: [mod],
      action: () => editor.chain().focus().toggleBold().run(),
      description: '加粗',
      macDescription: '⌘B',
    },
    {
      key: 'i',
      modifiers: [mod],
      action: () => editor.chain().focus().toggleItalic().run(),
      description: '斜体',
      macDescription: '⌘I',
    },
    {
      key: 'u',
      modifiers: [mod],
      action: () => editor.chain().focus().toggleUnderline().run(),
      description: '下划线',
      macDescription: '⌘U',
    },
    {
      key: 'x',
      modifiers: [mod, 'shift'],
      action: () => editor.chain().focus().toggleStrike().run(),
      description: '删除线',
      macDescription: '⌘⇧X',
    },
    {
      key: '0',
      modifiers: [mod, 'alt'],
      action: () => editor.chain().focus().setParagraph().run(),
      description: '正文',
      macDescription: '⌘⌥0',
    },
    {
      key: '1',
      modifiers: [mod, 'alt'],
      action: () => editor.chain().focus().toggleHeading({ level: 1 }).run(),
      description: '一级标题',
      macDescription: '⌘⌥1',
    },
    {
      key: '2',
      modifiers: [mod, 'alt'],
      action: () => editor.chain().focus().toggleHeading({ level: 2 }).run(),
      description: '二级标题',
      macDescription: '⌘⌥2',
    },
    {
      key: '3',
      modifiers: [mod, 'alt'],
      action: () => editor.chain().focus().toggleHeading({ level: 3 }).run(),
      description: '三级标题',
      macDescription: '⌘⌥3',
    },
    {
      key: '4',
      modifiers: [mod, 'alt'],
      action: () => editor.chain().focus().toggleHeading({ level: 4 }).run(),
      description: '四级标题',
      macDescription: '⌘⌥4',
    },
    {
      key: 'z',
      modifiers: [mod],
      action: () => editor.chain().focus().undo().run(),
      description: '撤销',
      macDescription: '⌘Z',
    },
    {
      key: 'z',
      modifiers: [mod, 'shift'],
      action: () => editor.chain().focus().redo().run(),
      description: '重做',
      macDescription: '⌘⇧Z',
    },
    {
      key: 'Enter',
      modifiers: [mod, 'shift'],
      action: () => editor.chain().focus().setHardBreak().run(),
      description: '强制换行',
      macDescription: '⌘⇧Enter',
    },
    {
      key: 'ArrowLeft',
      modifiers: [mod, 'alt'],
      action: () => editor.chain().focus().sinkListItem('listItem').run(),
      description: '减少缩进',
      macDescription: '⌘⌥←',
    },
    {
      key: 'ArrowRight',
      modifiers: [mod, 'alt'],
      action: () => editor.chain().focus().liftListItem('listItem').run(),
      description: '增加缩进',
      macDescription: '⌘⌥→',
    },
  ]
}

export function formatShortcut(shortcut: KeyboardShortcut): string {
  const isMac = typeof navigator !== 'undefined' && /Mac|iPod|iPhone|iPad/.test(navigator.platform)
  const parts: string[] = []

  const modifiers = shortcut.modifiers || []
  if (modifiers.includes('ctrl')) parts.push('Ctrl')
  if (modifiers.includes('meta') || (modifiers.includes('ctrl') && isMac)) {
    parts.push(isMac ? '⌘' : 'Ctrl')
  }
  if (modifiers.includes('alt')) parts.push(isMac ? '⌥' : 'Alt')
  if (modifiers.includes('shift')) parts.push(isMac ? '⇧' : 'Shift')

  let key = shortcut.key
  const specialKeys: Record<string, string> = {
    ArrowLeft: '←',
    ArrowRight: '→',
    ArrowUp: '↑',
    ArrowDown: '↓',
    Enter: 'Enter',
    Escape: 'Esc',
    Backspace: '⌫',
    Delete: '⌦',
  }

  if (specialKeys[key]) {
    key = specialKeys[key] as string
  } else if (key.length === 1) {
    key = key.toUpperCase()
  }

  parts.push(key)
  return parts.join(isMac ? '' : '+')
}

export function getShortcutDisplay(shortcut: KeyboardShortcut): string {
  const isMac = typeof navigator !== 'undefined' && /Mac|iPod|iPhone|iPad/.test(navigator.platform)
  if (isMac && shortcut.macDescription) {
    return `${shortcut.macDescription} ${shortcut.description}`
  }
  return `${formatShortcut(shortcut)} ${shortcut.description}`
}
