# DeepWrite 开发计划

本文档定义 DeepWrite 项目的开发路线图，包括短期、中期和长期目标。

---

## 项目当前状态

| 模块 | 完成度 | 状态 |
|------|--------|------|
| apps/web | 90% | 主力开发 |
| apps/admin | 40% | 框架完成 |
| services/gateway | 80% | 主力开发 |
| services/ai | 50% | 开发中 |
| services/worker | 60% | 开发中 |

---

## 短期计划 (1-4 周) - 进行中

### S1: DEditor Phase B ✅ 已完成

- [x] 快捷键映射 (10个快捷键)
- [x] 浮动工具栏 (8个格式按钮)
- [x] Slash 菜单优化 (分组/最近使用)

### S2: Admin 后台 ✅ 已完成

- [x] 管理员仪表盘
- [x] 用户管理 (列表/启用禁用/搜索)
- [x] 工作区管理 (列表/搜索/成员统计)
- [x] 布局组件 (侧边栏/导航)

### S3: 测试体系 ✅ 已完成

- [x] Vitest 配置
- [x] 单元测试 (19 passing)

---

### S3: 测试体系搭建 ✅ 已完成

**目标**：建立基础测试覆盖

**任务清单**：

- [x] **前端测试**
  - [x] Vitest 配置 (vitest.config.ts + setup.ts)
  - [x] Store 单元测试 (cache.test.ts - 13 tests)
  - [x] 组件基础测试 (useKeyboardShortcuts.test.ts - 6 tests)
  - [x] API Mock (预留)

**运行测试**：

```bash
pnpm test              # 监听模式
pnpm test:run          # 单次运行
pnpm test:coverage     # 覆盖率报告
```

**覆盖率**: 19 passing tests

---

## 中期计划 (1-2 月)

### M1: DEditor Phase C - 文档高级能力 ✅ 已完成

**目标**：提供接近 Google Docs 的完整文档体验

**任务清单**：

- [x] **文档辅助**
  - [x] 大纲导航 (标题自动生成侧边栏)
  - [x] 字数统计 + 阅读时间
  - [x] 查找替换 (正则/大小写/全词)

- [x] **版本管理**
  - [x] 本地草稿自动保存
  - [x] 手动版本快照
  - [x] 版本历史浏览与回滚

- [x] **导入导出**
  - [x] Markdown 导出
  - [x] Markdown 导入 (预留)
  - [x] HTML 导出 (预留)
  - [ ] PDF 导出 (待实现)

---

### M2: Worker 任务链路 ✅ 已完成

**目标**：实现异步任务流水线

**已实现任务链路**：

| 任务 | 状态 | 说明 |
|------|------|------|
| 文档解析 | ✅ | parse_document, index_document, export_document |
| 批量 AI | ✅ | chat_completion, batch_completion, generate_embedding |
| 通知推送 | ✅ | send_email, send_welcome, document_shared, task_reminder |

**文件清单**：

```
worker/tasks/
├── document_tasks.py   # 文档处理
├── ai_tasks.py         # AI 推理
└── notify_tasks.py      # 通知推送
```

---

### M3: AI 能力集成

**目标**：将 AI 服务接入 DeepWrite 工作流

**任务清单**：

- [ ] **Gateway 集成**
  - [ ] AI Provider 配置 API
  - [ ] Chat Completion 路由
  - [ ] 流式响应转发

- [ ] **前端 AI 能力**
  - [ ] AI 写作助手 (侧边栏)
  - [ ] 智能续写
  - [ ] 内容润色
  - [ ] 翻译支持

- [ ] **模型配置管理**
  - [ ] Provider 添加/编辑
  - [ ] 模型切换
  - [ ] 配额管理

---

### M4: CI/CD 流水线 ✅ 已完成

**目标**：自动化构建、测试、部署

**已实现**：

- [x] **GitHub Actions**
  - [x] PR 检查 (lint + typecheck + test)
  - [x] 主分支构建
  - [x] 构建产物上传

- [x] **Docker 支持**
  - [x] Web Dockerfile
  - [x] Gateway Dockerfile
  - [x] Worker Dockerfile
  - [x] docker-compose.yml

**文件清单**：

```
.github/workflows/ci.yml    # CI 工作流
.docker/
  ├── web.Dockerfile
  ├── gateway.Dockerfile
  └── worker.Dockerfile
docker-compose.yml          # 本地开发编排
```

---

## 长期规划 (2-3 月)

### L1: 全文搜索

- [ ] Elasticsearch 集成
- [ ] 文档索引
- [ ] 分词器配置
- [ ] 搜索 API

### L2: 协作文档增强

- [ ] 实时评论
- [ ] @提及通知
- [ ] 文档模板
- [ ] 权限细化

### L3: 移动端支持

- [ ] PWA 配置
- [ ] 响应式优化
- [ ] 离线支持

### L4: 高可用部署

- [ ] Kubernetes 配置
- [ ] 水平扩展
- [ ] 监控告警

---

## 技术债务

### 优先级 P0 (必须处理)

- [ ] 许可证声明 (LICENSE 文件)
- [ ] 测试覆盖率 > 60%
- [ ] API 文档完善

### 优先级 P1 (尽快处理)

- [ ] 环境配置文档化
- [ ] 错误码统一
- [ ] 日志规范

### 优先级 P2 (计划处理)

- [ ] 性能监控接入
- [ ] 安全审计
- [ ] 依赖更新流程

---

## 里程碑汇总

| 阶段 | 时间 | 里程碑 |
|------|------|--------|
| S1 | Week 1-2 | DEditor Phase B 完成 |
| S2 | Week 2-3 | Admin 框架完成 |
| S3 | Week 3-4 | 测试体系搭建 |
| M1 | Week 5-7 | DEditor Phase C 完成 |
| M2 | Week 5-9 | Worker 任务链路 |
| M3 | Week 7-10 | AI 能力集成 |
| M4 | Week 8-10 | CI/CD 流水线 |

---

## 开发规范

### 分支命名

```
feature/xxx          # 新功能
bugfix/xxx           # Bug 修复
hotfix/xxx           # 紧急修复
refactor/xxx         # 重构
docs/xxx             # 文档更新
```

### Commit 规范

```
<type>(<scope>): <subject>

Types: feat, fix, docs, style, refactor, test, chore
```

### PR 规范

- 标题清晰描述改动
- 包含变更说明
- 关联相关 Issue
- 通过 CI 检查

---

## 团队协作

### 代码审查清单

- [ ] 功能符合需求
- [ ] 代码规范遵循
- [ ] 测试覆盖充分
- [ ] 无敏感信息泄露
- [ ] 文档更新

### 发布流程

1. 功能分支开发
2. PR 审查
3. 合并到主分支
4. 标签版本
5. 构建发布
6. 部署验证
