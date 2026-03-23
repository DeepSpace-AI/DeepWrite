import logging
from datetime import datetime
from typing import Any

from worker.celery_app import celery_app

logger = logging.getLogger(__name__)


@celery_app.task(
    name="worker.ai.chat_completion",
    bind=True,
    max_retries=3,
    default_retry_delay=30,
    autoretry_for=(Exception,),
    retry_backoff=True,
)
def chat_completion(
    self,
    prompt: str,
    model: str = "gpt-3.5-turbo",
    options: dict[str, Any] | None = None,
) -> dict[str, Any]:
    """
    执行单次聊天完成请求。

    Args:
        prompt: 用户提示
        model: 模型名称
        options: 额外选项 (temperature, max_tokens 等)

    Returns:
        AI 响应结果
    """
    logger.info(f"[chat_completion] model={model}, prompt_len={len(prompt)}")

    try:
        result = {
            "status": "completed",
            "model": model,
            "prompt_tokens": len(prompt.split()),
            "completion_tokens": 50,
            "response": f"[AI Response to: {prompt[:50]}...]",
            "completed_at": datetime.now().isoformat(),
        }
        logger.info(f"[chat_completion] success model={model}")
        return result
    except Exception as exc:
        logger.error(f"[chat_completion] failed: {exc}")
        raise


@celery_app.task(
    name="worker.ai.batch_completion",
    bind=True,
    max_retries=2,
    default_retry_delay=60,
)
def batch_completion(
    self,
    prompts: list[str],
    model: str = "gpt-3.5-turbo",
    options: dict[str, Any] | None = None,
) -> dict[str, Any]:
    """
    批量执行聊天完成请求。

    Args:
        prompts: 提示列表
        model: 模型名称
        options: 额外选项

    Returns:
        批量处理结果
    """
    logger.info(f"[batch_completion] model={model}, count={len(prompts)}")

    try:
        results = []
        for i, prompt in enumerate(prompts):
            results.append({
                "index": i,
                "prompt": prompt[:100],
                "response": f"[Batch response {i}]",
                "status": "completed",
            })

        result = {
            "status": "completed",
            "model": model,
            "total": len(prompts),
            "successful": len(prompts),
            "results": results,
            "completed_at": datetime.now().isoformat(),
        }
        logger.info(f"[batch_completion] success total={len(prompts)}")
        return result
    except Exception as exc:
        logger.error(f"[batch_completion] failed: {exc}")
        raise


@celery_app.task(
    name="worker.ai.embedding",
    bind=True,
    max_retries=3,
    default_retry_delay=30,
)
def generate_embedding(
    self,
    text: str,
    model: str = "text-embedding-ada-002",
) -> dict[str, Any]:
    """
    生成文本嵌入向量。

    Args:
        text: 输入文本
        model: 嵌入模型

    Returns:
        嵌入结果
    """
    logger.info(f"[generate_embedding] model={model}, text_len={len(text)}")

    try:
        result = {
            "status": "completed",
            "model": model,
            "embedding": [0.1] * 1536,
            "tokens": len(text.split()),
            "completed_at": datetime.now().isoformat(),
        }
        logger.info(f"[generate_embedding] success model={model}")
        return result
    except Exception as exc:
        logger.error(f"[generate_embedding] failed: {exc}")
        raise
