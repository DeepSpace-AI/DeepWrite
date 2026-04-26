# DeepWrite 开发指南

本指南将帮助你搭建 DeepWrite 的本地开发环境并开始贡献代码。

## 环境要求

在开始之前，请确保你的本地机器已安装以下软件：

- **Go**: 1.22+
- **Python**: 3.12+
- **Node.js**: 20+ (推荐使用 pnpm)
- **Docker & Docker Compose**: 最新稳定版
- **Git**: 必须

## 本地开发环境搭建

### 1. 克隆项目
```bash
git clone https://github.com/your-org/deepwrite.git
cd deepwrite
```

### 2. 启动基础设施
使用 Docker Compose 启动数据库、缓存、消息队列等中间件：
```bash
docker-compose -f infra/docker/docker-compose.yml up -d postgres mongodb redis s3
```

### 3. 前端环境设置
```bash
cd frontend
pnpm install
pnpm dev
```
前端应用将运行在 [http://localhost:3000](http://localhost:3000)。

### 4. 后端服务启动
每个服务在 `services/` 目录下都有自己的独立环境。

#### Go 服务 (例如 project-service):
```bash
cd services/project-service
go mod download
go run main.go
```

#### Python 服务 (例如 ai-service):
```bash
cd services/ai-service
python -m venv venv
source venv/bin/activate  # Windows 使用 venv\Scripts\activate
pip install -r requirements.txt
python main.py
```

## 测试运行

### 前端测试
```bash
cd frontend
pnpm test
```

### 后端测试
- **Go**: `go test ./...`
- **Python**: `pytest`

## 常见问题 FAQ

**Q: 启动服务时提示数据库连接失败？**
A: 请检查 `docker ps` 确保 PostgreSQL 等基础镜像已成功运行，并检查各服务的 `.env` 配置文件（可参考 `.env.example`）。

**Q: 如何添加新的微服务？**
A: 请参考 `services/` 目录下的现有结构，确保包含 `Dockerfile` 以及标准化的健康检查接口。

**Q: 依赖项安装过慢？**
A: 建议配置国内镜像源（如 npm 淘宝镜像、Go goproxy 等）。
