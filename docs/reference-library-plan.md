# 文献库功能实施计划

## 一、需求概述

### 1.1 功能定位
- 学术引用管理 + 知识库系统
- 支持混合归属模式（用户私有 + 可分享到工作区）

### 1.2 核心功能 (MVP)

| 功能 | 优先级 | 说明 |
|------|--------|------|
| 文献数据模型 | P0 | 支持 Reference, Collection, Note |
| 文献 CRUD API | P0 | 基础管理接口 |
| DOI 元数据获取 | P0 | Crossref API 集成 |
| 搜索与过滤 | P0 | 标题/作者/标签/年份 |
| BibTeX 导入 | P1 | 解析 .bib 文件 |
| RIS 导入 | P1 | 解析 .ris 文件 |
| PDF 关联 | P1 | 关联 WorkspaceFile |
| 收藏集管理 | P1 | 分类组织文献 |
| PDF 元数据提取 | P2 | 后端 Worker 处理 |
| CSL 引用格式化 | P2 | citeproc-js 前端生成 |

### 1.3 技术选型

| 层级 | 技术 | 说明 |
|------|------|------|
| 数据模型 | Go + GORM | 复用现有架构 |
| 元数据查询 | Crossref API | DOI 查询 |
| BibTeX 解析 | github.com/nickng/bibtex | Go 库 |
| CSL 格式化 | citeproc-js | 前端生成 |
| PDF 提取 | Python pdfminer | Worker 异步任务 |

---

## 二、数据模型设计

### 2.1 Reference (文献)

```go
// services/gateway/models/reference/reference.go

package reference

import (
    "time"
    "gorm.io/datatypes"
)

const (
    ReferenceTypeArticle     ReferenceType = "article"
    ReferenceTypeBook        ReferenceType = "book"
    ReferenceTypeBookChapter ReferenceType = "book-chapter"
    ReferenceTypeConference  ReferenceType = "conference"
    ReferenceTypeThesis      ReferenceType = "thesis"
    ReferenceTypeReport      ReferenceType = "report"
    ReferenceTypeWeb         ReferenceType = "web"
    ReferenceTypePreprint    ReferenceType = "preprint"
    ReferenceTypeUnknown     ReferenceType = "unknown"
)

type ReferenceType string

type Author struct {
    Family  string `json:"family"`
    Given   string `json:"given"`
    Suffix  string `json:"suffix,omitempty"`
    Literal string `json:"literal,omitempty"`
    ORCID   string `json:"orcid,omitempty"`
}

type Reference struct {
    ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    OwnerID     string         `json:"owner_id" gorm:"type:uuid;not null;index"`
    WorkspaceID *string        `json:"workspace_id,omitempty" gorm:"type:uuid;index"`
    
    Title       string         `json:"title" gorm:"type:varchar(1024);not null"`
    Authors     datatypes.JSON `json:"authors" gorm:"type:jsonb;default:'[]'::jsonb"`
    Year        *int           `json:"year,omitempty" gorm:"index"`
    Source      string         `json:"source,omitempty" gorm:"type:varchar(255)"`          // 期刊/会议/出版社
    DOI         string         `json:"doi,omitempty" gorm:"type:varchar(255);uniqueIndex"`
    ISBN        string         `json:"isbn,omitempty" gorm:"type:varchar(20)"`
    URL         string         `json:"url,omitempty" gorm:"type:varchar(1024)"`
    Abstract    string         `json:"abstract,omitempty" gorm:"type:text"`
    Keywords    datatypes.JSON `json:"keywords,omitempty" gorm:"type:jsonb"`             // string[]
    Type        ReferenceType  `json:"type" gorm:"type:varchar(50);not null;default:'unknown'"`
    
    Volume      string         `json:"volume,omitempty" gorm:"type:varchar(50)"`
    Issue       string         `json:"issue,omitempty" gorm:"type:varchar(50)"`
    Pages       string         `json:"pages,omitempty" gorm:"type:varchar(50)"`
    Publisher   string         `json:"publisher,omitempty" gorm:"type:varchar(255)"`
    Language    string         `json:"language,omitempty" gorm:"type:varchar(20)"`
    
    FileID      *string        `json:"file_id,omitempty" gorm:"type:uuid;index"`
    CitationKey string         `json:"citation_key,omitempty" gorm:"type:varchar(255);index"` // vaswani2017
    BibtexRaw   string         `json:"bibtex_raw,omitempty" gorm:"type:text"`
    Metadata    datatypes.JSON `json:"metadata,omitempty" gorm:"type:jsonb"`                  // 扩展字段
    
    Starred     bool           `json:"starred" gorm:"not null;default:false;index"`
    Status      string         `json:"status" gorm:"type:varchar(50);not null;default:'active'"` // active, reviewed, archived
    
    CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
    DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (Reference) TableName() string {
    return "references"
}
```

### 2.2 Collection (收藏集)

```go
// services/gateway/models/reference/collection.go

package reference

import "time"

type Collection struct {
    ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    OwnerID     string    `json:"owner_id" gorm:"type:uuid;not null;index"`
    WorkspaceID *string   `json:"workspace_id,omitempty" gorm:"type:uuid;index"`
    ParentID    *string   `json:"parent_id,omitempty" gorm:"type:uuid;index"`
    
    Name        string    `json:"name" gorm:"type:varchar(255);not null"`
    Description string    `json:"description,omitempty" gorm:"type:text"`
    Color       string    `json:"color,omitempty" gorm:"type:varchar(20)"`
    Icon        string    `json:"icon,omitempty" gorm:"type:varchar(50)"`
    SortOrder   int       `json:"sort_order" gorm:"not null;default:0"`
    
    CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
    DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (Collection) TableName() string {
    return "collections"
}

type CollectionReference struct {
    CollectionID string    `json:"collection_id" gorm:"primaryKey;type:uuid"`
    ReferenceID  string    `json:"reference_id" gorm:"primaryKey;type:uuid"`
    SortOrder    int       `json:"sort_order" gorm:"not null;default:0"`
    AddedAt      time.Time `json:"added_at" gorm:"autoCreateTime"`
}

func (CollectionReference) TableName() string {
    return "collection_references"
}
```

### 2.3 ReferenceNote (文献笔记)

```go
// services/gateway/models/reference/note.go

package reference

import "time"

type ReferenceNote struct {
    ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    ReferenceID string    `json:"reference_id" gorm:"type:uuid;not null;index"`
    UserID      string    `json:"user_id" gorm:"type:uuid;not null;index"`
    
    Title       string    `json:"title,omitempty" gorm:"type:varchar(255)"`
    Content     string    `json:"content" gorm:"type:text;not null"`              // Markdown
    PageFrom    *int      `json:"page_from,omitempty"`
    PageTo      *int      `json:"page_to,omitempty"`
    
    CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
    DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (ReferenceNote) TableName() string {
    return "reference_notes"
}
```

---

## 三、API 接口设计

### 3.1 文献管理

```
GET    /api/v1/references                    # 列表 (分页/搜索/过滤)
GET    /api/v1/references/:id                # 详情
POST   /api/v1/references                    # 创建
PUT    /api/v1/references/:id                # 更新
DELETE /api/v1/references/:id                # 删除 (软删除)
POST   /api/v1/references/:id/file           # 关联 PDF
DELETE /api/v1/references/:id/file           # 移除 PDF
POST   /api/v1/references/batch              # 批量操作
```

### 3.2 元数据查询

```
GET    /api/v1/references/lookup/doi         # DOI 查询 -> Crossref
GET    /api/v1/references/lookup/isbn        # ISBN 查询 -> Open Library
POST   /api/v1/references/lookup/batch       # 批量查询
```

### 3.3 导入导出

```
POST   /api/v1/references/import/bibtex      # BibTeX 导入
POST   /api/v1/references/import/ris         # RIS 导入
GET    /api/v1/references/export/bibtex      # 导出 BibTeX
GET    /api/v1/references/export/ris         # 导出 RIS
GET    /api/v1/references/export/csl         # 导出 CSL JSON
```

### 3.4 收藏集

```
GET    /api/v1/collections                  # 列表 (树形)
POST   /api/v1/collections                  # 创建
PUT    /api/v1/collections/:id              # 更新
DELETE /api/v1/collections/:id              # 删除
PUT    /api/v1/collections/:id/references    # 设置收藏集文献
POST   /api/v1/collections/:id/references/:rid  # 添加文献
DELETE /api/v1/collections/:id/references/:rid  # 移除文献
```

### 3.5 笔记

```
GET    /api/v1/references/:id/notes         # 笔记列表
POST   /api/v1/references/:id/notes         # 创建笔记
PUT    /api/v1/notes/:id                    # 更新笔记
DELETE /api/v1/notes/:id                    # 删除笔记
```

---

## 四、请求/响应示例

### 4.1 文献列表

```typescript
// GET /api/v1/references?
//   q=transformer&
//   type=article&
//   year_from=2020&
//   year_to=2024&
//   collection_id=xxx&
//   starred=true&
//   sort_by=created_at&
//   sort_order=desc&
//   page=1&
//   page_size=20

interface ListReferencesResponse {
  items: Reference[]
  total: number
  page: number
  page_size: number
}
```

### 4.2 创建文献

```typescript
// POST /api/v1/references
interface CreateReferenceInput {
  title: string
  authors?: Author[]
  year?: number
  source?: string
  doi?: string
  isbn?: string
  url?: string
  abstract?: string
  keywords?: string[]
  type?: ReferenceType
  volume?: string
  issue?: string
  pages?: string
  publisher?: string
  language?: string
  file_id?: string
  workspace_id?: string
  citation_key?: string
  bibtex_raw?: string
}
```

### 4.3 DOI 查询响应

```typescript
// GET /api/v1/references/lookup/doi?doi=10.1038/nature12373
interface DoiLookupResponse {
  found: boolean
  reference?: {
    title: string
    authors: Author[]
    year: number
    source: string
    doi: string
    type: ReferenceType
    abstract?: string
    volume?: string
    issue?: string
    pages?: string
    publisher?: string
    url?: string
  }
}
```

### 4.4 BibTeX 导入

```typescript
// POST /api/v1/references/import/bibtex
// Content-Type: multipart/form-data
// file: .bib 文件

interface BibtexImportResponse {
  imported: number
  skipped: number
  references: Reference[]
  errors?: Array<{ line: number; message: string }>
}
```

---

## 五、前端类型定义

```typescript
// apps/web/src/types/reference.ts

export type ReferenceType = 
  | 'article'
  | 'book'
  | 'book-chapter'
  | 'conference'
  | 'thesis'
  | 'report'
  | 'web'
  | 'preprint'
  | 'unknown'

export interface Author {
  family: string
  given: string
  suffix?: string
  literal?: string
  orcid?: string
}

export interface Reference {
  id: string
  owner_id: string
  workspace_id?: string
  
  title: string
  authors: Author[]
  year?: number
  source?: string
  doi?: string
  isbn?: string
  url?: string
  abstract?: string
  keywords?: string[]
  type: ReferenceType
  
  volume?: string
  issue?: string
  pages?: string
  publisher?: string
  language?: string
  
  file_id?: string
  citation_key?: string
  bibtex_raw?: string
  metadata?: Record<string, unknown>
  
  starred: boolean
  status: 'active' | 'reviewed' | 'archived'
  
  created_at: string
  updated_at: string
}

export interface Collection {
  id: string
  owner_id: string
  workspace_id?: string
  parent_id?: string
  
  name: string
  description?: string
  color?: string
  icon?: string
  sort_order: number
  
  created_at: string
  updated_at: string
  
  // 前端计算
  children?: Collection[]
  reference_count?: number
}

export interface ReferenceNote {
  id: string
  reference_id: string
  user_id: string
  
  title?: string
  content: string
  page_from?: number
  page_to?: number
  
  created_at: string
  updated_at: string
}
```

---

## 六、数据库迁移

```sql
-- references 表
CREATE TABLE references (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    workspace_id UUID,
    
    title VARCHAR(1024) NOT NULL,
    authors JSONB DEFAULT '[]'::jsonb,
    year INTEGER,
    source VARCHAR(255),
    doi VARCHAR(255) UNIQUE,
    isbn VARCHAR(20),
    url VARCHAR(1024),
    abstract TEXT,
    keywords JSONB,
    type VARCHAR(50) NOT NULL DEFAULT 'unknown',
    
    volume VARCHAR(50),
    issue VARCHAR(50),
    pages VARCHAR(50),
    publisher VARCHAR(255),
    language VARCHAR(20),
    
    file_id UUID,
    citation_key VARCHAR(255),
    bibtex_raw TEXT,
    metadata JSONB,
    
    starred BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- 索引
CREATE INDEX idx_references_owner ON references(owner_id);
CREATE INDEX idx_references_workspace ON references(workspace_id);
CREATE INDEX idx_references_doi ON references(doi);
CREATE INDEX idx_references_year ON references(year);
CREATE INDEX idx_references_type ON references(type);
CREATE INDEX idx_references_starred ON references(starred);
CREATE INDEX idx_references_created ON references(created_at);
CREATE INDEX idx_references_citation_key ON references(citation_key);

-- 全文搜索索引
CREATE INDEX idx_references_search ON references 
USING GIN(to_tsvector('english', title || ' ' || COALESCE(abstract, '')));

-- collections 表
CREATE TABLE collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    workspace_id UUID,
    parent_id UUID REFERENCES collections(id) ON DELETE CASCADE,
    
    name VARCHAR(255) NOT NULL,
    description TEXT,
    color VARCHAR(20),
    icon VARCHAR(50),
    sort_order INTEGER NOT NULL DEFAULT 0,
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_collections_owner ON collections(owner_id);
CREATE INDEX idx_collections_workspace ON collections(workspace_id);
CREATE INDEX idx_collections_parent ON collections(parent_id);

-- collection_references 关联表
CREATE TABLE collection_references (
    collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    reference_id UUID NOT NULL REFERENCES references(id) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (collection_id, reference_id)
);

CREATE INDEX idx_collection_references_ref ON collection_references(reference_id);

-- reference_notes 表
CREATE TABLE reference_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference_id UUID NOT NULL REFERENCES references(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    
    title VARCHAR(255),
    content TEXT NOT NULL,
    page_from INTEGER,
    page_to INTEGER,
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_reference_notes_ref ON reference_notes(reference_id);
CREATE INDEX idx_reference_notes_user ON reference_notes(user_id);
```

---

## 七、实施计划

### Phase 1: 后端基础 (3-4 天)

**Day 1: 数据模型**
- [x] 创建 `models/reference/reference.go`
- [x] 创建 `models/reference/collection.go`
- [x] 创建 `models/reference/note.go`
- [x] 注册到 GORM AutoMigrate

**Day 2-3: API Handler**
- [x] 文献 CRUD Handler
- [x] DOI 查询 Handler (Crossref API)
- [x] 搜索过滤逻辑
- [x] 路由注册
- [x] Crossref 搜索 API

**Day 4: 导入功能**
- [x] BibTeX 解析 (`pkg/bibtex/bibtex.go`)
- [x] RIS 解析 (`pkg/ris/ris.go`)
- [x] BibTeX 导入 API
- [x] RIS 导入 API
- [x] BibTeX 导出 API
- [x] RIS 导出 API
- [x] CSL JSON 导出 API

### Phase 2: 前端集成 (2-3 天)

**Day 1: 类型与 API**
- [x] 创建 `types/reference.ts`
- [x] 创建 `api/reference.ts`
- [x] 创建 `stores/reference.ts`

**Day 2-3: 组件开发**
- [x] LibraryView 改造（包含卡片、列表视图）
- [x] 文献详情面板
- [x] DOI 查询对话框
- [x] 手动输入表单
- [x] 收藏集管理对话框
- [x] 搜索功能（防抖）
- [x] 收藏/取消收藏
- [x] 删除文献

### Phase 3: 高级功能 (2-3 天)

- [x] PDF 元数据提取 Worker 任务
- [x] Gateway Worker 客户端
- [x] JATS 标签处理（摘要渲染）
- [ ] CSL 引用格式化 (citeproc-js)
- [ ] 文献笔记 API
- [ ] 批量操作优化

---

## 八、依赖库

### 8.1 后端 Go

```go
// go.mod additions
require (
    github.com/nickng/bibtex v1.2.0  // BibTeX 解析
)
```

### 8.2 前端

```json
// package.json additions
{
  "dependencies": {
    "citeproc": "^2.4.63"
  }
}
```

---

## 九、实施进度

### 已完成 (2024-03-25)

| 模块 | 功能 | 状态 | 文件 |
|------|------|------|------|
| **后端** | Reference 数据模型 | ✅ | `models/reference/reference.go` |
| | Collection 数据模型 | ✅ | `models/reference/collection.go` |
| | ReferenceNote 数据模型 | ✅ | `models/reference/note.go` |
| | GORM AutoMigrate 注册 | ✅ | `bootstrap/database.go` |
| | 文献 CRUD API | ✅ | `handler/reference_handler.go` |
| | 收藏集 CRUD API | ✅ | `handler/collection_handler.go` |
| | DOI 查询 (Crossref) | ✅ | `pkg/crossref/client.go` |
| | Crossref 搜索 API | ✅ | `pkg/crossref/client.go` |
| | Semantic Scholar 搜索 API | ✅ | `pkg/semanticscholar/client.go` |
| | AI 检索 API | ✅ | `handler/search_handler.go` |
| | JATS 标签处理 | ✅ | `pkg/crossref/client.go` |
| | Worker 客户端 | ✅ | `pkg/worker/client.go` |
| | Worker 配置 | ✅ | `pkg/config/config.go` |
| | PDF 提取触发 API | ✅ | `handler/reference_handler.go` |
| | BibTeX 解析器 | ✅ | `pkg/bibtex/bibtex.go` |
| | BibTeX 导入 API | ✅ | `handler/import_handler.go` |
| | BibTeX 导出 API | ✅ | `handler/import_handler.go` |
| | RIS 解析器 | ✅ | `pkg/ris/ris.go` |
| | RIS 导入 API | ✅ | `handler/import_handler.go` |
| | RIS 导出 API | ✅ | `handler/import_handler.go` |
| | CSL JSON 导出 API | ✅ | `handler/import_handler.go` |
| **Worker** | PDF 元数据提取 | ✅ | `worker/tasks/pdf_tasks.py` |
| | DOI 提取 + Crossref 查询 | ✅ | `worker/tasks/pdf_tasks.py` |
| | 任务状态查询 API | ✅ | `worker/api/main.py` |
| **前端** | TypeScript 类型定义 | ✅ | `types/reference.ts` |
| | API 封装 | ✅ | `api/reference.ts` |
| | Pinia 状态管理 | ✅ | `stores/reference.ts` |
| | LibraryView 页面 | ✅ | `views/LibraryView.vue` |
| | 网格/列表视图 | ✅ | `views/LibraryView.vue` |
| | 搜索功能 | ✅ | `views/LibraryView.vue` |
| | 文献详情面板 | ✅ | `views/LibraryView.vue` |
| | DOI 查询对话框 | ✅ | `views/LibraryView.vue` |
| | 手动输入表单 | ✅ | `views/LibraryView.vue` |
| | 收藏集管理 | ✅ | `views/LibraryView.vue` |
| | 收藏/取消收藏 | ✅ | `views/LibraryView.vue` |
| | 删除文献 | ✅ | `views/LibraryView.vue` |
| | 导入对话框 (BibTeX/RIS) | ✅ | `views/LibraryView.vue` |
| | 导出下拉菜单 | ✅ | `views/LibraryView.vue` |
| | AI 检索界面 | ✅ | `views/LibraryView.vue` |
| | 批量导入搜索结果 | ✅ | `views/LibraryView.vue` |
| | CSL 引用格式化工具 | ✅ | `utils/citation.ts` |
| | 国际化 (中/英) | ✅ | `locales/zh.ts`, `locales/en.ts` |

### 待实现

| 功能 | 优先级 | 说明 |
|------|--------|------|
| 文献笔记 API | P2 | ReferenceNote CRUD |
| 批量操作 | P2 | 批量删除、批量添加到收藏集 |
| PDF 预览 | P3 | 在线 PDF 阅读器集成 |
| 引用预览组件 | P3 | 在详情面板显示格式化引用 |

### 已实现 API 端点

```
文献管理:
GET    /api/v1/references                    # 列表 (分页/搜索/过滤)
POST   /api/v1/references                    # 创建
GET    /api/v1/references/:id                # 详情
PUT    /api/v1/references/:id                # 更新
DELETE /api/v1/references/:id                # 删除
POST   /api/v1/references/:id/file           # 关联 PDF
DELETE /api/v1/references/:id/file           # 移除 PDF
GET    /api/v1/references/lookup/doi         # DOI 查询
GET    /api/v1/references/search             # Crossref 搜索
POST   /api/v1/references/:id/extract-pdf    # PDF 元数据提取
GET    /api/v1/references/tasks/:task_id     # 任务状态

导入导出:
POST   /api/v1/references/import/bibtex      # BibTeX 导入
POST   /api/v1/references/import/ris         # RIS 导入
GET    /api/v1/references/export/bibtex      # BibTeX 导出
GET    /api/v1/references/export/ris         # RIS 导出
GET    /api/v1/references/export/csl         # CSL JSON 导出

AI 检索:
GET    /api/v1/references/ai-search         # AI 检索 (支持 provider 参数)
POST   /api/v1/references/ai-search/doi     # DOI 批量检索
POST   /api/v1/references/ai-search/import  # 批量导入检索结果

收藏集:
GET    /api/v1/collections                   # 列表 (树形)
POST   /api/v1/collections                   # 创建
GET    /api/v1/collections/:id               # 详情
PUT    /api/v1/collections/:id               # 更新
DELETE /api/v1/collections/:id               # 删除
PUT    /api/v1/collections/:id/references    # 设置文献
POST   /api/v1/collections/:id/references/:rid  # 添加文献
DELETE /api/v1/collections/:id/references/:rid  # 移除文献

Worker:
POST   /tasks/pdf/extract                    # PDF 元数据提取
POST   /tasks/pdf/extract-and-lookup         # PDF + DOI 查询
GET    /tasks/:task_id/status                # 任务状态
```

---

## 十、文件结构

```
services/gateway/
├── models/reference/
│   ├── reference.go          ✅ 文献数据模型
│   ├── collection.go          ✅ 收藏集数据模型
│   └── note.go                ✅ 笔记数据模型
├── handler/
│   ├── reference_handler.go   ✅ 文献 API Handler
│   ├── collection_handler.go  ✅ 收藏集 API Handler
│   └── import_handler.go      ✅ 导入导出 API Handler
├── pkg/
│   ├── crossref/
│   │   └── client.go          ✅ Crossref API 客户端
│   ├── bibtex/
│   │   └── bibtex.go          ✅ BibTeX 解析/导出
│   ├── ris/
│   │   └── ris.go             ✅ RIS 解析/导出
│   └── worker/
│       └── client.go          ✅ Worker HTTP 客户端
├── bootstrap/
│   └── database.go            ✅ AutoMigrate 注册
├── pkg/config/
│   └── config.go              ✅ Worker 配置
└── routers/
    └── router.go              ✅ 路由注册

services/worker/
├── worker/tasks/
│   └── pdf_tasks.py           ✅ PDF 元数据提取任务
├── api/
│   └── main.py                ✅ Worker API
└── pyproject.toml             ✅ 依赖 (pdfminer.six, httpx)

apps/web/src/
├── types/
│   └── reference.ts           ✅ TypeScript 类型定义
├── api/
│   └── reference.ts           ✅ API 封装
├── stores/
│   └── reference.ts           ✅ Pinia 状态管理
├── utils/
│   └── citation.ts            ✅ CSL 引用格式化
├── views/
│   └── LibraryView.vue        ✅ 文献库页面
└── locales/
    ├── zh.ts                  ✅ 中文翻译
    └── en.ts                  ✅ 英文翻译
```