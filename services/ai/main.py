from __future__ import annotations

import logging
import time
from typing import Any

from fastapi import FastAPI, Header, HTTPException, Query, Request
from fastapi.responses import JSONResponse, Response, StreamingResponse
from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncEngine
import uvicorn

from config import AIServiceConfig, get_config
from crypto import init_crypto_service
from db import build_engine
from logging_middleware import RequestLoggingMiddleware, setup_logging
from model_repository import ModelConfigRepository
from provider_service import UnifiedProviderService
from schemas import (
    ChatCompletionsRequest,
    GenericModelRequest,
    ModelConfigResponse,
    TestConnectionRequest,
    TestConnectionResponse,
)

cfg = get_config()
setup_logging(cfg.log_level)
logger = logging.getLogger(__name__)

app = FastAPI(title=cfg.app_name)
app.add_middleware(
    RequestLoggingMiddleware,
    exclude_paths=["/", "/health", "/ready"],
)
app.state.db_engine = None
app.state.model_repository = None
app.state.provider_service = UnifiedProviderService(timeout_seconds=cfg.request_timeout_seconds)


def get_internal_token(config: AIServiceConfig, token: str | None) -> None:
    expected = config.internal_token.strip()
    if not expected:
        return
    if token != expected:
        raise HTTPException(status_code=401, detail="invalid internal token")


def _provider_service() -> UnifiedProviderService:
    return app.state.provider_service


def _repository() -> ModelConfigRepository:
    repository = app.state.model_repository
    if repository is None:
        raise HTTPException(status_code=500, detail="database is not configured")
    return repository


@app.on_event("startup")
async def startup() -> None:
    app.state.config = cfg
    if cfg.encryption_key:
        init_crypto_service(cfg.encryption_key)
    app.state.db_engine = build_engine(cfg)
    if app.state.db_engine is None:
        app.state.model_repository = None
    else:
        app.state.model_repository = ModelConfigRepository(app.state.db_engine, cfg)
    app.state.provider_service = UnifiedProviderService(timeout_seconds=cfg.request_timeout_seconds)
    await app.state.provider_service.start()


@app.on_event("shutdown")
async def shutdown() -> None:
    engine: AsyncEngine | None = app.state.db_engine
    if engine is not None:
        await engine.dispose()
    provider_service: UnifiedProviderService | None = app.state.provider_service
    if provider_service is not None:
        await provider_service.stop()


@app.get("/")
async def root() -> dict[str, str]:
    return {"service": cfg.app_name, "env": cfg.app_env, "status": "ok"}


@app.get("/health")
async def health() -> dict[str, str]:
    return {"status": "healthy"}


@app.get("/ready")
async def ready() -> dict[str, Any]:
    checks: dict[str, Any] = {
        "database": "unknown",
        "http_client": "unknown",
    }

    engine: AsyncEngine | None = app.state.db_engine
    if engine is not None:
        try:
            async with engine.connect() as conn:
                await conn.execute(text("SELECT 1"))
            checks["database"] = "ok"
        except Exception as e:
            checks["database"] = f"error: {e}"
    else:
        checks["database"] = "not_configured"

    provider_service: UnifiedProviderService | None = app.state.provider_service
    if provider_service is not None:
        try:
            client = provider_service._get_client()
            if client is not None:
                checks["http_client"] = "ok"
            else:
                checks["http_client"] = "not_initialized"
        except Exception as e:
            checks["http_client"] = f"error: {e}"
    else:
        checks["http_client"] = "not_initialized"

    all_healthy = all(
        v in ("ok", "not_configured", "unknown") for v in checks.values()
    )

    return {
        "status": "ready" if all_healthy else "not_ready",
        "checks": checks,
    }


@app.get("/internal/v1/models/{model}")
async def get_model_config(
    model: str,
    x_internal_token: str | None = Header(default=None),
) -> ModelConfigResponse:
    get_internal_token(cfg, x_internal_token)
    repo = _repository()
    model_config = await repo.get_by_model(model)
    return ModelConfigResponse(
        provider=model_config.provider,
        model=model_config.model_key,
        request_model=model_config.request_model,
        base_url=model_config.base_url,
        chat_completions_path=model_config.chat_completions_path,
        chat_responses_path=model_config.chat_responses_path,
        embeddings_path=model_config.embeddings_path,
        rerank_path=model_config.rerank_path,
        audio_speech_path=model_config.audio_speech_path,
        audio_transcriptions_path=model_config.audio_transcriptions_path,
        models_path=model_config.models_path,
    )


@app.post("/internal/v1/chat/completions", response_model=None)
async def internal_chat_completions(
    body: ChatCompletionsRequest,
    x_internal_token: str | None = Header(default=None),
) -> Response:
    get_internal_token(cfg, x_internal_token)
    payload = _build_upstream_payload(body)
    return await _forward_model_json_request(
        payload=payload,
        path_resolver=lambda config: config.chat_completions_path,
    )


@app.post("/internal/v1/chat/responses", response_model=None)
async def internal_chat_responses(
    body: GenericModelRequest,
    x_internal_token: str | None = Header(default=None),
) -> Response:
    get_internal_token(cfg, x_internal_token)
    payload = body.model_dump(exclude_none=True)
    return await _forward_model_json_request(
        payload=payload,
        path_resolver=lambda config: config.chat_responses_path,
    )


@app.post("/internal/v1/embeddings", response_model=None)
async def internal_embeddings(
    body: GenericModelRequest,
    x_internal_token: str | None = Header(default=None),
) -> Response:
    get_internal_token(cfg, x_internal_token)
    payload = body.model_dump(exclude_none=True)
    return await _forward_model_json_request(
        payload=payload,
        path_resolver=lambda config: config.embeddings_path,
    )


@app.post("/internal/v1/rerank", response_model=None)
async def internal_rerank(
    body: GenericModelRequest,
    x_internal_token: str | None = Header(default=None),
) -> Response:
    get_internal_token(cfg, x_internal_token)
    payload = body.model_dump(exclude_none=True)
    return await _forward_model_json_request(
        payload=payload,
        path_resolver=lambda config: config.rerank_path,
    )


@app.post("/internal/v1/audio/speech", response_model=None)
async def internal_audio_speech(
    request: Request,
    model: str = Query(min_length=1),
    x_internal_token: str | None = Header(default=None),
) -> Response:
    get_internal_token(cfg, x_internal_token)
    model_config = await _resolve_model_config(model)
    body = await request.body()
    content, content_type = await _provider_service().post_raw(
        model_config=model_config,
        path=model_config.audio_speech_path,
        body=body,
        content_type=request.headers.get("content-type"),
        accept=request.headers.get("accept"),
    )
    return Response(content=content, media_type=content_type)


@app.post("/internal/v1/audio/transcriptions", response_model=None)
async def internal_audio_transcriptions(
    request: Request,
    model: str = Query(min_length=1),
    x_internal_token: str | None = Header(default=None),
) -> Response:
    get_internal_token(cfg, x_internal_token)
    model_config = await _resolve_model_config(model)
    body = await request.body()
    content, content_type = await _provider_service().post_raw(
        model_config=model_config,
        path=model_config.audio_transcriptions_path,
        body=body,
        content_type=request.headers.get("content-type"),
        accept=request.headers.get("accept"),
    )
    return Response(content=content, media_type=content_type)


@app.get("/internal/v1/models", response_model=None)
async def internal_models(
    model: str = Query(min_length=1),
    x_internal_token: str | None = Header(default=None),
) -> Response:
    get_internal_token(cfg, x_internal_token)
    model_config = await _resolve_model_config(model)
    data = await _provider_service().get_json(
        model_config=model_config,
        path=model_config.models_path,
    )
    return JSONResponse(content=data)


@app.post("/internal/v1/test-connection", response_model=TestConnectionResponse)
async def test_connection(
    body: TestConnectionRequest,
    x_internal_token: str | None = Header(default=None),
) -> TestConnectionResponse:
    get_internal_token(cfg, x_internal_token)
    logger.info(f"[test_connection] Received request: base_url={body.base_url}, models_path={body.models_path}")

    from schemas import ModelConfig

    test_config = ModelConfig(
        provider="openai-compatible",
        model_key=body.model or "test-model",
        request_model=body.request_model or body.model or "test-model",
        base_url=body.base_url,
        api_key=body.api_key,
        organization=body.organization,
        chat_completions_path=body.chat_completions_path,
        models_path=body.models_path,
        extra_headers=body.extra_headers,
    )

    provider = _provider_service()
    start_time = time.perf_counter()

    try:
        await provider.test_connection(model_config=test_config, path=test_config.models_path)
        latency_ms = (time.perf_counter() - start_time) * 1000

        return TestConnectionResponse(
            success=True,
            message="连接成功",
            latency_ms=round(latency_ms, 2),
        )
    except HTTPException as e:
        latency_ms = (time.perf_counter() - start_time) * 1000
        return TestConnectionResponse(
            success=False,
            message="连接失败",
            latency_ms=round(latency_ms, 2),
            error_detail=e.detail,
        )
    except Exception as e:
        latency_ms = (time.perf_counter() - start_time) * 1000
        return TestConnectionResponse(
            success=False,
            message="连接失败",
            latency_ms=round(latency_ms, 2),
            error_detail=str(e),
        )


def _build_upstream_payload(body: ChatCompletionsRequest) -> dict[str, Any]:
    payload = body.model_dump(exclude_none=True)
    payload["stream"] = bool(body.stream)
    return payload


async def _resolve_model_config(model: str):
    repo = _repository()
    return await repo.get_by_model(model)


async def _forward_model_json_request(
    payload: dict[str, Any],
    path_resolver,
) -> Response:
    model = str(payload.get("model", "")).strip()
    if not model:
        raise HTTPException(status_code=422, detail="model is required")
    model_config = await _resolve_model_config(model)
    payload["model"] = model_config.request_model
    stream = bool(payload.get("stream"))
    provider = _provider_service()
    target_path = path_resolver(model_config)
    if stream:
        stream_iter = await provider.post_json(
            model_config=model_config,
            payload=payload,
            path=target_path,
            stream=True,
        )
        return StreamingResponse(stream_iter, media_type="text/event-stream")
    data = await provider.post_json(
        model_config=model_config,
        payload=payload,
        path=target_path,
        stream=False,
    )
    return JSONResponse(content=data)


if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host=cfg.host,
        port=cfg.port,
        reload=cfg.reload,
        log_level=cfg.log_level,
    )
