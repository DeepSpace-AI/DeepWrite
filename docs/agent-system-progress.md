# Agent 管理系统开发进度

## 目标

实现完整的 **Agent 管理系统**，包括：
1. 独立的 AI 对话窗口（类似 ChatGPT/Gemini，不强制绑定工作区）
2. 工作区数据动态引入（对话中选择/拖拽引用）
3. 全局对话历史
4. 完整的后端对话服务（SSE 流式响应）

---

## 已完成

### Phase 1: 核心 Chat 组件重构 ✅

**后端：** 无需改动

**前端：**
- `types/chat.ts` - 统一消息类型定义（MessagePart 体系）
- `composables/useChatStream.ts` - SSE 流式响应处理
- `stores/chat.ts` - Chat Store（消息状态管理）
- `components/chat/` - Chat 组件族
  - `ChatMessage.vue` - 消息组件（支持多 part）
  - `ChatMessageText.vue` - 文本消息（含流式光标）
  - `ChatMessageReasoning.vue` - 思考过程（可折叠）
  - `ChatMessageError.vue` - 错误消息
  - `ChatMessageToolCall.vue` - 工具调用
  - `ChatMessageList.vue` - 消息列表
  - `ChatInput.vue` - 输入组件

### Phase 2: 对话列表功能 ✅

**后端改动：**

| 文件 | 改动 |
|------|------|
| `models/agent/session.go` | 新增 `Pinned`, `Archived`, `GroupID`, `Tags` 字段 |
| `models/agent/session_utils.go` | 新增 `SessionListFilter`, `PinSession`, `UnpinSession`, `UnarchiveSession` |
| `handler/session_handler.go` | List 支持过滤参数，新增 Pin/Unpin/Unarchive 方法 |
| `routers/router.go` | 新增 `/sessions/:id/pin`, `/unpin`, `/unarchive` 路由 |

**前端改动：**

| 文件 | 说明 |
|------|------|
| `api/agent.ts` | Session 类型新增字段，新增 pin/unpin/unarchive API |
| `stores/session.ts` | 新建 SessionStore |
| `components/session/SessionList.vue` | 会话列表组件 |
| `components/session/SessionSearch.vue` | 会话搜索组件 |

**功能：**
- 按时间排序（置顶优先）
- 搜索对话
- 置顶/取消置顶
- 归档/恢复
- 删除对话

### Phase 3: Tool Calls + 消息编辑 ✅

**前端改动：**

| 文件 | 说明 |
|------|------|
| `components/chat/ChatMessageToolCall.vue` | 工具调用显示组件 |
| `components/chat/ChatMessage.vue` | 添加编辑/重新生成/删除按钮 |
| `components/chat/ChatMessageList.vue` | 支持 showActions 和事件传递 |
| `stores/chat.ts` | 新增 `deleteMessagesAfter`, `regenerateFromMessage` |

**功能：**
- Tool Calls 显示（工具名、参数、结果、状态）
- 消息编辑（用户消息编辑后重新发送）
- 重新生成（从 assistant 消息重新生成）
- 删除消息

### Phase 4: 对话分组 + 导出 + 分享 ✅

**后端改动：**

| 文件 | 说明 |
|------|------|
| `models/agent/session_group.go` | SessionGroup + SessionShare 模型 |
| `models/agent/session_group_utils.go` | 分组/分享 CRUD 方法 |
| `handler/session_group_handler.go` | 分组 API Handler |
| `handler/session_share_handler.go` | 分享 API Handler |
| `handler/session_export_handler.go` | 导出 API Handler |
| `routers/router.go` | 新增分组/分享/导出路由 |

**前端改动：**

| 文件 | 说明 |
|------|------|
| `api/agent.ts` | 新增 SessionGroup/SessionShare 类型和 API |
| `composables/useSessionActions.ts` | 导出/分享逻辑封装 |
| `components/session/SessionGroupList.vue` | 分组管理组件 |
| `components/session/SessionShareDialog.vue` | 分享对话框组件 |

**功能：**
- 创建/编辑/删除对话分组
- 导出对话为 Markdown / JSON
- 创建分享链接
- 复制分享链接
- 查看分享（公开访问）

---

## 待开发

### Phase 5: 多模态支持

- 图像消息发送和显示
- 语音消息发送和显示
- 文件上传和解析

---

## 关键文件

### 后端

```
services/gateway/
├── handler/
│   ├── chat_handler.go           # 对话执行接口
│   ├── session_handler.go         # Session CRUD + Pin/Archive
│   ├── session_group_handler.go   # Session Group CRUD
│   ├── session_share_handler.go   # Session Share CRUD
│   ├── session_export_handler.go  # 导出 Markdown/JSON
│   └── memory_handler.go          # Memory CRUD
├── models/agent/
│   ├── session.go                 # Session 模型
│   ├── session_group.go           # SessionGroup + SessionShare 模型
│   ├── session_utils.go           # Session 查询/更新方法
│   └── session_group_utils.go     # 分组/分享方法
└── routers/router.go              # 路由注册
```

### 前端

```
apps/web/src/
├── types/chat.ts                  # 消息类型定义
├── stores/
│   ├── chat.ts                    # Chat Store
│   └── session.ts                 # Session Store
├── composables/
│   ├── useChatStream.ts           # SSE 流处理
│   └── useSessionActions.ts       # 导出/分享逻辑
├── components/
│   ├── chat/                      # Chat 组件族
│   └── session/                   # Session 组件族
└── api/agent.ts                   # API 接口
```

---

## 数据结构

### 消息（ChatMessage）

```typescript
interface ChatMessage {
  id: string
  role: 'user' | 'assistant' | 'system'
  parts: MessagePart[]  // 支持多类型内容
  status: 'pending' | 'streaming' | 'done' | 'error'
  createdAt: Date
}

type MessagePart = TextPart | ReasoningPart | ToolCallPart | ImagePart | AudioPart | ErrorPart
```

### 会话（Session）

```typescript
interface Session {
  id: string
  agent_id: string
  workspace_id: string | null
  title: string
  pinned: boolean
  archived: boolean
  group_id: string | null
  tags: string[]
  last_message_at: string
}
```

### 会话分组（SessionGroup）

```typescript
interface SessionGroup {
  id: string
  user_id: string
  name: string
  description: string
  color: string      // 十六进制颜色
  icon: string       // 图标名称
  sort_order: number
}
```

### 会话分享（SessionShare）

```typescript
interface SessionShare {
  id: string
  session_id: string
  user_id: string
  share_token: string    // 32 字符随机 token
  title: string
  expires_at: string | null
  view_count: number
  allow_copy: boolean
  is_public: boolean
}
```

---

## API 路由

```
# 会话
GET    /api/v1/sessions
POST   /api/v1/sessions
GET    /api/v1/sessions/:id
PUT    /api/v1/sessions/:id
DELETE /api/v1/sessions/:id
PUT    /api/v1/sessions/:id/pin
PUT    /api/v1/sessions/:id/unpin
PUT    /api/v1/sessions/:id/archive
PUT    /api/v1/sessions/:id/unarchive
POST   /api/v1/sessions/:id/chat/stream

# 分组
GET    /api/v1/session-groups
POST   /api/v1/session-groups
PUT    /api/v1/session-groups/:id
DELETE /api/v1/session-groups/:id

# 分享
GET    /api/v1/session-shares
POST   /api/v1/session-shares
DELETE /api/v1/session-shares/:id
GET    /api/v1/shares/:token    # 公开访问

# 导出
GET    /api/v1/sessions/:id/export/markdown
GET    /api/v1/sessions/:id/export/json
```

---

## 注意事项

### 数据库迁移

Phase 2 和 Phase 4 新增了数据库字段和表，需要执行迁移：

```sql
-- Session 表新增字段
ALTER TABLE dw_sessions ADD COLUMN pinned BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE dw_sessions ADD COLUMN archived BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE dw_sessions ADD COLUMN group_id UUID;
ALTER TABLE dw_sessions ADD COLUMN tags JSONB NOT NULL DEFAULT '[]'::jsonb;

-- 创建索引
CREATE INDEX idx_sessions_pinned ON dw_sessions(pinned);
CREATE INDEX idx_sessions_archived ON dw_sessions(archived);
CREATE INDEX idx_sessions_group_id ON dw_sessions(group_id);

-- SessionGroup 表
CREATE TABLE dw_session_groups (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  color VARCHAR(20) DEFAULT '#6366f1',
  icon VARCHAR(50),
  sort_order INT DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_session_groups_user_id ON dw_session_groups(user_id);

-- SessionShare 表
CREATE TABLE dw_session_shares (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id UUID NOT NULL,
  user_id UUID NOT NULL,
  share_token VARCHAR(32) UNIQUE NOT NULL,
  title VARCHAR(255),
  expires_at TIMESTAMP,
  view_count INT DEFAULT 0,
  allow_copy BOOLEAN DEFAULT TRUE,
  is_public BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_session_shares_session_id ON dw_session_shares(session_id);
CREATE INDEX idx_session_shares_user_id ON dw_session_shares(user_id);
CREATE UNIQUE INDEX idx_session_shares_token ON dw_session_shares(share_token);
```