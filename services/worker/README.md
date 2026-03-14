# DeepWrite Worker（FastAPI + Celery）

本文档说明当前 Worker 中 Celery 的接入方式、任务是如何被注册和执行的，以及后续扩展任务的推荐流程。

## 1. 当前目录结构

```text
services/worker/
  api/
    main.py
  config/
    config.py
  worker/
    __init__.py
    celery_app.py
    tasks/
      __init__.py
      ping_tasks.py
      math_tasks.py
  config.yaml
  main.py
  pyproject.toml
```

关键文件说明：

- `worker/celery_app.py`：创建 Celery 应用，读取 broker/backend/queue 配置。
- `worker/tasks/`：多文件任务目录，按领域拆分任务实现。
- `api/main.py`：示例接口里演示如何把任务投递到队列（`ping.delay(...)`）。
- `main.py`：一键启动入口，支持同时启动 API + Celery。

## 2. Celery 是如何“添加任务”的

### 2.1 Celery 应用初始化

在 `worker/celery_app.py` 中，`celery_app = Celery(...)` 会完成以下事情：

1. 绑定消息中间件（`broker`）和结果后端（`backend`）。
2. 指定默认队列 `task_default_queue`。
3. 通过 `include=["worker.tasks"]` + `autodiscover_tasks(...)` 加载任务包。
4. `worker/tasks/__init__.py` 会自动扫描并导入 `worker/tasks/*.py`，完成任务注册。

### 2.2 任务注册

在 `worker/tasks/` 目录下，任何被装饰器包裹的函数都会被注册为任务：

```python
@celery_app.task(name="worker.add")
def add(a: int, b: int) -> int:
    return a + b
```

- `name` 是任务在 Celery 中的唯一标识，建议统一使用 `worker.xxx` 命名空间。
- 任务函数参数和返回值应尽量使用可序列化数据（基础类型、字典、列表）。

### 2.3 任务投递

任意业务代码（比如 API 层）可通过 `delay/apply_async` 投递：

```python
result = ping.delay("hello")
```

- `result.id` 是任务 ID，可用于追踪执行状态。
- 若需要自定义路由、延时执行、重试策略，推荐使用 `apply_async(...)`。

## 3. 如何扩展新任务（推荐流程）

### 步骤 1：在 `worker/tasks/` 下新增一个任务文件

示例：新建 `worker/tasks/document_tasks.py`

```python
@celery_app.task(name="worker.document.parse")
def parse_document(file_id: str) -> dict[str, str]:
    return {"file_id": file_id, "status": "parsed"}
```

建议：

- 任务名遵循 `worker.<domain>.<action>`，例如 `worker.ai.generate`。
- 任务逻辑尽量无副作用、可重试、幂等。
- 文件名建议使用 `<domain>_tasks.py`，便于团队快速检索。

### 步骤 2：在业务入口投递任务

例如在 API 中：

```python
task = parse_document.delay(file_id)
return {"task_id": task.id}
```

如果需要指定队列和参数：

```python
task = parse_document.apply_async(
    args=[file_id],
    queue="worker.document",
    countdown=3,
)
```

### 步骤 3：按任务类型拆分队列（可选但推荐）

当任务种类变多后，建议按领域拆队列：

- `worker.default`
- `worker.ai`
- `worker.document`
- `worker.notify`

并按队列启动不同 worker 进程，提高隔离性和吞吐稳定性。

## 4. 启动方式

### 4.1 一键启动 API + Celery

```bash
uv run python main.py
```

等价于 `--mode all`。

### 4.2 仅启动 API

```bash
uv run python main.py --mode api
```

### 4.3 仅启动 Celery

```bash
uv run python main.py --mode celery
```

## 5. 配置项说明（Celery 相关）

`config.yaml`：

```yaml
CELERY:
  BROKER_URL: redis://localhost:6379/1
  RESULT_BACKEND: redis://localhost:6379/2
  DEFAULT_QUEUE: worker.default
```

环境变量可覆盖：

- `CELERY_BROKER_URL`
- `CELERY_RESULT_BACKEND`
- `CELERY_DEFAULT_QUEUE`

## 6. 后续扩展建议

建议你在下个阶段优先增加以下能力：

1. 任务重试策略（`autoretry_for`、`retry_backoff`、`max_retries`）。
2. 任务超时控制（软超时/硬超时）。
3. 任务状态与结果持久化（配合业务表记录）。
4. 幂等键与防重入（避免重复消费造成脏数据）。
5. 按队列拆 worker 实例并设置并发参数。

---
