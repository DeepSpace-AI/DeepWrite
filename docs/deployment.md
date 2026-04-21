# DeepWrite 部署指南

## 架构概览

```
                    ┌─────────────────────────────────────┐
                    │            Nginx (80/443)           │
                    └───────────────┬─────────────────────┘
                                    │
        ┌───────────────────────────┼───────────────────────────┐
        │                           │                           │
        ▼                           ▼                           ▼
┌───────────────┐          ┌───────────────┐          ┌───────────────┐
│   Web (5173)  │          │ Admin (5174)  │          │ Gateway (8080)│
│   Vue 3 App   │          │   Vue 3 App   │          │   Go + Gin    │
└───────────────┘          └───────────────┘          └───────┬───────┘
                                                              │
                    ┌─────────────────────────────────────────┤
                    │                                         │
                    ▼                                         ▼
          ┌─────────────────┐                     ┌─────────────────┐
          │ AI Service(8002)│                     │ Worker (8000/1) │
          │   Python/FastAPI│                     │  Python/Celery  │
          └────────┬────────┘                     └────────┬────────┘
                   │                                       │
                   └──────────────────┬───────────────────┘
                                      │
                    ┌─────────────────┴─────────────────┐
                    │                                   │
                    ▼                                   ▼
          ┌─────────────────┐                 ┌─────────────────┐
          │   PostgreSQL    │                 │     Redis       │
          │      (5432)     │                 │     (6379)      │
          └─────────────────┘                 └─────────────────┘
```

## 快速开始

### 1. 服务器准备

```bash
# 在服务器上运行
curl -fsSL https://raw.githubusercontent.com/your-org/deepwrite/main/scripts/server-setup.sh | bash
```

或手动设置：

```bash
# 克隆仓库
git clone https://github.com/your-org/deepwrite.git /opt/deepwrite
cd /opt/deepwrite

# 复制环境配置
cp .env.example .env.production

# 编辑配置
nano .env.production
```

### 2. 配置 GitHub Secrets

在 GitHub 仓库的 Settings -> Secrets and variables -> Actions 中添加：

| Secret | 描述 |
|--------|------|
| `DEPLOY_HOST` | 服务器 IP 或域名 |
| `DEPLOY_USER` | SSH 用户名 (如 `root` 或 `ubuntu`) |
| `DEPLOY_KEY` | SSH 私钥 |
| `DEPLOY_PATH` | 部署路径 (默认 `/opt/deepwrite`) |
| `DB_PASSWORD` | 数据库密码 |
| `JWT_SECRET` | JWT 密钥 (至少32字符) |

### 3. 部署

**自动部署**（推送到 main 分支或创建 tag）：
```bash
git push origin main
# 或创建版本标签
git tag v1.0.0
git push origin v1.0.0
```

**手动部署**：
```bash
# 在服务器上
cd /opt/deepwrite
./scripts/deploy.sh
```

## 镜像说明

| 镜像 | 用途 | 端口 |
|------|------|------|
| `ghcr.io/your-org/deepwrite/web` | 前端应用 | 5173 |
| `ghcr.io/your-org/deepwrite/admin` | 管理后台 | 5174 |
| `ghcr.io/your-org/deepwrite/gateway` | API 网关 | 8080 |
| `ghcr.io/your-org/deepwrite/worker` | Celery Worker | 8000/8001 |
| `ghcr.io/your-org/deepwrite/ai` | AI 服务 | 8002 |

## 本地开发

```bash
# 启动开发环境
pnpm dev

# 或使用 Docker Compose
docker compose up -d
```

## 常用命令

```bash
# 查看服务状态
docker compose -f docker-compose.prod.yml ps

# 查看日志
docker compose -f docker-compose.prod.yml logs -f gateway

# 重启服务
docker compose -f docker-compose.prod.yml restart gateway

# 数据库备份
docker compose -f docker-compose.prod.yml exec -T postgres \
  pg_dump -U deepwrite deepwrite > backup.sql

# 数据库恢复
docker compose -f docker-compose.prod.yml exec -T postgres \
  psql -U deepwrite deepwrite < backup.sql
```

## 环境变量

| 变量 | 必需 | 描述 |
|------|------|------|
| `DB_USER` | 否 | 数据库用户名 (默认: deepwrite) |
| `DB_PASSWORD` | 是 | 数据库密码 |
| `DB_NAME` | 否 | 数据库名 (默认: deepwrite) |
| `JWT_SECRET` | 是 | JWT 签名密钥 |
| `IMAGE_TAG` | 否 | 镜像标签 (默认: latest) |
| `GITHUB_REPOSITORY` | 否 | GitHub 仓库名 |

## 故障排除

### 服务无法启动
```bash
# 检查日志
docker compose -f docker-compose.prod.yml logs

# 检查容器状态
docker compose -f docker-compose.prod.yml ps
```

### 数据库连接失败
```bash
# 检查数据库状态
docker compose -f docker-compose.prod.yml exec postgres pg_isready

# 测试连接
docker compose -f docker-compose.prod.yml exec gateway \
  nc -zv postgres 5432
```

### 清理和重启
```bash
# 停止所有服务
docker compose -f docker-compose.prod.yml down

# 清理镜像缓存
docker system prune -a

# 重新部署
./scripts/deploy.sh
```