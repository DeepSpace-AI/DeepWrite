import logging
from datetime import date, datetime, timedelta
from typing import Any
from uuid import UUID

from fastapi import Depends, HTTPException, Request
from pydantic import BaseModel
from sqlalchemy import func, select
from sqlalchemy.ext.asyncio import AsyncSession

from models import AIRequestLog, AIUsageStats

logger = logging.getLogger(__name__)


async def get_db(request: Request) -> AsyncSession:
    return request.state.db


async def get_workspace_id(request: Request) -> str:
    workspace_id = request.headers.get("X-Workspace-Id", "")
    if not workspace_id:
        raise HTTPException(status_code=400, detail="X-Workspace-Id header is required")
    return workspace_id


async def get_user_id(request: Request) -> str:
    user_id = request.headers.get("X-User-Id", "")
    if not user_id:
        raise HTTPException(status_code=400, detail="X-User-Id header is required")
    return user_id


class StatsResponse(BaseModel):
    total_requests: int
    success_requests: int
    failed_requests: int
    total_tokens: int
    prompt_tokens: int
    completion_tokens: int
    avg_latency_ms: float
    by_model: list[dict]
    by_capability: list[dict]


async def get_workspace_stats(
    request: Request,
    db: AsyncSession = Depends(get_db),
    workspace_id: str = Depends(get_workspace_id),
    start_date: str | None = None,
    end_date: str | None = None,
    model: str | None = None,
) -> StatsResponse:
    try:
        start = datetime.strptime(start_date, "%Y-%m-%d") if start_date else datetime.utcnow() - timedelta(days=30)
        end = datetime.strptime(end_date, "%Y-%m-%d") + timedelta(days=1) if end_date else datetime.utcnow()
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid date format, use YYYY-MM-DD")

    query = select(AIUsageStats).where(
        AIUsageStats.workspace_id == UUID(workspace_id),
        AIUsageStats.date >= start.date(),
        AIUsageStats.date <= end.date(),
    )
    if model:
        query = query.where(AIUsageStats.model == model)

    result = await db.execute(query)
    stats = result.scalars().all()

    total_requests = sum(s.request_count for s in stats)
    success_requests = sum(s.success_count for s in stats)
    failed_requests = sum(s.error_count + s.timeout_count for s in stats)
    total_tokens = sum(s.total_tokens for s in stats)
    prompt_tokens = sum(s.prompt_tokens for s in stats)
    completion_tokens = sum(s.completion_tokens for s in stats)
    total_latency = sum(s.total_latency_ms for s in stats)

    avg_latency = total_latency / success_requests if success_requests > 0 else 0

    model_stats = {}
    for s in stats:
        if s.model not in model_stats:
            model_stats[s.model] = {"model": s.model, "requests": 0, "tokens": 0}
        model_stats[s.model]["requests"] += s.request_count
        model_stats[s.model]["tokens"] += s.total_tokens

    capability_stats = {}
    for s in stats:
        if s.capability not in capability_stats:
            capability_stats[s.capability] = {"capability": s.capability, "requests": 0, "tokens": 0}
        capability_stats[s.capability]["requests"] += s.request_count
        capability_stats[s.capability]["tokens"] += s.total_tokens

    return StatsResponse(
        total_requests=total_requests,
        success_requests=success_requests,
        failed_requests=failed_requests,
        total_tokens=total_tokens,
        prompt_tokens=prompt_tokens,
        completion_tokens=completion_tokens,
        avg_latency_ms=avg_latency,
        by_model=list(model_stats.values()),
        by_capability=list(capability_stats.values()),
    )


async def get_user_stats(
    request: Request,
    db: AsyncSession = Depends(get_db),
    user_id: str = Depends(get_user_id),
    start_date: str | None = None,
    end_date: str | None = None,
) -> StatsResponse:
    try:
        start = datetime.strptime(start_date, "%Y-%m-%d") if start_date else datetime.utcnow() - timedelta(days=30)
        end = datetime.strptime(end_date, "%Y-%m-%d") + timedelta(days=1) if end_date else datetime.utcnow()
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid date format, use YYYY-MM-DD")

    query = select(AIUsageStats).where(
        AIUsageStats.user_id == UUID(user_id),
        AIUsageStats.date >= start.date(),
        AIUsageStats.date <= end.date(),
    )

    result = await db.execute(query)
    stats = result.scalars().all()

    total_requests = sum(s.request_count for s in stats)
    success_requests = sum(s.success_count for s in stats)
    failed_requests = sum(s.error_count + s.timeout_count for s in stats)
    total_tokens = sum(s.total_tokens for s in stats)
    prompt_tokens = sum(s.prompt_tokens for s in stats)
    completion_tokens = sum(s.completion_tokens for s in stats)
    total_latency = sum(s.total_latency_ms for s in stats)

    avg_latency = total_latency / success_requests if success_requests > 0 else 0

    model_stats = {}
    for s in stats:
        if s.model not in model_stats:
            model_stats[s.model] = {"model": s.model, "requests": 0, "tokens": 0}
        model_stats[s.model]["requests"] += s.request_count
        model_stats[s.model]["tokens"] += s.total_tokens

    capability_stats = {}
    for s in stats:
        if s.capability not in capability_stats:
            capability_stats[s.capability] = {"capability": s.capability, "requests": 0, "tokens": 0}
        capability_stats[s.capability]["requests"] += s.request_count
        capability_stats[s.capability]["tokens"] += s.total_tokens

    return StatsResponse(
        total_requests=total_requests,
        success_requests=success_requests,
        failed_requests=failed_requests,
        total_tokens=total_tokens,
        prompt_tokens=prompt_tokens,
        completion_tokens=completion_tokens,
        avg_latency_ms=avg_latency,
        by_model=list(model_stats.values()),
        by_capability=list(capability_stats.values()),
    )