import logging
from datetime import datetime, date
from uuid import UUID

from fastapi import Request
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from starlette.middleware.base import BaseHTTPMiddleware

from config import get_config
from models import AIUsageStats

logger = logging.getLogger(__name__)


class RequestLoggerMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        response = await call_next(request)
        return response


async def update_usage_stats(
    db: AsyncSession,
    user_id: str,
    workspace_id: str,
    model: str,
    provider: str,
    capability: str,
    prompt_tokens: int,
    completion_tokens: int,
    total_tokens: int,
    latency_ms: int,
    status: str,
):
    try:
        today = date.today()

        result = await db.execute(
            select(AIUsageStats).where(
                AIUsageStats.user_id == UUID(user_id),
                AIUsageStats.workspace_id == UUID(workspace_id),
                AIUsageStats.model == model,
                AIUsageStats.capability == capability,
                AIUsageStats.date == today,
            )
        )
        stats = result.scalar_one_or_none()

        if stats:
            stats.request_count += 1
            stats.total_tokens += total_tokens
            stats.prompt_tokens += prompt_tokens
            stats.completion_tokens += completion_tokens
            stats.total_latency_ms += latency_ms
            stats.avg_latency_ms = stats.total_latency_ms // stats.request_count

            if status == "success":
                stats.success_count += 1
            elif status == "timeout":
                stats.timeout_count += 1
            else:
                stats.error_count += 1
        else:
            stats = AIUsageStats(
                user_id=UUID(user_id),
                workspace_id=UUID(workspace_id),
                model=model,
                provider=provider,
                capability=capability,
                date=today,
                request_count=1,
                success_count=1 if status == "success" else 0,
                error_count=1 if status not in ("success", "timeout") else 0,
                timeout_count=1 if status == "timeout" else 0,
                prompt_tokens=prompt_tokens,
                completion_tokens=completion_tokens,
                total_tokens=total_tokens,
                total_latency_ms=latency_ms,
                avg_latency_ms=latency_ms,
            )
            db.add(stats)

        await db.commit()
    except Exception as e:
        logger.error(f"Failed to update usage stats: {e}")