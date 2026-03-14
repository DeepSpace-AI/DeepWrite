# DeepWrite Worker 端开发方案（v1）

## 1. 目标与范围

### 1.1 目标

建设 Worker 服务，作为 DeepWrite 的 AI 服务层与任务执行层，满足以下目标：

- 通过 gRPC 与 Gateway 通讯，提供任务提交、取消、查询、流式结果订阅、健康检查能力。
- 以 Celery 作为异步任务队列执行引擎。
- 与 Gateway 共享 PostgreSQL 与 Redis。
- 支持按 workspace 和 user 维度的基础并发限制与配额控制。
- 提供可水平扩展、可追踪、可重试、可幂等的任务执行体系。

### 1.2 一期任务类型

- LLM 文本生成/改写/总结
- 文档结构化解析与清洗
- Embedding 与检索索引
- 异步通知（含邮件触达）
- 批量导入导出
- 定时任务（仅预留任务类型与队列，不启用 Celery Beat 进程）

### 1.3 非目标（v1 不做）

- 直接由前端连接 Worker；前端仍只连接 Gateway。
- 供应商强绑定；先做 Provider 抽象，不锁定 OpenAI/Azure/Anthropic/本地模型。
- k8s 首发；以本地裸进程开发为主，容器化仅提供附录参考。

## 2. 架构总览

### 2.1 组件

- Gateway（Go）
- Worker API（Python，gRPC Server）
- Celery Worker（Python，多进程 prefork）
- Redis（Broker + Pub/Sub）
- PostgreSQL（任务权威状态 + 业务数据）

### 2.2 主链路

1. Gateway 接收前端请求，完成鉴权与业务校验。
2. Gateway 写入任务表（状态 pending），生成 idempotency key。
3. Gateway 通过 gRPC SubmitTask 向 Worker API 提交。
4. Worker API 将任务入 Celery 指定队列，返回 task_id/queue 信息。
5. Celery Worker 消费并执行，按阶段回写 PostgreSQL（queued/running/succeeded/failed/canceled）。
6. Worker 在进度和结果事件发生时，调用 Gateway 内部 HTTP 回调接口。
7. Worker 同时发布 Redis Pub/Sub 事件，Gateway 订阅后推送到前端（SSE/WebSocket）。

### 2.3 流式输出路径

- AI token 流由 Worker 侧产生。
- Worker 将 chunk 通过 HTTP 回调发送到 Gateway。
- Gateway 再转发给前端。
- 前端只感知 Gateway 的统一流式接口。

## 3. 技术选型与原则

### 3.1 Python 技术栈建议

- gRPC: grpcio + grpcio-tools
- 配置与模型: pydantic-settings
- ORM: SQLAlchemy 2.x（async engine 可选）
- 任务队列: Celery 5.x
- Redis 客户端: redis-py
- 日志: structlog 或标准 logging + JSON formatter
- 可观测: OpenTelemetry（分阶段接入）

### 3.2 关键原则

- PostgreSQL 是任务状态与结果权威来源。
- Redis 仅承担队列与短时消息分发，不作为最终真相。
- 所有任务必须具备幂等键，重试不可产生重复副作用。
- 状态机单向推进，禁止回退。

## 4. 状态机与执行语义

### 4.1 状态流转

pending -> queued -> running -> succeeded/failed/canceled

说明：

- pending: Gateway 落库成功，尚未被 Worker 接收。
- queued: Worker 接收并成功投递 Celery。
- running: Celery 开始执行。
- succeeded: 执行成功并持久化结果。
- failed: 执行失败且超过重试上限。
- canceled: 用户或系统取消，执行中断。

### 4.2 重试策略

- 指数退避：5s, 15s, 45s, 135s, 405s
- 最大重试次数：5
- 超限后写 failed，并记录 error_code 与 error_message
- 可重试错误分类示例：
  - 上游模型 429/5xx
  - 网络超时
  - 暂时性存储不可用
- 不可重试错误分类示例：
  - 参数校验失败
  - 权限不足
  - 资源不存在

### 4.3 幂等策略

- 幂等键建议：workspace_id + task_type + payload_hash + client_request_id
- 唯一索引：idempotency_key
- 重复提交命中时直接返回既有 task_id 和当前状态

## 5. gRPC 协议草案

建议 proto 文件位置：services/worker/proto/worker/v1/worker.proto

```proto
syntax = "proto3";

package deepwrite.worker.v1;

option go_package = "github.com/deepwrite/services/gateway/pkg/workerpb;workerpb";

message TaskPayload {
  string task_type = 1;
  string content_type = 2;
  bytes body = 3;
}

message SubmitTaskRequest {
  string request_id = 1;
  string idempotency_key = 2;
  string workspace_id = 3;
  string user_id = 4;
  int32 priority = 5;
  string callback_url = 6;
  map<string, string> metadata = 7;
  TaskPayload payload = 8;
}

message SubmitTaskResponse {
  string task_id = 1;
  string celery_task_id = 2;
  string queue = 3;
  string status = 4;
  int64 accepted_at_unix_ms = 5;
}

message CancelTaskRequest {
  string task_id = 1;
  string reason = 2;
}

message CancelTaskResponse {
  string task_id = 1;
  string status = 2;
}

message GetTaskRequest {
  string task_id = 1;
}

message TaskInfo {
  string task_id = 1;
  string status = 2;
  string task_type = 3;
  string workspace_id = 4;
  string user_id = 5;
  int32 progress = 6;
  string error_code = 7;
  string error_message = 8;
  int64 created_at_unix_ms = 9;
  int64 updated_at_unix_ms = 10;
}

message GetTaskResponse {
  TaskInfo task = 1;
}

message StreamTaskRequest {
  string task_id = 1;
}

message StreamEvent {
  string task_id = 1;
  string event_type = 2; // status | chunk | usage | done | error
  int32 progress = 3;
  bytes chunk = 4;
  map<string, string> metadata = 5;
  int64 ts_unix_ms = 6;
}

message HealthCheckRequest {}

message HealthCheckResponse {
  string status = 1;
  string version = 2;
  map<string, string> dependencies = 3;
}

message CapabilityRequest {}

message CapabilityResponse {
  repeated string task_types = 1;
  repeated string model_providers = 2;
  int32 max_concurrency = 3;
}

service WorkerService {
  rpc SubmitTask(SubmitTaskRequest) returns (SubmitTaskResponse);
  rpc CancelTask(CancelTaskRequest) returns (CancelTaskResponse);
  rpc GetTask(GetTaskRequest) returns (GetTaskResponse);
  rpc StreamTask(StreamTaskRequest) returns (stream StreamEvent);
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
  rpc GetCapabilities(CapabilityRequest) returns (CapabilityResponse);
}
```

### 5.1 鉴权建议（内部调用）

- gRPC Metadata 传递静态 token，例如 x-internal-token。
- Worker 校验 token，不通过直接返回 unauthenticated。
- 生产环境可平滑升级到 mTLS（保留接口扩展点）。

## 6. Celery 设计

### 6.1 Broker 与 Backend

- Broker: Redis
- Result backend: Redis 可用于执行态追踪，但最终状态以 PostgreSQL 为准

### 6.2 队列划分

- ai.generate
- doc.parse
- embedding.index
- notify.dispatch
- batch.import_export
- schedule.periodic

### 6.3 路由策略

- 按 task_type 映射固定队列。
- 优先级字段写入 Celery headers。
- 对 ai.generate 配置单独并发上限，避免挤占轻任务。

### 6.4 worker 并发

- 并发池：prefork
- 开发默认并发：CPU 核数或 4（二者取小）
- 生产建议：按队列启动多个 worker 进程并水平扩容

### 6.5 取消语义

- CancelTask 调用后：
  - 对 queued 任务执行 revoke(terminate=false)
  - 对 running 任务发送终止信号并更新 canceled
- 任务代码需周期检查 cancel flag，保证长任务可中断

## 7. 数据库表草案（与 Gateway 共库）

建议新增以下表，命名前缀 dw_task_，避免与现有业务表混淆。

### 7.1 dw_task_jobs

用途：任务主表（权威状态）

字段建议：

- id: uuid, pk
- workspace_id: uuid, not null, index
- user_id: uuid, not null, index
- task_type: varchar(64), not null, index
- status: varchar(32), not null, index
- priority: int, not null, default 50
- idempotency_key: varchar(128), not null, unique
- celery_task_id: varchar(128), null, index
- payload_json: jsonb, not null
- result_json: jsonb, null
- progress: int, not null, default 0
- error_code: varchar(64), null
- error_message: text, null
- retry_count: int, not null, default 0
- max_retries: int, not null, default 5
- started_at: timestamptz, null
- finished_at: timestamptz, null
- created_at: timestamptz, not null
- updated_at: timestamptz, not null

约束建议：

- check(progress >= 0 and progress <= 100)
- check(status in ('pending','queued','running','succeeded','failed','canceled'))

### 7.2 dw_task_events

用途：记录状态、chunk、错误、审计事件

字段建议：

- id: bigserial, pk
- task_id: uuid, not null, index
- event_type: varchar(32), not null
- seq: bigint, not null
- payload_json: jsonb, not null
- created_at: timestamptz, not null

索引建议：

- unique(task_id, seq)
- index(task_id, created_at)

### 7.3 dw_task_quotas

用途：workspace/user 级配额与并发限制

字段建议：

- id: uuid, pk
- scope_type: varchar(16), not null // workspace | user
- scope_id: uuid, not null
- max_concurrency: int, not null
- daily_token_budget: bigint, null
- daily_task_budget: int, null
- enabled: bool, not null, default true
- created_at: timestamptz, not null
- updated_at: timestamptz, not null

索引建议：

- unique(scope_type, scope_id)

### 7.4 迁移落地建议

- Gateway 使用 GORM AutoMigrate 仅管理现有业务表。
- Worker 任务表建议使用 Alembic 显式迁移，避免跨语言 ORM 竞争迁移。
- 在 docs 中维护 SQL 变更清单，严格版本化。

## 8. Redis 键空间约定

统一前缀建议：dw:worker:

- dw:worker:lock:{task_id} 任务幂等锁
- dw:worker:quota:{scope_type}:{scope_id} 并发令牌
- dw:worker:stream:{task_id} 流式事件短时缓存
- dw:worker:pubsub:task-events 事件发布通道

注意：

- 键需设置 TTL，避免泄漏。
- 与 gateway: 前缀隔离，防止冲突。

## 9. Gateway 改造清单

### 9.1 配置项

新增配置段建议：

- WORKER_GRPC:
  - HOST
  - PORT
  - TIMEOUT_MS
  - INTERNAL_TOKEN
  - ENABLE_TLS
- WORKER_CALLBACK:
  - ENABLED
  - VERIFY_TOKEN
  - ALLOWED_IPS

### 9.2 代码改造点

- 新增 Worker gRPC client（连接池 + 超时 + 重试）。
- 新增任务 API：创建、查询、取消、结果订阅。
- 新增内部回调 API：
  - POST /internal/worker/task-events
- 新增 Redis 订阅器：监听 dw:worker:pubsub:task-events 并转推前端。
- 新增鉴权中间件：校验 worker 回调 token。

### 9.3 API 行为约束

- 对前端暴露统一任务状态字段，不泄漏 Celery 术语。
- 对流式 chunk 提供序号与终止事件，保证前端可重组。

## 10. Worker 工程目录建议

建议目录：services/worker

```text
services/worker/
  pyproject.toml
  README.md
  .env.example
  main.py                      # gRPC server 入口
  celery_app.py                # Celery app 初始化
  proto/
    worker/v1/worker.proto
  generated/
    worker_pb2.py
    worker_pb2_grpc.py
  app/
    config.py
    logging.py
    db.py
    redis_client.py
    security.py
    grpc_server/
      service.py
      interceptors.py
    callbacks/
      gateway_client.py
    tasks/
      base.py
      ai_generate.py
      doc_parse.py
      embedding_index.py
      notify_dispatch.py
      batch_import_export.py
      schedule_periodic.py
    providers/
      base.py
      factory.py
      openai_compatible.py
    repositories/
      task_job_repo.py
      task_event_repo.py
      quota_repo.py
    scheduler/
      router.py
      quota_guard.py
      idempotency.py
  migrations/
    versions/
  tests/
    unit/
    integration/
```

## 11. 配置规范（worker）

建议 .env 字段：

- WORKER_ENV=development
- WORKER_GRPC_HOST=0.0.0.0
- WORKER_GRPC_PORT=50051
- WORKER_INTERNAL_TOKEN=<secret>
- DATABASE_DSN=postgresql+psycopg://...
- REDIS_URL=redis://localhost:6379/0
- CELERY_BROKER_URL=redis://localhost:6379/1
- CELERY_RESULT_BACKEND=redis://localhost:6379/2
- GATEWAY_CALLBACK_URL=http://localhost:8080/internal/worker/task-events
- GATEWAY_CALLBACK_TOKEN=<secret>
- MAX_WORKSPACE_CONCURRENCY=8
- MAX_USER_CONCURRENCY=3
- STREAM_CHUNK_MAX_BYTES=4096

## 12. 可观测与审计

### 12.1 日志

- 每个任务统一 trace_id、task_id、workspace_id、user_id。
- 结构化日志 JSON，便于检索。

### 12.2 指标

- task_submit_total{task_type}
- task_running_gauge{task_type}
- task_latency_seconds{task_type,status}
- task_retry_total{task_type}
- callback_fail_total

### 12.3 审计

- 关键状态变更必须写 dw_task_events。
- 错误堆栈写内部日志，外部仅返回 error_code 与安全摘要。

## 13. 测试策略

### 13.1 单元测试

- 幂等命中逻辑
- 状态机推进合法性
- 重试分类与退避计算
- 配额控制策略

### 13.2 集成测试

- gRPC SubmitTask -> Celery 消费 -> DB 状态推进
- Worker 回调 Gateway 成功/失败重试
- Redis pub/sub 到 Gateway 转推链路

### 13.3 端到端冒烟

- 前端提交生成任务
- 实时收到进度/文本 chunk
- 最终结果落库可查询

## 14. 里程碑与验收

### M1：协议与骨架（1 周）

交付：

- proto 定稿与双端代码生成
- worker 工程骨架
- gateway gRPC client 骨架

验收：

- HealthCheck 与 GetCapabilities 可联通

### M2：核心任务链路（1-2 周）

交付：

- dw_task_jobs 与 dw_task_events 迁移
- SubmitTask/GetTask/CancelTask 可用
- Celery 队列路由与重试策略生效

验收：

- pending -> queued -> running -> succeeded 全链路跑通
- 失败任务可按策略重试并正确终态

### M3：流式与回调（1 周）

交付：

- Worker 回调 Gateway 内部接口
- Redis pub/sub 转推链路
- 前端可消费 Gateway 流式输出

验收：

- 流式 chunk 顺序一致
- 回调失败可重试，最终一致

### M4：配额与稳定性（1 周）

交付：

- workspace/user 并发配额
- 压测脚本与基线报告
- 可观测指标与告警规则

验收：

- 过载下系统可降级，不出现状态错乱或数据重复

## 15. 风险与规避

- 跨语言迁移冲突：任务表由 Worker 专管迁移，避免 GORM/Alembic 双写。
- 流式乱序：chunk 必带 seq，Gateway 做有序转发。
- 取消不彻底：任务函数中必须插入 cancel checkpoint。
- 共享 Redis 污染：前缀隔离 + 独立 DB 编号。
- 模型供应商切换成本：统一 Provider 抽象层。

## 16. 开发启动清单（按顺序）

1. 定稿 proto 与错误码字典。
2. 在 worker 建立 gRPC server + Celery app + 配置加载。
3. 落地 dw_task_jobs 和 dw_task_events 迁移。
4. 实现 SubmitTask/GetTask/CancelTask。
5. 接入 ai.generate 第一条任务并跑通全链路。
6. 增加 HTTP 回调与 Redis pub/sub 转推。
7. 增加 quota_guard 与幂等锁。
8. 完成单测、集成测试、冒烟测试。

## 17. 附录：本地开发进程建议

- 进程 1：Gateway
- 进程 2：Worker gRPC API
- 进程 3：Celery worker（prefork）
- 基础依赖：PostgreSQL、Redis

示例命令（占位）：

```bash
# worker gRPC API
uv run python main.py

# celery worker
uv run celery -A celery_app.app worker -Q ai.generate,doc.parse,embedding.index,notify.dispatch,batch.import_export,schedule.periodic -l info --pool=prefork
```

---

该方案已按当前确认条件编写，可直接作为 v1 实施基线。若后续决定引入 Celery Beat、mTLS 或 k8s，可在本文件上增量扩展为 v1.1。
