import json
import logging
import time
from abc import ABC, abstractmethod
from typing import Any, AsyncIterator

from models import ProviderModel

logger = logging.getLogger(__name__)


class BaseAIClient(ABC):
    def __init__(self, config: ProviderModel, decrypted_api_key: str):
        self.config = config
        self.api_key = decrypted_api_key
        self.base_url = config.base_url.rstrip("/")
        self.organization = config.organization
        self.extra_headers = config.extra_headers or {}

    def _build_headers(self) -> dict[str, str]:
        headers = {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {self.api_key}",
        }
        if self.organization:
            headers["OpenAI-Organization"] = self.organization
        for key, value in self.extra_headers.items():
            if key and value:
                headers[key] = str(value)
        return headers

    @abstractmethod
    async def chat_completions(
        self, messages: list[dict], model: str, stream: bool = False, **kwargs
    ) -> dict | AsyncIterator[dict]:
        pass

    @abstractmethod
    async def embeddings(self, input_texts: list[str], model: str, **kwargs) -> dict:
        pass

    @abstractmethod
    async def rerank(
        self, query: str, documents: list[str], model: str, **kwargs
    ) -> dict:
        pass

    @abstractmethod
    async def speech(self, input_text: str, model: str, voice: str, **kwargs) -> bytes:
        pass

    @abstractmethod
    async def transcriptions(self, audio_data: bytes, model: str, **kwargs) -> dict:
        pass

    def _parse_error_response(self, response_body: bytes, status_code: int) -> dict:
        try:
            error_data = json.loads(response_body)
            if "error" in error_data:
                error = error_data["error"]
                return {
                    "error_code": error.get("code", "unknown"),
                    "error_message": error.get("message", "Unknown error"),
                    "status_code": status_code,
                }
            return {
                "error_code": "unknown",
                "error_message": response_body.decode("utf-8", errors="ignore")[:500],
                "status_code": status_code,
            }
        except Exception:
            return {
                "error_code": "parse_error",
                "error_message": response_body.decode("utf-8", errors="ignore")[:500],
                "status_code": status_code,
            }


class ChatCompletionResponse:
    def __init__(
        self,
        content: str = "",
        model: str = "",
        prompt_tokens: int = 0,
        completion_tokens: int = 0,
        total_tokens: int = 0,
        finish_reason: str = "",
        raw_response: dict | None = None,
    ):
        self.content = content
        self.model = model
        self.prompt_tokens = prompt_tokens
        self.completion_tokens = completion_tokens
        self.total_tokens = total_tokens
        self.finish_reason = finish_reason
        self.raw_response = raw_response

    def to_dict(self) -> dict:
        return {
            "content": self.content,
            "model": self.model,
            "prompt_tokens": self.prompt_tokens,
            "completion_tokens": self.completion_tokens,
            "total_tokens": self.total_tokens,
            "finish_reason": self.finish_reason,
        }


class EmbeddingResponse:
    def __init__(
        self,
        embeddings: list[list[float]],
        model: str = "",
        prompt_tokens: int = 0,
        total_tokens: int = 0,
    ):
        self.embeddings = embeddings
        self.model = model
        self.prompt_tokens = prompt_tokens
        self.total_tokens = total_tokens

    def to_dict(self) -> dict:
        return {
            "embeddings": self.embeddings,
            "model": self.model,
            "prompt_tokens": self.prompt_tokens,
            "total_tokens": self.total_tokens,
        }


class RerankResponse:
    def __init__(self, results: list[dict], model: str = ""):
        self.results = results
        self.model = model

    def to_dict(self) -> dict:
        return {
            "results": self.results,
            "model": self.model,
        }