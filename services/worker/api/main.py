from fastapi import FastAPI, HTTPException

from config.config import get_config
from worker.tasks import ping

cfg = get_config()
app = FastAPI(title=cfg.app_name)

@app.get("/")
async def root():
    return {"service": cfg.app_name, "env": cfg.app_env, "status": "ok"}


@app.post("/tasks/ping")
async def enqueue_ping(payload: str = "pong"):
    try:
        result = ping.delay(payload)
    except Exception as exc:
        raise HTTPException(status_code=503, detail=f"任务入队失败: {exc}") from exc
    return {"task_id": result.id, "task_name": "worker.ping", "queue": cfg.celery_default_queue}
