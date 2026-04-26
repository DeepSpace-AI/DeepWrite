# DeepWrite - AI辅助科研工作平台设计文档

**版本**：1.0  
**日期**：2026-04-25  
**状态**：待审核

---

## 目录

1. [项目概述](#项目概述)
2. [系统架构](#系统架构)
3. [功能模块](#功能模块)
4. [数据库设计](#数据库设计)
5. [API设计](#api设计)
6. [用户流程](#用户流程)
7. [界面设计](#界面设计)
8. [部署架构](#部署架构)
9. [安全设计](#安全设计)
10. [测试策略](#测试策略)
11. [监控与运维](#监控与运维)
12. [开发计划](#开发计划)
13. [风险评估](#风险评估)
14. [成本估算](#成本估算)
15. [商业模式](#商业模式)
16. [法律合规](#法律合规)

---

## 项目概述

### 项目背景

科研工作者在论文写作过程中需要使用多个工具（文献管理、写作、代码、图像等），流程分散，效率低下。现有的AI工具虽然强大，但存在幻觉问题，在学术场景中不够可靠。

### 项目目标

构建一个AI辅助科研工作平台，全流程深度参与科研论文写作的全过程，同时确保学术诚信和内容准确性。

### 目标用户

- 医学研究人员
- 计算机科学研究人员
- 机械工程研究人员
- 人文社会科学研究人员
- 其他学术领域研究者

### 核心特性

1. **全流程覆盖**：选题、文献、写作、代码、图像、投递
2. **AI深度集成**：智能辅助，但强调零幻觉、强上下文
3. **通用型设计**：适配不同学科领域
4. **SaaS商业模式**：Freemium模式，支持商业化

---

## 系统架构

### 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        客户端层                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                      Web App                             │  │
│  │                  (Next.js 16+)                           │  │
│  │  • App Router  • React Server Components  • TailwindCSS  │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      API Gateway                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  • 认证授权  • 限流  • 路由  • 日志  • 监控              │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│  用户服务    │      │  项目服务    │      │  AI服务      │
│ • 注册登录   │      │ • 项目管理   │      │ • 论文润色   │
│ • 用户资料   │      │ • 团队协作   │      │ • 文献分析   │
│ • 权限管理   │      │ • 版本控制   │      │ • 代码生成   │
│ • 订阅计费   │      │ • 文件存储   │      │ • 图像生成   │
└──────────────┘      └──────────────┘      └──────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      数据层                                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │PostgreSQL│  │  Redis   │  │MongoDB   │  │ S3/Minio │       │
│  │   18+    │  │  缓存    │  │ 文档数据 │  │ 文件存储 │       │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘       │
└─────────────────────────────────────────────────────────────────┘
```

### 技术栈

**前端**：
- Next.js 16+（App Router）
- React 19+ + TypeScript
- TailwindCSS 4.x + Shadcn/ui
- TipTap/ProseMirror（富文本编辑器）
- Monaco Editor（代码编辑器）

**后端**：
- API Gateway: Kong/Nginx
- **Go 服务**（高性能、高并发场景）：
  - 用户服务: Go 1.22+ + Gin/Echo
  - 项目服务: Go 1.22+ + Gin/Echo
  - 写作服务: Go 1.22+ + Gin/Echo
  - 图像服务: Go 1.22+ + Gin/Echo
  - 存储服务: Go 1.22+ + Gin/Echo
- **Python 服务**（AI/ML场景）：
  - 文献服务: Python 3.12+ + FastAPI
  - 代码服务: Python 3.12+ + FastAPI + Docker
  - AI服务: Python 3.12+ + FastAPI

**数据存储**：
- PostgreSQL 18+: 用户数据、项目元数据、订阅信息
- Redis 8.x: 缓存、会话、实时协作
- MongoDB 8.x: 文档内容、版本历史
- S3/Minio: 文件存储（论文、图像、代码）

**基础设施**：
- Docker + Docker Compose（开发）
- Kubernetes 1.30+（生产）
- GitHub Actions（CI/CD）
- Prometheus + Grafana（监控）

### 核心服务划分

| 服务名称 | 职责 | 技术栈 |
|---------|------|--------|
| **用户服务** | 用户认证、授权、订阅计费 | Go 1.22+ + PostgreSQL 18+ |
| **项目服务** | 科研项目管理、团队协作 | Go 1.22+ + PostgreSQL 18+ |
| **文献服务** | 文献检索、分析、管理、引用 | Python 3.12+ + MongoDB 8.x |
| **写作服务** | 论文编辑、润色、格式管理 | Go 1.22+ + PostgreSQL 18+ |
| **代码服务** | 代码编辑、运行、结果存储 | Python 3.12+ + Docker |
| **图像服务** | 科研图像绘制、模板管理 | Go 1.22+ + Canvas |
| **AI服务** | 统一AI能力接口 | Python 3.12+ + FastAPI |
| **存储服务** | 文件上传、下载、版本管理 | Go 1.22+ + S3 |

---

## 功能模块

### 1. 用户与权限系统

**核心功能**：
- 用户注册/登录（邮箱、手机、OAuth）
- 用户资料管理
- 团队创建与管理
- 角色权限控制（管理员、成员、访客）
- 订阅计费管理

**Freemium分层**：

| 功能 | 免费版 | 专业版 | 企业版 |
|------|--------|--------|--------|
| 项目数量 | 3个 | 无限制 | 无限制 |
| 存储空间 | 1GB | 50GB | 500GB |
| AI调用次数 | 100次/月 | 1000次/月 | 无限制 |
| 团队成员 | 1人 | 5人 | 无限制 |
| 高级功能 | ✗ | ✓ | ✓ |

### 2. 科研选题系统

**核心功能**：
- 研究热点分析
- 选题推荐（基于用户兴趣和领域）
- 可行性评估
- 创新性分析
- 竞争态势分析

**AI能力**：
- 分析近5年文献趋势
- 识别研究空白
- 评估选题价值
- 生成选题报告

### 3. 文献管理系统

**核心功能**：
- 文献检索（Google Scholar、PubMed、CNKI等）
- 文献导入（DOI、PDF、BibTeX）
- 文献分类与标签
- 文献笔记与标注
- 引用管理（GB/T 7714、APA、MLA等格式）

**AI能力**：
- 文献摘要生成
- 文献关系分析
- 关键词提取
- 引用推荐

### 4. 论文写作系统

**核心原则**：零幻觉、强上下文

**架构设计**：

```
用户输入 → 上下文分析器 → 文献检索器 → RAG生成器 → 验证器 → 输出
```

**关键功能**：
- 上下文感知写作
- 文献支持的生成
- 引用追踪
- 幻觉检测
- 续写模式（严格/扩展/创意）

**验证流程**：
1. 生成前验证：检查文献库、上下文、指令
2. 生成中验证：实时检查内容与文献一致性
3. 生成后验证：完整性、一致性、格式检查

### 5. 代码编辑系统

**核心功能**：
- 代码编辑器（Monaco Editor）
- 多语言支持（Python、R、MATLAB、Julia）
- 代码运行环境（Docker容器）
- 运行结果展示
- 代码版本管理

**AI能力**：
- 代码生成
- 代码解释
- 错误修复
- 性能优化建议

### 6. 科研图像系统

**核心功能**：
- AI图像生成（GPT IMAGE 2）
- 矢量图编辑（Canvas）
- 图像模板库
- 图像导出（SVG、PNG、PDF、EPS）
- 图像版本管理

**技术实现**：
- 前端：HTML5 Canvas + Fabric.js/SVG.js
- 后端：Node.js + Sharp（图像处理）
- AI：OpenAI GPT IMAGE 2 API
- 存储：SVG文件存储在S3，元数据存储在PostgreSQL

### 7. 审稿投递系统

**核心功能**：
- 期刊推荐（基于论文主题和质量）
- 投稿格式检查
- 审稿意见管理
- 修改版本跟踪
- 投稿状态追踪

**AI能力**：
- 期刊匹配分析
- 审稿意见解读
- 修改建议生成
- 投稿成功率预测

---

## 数据库设计

### 数据库选型与职责

| 数据库 | 版本 | 用途 | 数据类型 |
|--------|------|------|----------|
| PostgreSQL | 18+ | 核心业务数据 | 用户、项目、订阅、元数据 |
| Redis | 8.x | 缓存与实时功能 | 会话、协作状态、临时数据 |
| MongoDB | 8.x | 文档存储 | 论文内容、版本历史、文献 |
| S3/Minio | - | 文件存储 | PDF、图像、代码文件 |

### PostgreSQL 数据模型

#### 用户与权限

```sql
-- 用户表
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    institution VARCHAR(255),
    research_fields TEXT[],
    subscription_tier VARCHAR(50) DEFAULT 'free',
    subscription_expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 团队表
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 团队成员表
CREATE TABLE team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID REFERENCES teams(id),
    user_id UUID REFERENCES users(id),
    role VARCHAR(50) NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, user_id)
);
```

#### 科研项目

```sql
-- 项目表
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(500) NOT NULL,
    description TEXT,
    owner_id UUID REFERENCES users(id),
    team_id UUID REFERENCES teams(id),
    status VARCHAR(50) DEFAULT 'active',
    research_field VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 项目成员表
CREATE TABLE project_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    user_id UUID REFERENCES users(id),
    role VARCHAR(50) NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, user_id)
);
```

#### 文献管理

```sql
-- 文献表
CREATE TABLE references (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    title VARCHAR(1000) NOT NULL,
    authors TEXT[],
    year INTEGER,
    journal VARCHAR(500),
    doi VARCHAR(255),
    url TEXT,
    abstract TEXT,
    keywords TEXT[],
    citation_key VARCHAR(100),
    bibtex TEXT,
    pdf_url TEXT,
    notes TEXT,
    tags TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 订阅计费

```sql
-- 订阅计划表
CREATE TABLE subscription_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    tier VARCHAR(50) NOT NULL,
    price_monthly DECIMAL(10,2),
    price_yearly DECIMAL(10,2),
    features JSONB NOT NULL,
    limits JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 订阅记录表
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    plan_id UUID REFERENCES subscription_plans(id),
    status VARCHAR(50) NOT NULL,
    current_period_start TIMESTAMP,
    current_period_end TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 使用量记录表
CREATE TABLE usage_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    resource_type VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### MongoDB 数据模型

#### 论文文档

```json
{
  "_id": "ObjectId",
  "project_id": "UUID",
  "title": "论文标题",
  "content": {
    "abstract": "摘要内容",
    "introduction": "引言内容",
    "methods": "方法内容",
    "results": "结果内容",
    "discussion": "讨论内容",
    "conclusion": "结论内容"
  },
  "metadata": {
    "word_count": 5000,
    "citation_count": 25,
    "last_edited_by": "user_id",
    "format": "markdown"
  },
  "versions": [
    {
      "version_id": "v1",
      "content": {},
      "created_at": "2024-01-01T00:00:00Z",
      "created_by": "user_id",
      "message": "初稿完成"
    }
  ],
  "citations": [
    {
      "reference_id": "UUID",
      "citation_key": "author2024",
      "positions": [
        { "section": "introduction", "paragraph": 2, "offset": 150 }
      ]
    }
  ],
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Redis 数据结构

```
# 用户会话
session:{session_id} -> {user_id, expires_at, metadata}

# 实时协作状态
collaboration:{document_id} -> {active_users, cursors, locks}

# 使用量计数器
usage:{user_id}:{resource_type}:{YYYY-MM-DD} -> count

# 缓存
cache:{entity_type}:{entity_id} -> serialized_data
```

---

## API设计

### API端点设计

#### 用户服务 `/api/users`

```
POST   /api/users/register          # 用户注册
POST   /api/users/login             # 用户登录
POST   /api/users/logout            # 用户登出
GET    /api/users/me                 # 获取当前用户信息
PUT    /api/users/me                 # 更新用户信息

# 团队管理
POST   /api/teams                   # 创建团队
GET    /api/teams                   # 获取用户的团队列表
GET    /api/teams/:id               # 获取团队详情
PUT    /api/teams/:id               # 更新团队信息
DELETE /api/teams/:id               # 删除团队

# 订阅管理
GET    /api/subscriptions/plans     # 获取订阅计划列表
POST   /api/subscriptions           # 创建订阅
GET    /api/subscriptions/me        # 获取当前订阅
GET    /api/subscriptions/usage     # 获取使用量
```

#### 项目服务 `/api/projects`

```
POST   /api/projects                # 创建项目
GET    /api/projects                # 获取项目列表
GET    /api/projects/:id            # 获取项目详情
PUT    /api/projects/:id            # 更新项目
DELETE /api/projects/:id            # 删除项目
```

#### 文献服务 `/api/references`

```
POST   /api/projects/:pid/references           # 添加文献
GET    /api/projects/:pid/references            # 获取文献列表
GET    /api/projects/:pid/references/:id        # 获取文献详情
PUT    /api/projects/:pid/references/:id        # 更新文献
DELETE /api/projects/:pid/references/:id        # 删除文献
POST   /api/projects/:pid/references/import     # 批量导入
POST   /api/projects/:pid/references/search     # 搜索文献
```

#### 写作服务 `/api/documents`

```
POST   /api/projects/:pid/documents            # 创建文档
GET    /api/projects/:pid/documents            # 获取文档列表
GET    /api/projects/:pid/documents/:id        # 获取文档详情
PUT    /api/projects/:pid/documents/:id        # 更新文档
DELETE /api/projects/:pid/documents/:id        # 删除文档

# AI写作辅助
POST   /api/projects/:pid/documents/:id/ai/generate   # 生成内容
POST   /api/projects/:pid/documents/:id/ai/polish     # 润色内容
POST   /api/projects/:pid/documents/:id/ai/continue   # 续写内容

# 引用管理
POST   /api/projects/:pid/documents/:id/citations     # 插入引用
GET    /api/projects/:pid/documents/:id/citations     # 获取引用列表
```

#### 代码服务 `/api/code`

```
POST   /api/projects/:pid/code/files            # 创建代码文件
GET    /api/projects/:pid/code/files            # 获取代码文件列表
PUT    /api/projects/:pid/code/files/:id        # 更新代码文件
POST   /api/projects/:pid/code/run              # 运行代码
GET    /api/projects/:pid/code/run/:runid       # 获取运行结果
```

#### 图像服务 `/api/images`

```
POST   /api/projects/:pid/images                # 上传图像
GET    /api/projects/:pid/images                # 获取图像列表
POST   /api/projects/:pid/images/generate       # AI生成图像
POST   /api/projects/:pid/images/:id/edit       # 编辑图像
POST   /api/projects/:pid/images/:id/export     # 导出图像
```

#### AI服务 `/api/ai`

```
POST   /api/ai/chat                             # 聊天对话
POST   /api/ai/academic/write                   # 学术写作
POST   /api/ai/academic/polish                  # 学术润色
POST   /api/ai/references/analyze               # 分析文献
POST   /api/ai/code/generate                    # 生成代码
POST   /api/ai/images/generate                  # 生成图像
```

### API响应格式

```typescript
// 成功响应
interface ApiResponse<T> {
  success: true;
  data: T;
  meta?: {
    page?: number;
    limit?: number;
    total?: number;
  };
}

// 错误响应
interface ApiError {
  success: false;
  error: {
    code: string;
    message: string;
    details?: any;
  };
}
```

---

## 用户流程

### 核心用户旅程

```
用户注册/登录 → 创建/选择项目 → 选题阶段 → 写作阶段 → 投递阶段 → 论文完成
```

### 详细用户流程

#### 1. 注册与登录流程

```
新用户访问 → 选择注册方式 → 填写基本信息 → 邮箱验证 → 完善个人资料 → 选择订阅计划 → 进入工作台
```

#### 2. 项目创建流程

```
点击"新建项目" → 填写项目信息 → 选择项目类型 → 邀请团队成员 → 项目创建完成 → 进入项目工作区
```

#### 3. 选题阶段流程

```
进入选题模块 → 输入研究兴趣/关键词 → AI分析研究热点 → 查看选题推荐列表 → 选择感兴趣的选题 → 查看选题详情 → 确认选题 → 生成研究大纲
```

#### 4. 文献管理流程

```
进入文献模块 → 选择导入方式 → 导入文献 → 文献自动分析 → 文献整理 → 文献引用
```

#### 5. 论文写作流程

```
进入写作模块 → 选择写作模式 → 开始写作 → 手动写作/AI辅助写作 → 插入引用 → 格式调整 → 版本保存
```

#### 6. 代码编辑流程

```
进入代码模块 → 选择代码语言 → 编写/生成代码 → 运行代码 → 查看结果 → 保存代码
```

#### 7. 图像绘制流程

```
进入图像模块 → 选择图像类型 → 选择创建方式 → 创建/编辑图像 → 编辑图像 → 导出图像
```

#### 8. 审稿投递流程

```
进入投递模块 → 选择投稿论文 → AI分析论文 → 获取期刊推荐 → 选择目标期刊 → 准备投稿材料 → 提交投稿 → 追踪投稿状态 → 处理审稿意见 → 提交修改版本
```

---

## 界面设计

### 整体布局

```
┌─────────────────────────────────────────────────────────────────┐
│  顶部导航栏：Logo │ 项目选择器 │ 搜索 │ 通知 │ 用户头像 │ 设置  │
├─────────────────────────────────────────────────────────────────┤
│  左侧边栏  │                    主内容区                        │
│  ┌─────────┐  ┌────────────────────────────────────────────┐  │
│  │ 工作台   │  │                                            │  │
│  │ • 项目   │  │              动态内容区                     │  │
│  │ • 文献   │  │                                            │  │
│  │ • 写作   │  │                                            │  │
│  │ • 代码   │  │                                            │  │
│  │ • 图像   │  │                                            │  │
│  │ • 投递   │  │                                            │  │
│  └─────────┘  └────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 核心页面

1. **工作台（Dashboard）**：最近项目、待办事项、使用统计
2. **项目工作区**：项目导航、进度概览、最近活动
3. **文献管理页面**：搜索、分类、文献列表
4. **论文写作页面**：工具栏、文档大纲、编辑区、AI助手
5. **代码编辑页面**：文件列表、代码编辑器、运行结果
6. **图像绘制页面**：模板库、画布、AI生成
7. **审稿投递页面**：论文分析、推荐期刊

---

## 部署架构

### 部署环境

- **开发环境**：Docker Compose
- **生产环境**：Kubernetes

### Docker Compose（开发环境）

```yaml
version: '3.8'

services:
  # 前端
  frontend:
    build: ./frontend
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://api-gateway:8080

  # API网关
  api-gateway:
    image: kong:latest
    ports:
      - "8080:8000"
      - "8443:8443"
    environment:
      - KONG_DATABASE=postgres
      - KONG_PG_HOST=kong-db

  # Go服务
  user-service:
    build: ./services/user-service
    ports:
      - "8081:8081"
    environment:
      - DATABASE_URL=postgresql://user:password@user-db:5432/users
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - user-db
      - redis

  project-service:
    build: ./services/project-service
    ports:
      - "8082:8082"
    environment:
      - DATABASE_URL=postgresql://user:password@project-db:5432/projects
      - REDIS_URL=redis://redis:6379
    depends_on:
      - project-db
      - redis

  writing-service:
    build: ./services/writing-service
    ports:
      - "8084:8084"
    environment:
      - DATABASE_URL=postgresql://user:password@writing-db:5432/writing
      - MONGODB_URL=mongodb://writing-doc-db:27017/documents
      - REDIS_URL=redis://redis:6379
    depends_on:
      - writing-db
      - writing-doc-db
      - redis

  image-service:
    build: ./services/image-service
    ports:
      - "8086:8086"
    environment:
      - DATABASE_URL=postgresql://user:password@image-db:5432/images
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    depends_on:
      - image-db

  storage-service:
    build: ./services/storage-service
    ports:
      - "8088:8088"
    environment:
      - S3_ENDPOINT=http://minio:9000
      - S3_ACCESS_KEY=${S3_ACCESS_KEY}
      - S3_SECRET_KEY=${S3_SECRET_KEY}
    depends_on:
      - minio

  # Python服务
  reference-service:
    build: ./services/reference-service
    ports:
      - "8083:8083"
    environment:
      - MONGODB_URL=mongodb://reference-db:27017/references
      - REDIS_URL=redis://redis:6379
    depends_on:
      - reference-db
      - redis

  code-service:
    build: ./services/code-service
    ports:
      - "8085:8085"
    environment:
      - DATABASE_URL=postgresql://user:password@code-db:5432/code
      - DOCKER_HOST=unix:///var/run/docker.sock
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    depends_on:
      - code-db

  ai-service:
    build: ./services/ai-service
    ports:
      - "8087:8087"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - REDIS_URL=redis://redis:6379
    depends_on:
      - redis

  # 数据库
  user-db:
    image: postgres:18
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=users
    volumes:
      - user-db-data:/var/lib/postgresql/data

  project-db:
    image: postgres:18
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=projects
    volumes:
      - project-db-data:/var/lib/postgresql/data

  writing-db:
    image: postgres:18
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=writing
    volumes:
      - writing-db-data:/var/lib/postgresql/data

  code-db:
    image: postgres:18
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=code
    volumes:
      - code-db-data:/var/lib/postgresql/data

  image-db:
    image: postgres:18
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=images
    volumes:
      - image-db-data:/var/lib/postgresql/data

  reference-db:
    image: mongo:8
    volumes:
      - reference-db-data:/data/db

  writing-doc-db:
    image: mongo:8
    volumes:
      - writing-doc-db-data:/data/db

  # 缓存
  redis:
    image: redis:8-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data

  # 对象存储
  minio:
    image: minio/minio
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ROOT_USER=${S3_ACCESS_KEY}
      - MINIO_ROOT_PASSWORD=${S3_SECRET_KEY}
    command: server /data --console-address ":9001"
    volumes:
      - minio-data:/data

  # Kong数据库
  kong-db:
    image: postgres:18
    environment:
      - POSTGRES_USER=kong
      - POSTGRES_PASSWORD=kong
      - POSTGRES_DB=kong
    volumes:
      - kong-db-data:/var/lib/postgresql/data

volumes:
  user-db-data:
  project-db-data:
  writing-db-data:
  code-db-data:
  image-db-data:
  reference-db-data:
  writing-doc-db-data:
  redis-data:
  minio-data:
  kong-db-data:
```

### CI/CD流程

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run tests
        run: npm test

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - name: Build Docker images
        run: docker-compose build
      - name: Push to registry
        run: docker-compose push

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - name: Deploy to Kubernetes
        run: kubectl apply -f k8s/
```

---

## 安全设计

### 认证与授权

#### JWT认证

```typescript
interface JwtPayload {
  sub: string;        // 用户ID
  email: string;      // 用户邮箱
  role: string;       // 用户角色
  tier: string;       // 订阅等级
  iat: number;        // 签发时间
  exp: number;        // 过期时间
}

const tokenConfig = {
  accessToken: { expiresIn: '15m', algorithm: 'RS256' },
  refreshToken: { expiresIn: '7d', algorithm: 'RS256' },
};
```

#### RBAC权限模型

```typescript
enum Role {
  SUPER_ADMIN = 'super_admin',
  ADMIN = 'admin',
  USER = 'user',
  VIEWER = 'viewer',
}

enum Permission {
  PROJECT_CREATE = 'project:create',
  PROJECT_READ = 'project:read',
  PROJECT_UPDATE = 'project:update',
  PROJECT_DELETE = 'project:delete',
  // ... 其他权限
}
```

### 数据安全

- 传输加密：TLS 1.3
- 存储加密：AES-256-GCM
- 字段级加密：敏感字段单独加密
- 数据脱敏：日志和展示中脱敏

### API安全

- 限流：基于用户订阅等级
- 输入验证：严格验证所有输入
- 输出编码：防止XSS
- CORS：限制跨域请求

### 安全头部

```typescript
const securityHeaders = {
  'X-Frame-Options': 'DENY',
  'X-Content-Type-Options': 'nosniff',
  'X-XSS-Protection': '1; mode=block',
  'Strict-Transport-Security': 'max-age=31536000; includeSubDomains',
  'Content-Security-Policy': "default-src 'self';",
  'Referrer-Policy': 'strict-origin-when-cross-origin',
};
```

---

## 测试策略

### 测试层次

```
单元测试（80%+）→ 集成测试（70%+）→ E2E测试（核心流程100%）
```

### 测试工具

**前端测试**：
- 单元测试：Jest + React Testing Library
- E2E测试：Playwright

**Go服务测试**：
- 单元测试：Go内置testing包 + testify
- 集成测试：Go内置testing包 + httptest
- API测试：Go内置testing包 + httpexpect

**Python服务测试**：
- 单元测试：pytest
- 集成测试：pytest + httpx
- API测试：pytest + FastAPI TestClient

### 测试覆盖率目标

| 测试类型 | 覆盖率目标 | 工具 |
|---------|-----------|------|
| 前端单元测试 | 80%+ | Jest + React Testing Library |
| Go服务单元测试 | 80%+ | Go testing + testify |
| Python服务单元测试 | 80%+ | pytest |
| 集成测试 | 70%+ | 各语言集成测试工具 |
| E2E测试 | 核心流程100% | Playwright |

---

## 监控与运维

### 监控架构

```
Prometheus → Grafana → 告警通知
    ↓
日志系统（ELK Stack）
    ↓
追踪系统（Jaeger）
```

### 监控指标

- 系统指标：CPU、内存、磁盘、网络
- 应用指标：请求量、响应时间、错误率
- 业务指标：用户活跃度、AI调用量、订阅转化率

### 告警规则

- 系统告警：CPU>80%、内存>80%、磁盘>80%
- 应用告警：错误率>5%、响应时间>1s、服务宕机
- 业务告警：AI错误率>10%、AI响应时间>10s

---

## 开发计划

### 开发阶段

| 阶段 | 周期 | 内容 |
|------|------|------|
| Phase 1: 基础架构 | 4周 | 用户服务、项目服务、基础UI |
| Phase 2: 核心功能 | 6周 | 文献、写作、AI、协作 |
| Phase 3: 扩展功能 | 4周 | 代码、图像、投递 |
| Phase 4: 商业化 | 3周 | 订阅、支付、分析 |
| Phase 5: 优化上线 | 3周 | 性能、安全、监控 |

### 里程碑

| 里程碑 | 时间 | 交付物 |
|--------|------|--------|
| M1: 基础架构完成 | 第4周 | 用户、项目服务，基础UI |
| M2: 核心功能完成 | 第10周 | 文献、写作、AI、协作 |
| M3: 扩展功能完成 | 第14周 | 代码、图像、投递 |
| M4: 商业化完成 | 第17周 | 订阅、支付、分析 |
| M5: 上线准备完成 | 第20周 | 性能、安全、监控 |

### 资源需求

| 角色 | 人数 | 职责 |
|------|------|------|
| 产品经理 | 1 | 需求管理、用户研究 |
| 前端开发 | 2 | UI开发、交互实现（TypeScript/React） |
| Go后端开发 | 2 | Go服务开发、API设计（用户、项目、写作、图像、存储） |
| Python后端开发 | 1 | Python服务开发（文献、代码、AI服务） |
| AI工程师 | 1 | AI模型集成、优化 |
| DevOps | 1 | 部署、运维、监控 |
| 测试工程师 | 1 | 测试、质量保证 |

**总计**：9人，20周

---

## 风险评估

### 技术风险

| 风险 | 可能性 | 影响 | 等级 | 缓解措施 |
|------|--------|------|------|----------|
| AI API不稳定 | 中 | 高 | 6 | 多供应商备份、降级策略 |
| 实时协作冲突 | 中 | 中 | 4 | 使用成熟方案（Yjs/CRDT） |
| 性能瓶颈 | 中 | 高 | 6 | 提前压测、缓存策略 |
| 数据安全泄露 | 低 | 高 | 3 | 加密存储、访问控制 |
| 第三方服务中断 | 中 | 中 | 4 | 多供应商、本地备份 |

### 业务风险

| 风险 | 可能性 | 影响 | 等级 | 缓解措施 |
|------|--------|------|------|----------|
| 用户需求变化 | 高 | 中 | 6 | 敏捷开发、快速迭代 |
| 竞争对手抢先 | 中 | 高 | 6 | 差异化定位、快速上线 |
| 商业模式不成立 | 低 | 高 | 3 | 市场验证、灵活调整 |
| 法律合规问题 | 低 | 中 | 2 | 法律咨询、合规审查 |
| 资金不足 | 低 | 高 | 3 | 成本控制、融资规划 |

---

## 成本估算

### 开发成本

- 人力成本（20周）：2,000,000元
- 其他开发成本：100,000元
- **开发总成本**：2,100,000元

### 运营成本（月度）

- 基础设施成本：8,400元/月
- AI API成本：23,050元/月
- 其他运营成本：1,500元/月
- **月度运营总成本**：33,000元/月

### 收入预测

| 月份 | 订阅收入 | AI使用收入 | 总收入 |
|------|---------|-----------|--------|
| 1 | 5,000 | 2,000 | 7,000 |
| 3 | 30,000 | 10,000 | 40,000 |
| 6 | 100,000 | 35,000 | 135,000 |
| 12 | 400,000 | 150,000 | 550,000 |
| 24 | 1,500,000 | 500,000 | 2,000,000 |

### 盈亏平衡点

- 固定成本：33,000元/月
- 边际成本：约5元/付费用户/月
- 平均收入：约100元/付费用户/月
- **盈亏平衡点**：约347个付费用户

---

## 商业模式

### 价值主张

对于科研人员来说，DeepWrite是一个：
- 一站式科研工作平台
- AI驱动的智能助手
- 协作与知识管理平台
- 学术规范保障

### 收入模式

1. **订阅收入（主要）**：免费版、专业版（¥99/月）、企业版（¥299/月）
2. **使用量收入（补充）**：AI调用、存储空间等
3. **增值服务收入**：专属客服、定制开发、培训服务

### 客户获取策略

1. 内容营销：学术博客、视频教程、白皮书
2. 社区运营：学术社区、用户社群、KOL合作
3. 渠道合作：高校合作、期刊合作、科研工具集成
4. 产品驱动增长：免费增值、邀请奖励、口碑传播

### 竞争优势

1. 一站式平台：覆盖科研全流程
2. AI深度集成：深度理解科研场景
3. 学术规范保障：零幻觉设计
4. 本地化优势：中文支持、国内文献数据库

---

## 法律合规

### 合规框架

1. **数据隐私合规**：个人信息保护法（PIPL）、数据安全法、网络安全法
2. **学术诚信合规**：学术不端行为界定、AI生成内容标注、引用规范
3. **知识产权合规**：版权保护、商标注册、专利申请
4. **商业合规**：经营许可证、税务合规、消费者权益保护

### 数据隐私合规

- 最小必要原则：只收集必要的个人信息
- 目的限制原则：只用于明确告知的目的
- 存储期限原则：不超过必要期限
- 安全保障原则：采取必要安全措施
- 主体权利原则：保障用户知情、决定、查阅、复制、更正、删除等权利

### 学术诚信合规

- AI内容标注：标注AI生成/辅助/润色的内容
- 抄袭检测：检测抄袭和重复内容
- 引用检查：确保引用格式正确、完整

### 商业合规

- ICP许可证：互联网信息服务许可证
- 增值电信许可证：B25信息服务业务
- 税务合规：增值税、企业所得税、个人所得税

---

## 附录

### 术语表

| 术语 | 定义 |
|------|------|
| SaaS | Software as a Service，软件即服务 |
| Freemium | 免费增值模式，基础功能免费，高级功能付费 |
| RBAC | Role-Based Access Control，基于角色的访问控制 |
| JWT | JSON Web Token，JSON网络令牌 |
| RAG | Retrieval-Augmented Generation，检索增强生成 |
| CRDT | Conflict-free Replicated Data Type，无冲突复制数据类型 |

### 参考资料

- Next.js官方文档：https://nextjs.org/docs
- PostgreSQL官方文档：https://www.postgresql.org/docs/
- Kubernetes官方文档：https://kubernetes.io/docs/
- OpenAI API文档：https://platform.openai.com/docs/
- Anthropic API文档：https://docs.anthropic.com/

---

**文档版本历史**

| 版本 | 日期 | 变更说明 |
|------|------|----------|
| 1.0 | 2026-04-25 | 初始版本 |
| 1.1 | 2026-04-25 | 技术栈调整：后端从Node.js改为Go+Python，更新服务划分、Docker Compose配置、测试策略和资源需求 |
