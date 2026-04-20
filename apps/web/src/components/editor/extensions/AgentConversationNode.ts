import { Node, mergeAttributes } from '@tiptap/core'
import { VueNodeViewRenderer } from '@tiptap/vue-3'
import AgentConversationNodeView from './AgentConversationNodeView.vue'

export interface AgentConversationNodeOptions {
  HTMLAttributes: Record<string, string>
}

declare module '@tiptap/core' {
  interface Commands<ReturnType> {
    agentConversation: {
      insertAgentConversation: (attrs?: Partial<AgentConversationAttrs>) => ReturnType
      updateAgentConversation: (attrs: Partial<AgentConversationAttrs>) => ReturnType
    }
  }
}

export interface AgentConversationAttrs {
  agentId: string | null
  workspaceId: string
  sessionId: string | null
  collapsed: boolean
}

export const AgentConversationNode = Node.extend<AgentConversationNodeOptions>({
  name: 'agentConversation',

  group: 'block',

  atom: true,

  draggable: true,

  isolating: true,

  selectable: true,

  addOptions() {
    return {
      HTMLAttributes: {},
    }
  },

  addAttributes() {
    return {
      agentId: {
        default: null,
        parseHTML: (el) => el.getAttribute('data-agent-id'),
        renderHTML: (attrs) => {
          if (!attrs.agentId) return {}
          return { 'data-agent-id': attrs.agentId }
        },
      },
      workspaceId: {
        default: '',
        parseHTML: (el) => el.getAttribute('data-workspace-id'),
        renderHTML: (attrs) => {
          if (!attrs.workspaceId) return {}
          return { 'data-workspace-id': attrs.workspaceId }
        },
      },
      sessionId: {
        default: null,
        parseHTML: (el) => el.getAttribute('data-session-id'),
        renderHTML: (attrs) => {
          if (!attrs.sessionId) return {}
          return { 'data-session-id': attrs.sessionId }
        },
      },
      collapsed: {
        default: false,
        parseHTML: (el) => el.getAttribute('data-collapsed') === 'true',
        renderHTML: (attrs) => ({
          'data-collapsed': attrs.collapsed ? 'true' : 'false',
        }),
      },
    }
  },

  parseHTML() {
    return [
      {
        tag: 'div[data-type="agent-conversation"]',
      },
    ]
  },

  renderHTML({ HTMLAttributes }) {
    return [
      'div',
      mergeAttributes(this.options.HTMLAttributes, HTMLAttributes, {
        'data-type': 'agent-conversation',
      }),
    ]
  },

  addNodeView() {
    return VueNodeViewRenderer(AgentConversationNodeView)
  },

  addCommands() {
    return {
      insertAgentConversation:
        (attrs) =>
        ({ commands }) =>
          commands.insertContent({
            type: this.name,
            attrs: {
              agentId: attrs?.agentId ?? null,
              workspaceId: attrs?.workspaceId ?? '',
              sessionId: attrs?.sessionId ?? null,
              collapsed: attrs?.collapsed ?? false,
            },
          }),
      updateAgentConversation:
        (attrs) =>
        ({ commands }) =>
          commands.updateAttributes(this.name, attrs),
    }
  },
})

export default AgentConversationNode