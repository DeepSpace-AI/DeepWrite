import type { Editor } from '@tiptap/vue-3'

export function exportToHtml(editor: Editor | null): string {
  if (!editor) return ''

  const html = editor.getHTML()
  return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Document</title>
  <style>
    body { max-width: 800px; margin: 40px auto; padding: 0 20px; font-family: system-ui, sans-serif; }
    h1, h2, h3 { margin-top: 1.5em; }
    p { line-height: 1.7; }
    code { background: #f4f4f4; padding: 0.2em 0.4em; border-radius: 3px; }
    pre { background: #f4f4f4; padding: 1em; border-radius: 5px; overflow-x: auto; }
    blockquote { border-left: 4px solid #ddd; margin-left: 0; padding-left: 1em; color: #666; }
  </style>
</head>
<body>
${html}
</body>
</html>`
}

export function exportToMarkdown(editor: Editor | null): string {
  if (!editor) return ''

  let markdown = ''
  const doc = editor.state.doc

  doc.descendants((node, _pos) => {
    switch (node.type.name) {
      case 'heading': {
        const level = node.attrs.level
        const text = node.textContent
        markdown += `${'#'.repeat(level)} ${text}\n\n`
        break
      }
      case 'paragraph': {
        const text = node.textContent
        if (text) markdown += `${text}\n\n`
        break
      }
      case 'bulletList': {
        node.forEach((item) => {
          markdown += `- ${item.textContent}\n`
        })
        markdown += '\n'
        break
      }
      case 'orderedList': {
        let idx = 1
        node.forEach((item) => {
          markdown += `${idx}. ${item.textContent}\n`
          idx++
        })
        markdown += '\n'
        break
      }
      case 'taskList': {
        node.forEach((item) => {
          const checked = item.attrs.checked ? '[x]' : '[ ]'
          markdown += `- ${checked} ${item.textContent}\n`
        })
        markdown += '\n'
        break
      }
      case 'codeBlock': {
        const lang = node.attrs.language || ''
        markdown += `\`\`\`${lang}\n${node.textContent}\n\`\`\`\n\n`
        break
      }
      case 'blockquote': {
        const text = node.textContent
        markdown += `> ${text}\n\n`
        break
      }
      case 'horizontalRule': {
        markdown += '---\n\n'
        break
      }
      default:
        break
    }
    return true
  })

  return markdown.trim()
}

export function copyToClipboard(text: string): Promise<void> {
  if (navigator.clipboard && navigator.clipboard.writeText) {
    return navigator.clipboard.writeText(text)
  }

  return new Promise((resolve, reject) => {
    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    try {
      document.execCommand('copy')
      resolve()
    } catch (err) {
      reject(err)
    } finally {
      document.body.removeChild(textarea)
    }
  })
}
