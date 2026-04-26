# DeepWrite - AI辅助科研工作平台

DeepWrite 是一款专为科研人员设计的一站式科研论文写作平台。它集成了文献管理、AI辅助写作、代码编辑与运行、以及专业的图像绘制功能，旨在通过人工智能技术提升学术生产力。

## 核心功能

- **文献管理**：集成主流文献库，支持智能引用与摘要提取。
- **AI 写作辅助**：提供语法检查、风格优化、自动润色及大纲生成。
- **集成开发环境**：支持在平台内直接编写并运行论文相关代码。
- **科学绘图**：内置专业绘图工具，支持生成符合学术标准的图表。

## 技术栈

### 前端
- **框架**: Next.js 16+ / React 19+
- **语言**: TypeScript
- **样式**: TailwindCSS
- **组件库**: Shadcn/ui

### 后端 (微服务架构)
- **核心业务**: Go 1.22+ (Gin)
- **AI 逻辑**: Python 3.12+ (FastAPI)

### 基础设施 & 存储
- **关系型数据库**: PostgreSQL 18+
- **文档数据库**: MongoDB 8.x
- **缓存**: Redis 8.x
- **对象存储**: S3 (Compatible)
- **编排 & 网关**: Docker / Kubernetes / Kong

## 目录结构

```text
.
├── docs/               # 项目文档 (贡献指南、开发手册等)
├── frontend/           # Next.js 前端应用
├── infra/              # 基础设施配置 (Docker, K8s, Nginx)
├── scripts/            # 运维与开发辅助脚本
└── services/           # 后端微服务
    ├── ai-service/     # AI 处理服务
    ├── code-service/   # 代码执行与管理服务
    ├── user-service/   # 用户与权限服务
    └── ...             # 其他业务服务
```

## 快速开始

### 依赖环境
- Docker & Docker Compose
- Node.js 20+ (仅本地开发前端)
- Go 1.22+ / Python 3.12+ (仅本地独立运行服务)

### 启动项目
使用 Docker Compose 快速启动整套开发环境：

```bash
docker-compose -f infra/docker/docker-compose.yml up -d
```

启动后，访问 [http://localhost:3000](http://localhost:3000) 进入前端界面。

## 相关文档

- [开发指南](docs/DEVELOPMENT.md)
- [贡献指南](docs/CONTRIBUTING.md)
- [API 文档](docs/API.md) (待补充)

## 许可协议
[LICENSE](LICENSE)
