from __future__ import annotations

from typing import Any

from pydantic import BaseModel, ConfigDict, Field


class ChatMessage(BaseModel):
    role: str
    content: Any
    name: str | None = None
    tool_call_id: str | None = None


class ChatCompletionsRequest(BaseModel):
    model_config = ConfigDict(extra="allow")

    model: str = Field(min_length=1)
    messages: list[ChatMessage]
    stream: bool = False
    temperature: float | None = None
    top_p: float | None = None
    max_tokens: int | None = None
    presence_penalty: float | None = None
    frequency_penalty: float | None = None
    user: str | None = None
    stop: str | list[str] | None = None
    tools: list[dict[str, Any]] | None = None
    tool_choice: str | dict[str, Any] | None = None


class GenericModelRequest(BaseModel):
    model_config = ConfigDict(extra="allow")

    model: str = Field(min_length=1)


class ModelConfig(BaseModel):
    provider: str = "openai-compatible"
    model_key: str
    request_model: str
    base_url: str
    api_key: str
    organization: str | None = None
    chat_completions_path: str = "/chat/completions"
    chat_path: str = "/chat/completions"
    chat_responses_path: str = "/responses"
    embeddings_path: str = "/embeddings"
    rerank_path: str = "/rerank"
    audio_speech_path: str = "/audio/speech"
    audio_transcriptions_path: str = "/audio/transcriptions"
    models_path: str = "/models"
    extra_headers: dict[str, str] = Field(default_factory=dict)


class ModelConfigResponse(BaseModel):
    provider: str
    model: str
    request_model: str
    base_url: str
    chat_completions_path: str
    chat_responses_path: str
    embeddings_path: str
    rerank_path: str
    audio_speech_path: str
    audio_transcriptions_path: str
    models_path: str
