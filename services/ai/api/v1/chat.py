import json
import logging
import time
import uuid
from typing import Any

from fastapi import Depends, HTTPException, Request
from fastapi.responses import StreamingResponse
from pydantic import BaseModel, Field
from sqlalchemy.ext.asyncio import AsyncSession

from core import decrypt_api_key, model_config_cache, QuotaChecker
from models import AIRequestLog
from services import create_client

logger = logging.getLogger(__name__)


class ChatMessage(BaseModel):
    role: str
    content: str | list[dict] | None = None
    name: str | None = None


class ChatCompletionRequest(BaseModel):
    model: str
    messages: list[ChatMessage]
    temperature: float | None = None
    top_p: float | None = None
    n: int | None = None
    stream: bool = False
    stop: str | list[str] | None = None
    max_tokens: int | None = None
    presence_penalty: float | None = None
    frequency_penalty: float | None = None
    user: str | None = None


class ChatCompletionChoice(BaseModel):
    index: int
    message: dict
    finish_reason: str


class ChatCompletionResponse(BaseModel):
    id: str
    object: str = "chat.completion"
    created: int
    model: str
    choices: list[ChatCompletionChoice]
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


async def chat_completions(
    request: ChatCompletionRequest,
    db: AsyncSession = Depends(get_db),
    workspace_id: str = Depends(get_workspace_id),
    user_id: str = Depends(get_user_id),
) -> ChatCompletionResponse | StreamingResponse:
    model_config = await model_config_cache.get_model(db, request.model)
    if not model_config:
        raise HTTPException(status_code=404, detail=f"Model {request.model} not found or not enabled")

    if not model_config.supports_chat_completions:
        raise HTTPException(status_code=400, detail=f"Model {request.model} does not support chat completions")

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
    messages = [{"role": m.role, "content": m.content, "name": m.name} for m in request.messages if m.content]

    if request.stream:
        return StreamingResponse(
            _stream_chat_response(
                db, client, request, messages, model_config, workspace_id, user_id, request_id, start_time
            ),
            media_type="text/event-stream",
        )

    try:
        result = await client.chat_completions(
            messages=messages,
            model=request.model,
            stream=False,
            temperature=request.temperature,
            top_p=request.top_p,
            max_tokens=request.max_tokens,
            presence_penalty=request.presence_penalty,
            frequency_penalty=request.frequency_penalty,
            stop=request.stop,
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
                capability="chat",
                prompt_tokens=0,
                completion_tokens=0,
                total_tokens=0,
                latency_ms=int((time.time() - start_time) * 1000),
                status="failed",
                error_code=error.get("error_code"),
                error_message=error.get("error_message"),
                is_stream=False,
            )
            raise HTTPException(status_code=error.get("status_code", 500), detail=error.get("error_message"))

        prompt_tokens = result.get("prompt_tokens", 0)
        completion_tokens = result.get("completion_tokens", 0)
        total_tokens = result.get("total_tokens", 0)
        latency_ms = int((time.time() - start_time) * 1000)

        await _log_request(
            db=db,
            request_id=request_id,
            user_id=user_id,
            workspace_id=workspace_id,
            model=request.model,
            provider=model_config.provider,
            capability="chat",
            prompt_tokens=prompt_tokens,
            completion_tokens=completion_tokens,
            total_tokens=total_tokens,
            latency_ms=latency_ms,
            status="success",
            error_code=None,
            error_message=None,
            is_stream=False,
        )

        await quota_checker.consume_quota(workspace_id, total_tokens)

        return ChatCompletionResponse(
            id=request_id,
            created=int(time.time()),
            model=result.get("model", request.model),
            choices=[
                ChatCompletionChoice(
                    index=0,
                    message={"role": "assistant", "content": result.get("content", "")},
                    finish_reason=result.get("finish_reason", "stop"),
                )
            ],
            usage={
                "prompt_tokens": prompt_tokens,
                "completion_tokens": completion_tokens,
                "total_tokens": total_tokens,
            },
        )
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Chat completion error: {e}")
        await _log_request(
            db=db,
            request_id=request_id,
            user_id=user_id,
            workspace_id=workspace_id,
            model=request.model,
            provider=model_config.provider,
            capability="chat",
            prompt_tokens=0,
            completion_tokens=0,
            total_tokens=0,
            latency_ms=int((time.time() - start_time) * 1000),
            status="failed",
            error_code="internal_error",
            error_message=str(e)[:500],
            is_stream=False,
        )
        raise HTTPException(status_code=500, detail=str(e))


async def _stream_chat_response(
    db: AsyncSession,
    client,
    request: ChatCompletionRequest,
    messages: list[dict],
    model_config,
    workspace_id: str,
    user_id: str,
    request_id: str,
    start_time: float,
):
    total_content = ""
    prompt_tokens = 0
    completion_tokens = 0
    error_occurred = False
    error_info = None

    try:
        stream = await client.chat_completions(
            messages=messages,
            model=request.model,
            stream=True,
            temperature=request.temperature,
            top_p=request.top_p,
            max_tokens=request.max_tokens,
        )

        async for chunk in stream:
            if "error" in chunk:
                error_occurred = True
                error_info = chunk["error"]
                yield f"data: {json.dumps(chunk)}\n\n"
                break

            choices = chunk.get("choices", [])
            if choices:
                delta = choices[0].get("delta", {})
                content = delta.get("content", "")
                if content:
                    total_content += content

            yield f"data: {json.dumps(chunk)}\n\n"

        yield "data: [DONE]\n\n"

        latency_ms = int((time.time() - start_time) * 1000)
        completion_tokens = len(total_content.split())

        status = "failed" if error_occurred else "success"
        error_code = error_info.get("error_code") if error_info else None
        error_message = error_info.get("error_message") if error_info else None

        await _log_request(
            db=db,
            request_id=request_id,
            user_id=user_id,
            workspace_id=workspace_id,
            model=request.model,
            provider=model_config.provider,
            capability="chat",
            prompt_tokens=prompt_tokens,
            completion_tokens=completion_tokens,
            total_tokens=prompt_tokens + completion_tokens,
            latency_ms=latency_ms,
            status=status,
            error_code=error_code,
            error_message=error_message,
            is_stream=True,
        )

        if not error_occurred:
            quota_checker = QuotaChecker(db)
            await quota_checker.consume_quota(workspace_id, prompt_tokens + completion_tokens)

    except Exception as e:
        logger.error(f"Stream chat error: {e}")
        error_data = {"error": {"message": str(e), "code": "internal_error"}}
        yield f"data: {json.dumps(error_data)}\n\n"


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
    is_stream: bool,
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
            is_stream=is_stream,
        )
        db.add(log)
        await db.commit()
    except Exception as e:
        logger.error(f"Failed to log request: {e}")