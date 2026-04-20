from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

from config.config import get_config
from worker.tasks import ping, extract_pdf_metadata, extract_and_lookup_doi

cfg = get_config()
app = FastAPI(title=cfg.app_name)


class ExtractPdfRequest(BaseModel):
    file_id: str
    reference_id: str | None = None
    content: str  # Base64 encoded PDF content


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


@app.post("/tasks/pdf/extract")
async def enqueue_extract_pdf(request: ExtractPdfRequest):
    """
    提取 PDF 元数据任务。

    仅从 PDF 中提取 DOI、可能的年份等信息。
    """
    try:
        result = extract_pdf_metadata.delay(request.file_id, request.content)
    except Exception as exc:
        raise HTTPException(status_code=503, detail=f"任务入队失败: {exc}") from exc
    return {
        "task_id": result.id,
        "task_name": "worker.pdf.extract_metadata",
        "queue": cfg.celery_default_queue,
    }


@app.post("/tasks/pdf/extract-and-lookup")
async def enqueue_extract_and_lookup(request: ExtractPdfRequest):
    """
    提取 PDF 元数据并查询 Crossref。

    从 PDF 提取 DOI，然后查询 Crossref 获取完整元数据。
    如果提供了 reference_id，完成后会回调更新文献信息。
    """
    try:
        result = extract_and_lookup_doi.delay(
            request.file_id,
            request.content,
            reference_id=request.reference_id,
        )
    except Exception as exc:
        raise HTTPException(status_code=503, detail=f"任务入队失败: {exc}") from exc
    return {
        "task_id": result.id,
        "task_name": "worker.pdf.extract_and_lookup",
        "queue": cfg.celery_default_queue,
    }


@app.get("/tasks/{task_id}/status")
async def get_task_status(task_id: str):
    """
    查询任务状态和结果。
    """
    from celery.result import AsyncResult
    from worker.celery_app import celery_app

    task_result = AsyncResult(task_id, app=celery_app)

    response = {
        "task_id": task_id,
        "status": task_result.status,
        "ready": task_result.ready(),
    }

    if task_result.ready():
        if task_result.successful():
            response["result"] = task_result.result
        else:
            response["error"] = str(task_result.result)

    return response
