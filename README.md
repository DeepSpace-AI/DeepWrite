<div align="center">
  <img src="docs/logo.png" alt="DeepWrite Logo" width="220" />

# DeepWrite

沉浸式 AI 协作科研工作平台。

[![Node](https://img.shields.io/badge/Node.js-20+-green.svg)](https://nodejs.org)
[![Go](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://go.dev)
[![Python](https://img.shields.io/badge/Python-3.13+-yellow.svg)](https://python.org)
[![Vue](https://img.shields.io/badge/Vue-3.5-42b883.svg)](https://vuejs.org)

## 项目简介

DeepWrite 是一个 Monorepo 项目，目标是提供从内容创作到协作编辑的完整工作流：

- **产品定位**：沉浸式 AI 协作科研工作平台
- **官网地址**：https://deepwrite.work
- **核心能力**：文档编辑与版本管理、工作区协作、Yjs 实时协作、AI 能力集成

## 项目完成度

| 模块 | 完成度 | 状态 | 说明 |
|------|--------|------|------|
| `apps/web` | **85%** | ✅ 主力开发 | 完整业务前端，含协作编辑 |
| `apps/admin` | **15%** | 🔨 框架阶段 | 基础脚手架 |
| `services/gateway` | **80%** | ✅ 主力开发 | API + 协作 WebSocket |
| `services/ai` | **50%** | 🔨 开发中 | 模型网关框架 |
| `services/worker` | **20%** | 🔨 初始化 | 入口已就绪 |

### 功能完成清单

#### apps/web (85%)
- ✅ 用户认证 (JWT + 自动刷新)
- ✅ 工作区管理 (CRUD + 成员 + 邀请)
- ✅ 富文本编辑器 DEditor Phase A (34工具 + Slash菜单)
- ✅ 实时协作 (Yjs + WebSocket)
- ✅ 版本管理 (快照 + 历史)
- ✅ 国际化 (中英文)
- 🔧 DEditor Phase B/C (快捷键/导入导出)

#### services/gateway (80%)
- ✅ 认证/用户/工作区/文档 API
- ✅ 协作 WebSocket (Token + Room + 广播)
- ✅ 文件上传下载 (S3/本地)
- ✅ 邮件发送 (SMTP)
- ✅ 缓存 (Redis)
- 🔧 AI Provider 配置管理

#### services/ai (50%)
- ✅ 统一模型网关 (FastAPI)
- ✅ 多 Provider 支持
- ✅ 流式响应 (SSE)
- 🔧 模型配置持久化

#### services/worker (20%)
- ✅ Celery 任务框架
- ✅ API + Worker 启动入口
- 🔧 任务链路 (文档处理/AI/通知)

## 技术栈

| 层级 | 技术 | 版本 |
|------|------|------|
| **前端框架** | Vue 3 + TypeScript | ^3.5.28 / ~5.9.3 |
| **构建工具** | Vite | ^7.3.1 |
| **状态管理** | Pinia | ^3.0.4 |
| **编辑器** | TipTap + Yjs | ^3.20.1 / ^13.6.29 |
| **UI 框架** | TailwindCSS + DaisyUI | ^4.2.1 / ^5.5.19 |
| **后端服务** | Go + Gin + GORM | - |
| **异步任务** | Python + Celery + Redis | - |
| **数据库** | PostgreSQL 14+ / Redis 6+ | - |

## 仓库结构

```text
.
├─apps/
│  ├─web/       # 主站前端 (Vue 3 + TipTap + Yjs)     [85%]
│  └─admin/     # 管理端前端 (开发中)                   [15%]
├─services/
│  ├─gateway/   # Go 网关服务 (API + 协作 WebSocket)   [80%]
│  ├─ai/        # AI 模型网关 (FastAPI)               [50%]
│  └─worker/    # Python Worker (Celery)              [20%]
├─docs/         # 项目文档
└─docs/         # 配置文件与环境变量示例
```

## 环境要求

| 依赖 | 版本 | 说明 |
|------|------|------|
| Node.js | ^20.19.0 或 >=22.12.0 | 前端构建 |
| pnpm | 10.x | 包管理 |
| Go | 1.25+ | 网关服务 |
| Python | 3.13+ | AI/Worker 服务 |
| PostgreSQL | 14+ | 主数据库 |
| Redis | 6+ | 缓存/消息队列 |

## 快速开始

### 1. 安装依赖

```bash
pnpm install
```

### 2. 配置前端环境变量

```bash
# Windows
Copy-Item apps/web/.env.example apps/web/.env.local

# macOS / Linux
cp apps/web/.env.example apps/web/.env.local
```

必需配置：

```env
VITE_GATEWAY_BASE_URL=http://localhost:8080
VITE_CACHE_SECRET=change-this-to-a-long-random-string
```

### 3. 配置网关服务

1. 复制 `services/gateway/config_example.yaml` 为 `config.development.yaml`
2. 填写数据库、Redis、对象存储、JWT、邮件等配置
3. 确保 PostgreSQL 与 Redis 已就绪

### 4. 启动开发环境

```bash
# 启动全部模块 (需配置所有服务)
pnpm dev

# 仅启动前端
pnpm dev:web

# 仅启动网关 (Go)
pnpm dev:gateway

# 启动 AI 服务
pnpm dev:ai

# 启动 Worker
pnpm dev:worker
```

## 常用命令

```bash
pnpm dev            # Turbo 并行开发
pnpm build          # 构建所有项目
pnpm lint           # Lint 检查
pnpm format         # 代码格式化
pnpm gen:openapi    # 生成 Swagger 文档 (gateway)
```

## 服务端口

| 服务 | 地址 | 说明 |
|------|------|------|
| Web 前端 | http://localhost:5173 | 主站 |
| Gateway API | http://localhost:8080 | REST API |
| Gateway Swagger | http://localhost:8080/swagger/index.html | API 文档 |
| AI 服务 | http://localhost:8000 | AI 模型网关 |
| Worker API | http://localhost:8001 | 任务队列 API |

## 子项目说明

### apps/web - 主站前端

**核心功能**：
- 用户认证 (登录/注册/JWT 会话)
- 工作区与文件管理 (CRUD/权限/邀请)
- 文档编辑与版本能力 (DEditor)
- Yjs 实时协作编辑 (WebSocket)

**关键模块**：
- `src/collab/` - Yjs WebSocket Provider
- `src/components/editor/` - DEditor 富文本编辑器
- `src/stores/` - Pinia 状态管理
- `src/api/` - Axios HTTP 封装

### apps/admin - 管理端前端

**当前状态**：基础框架，仅包含登录/布局

**规划功能**：用户管理、工作区管理、审计日志、系统配置

### services/gateway - Go 网关服务

**核心能力**：
- RESTful API (`/api/v1/*`)
- 实时协作 WebSocket (`/api/v1/documents/:id/collab/ws`)
- JWT 鉴权中间件
- 文件存储 (S3/本地)
- 邮件发送 (SMTP)
- Redis 缓存

**关键模块**：
- `handler/` - 业务处理器 (auth/workspace/document/collab)
- `models/` - 数据模型 (GORM)
- `pkg/collab/` - Yjs 协议解析与广播

### services/ai - AI 模型网关

**核心能力**：
- 统一 Provider 接口 (OpenAI/Claude/自定义)
- 流式响应 (SSE)
- 模型配置管理

**API 端点**：
```
/internal/v1/chat/completions
/internal/v1/embeddings
/internal/v1/rerank
/internal/v1/audio/speech
/internal/v1/audio/transcriptions
```

### services/worker - Python Worker

**当前状态**：Celery 框架就绪，任务待实现

**规划任务**：
- 文档解析与导入导出
- AI 批量推理
- 异步通知推送
- 定时任务

详细方案见 `docs/worker-development-plan.md`

## 开发建议

### 代码规范

- 新增接口后执行 `pnpm gen:openapi` 同步 Swagger
- Go 代码遵循标准项目布局
- 前端使用 TypeScript 严格模式

### 测试覆盖

- 前端：`apps/web` 补充 Vitest 单元测试
- 后端：`services/gateway` 补充 Go 测试
- E2E：Playwright 端到端测试

### CI/CD

- 类型检查 + Lint + 构建
- 自动化测试
- 镜像构建 (可选)

## 开发计划

详见 [DEVELOPMENT.md](./DEVELOPMENT.md)，包含：
- 短期里程碑 (1-4 周)
- 中期目标 (1-2 月)
- 长期规划 (2-3 月)

## 贡献

欢迎通过 Issue / PR 参与改进。

建议流程：

1. Fork 并创建功能分支
2. 提交最小可验证改动
3. 提交 PR 并补充必要说明

## License

MIT License
