import json
import logging
from typing import Any, AsyncIterator

import httpx

from models import ProviderModel
from services.base_client import BaseAIClient, ChatCompletionResponse, EmbeddingResponse, RerankResponse

logger = logging.getLogger(__name__)


class OpenAICompatibleClient(BaseAIClient):
    async def chat_completions(
        self,
        messages: list[dict],
        model: str,
        stream: bool = False,
        **kwargs,
    ) -> dict | AsyncIterator[dict]:
        url = f"{self.base_url}{self.config.chat_completions_path}"

        request_model = self.config.request_model or model
        payload = {
            "model": request_model,
            "messages": messages,
            "stream": stream,
            **kwargs,
        }

        headers = self._build_headers()

        async with httpx.AsyncClient(timeout=120.0) as client:
            if stream:
                return self._stream_chat(client, url, payload, headers)

            response = await client.post(url, json=payload, headers=headers)
            return self._parse_chat_response(response)

    async def _stream_chat(
        self, client: httpx.AsyncClient, url: str, payload: dict, headers: dict
    ) -> AsyncIterator[dict]:
        async with client.stream("POST", url, json=payload, headers=headers) as response:
            if response.status_code >= 400:
                error_body = await response.aread()
                error_info = self._parse_error_response(error_body, response.status_code)
                yield {"error": error_info}
                return

            async for line in response.aiter_lines():
                if line.startswith("data: "):
                    data = line[6:]
                    if data == "[DONE]":
                        break
                    try:
                        chunk = json.loads(data)
                        yield chunk
                    except json.JSONDecodeError:
                        continue

    def _parse_chat_response(self, response: httpx.Response) -> dict:
        if response.status_code >= 400:
            error_info = self._parse_error_response(response.content, response.status_code)
            return {"error": error_info}

        data = response.json()
        choices = data.get("choices", [])
        usage = data.get("usage", {})

        content = ""
        finish_reason = ""
        if choices:
            message = choices[0].get("message", {})
            content = message.get("content", "")
            finish_reason = choices[0].get("finish_reason", "")

        return {
            "content": content,
            "model": data.get("model", ""),
            "prompt_tokens": usage.get("prompt_tokens", 0),
            "completion_tokens": usage.get("completion_tokens", 0),
            "total_tokens": usage.get("total_tokens", 0),
            "finish_reason": finish_reason,
            "raw": data,
        }

    async def embeddings(self, input_texts: list[str], model: str, **kwargs) -> dict:
        url = f"{self.base_url}{self.config.embeddings_path}"

        request_model = self.config.request_model or model
        payload = {
            "model": request_model,
            "input": input_texts,
            **kwargs,
        }

        headers = self._build_headers()

        async with httpx.AsyncClient(timeout=60.0) as client:
            response = await client.post(url, json=payload, headers=headers)
            return self._parse_embeddings_response(response)

    def _parse_embeddings_response(self, response: httpx.Response) -> dict:
        if response.status_code >= 400:
            error_info = self._parse_error_response(response.content, response.status_code)
            return {"error": error_info}

        data = response.json()
        embeddings = []
        for item in data.get("data", []):
            embeddings.append(item.get("embedding", []))

        usage = data.get("usage", {})

        return {
            "embeddings": embeddings,
            "model": data.get("model", ""),
            "prompt_tokens": usage.get("prompt_tokens", 0),
            "total_tokens": usage.get("total_tokens", 0),
        }

    async def rerank(
        self, query: str, documents: list[str], model: str, **kwargs
    ) -> dict:
        url = f"{self.base_url}{self.config.rerank_path}"

        request_model = self.config.request_model or model
        payload = {
            "model": request_model,
            "query": query,
            "documents": documents,
            **kwargs,
        }

        headers = self._build_headers()

        async with httpx.AsyncClient(timeout=60.0) as client:
            response = await client.post(url, json=payload, headers=headers)
            return self._parse_rerank_response(response)

    def _parse_rerank_response(self, response: httpx.Response) -> dict:
        if response.status_code >= 400:
            error_info = self._parse_error_response(response.content, response.status_code)
            return {"error": error_info}

        data = response.json()
        results = data.get("results", [])

        return {
            "results": results,
            "model": data.get("model", ""),
        }

    async def speech(self, input_text: str, model: str, voice: str = "alloy", **kwargs) -> bytes:
        url = f"{self.base_url}{self.config.audio_speech_path}"

        request_model = self.config.request_model or model
        payload = {
            "model": request_model,
            "input": input_text,
            "voice": voice,
            **kwargs,
        }

        headers = self._build_headers()

        async with httpx.AsyncClient(timeout=120.0) as client:
            response = await client.post(url, json=payload, headers=headers)
            if response.status_code >= 400:
                error_info = self._parse_error_response(response.content, response.status_code)
                return json.dumps({"error": error_info}).encode()
            return response.content

    async def transcriptions(self, audio_data: bytes, model: str, **kwargs) -> dict:
        url = f"{self.base_url}{self.config.audio_transcriptions_path}"

        request_model = self.config.request_model or model

        headers = {"Authorization": f"Bearer {self.api_key}"}
        if self.organization:
            headers["OpenAI-Organization"] = self.organization

        files = {"file": ("audio.mp3", audio_data, "audio/mpeg")}
        data = {"model": request_model, **kwargs}

        async with httpx.AsyncClient(timeout=120.0) as client:
            response = await client.post(url, files=files, data=data, headers=headers)
            if response.status_code >= 400:
                error_info = self._parse_error_response(response.content, response.status_code)
                return {"error": error_info}

            return response.json()