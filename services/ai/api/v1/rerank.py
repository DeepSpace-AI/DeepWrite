import logging
import time
import uuid
from typing import Any

from fastapi import Depends, HTTPException, Request
from pydantic import BaseModel, Field
from sqlalchemy.ext.asyncio import AsyncSession

from core import decrypt_api_key, model_config_cache, QuotaChecker
from models import AIRequestLog
from services import create_client

logger = logging.getLogger(__name__)


class RerankRequest(BaseModel):
    model: str
    query: str
    documents: list[str]
    top_n: int | None = None
    return_documents: bool | None = True


class RerankResult(BaseModel):
    index: int
    relevance_score: float
    document: str | None = None


class RerankResponse(BaseModel):
    object: str = "rerank"
    model: str
    results: list[RerankResult]
    usage: dict


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


async def rerank(
    request: RerankRequest,
    db: AsyncSession = Depends(get_db),
    workspace_id: str = Depends(get_workspace_id),
    user_id: str = Depends(get_user_id),
) -> RerankResponse:
    model_config = await model_config_cache.get_model(db, request.model)
    if not model_config:
        raise HTTPException(status_code=404, detail=f"Model {request.model} not found or not enabled")

    if not model_config.supports_rerank:
        raise HTTPException(status_code=400, detail=f"Model {request.model} does not support rerank")

    quota_checker = QuotaChecker(db)
    quota_result = await quota_checker.check_quota(workspace_id, request.model)
    if not quota_result.allowed:
        raise HTTPException(status_code=429, detail=quota_result.reason)

    try:
        decrypted_api_key = decrypt_api_key(model_config.api_key)
    except Exception as e:
        logger.error(f"Failed to decrypt API key: {e}")
        raise HTTPException(status_code=500, detail="Failed to decrypt API key")

    client = create_client(model_config, decrypted_api_key)

    request_id = str(uuid.uuid4())
    start_time = time.time()

    try:
        result = await client.rerank(
            query=request.query,
            documents=request.documents,
            model=request.model,
            top_n=request.top_n,
        )

        if "error" in result:
            error = result["error"]
            await _log_request(
                db=db,
                request_id=request_id,
                user_id=user_id,
                workspace_id=workspace_id,
                model=request.model,
                provider=model_config.provider,
                capability="rerank",
                prompt_tokens=0,
                completion_tokens=0,
                total_tokens=0,
                latency_ms=int((time.time() - start_time) * 1000),
                status="failed",
                error_code=error.get("error_code"),
                error_message=error.get("error_message"),
            )
            raise HTTPException(status_code=error.get("status_code", 500), detail=error.get("error_message"))

        latency_ms = int((time.time() - start_time) * 1000)
        total_tokens = len(request.query.split()) + sum(len(d.split()) for d in request.documents)

        await _log_request(
            db=db,
            request_id=request_id,
            user_id=user_id,
            workspace_id=workspace_id,
            model=request.model,
            provider=model_config.provider,
            capability="rerank",
            prompt_tokens=total_tokens,
            completion_tokens=0,
            total_tokens=total_tokens,
            latency_ms=latency_ms,
            status="success",
            error_code=None,
            error_message=None,
        )

        await quota_checker.consume_quota(workspace_id, total_tokens)

        results = []
        raw_results = result.get("results", [])
        for i, r in enumerate(raw_results[: request.top_n or len(raw_results)]):
            results.append(
                RerankResult(
                    index=r.get("index", i),
                    relevance_score=r.get("relevance_score", 0.0),
                    document=request.documents[r.get("index", i)] if request.return_documents else None,
                )
            )

        return RerankResponse(
            model=result.get("model", request.model),
            results=results,
            usage={"total_tokens": total_tokens},
        )
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Rerank error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


async def _log_request(
    db: AsyncSession,
    request_id: str,
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
    error_code: str | None,
    error_message: str | None,
):
    from uuid import UUID

    try:
        log = AIRequestLog(
            request_id=request_id,
            user_id=UUID(user_id),
            workspace_id=UUID(workspace_id),
            model=model,
            provider=provider,
            capability=capability,
            prompt_tokens=prompt_tokens,
            completion_tokens=completion_tokens,
            total_tokens=total_tokens,
            latency_ms=latency_ms,
            status=status,
            error_code=error_code,
            error_message=error_message[:500] if error_message else None,
            is_stream=False,
        )
        db.add(log)
        await db.commit()
    except Exception as e:
        logger.error(f"Failed to log request: {e}")