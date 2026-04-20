from __future__ import annotations

import json
from dataclasses import dataclass
from typing import Any

from fastapi import HTTPException
from sqlalchemy import text
from sqlalchemy.engine import RowMapping
from sqlalchemy.ext.asyncio import AsyncEngine

from config import AIServiceConfig
from crypto import decrypt_api_key
from schemas import ModelConfig


@dataclass(slots=True)
class ColumnProfile:
    model: str
    base_url: str
    api_key: str
    provider: str | None
    request_model: str | None
    organization: str | None
    chat_completions_path: str | None
    chat_responses_path: str | None
    embeddings_path: str | None
    rerank_path: str | None
    audio_speech_path: str | None
    audio_transcriptions_path: str | None
    models_path: str | None
    extra_headers: str | None
    enabled: str | None


class ModelConfigRepository:
    def __init__(self, engine: AsyncEngine, config: AIServiceConfig):
        self._engine = engine
        self._config = config
        self._table = config.validated_model_table()
        self._profile: ColumnProfile | None = None

    async def get_by_model(self, model: str) -> ModelConfig:
        profile = await self._get_profile()
        query = self._build_query(profile)
        async with self._engine.connect() as conn:
            result = await conn.execute(text(query), {"model": model})
            row = result.mappings().first()
        if row is None:
            raise HTTPException(status_code=404, detail=f"Model config not found for: {model}")
        return self._to_model_config(profile, row)

    async def _get_profile(self) -> ColumnProfile:
        if self._profile is not None:
            return self._profile
        async with self._engine.connect() as conn:
            result = await conn.execute(
                text(
                    """
                    SELECT column_name
                    FROM information_schema.columns
                    WHERE table_schema = current_schema()
                      AND table_name = :table_name
                    """
                ),
                {"table_name": self._table},
            )
            columns = {str(row[0]).lower() for row in result.fetchall()}
        if not columns:
            raise HTTPException(status_code=500, detail=f"Model config table not found: {self._table}")
        self._profile = ColumnProfile(
            model=self._pick_required(columns, "model", "model_key", "model_name", "model_code", "name"),
            base_url=self._pick_required(columns, "base_url", "api_base_url", "endpoint", "api_url", "url"),
            api_key=self._pick_required(columns, "api_key", "api_token", "access_key", "key"),
            provider=self._pick_optional(columns, "provider", "provider_name"),
            request_model=self._pick_optional(columns, "request_model", "upstream_model", "target_model"),
            organization=self._pick_optional(columns, "organization", "org_id"),
            chat_completions_path=self._pick_optional(
                columns,
                "chat_completions_path",
                "chat_path",
                "chat_endpoint",
                "endpoint_path",
            ),
            chat_responses_path=self._pick_optional(columns, "chat_responses_path", "responses_path"),
            embeddings_path=self._pick_optional(columns, "embeddings_path", "embedding_path"),
            rerank_path=self._pick_optional(columns, "rerank_path", "reranking_path"),
            audio_speech_path=self._pick_optional(columns, "audio_speech_path", "speech_path", "tts_path"),
            audio_transcriptions_path=self._pick_optional(
                columns,
                "audio_transcriptions_path",
                "transcriptions_path",
                "audio_transcribe_path",
            ),
            models_path=self._pick_optional(columns, "models_path", "model_list_path"),
            extra_headers=self._pick_optional(columns, "extra_headers", "headers", "extra_config"),
            enabled=self._pick_optional(columns, "is_enabled", "enabled", "status"),
        )
        return self._profile

    @staticmethod
    def _pick_required(columns: set[str], *candidates: str) -> str:
        value = ModelConfigRepository._pick_optional(columns, *candidates)
        if value is None:
            names = ", ".join(candidates)
            raise HTTPException(status_code=500, detail=f"Missing required config columns: {names}")
        return value

    @staticmethod
    def _pick_optional(columns: set[str], *candidates: str) -> str | None:
        for name in candidates:
            if name in columns:
                return name
        return None

    def _build_query(self, profile: ColumnProfile) -> str:
        conditions = [f"{profile.model} = :model"]
        if profile.enabled is not None:
            conditions.append(
                f"LOWER(CAST({profile.enabled} AS TEXT)) IN ('true', '1', 'active', 'enabled', 'on')"
            )
        where_clause = " AND ".join(conditions)
        return f"SELECT * FROM {self._table} WHERE {where_clause} LIMIT 1"

    def _to_model_config(self, profile: ColumnProfile, row: RowMapping) -> ModelConfig:
        provider = self._read_str(row, profile.provider, "openai-compatible")
        model_key = self._read_str(row, profile.model, "")
        request_model = self._read_str(row, profile.request_model, model_key)
        base_url = self._read_str(row, profile.base_url, "")
        api_key_encrypted = self._read_str(row, profile.api_key, "")
        api_key = decrypt_api_key(api_key_encrypted)
        organization = self._read_str(row, profile.organization, "") or None
        chat_completions_path = self._normalize_path(
            self._read_str(row, profile.chat_completions_path, self._config.default_chat_completions_path)
        )
        chat_responses_path = self._normalize_path(
            self._read_str(row, profile.chat_responses_path, self._config.default_chat_responses_path)
        )
        embeddings_path = self._normalize_path(
            self._read_str(row, profile.embeddings_path, self._config.default_embeddings_path)
        )
        rerank_path = self._normalize_path(self._read_str(row, profile.rerank_path, self._config.default_rerank_path))
        audio_speech_path = self._normalize_path(
            self._read_str(row, profile.audio_speech_path, self._config.default_audio_speech_path)
        )
        audio_transcriptions_path = self._normalize_path(
            self._read_str(
                row,
                profile.audio_transcriptions_path,
                self._config.default_audio_transcriptions_path,
            )
        )
        models_path = self._normalize_path(
            self._read_str(row, profile.models_path, self._config.default_models_path)
        )
        headers_raw = self._read_value(row, profile.extra_headers)
        extra_headers = self._parse_headers(headers_raw)
        if not base_url.strip() or not api_key.strip():
            raise HTTPException(status_code=500, detail=f"Invalid model config for: {model_key}")
        return ModelConfig(
            provider=provider,
            model_key=model_key,
            request_model=request_model,
            base_url=base_url.rstrip("/"),
            api_key=api_key,
            organization=organization,
            chat_completions_path=chat_completions_path,
            chat_path=chat_completions_path,
            chat_responses_path=chat_responses_path,
            embeddings_path=embeddings_path,
            rerank_path=rerank_path,
            audio_speech_path=audio_speech_path,
            audio_transcriptions_path=audio_transcriptions_path,
            models_path=models_path,
            extra_headers=extra_headers,
        )

    @staticmethod
    def _read_value(row: RowMapping, column: str | None) -> Any:
        if column is None:
            return None
        return row.get(column)

    @staticmethod
    def _read_str(row: RowMapping, column: str | None, default: str) -> str:
        if column is None:
            return default
        value = row.get(column)
        if value is None:
            return default
        return str(value)

    @staticmethod
    def _parse_headers(value: Any) -> dict[str, str]:
        if value is None:
            return {}
        if isinstance(value, dict):
            return {str(k): str(v) for k, v in value.items()}
        if isinstance(value, str):
            text = value.strip()
            if not text:
                return {}
            try:
                parsed = json.loads(text)
            except json.JSONDecodeError:
                return {}
            if isinstance(parsed, dict):
                return {str(k): str(v) for k, v in parsed.items()}
        return {}

    @staticmethod
    def _normalize_path(path: str) -> str:
        value = path.strip()
        if not value:
            return "/"
        if value.startswith("/"):
            return value
        return f"/{value}"
