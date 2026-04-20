import logging
import time
import uuid

from fastapi import Depends, HTTPException, Request
from fastapi.responses import Response
from pydantic import BaseModel
from sqlalchemy.ext.asyncio import AsyncSession

from core import decrypt_api_key, model_config_cache, QuotaChecker
from models import AIRequestLog

logger = logging.getLogger(__name__)


class SpeechRequest(BaseModel):
    model: str
    input: str
    voice: str = "alloy"
    response_format: str | None = "mp3"
    speed: float | None = 1.0


class TranscriptionResponse(BaseModel):
    text: str


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


async def speech(
    request: SpeechRequest,
    db: AsyncSession = Depends(get_db),
    workspace_id: str = Depends(get_workspace_id),
    user_id: str = Depends(get_user_id),
) -> Response:
    from services import create_client

    model_config = await model_config_cache.get_model(db, request.model)
    if not model_config:
        raise HTTPException(status_code=404, detail=f"Model {request.model} not found or not enabled")

    if not model_config.supports_audio_speech:
        raise HTTPException(status_code=400, detail=f"Model {request.model} does not support audio speech")

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
        audio_data = await client.speech(
            input_text=request.input,
            model=request.model,
            voice=request.voice,
            response_format=request.response_format,
            speed=request.speed,
        )

        latency_ms = int((time.time() - start_time) * 1000)
        total_tokens = len(request.input.split())

        await _log_request(
            db=db,
            request_id=request_id,
            user_id=user_id,
            workspace_id=workspace_id,
            model=request.model,
            provider=model_config.provider,
            capability="tts",
            prompt_tokens=total_tokens,
            completion_tokens=0,
            total_tokens=total_tokens,
            latency_ms=latency_ms,
            status="success",
            error_code=None,
            error_message=None,
        )

        await quota_checker.consume_quota(workspace_id, total_tokens)

        content_type = "audio/mpeg"
        if request.response_format == "wav":
            content_type = "audio/wav"
        elif request.response_format == "opus":
            content_type = "audio/opus"
        elif request.response_format == "aac":
            content_type = "audio/aac"
        elif request.response_format == "flac":
            content_type = "audio/flac"

        return Response(
            content=audio_data,
            media_type=content_type,
            headers={
                "X-Request-Id": request_id,
                "X-Model": request.model,
            },
        )
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Speech synthesis error: {e}")
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