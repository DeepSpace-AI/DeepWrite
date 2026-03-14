from __future__ import annotations

from collections.abc import AsyncIterator
from typing import Any

import httpx
from fastapi import HTTPException

from schemas import ModelConfig


class UnifiedProviderService:
    def __init__(self, timeout_seconds: int):
        self._timeout_seconds = timeout_seconds

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
        timeout = httpx.Timeout(self._timeout_seconds)
        async with httpx.AsyncClient(timeout=timeout) as client:
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
        timeout = httpx.Timeout(self._timeout_seconds, read=None)
        async with httpx.AsyncClient(timeout=timeout) as client:
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
        timeout = httpx.Timeout(self._timeout_seconds)
        async with httpx.AsyncClient(timeout=timeout) as client:
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
        timeout = httpx.Timeout(self._timeout_seconds, read=None)
        async with httpx.AsyncClient(timeout=timeout) as client:
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
