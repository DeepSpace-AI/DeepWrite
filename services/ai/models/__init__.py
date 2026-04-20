from datetime import datetime
from uuid import UUID

from sqlalchemy import DateTime, Integer, String, Boolean, JSON
from sqlalchemy.dialects.postgresql import UUID as PGUUID
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column


class Base(DeclarativeBase):
    pass


class Provider(Base):
    __tablename__ = "dw_ai_providers"

    id: Mapped[UUID] = mapped_column(PGUUID(as_uuid=True), primary_key=True)
    name: Mapped[str] = mapped_column(String(64), unique=True, nullable=False)
    base_url: Mapped[str] = mapped_column(String(1024), nullable=False)
    api_key: Mapped[str] = mapped_column(String, nullable=False)
    organization: Mapped[str | None] = mapped_column(String(255))
    chat_completions_path: Mapped[str] = mapped_column(String(255), nullable=False, default="/chat/completions")
    chat_responses_path: Mapped[str] = mapped_column(String(255), nullable=False, default="/responses")
    embeddings_path: Mapped[str] = mapped_column(String(255), nullable=False, default="/embeddings")
    rerank_path: Mapped[str] = mapped_column(String(255), nullable=False, default="/rerank")
    audio_speech_path: Mapped[str] = mapped_column(String(255), nullable=False, default="/audio/speech")
    audio_transcriptions_path: Mapped[str] = mapped_column(String(255), nullable=False, default="/audio/transcriptions")
    models_path: Mapped[str] = mapped_column(String(255), nullable=False, default="/models")
    extra_headers: Mapped[dict] = mapped_column(JSON, nullable=False, default=dict)
    enabled: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    created_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)
    updated_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)


class ProviderModel(Base):
    __tablename__ = "dw_ai_models"

    id: Mapped[UUID] = mapped_column(PGUUID(as_uuid=True), primary_key=True)
    provider_id: Mapped[UUID | None] = mapped_column(PGUUID(as_uuid=True), index=True)
    model: Mapped[str] = mapped_column(String(120), unique=True, nullable=False)
    request_model: Mapped[str] = mapped_column(String(120), nullable=False)
    supports_chat_completions: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    supports_chat_responses: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    supports_embeddings: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    supports_rerank: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    supports_audio_speech: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    supports_audio_transcriptions: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    supports_models: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    provider: Mapped[str] = mapped_column(String(64), nullable=False, default="openai-compatible")
    base_url: Mapped[str] = mapped_column(String(1024), nullable=False)
    api_key: Mapped[str] = mapped_column(String, nullable=False)
    organization: Mapped[str | None] = mapped_column(String(255))
    chat_completions_path: Mapped[str] = mapped_column(String(255), nullable=False, default="/chat/completions")
    chat_responses_path: Mapped[str] = mapped_column(String(255), nullable=False, default="/responses")
    embeddings_path: Mapped[str] = mapped_column(String(255), nullable=False, default="/embeddings")
    rerank_path: Mapped[str] = mapped_column(String(255), nullable=False, default="/rerank")
    audio_speech_path: Mapped[str] = mapped_column(String(255), nullable=False, default="/audio/speech")
    audio_transcriptions_path: Mapped[str] = mapped_column(String(255), nullable=False, default="/audio/transcriptions")
    models_path: Mapped[str] = mapped_column(String(255), nullable=False, default="/models")
    extra_headers: Mapped[dict] = mapped_column(JSON, nullable=False, default=dict)
    enabled: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    created_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)
    updated_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)


from models.request_log import AIRequestLog
from models.usage_stats import AIUsageStats
from models.workspace_quota import AIWorkspaceQuota

__all__ = ["Base", "Provider", "ProviderModel", "AIRequestLog", "AIUsageStats", "AIWorkspaceQuota"]