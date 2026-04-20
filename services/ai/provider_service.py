from __future__ import annotations

import logging
from collections.abc import AsyncIterator
from typing import Any

import httpx
from fastapi import HTTPException

from schemas import ModelConfig

logger = logging.getLogger(__name__)


class UnifiedProviderService:
    def __init__(
        self,
        timeout_seconds: int,
        max_connections: int = 100,
        max_keepalive_connections: int = 20,
        keepalive_expiry: float = 30.0,
    ):
        self._timeout_seconds = timeout_seconds
        self._max_connections = max_connections
        self._max_keepalive_connections = max_keepalive_connections
        self._keepalive_expiry = keepalive_expiry
        self._client: httpx.AsyncClient | None = None

    async def start(self) -> None:
        if self._client is not None:
            return
        limits = httpx.Limits(
            max_connections=self._max_connections,
            max_keepalive_connections=self._max_keepalive_connections,
            keepalive_expiry=self._keepalive_expiry,
        )
        timeout = httpx.Timeout(self._timeout_seconds, read=None)
        self._client = httpx.AsyncClient(limits=limits, timeout=timeout)

    async def stop(self) -> None:
        if self._client is None:
            return
        await self._client.aclose()
        self._client = None

    def _get_client(self) -> httpx.AsyncClient:
        if self._client is None:
            raise HTTPException(status_code=500, detail="HTTP client not initialized")
        return self._client

    async def post_json(
        self,
        model_config: ModelConfig,
        payload: dict[str, Any],
        path: str,
        stream: bool = False,
    ) -> Any:
        if stream:
            return self._stream_request(model_config, payload, path)
        return await self._single_request(model_config, payload, path)

    async def get_json(
        self,
        model_config: ModelConfig,
        path: str,
        params: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        endpoint = self._build_endpoint(model_config.base_url, path)
        headers = self._build_headers(model_config)
        client = self._get_client()
        try:
            response = await client.get(endpoint, params=params, headers=headers)
        except httpx.HTTPError as exc:
            raise HTTPException(status_code=502, detail=f"Provider request failed: {exc}") from exc
        if response.status_code >= 400:
            raise HTTPException(status_code=502, detail=self._error_detail(response))
        try:
            data = response.json()
        except ValueError as exc:
            raise HTTPException(status_code=502, detail="Provider response is not valid JSON") from exc
        if not isinstance(data, dict):
            raise HTTPException(status_code=502, detail="Provider response JSON must be an object")
        return data

    async def test_connection(
        self,
        model_config: ModelConfig,
        path: str,
    ) -> bool:
        endpoint = self._build_endpoint(model_config.base_url, path)
        headers = self._build_headers(model_config)
        client = self._get_client()
        logger.info(f"[test_connection] 请求 endpoint: {endpoint}")
        logger.info(f"[test_connection] headers: {list(headers.keys())}")
        try:
            response = await client.get(endpoint, headers=headers)
        except httpx.HTTPError as exc:
            logger.error(f"[test_connection] HTTP 错误: {exc}")
            raise HTTPException(status_code=502, detail=f"连接失败: {exc}") from exc
        logger.info(f"[test_connection] 状态码: {response.status_code}")
        logger.info(f"[test_connection] 响应体: {response.text[:500]}")
        if response.status_code >= 400:
            raise HTTPException(status_code=502, detail=self._error_detail(response))
        try:
            data = response.json()
            if isinstance(data, dict):
                error_msg = data.get("detail") or data.get("error") or data.get("message")
                if error_msg and not data.get("data") and not data.get("models"):
                    raise HTTPException(status_code=502, detail=f"Provider 返回错误: {error_msg}")
        except ValueError:
            pass
        return True

    async def post_raw(
        self,
        model_config: ModelConfig,
        path: str,
        body: bytes,
        content_type: str | None,
        accept: str | None,
    ) -> tuple[bytes, str]:
        endpoint = self._build_endpoint(model_config.base_url, path)
        headers = self._build_headers(model_config)
        headers.pop("Content-Type", None)
        if content_type:
            headers["Content-Type"] = content_type
        if accept:
            headers["Accept"] = accept
        client = self._get_client()
        try:
            response = await client.post(endpoint, content=body, headers=headers)
        except httpx.HTTPError as exc:
            raise HTTPException(status_code=502, detail=f"Provider request failed: {exc}") from exc
        if response.status_code >= 400:
            raise HTTPException(status_code=502, detail=self._error_detail(response))
        response_content_type = response.headers.get("content-type", "application/octet-stream")
        return response.content, response_content_type

    async def _single_request(
        self,
        model_config: ModelConfig,
        payload: dict[str, Any],
        path: str,
    ) -> dict[str, Any]:
        endpoint = self._build_endpoint(model_config.base_url, path)
        headers = self._build_headers(model_config)
        client = self._get_client()
        try:
            response = await client.post(endpoint, json=payload, headers=headers)
        except httpx.HTTPError as exc:
            raise HTTPException(status_code=502, detail=f"Provider request failed: {exc}") from exc
        if response.status_code >= 400:
            raise HTTPException(status_code=502, detail=self._error_detail(response))
        try:
            data = response.json()
        except ValueError as exc:
            raise HTTPException(status_code=502, detail="Provider response is not valid JSON") from exc
        if not isinstance(data, dict):
            raise HTTPException(status_code=502, detail="Provider response JSON must be an object")
        return data

    async def _stream_request(
        self,
        model_config: ModelConfig,
        payload: dict[str, Any],
        path: str,
    ) -> AsyncIterator[bytes]:
        endpoint = self._build_endpoint(model_config.base_url, path)
        headers = self._build_headers(model_config)
        client = self._get_client()
        try:
            async with client.stream("POST", endpoint, json=payload, headers=headers) as response:
                if response.status_code >= 400:
                    body = await response.aread()
                    detail = body.decode("utf-8", errors="ignore") or self._error_detail_from_status(
                        response.status_code
                    )
                    raise HTTPException(status_code=502, detail=detail)
                async for chunk in response.aiter_raw():
                    if chunk:
                        yield chunk
        except HTTPException:
            raise
        except httpx.HTTPError as exc:
            raise HTTPException(status_code=502, detail=f"Provider stream failed: {exc}") from exc

    @staticmethod
    def _build_headers(model_config: ModelConfig) -> dict[str, str]:
        headers: dict[str, str] = {
            "Authorization": f"Bearer {model_config.api_key}",
            "Accept": "application/json",
            "Content-Type": "application/json",
        }
        if model_config.organization:
            headers["OpenAI-Organization"] = model_config.organization
        headers.update(model_config.extra_headers)
        return headers

    @staticmethod
    def _build_endpoint(base_url: str, path: str) -> str:
        normalized_path = path if path.startswith("/") else f"/{path}"
        return f"{base_url.rstrip('/')}{normalized_path}"

    @staticmethod
    def _error_detail(response: httpx.Response) -> str:
        detail = response.text.strip()
        if detail:
            return detail
        return UnifiedProviderService._error_detail_from_status(response.status_code)

    @staticmethod
    def _error_detail_from_status(status_code: int) -> str:
        return f"Provider request failed with status {status_code}"