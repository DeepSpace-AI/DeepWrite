from __future__ import annotations

import os
import re
from dataclasses import dataclass, field
from pathlib import Path
from threading import RLock
from typing import Any

import yaml

DEFAULT_CONFIG_PATH = Path(__file__).resolve().parent / "config.yaml"
IDENTIFIER_PATTERN = re.compile(r"^[A-Za-z_][A-Za-z0-9_]*$")


def _to_bool(value: Any, default: bool) -> bool:
    if isinstance(value, bool):
        return value
    if value is None:
        return default
    text = str(value).strip().lower()
    if text in {"1", "true", "yes", "on"}:
        return True
    if text in {"0", "false", "no", "off"}:
        return False
    return default


def _to_int(value: Any, default: int) -> int:
    try:
        return int(value)
    except (TypeError, ValueError):
        return default


def _to_str(value: Any, default: str) -> str:
    if value is None:
        return default
    return str(value)


def _deep_get(data: dict[str, Any], *keys: str, default: Any = None) -> Any:
    current: Any = data
    for key in keys:
        if not isinstance(current, dict):
            return default
        if key not in current:
            return default
        current = current[key]
    return current


def _load_config_file(path: Path) -> dict[str, Any]:
    if not path.exists():
        return {}
    content = path.read_text(encoding="utf-8")
    if not content.strip():
        return {}
    loaded = yaml.safe_load(content)
    if isinstance(loaded, dict):
        return loaded
    return {}


def normalize_database_dsn(dsn: str) -> str:
    value = dsn.strip()
    if value.startswith("postgresql+psycopg://"):
        return value.replace("postgresql+psycopg://", "postgresql+asyncpg://", 1)
    if value.startswith("postgresql://"):
        return value.replace("postgresql://", "postgresql+asyncpg://", 1)
    return value


@dataclass(frozen=True, slots=True)
class AIServiceConfig:
    app_name: str = "ai-service"
    app_env: str = "development"
    host: str = "0.0.0.0"
    port: int = 8010
    reload: bool = True
    log_level: str = "info"
    internal_token: str = ""
    database_dsn: str = ""
    encryption_key: str = ""
    request_timeout_seconds: int = 120
    model_config_table: str = "dw_ai_models"
    default_chat_completions_path: str = "/chat/completions"
    default_chat_responses_path: str = "/responses"
    default_embeddings_path: str = "/embeddings"
    default_rerank_path: str = "/rerank"
    default_audio_speech_path: str = "/audio/speech"
    default_audio_transcriptions_path: str = "/audio/transcriptions"
    default_models_path: str = "/models"
    raw: dict[str, Any] = field(default_factory=dict, repr=False)

    @classmethod
    def from_source(cls, data: dict[str, Any], env: dict[str, str]) -> "AIServiceConfig":
        defaults = cls()
        app_name = _to_str(data.get("APP_NAME"), defaults.app_name)
        app_env = _to_str(data.get("APP_ENV"), defaults.app_env)
        host = _to_str(_deep_get(data, "SERVER", "HOST", default=data.get("HOST")), defaults.host)
        port = _to_int(_deep_get(data, "SERVER", "PORT", default=data.get("PORT")), defaults.port)
        reload = _to_bool(_deep_get(data, "SERVER", "RELOAD", default=data.get("RELOAD")), defaults.reload)
        log_level = _to_str(_deep_get(data, "LOG", "LEVEL", default=data.get("LOG_LEVEL")), defaults.log_level)
        internal_token = _to_str(data.get("INTERNAL_TOKEN"), defaults.internal_token)
        database_dsn = _to_str(data.get("DATABASE_DSN"), defaults.database_dsn)
        encryption_key = _to_str(data.get("ENCRYPTION_KEY"), defaults.encryption_key)
        request_timeout_seconds = _to_int(
            _deep_get(data, "HTTP", "REQUEST_TIMEOUT_SECONDS", default=data.get("REQUEST_TIMEOUT_SECONDS")),
            defaults.request_timeout_seconds,
        )
        model_config_table = _to_str(
            _deep_get(data, "MODEL", "CONFIG_TABLE", default=data.get("MODEL_CONFIG_TABLE")),
            defaults.model_config_table,
        )
        default_chat_completions_path = _to_str(
            _deep_get(
                data,
                "MODEL",
                "DEFAULT_CHAT_COMPLETIONS_PATH",
                default=_deep_get(data, "MODEL", "DEFAULT_CHAT_PATH", default=data.get("DEFAULT_CHAT_PATH")),
            ),
            defaults.default_chat_completions_path,
        )
        default_chat_responses_path = _to_str(
            _deep_get(data, "MODEL", "DEFAULT_CHAT_RESPONSES_PATH", default=data.get("DEFAULT_CHAT_RESPONSES_PATH")),
            defaults.default_chat_responses_path,
        )
        default_embeddings_path = _to_str(
            _deep_get(data, "MODEL", "DEFAULT_EMBEDDINGS_PATH", default=data.get("DEFAULT_EMBEDDINGS_PATH")),
            defaults.default_embeddings_path,
        )
        default_rerank_path = _to_str(
            _deep_get(data, "MODEL", "DEFAULT_RERANK_PATH", default=data.get("DEFAULT_RERANK_PATH")),
            defaults.default_rerank_path,
        )
        default_audio_speech_path = _to_str(
            _deep_get(data, "MODEL", "DEFAULT_AUDIO_SPEECH_PATH", default=data.get("DEFAULT_AUDIO_SPEECH_PATH")),
            defaults.default_audio_speech_path,
        )
        default_audio_transcriptions_path = _to_str(
            _deep_get(
                data,
                "MODEL",
                "DEFAULT_AUDIO_TRANSCRIPTIONS_PATH",
                default=data.get("DEFAULT_AUDIO_TRANSCRIPTIONS_PATH"),
            ),
            defaults.default_audio_transcriptions_path,
        )
        default_models_path = _to_str(
            _deep_get(data, "MODEL", "DEFAULT_MODELS_PATH", default=data.get("DEFAULT_MODELS_PATH")),
            defaults.default_models_path,
        )

        app_name = _to_str(env.get("AI_APP_NAME", app_name), app_name)
        app_env = _to_str(env.get("AI_ENV", app_env), app_env)
        host = _to_str(env.get("AI_HOST", host), host)
        port = _to_int(env.get("AI_PORT", port), port)
        reload = _to_bool(env.get("AI_RELOAD", reload), reload)
        log_level = _to_str(env.get("AI_LOG_LEVEL", log_level), log_level)
        internal_token = _to_str(env.get("AI_INTERNAL_TOKEN", internal_token), internal_token)
        database_dsn = _to_str(env.get("DATABASE_DSN", database_dsn), database_dsn)
        encryption_key = _to_str(env.get("AI_ENCRYPTION_KEY", encryption_key), encryption_key)
        request_timeout_seconds = _to_int(
            env.get("AI_REQUEST_TIMEOUT_SECONDS", request_timeout_seconds),
            request_timeout_seconds,
        )
        model_config_table = _to_str(env.get("AI_MODEL_CONFIG_TABLE", model_config_table), model_config_table)
        default_chat_completions_path = _to_str(
            env.get(
                "AI_DEFAULT_CHAT_COMPLETIONS_PATH",
                env.get("AI_DEFAULT_CHAT_PATH", default_chat_completions_path),
            ),
            default_chat_completions_path,
        )
        default_chat_responses_path = _to_str(
            env.get("AI_DEFAULT_CHAT_RESPONSES_PATH", default_chat_responses_path),
            default_chat_responses_path,
        )
        default_embeddings_path = _to_str(
            env.get("AI_DEFAULT_EMBEDDINGS_PATH", default_embeddings_path),
            default_embeddings_path,
        )
        default_rerank_path = _to_str(env.get("AI_DEFAULT_RERANK_PATH", default_rerank_path), default_rerank_path)
        default_audio_speech_path = _to_str(
            env.get("AI_DEFAULT_AUDIO_SPEECH_PATH", default_audio_speech_path),
            default_audio_speech_path,
        )
        default_audio_transcriptions_path = _to_str(
            env.get("AI_DEFAULT_AUDIO_TRANSCRIPTIONS_PATH", default_audio_transcriptions_path),
            default_audio_transcriptions_path,
        )
        default_models_path = _to_str(env.get("AI_DEFAULT_MODELS_PATH", default_models_path), default_models_path)

        return cls(
            app_name=app_name,
            app_env=app_env,
            host=host,
            port=port,
            reload=reload,
            log_level=log_level,
            internal_token=internal_token,
            database_dsn=normalize_database_dsn(database_dsn),
            encryption_key=encryption_key,
            request_timeout_seconds=request_timeout_seconds,
            model_config_table=model_config_table,
            default_chat_completions_path=default_chat_completions_path,
            default_chat_responses_path=default_chat_responses_path,
            default_embeddings_path=default_embeddings_path,
            default_rerank_path=default_rerank_path,
            default_audio_speech_path=default_audio_speech_path,
            default_audio_transcriptions_path=default_audio_transcriptions_path,
            default_models_path=default_models_path,
            raw=data,
        )

    def validated_model_table(self) -> str:
        table = self.model_config_table.strip()
        if not IDENTIFIER_PATTERN.fullmatch(table):
            raise ValueError(f"Invalid AI model config table name: {table}")
        return table


_config_lock = RLock()
_cached_config: AIServiceConfig | None = None


def load_config(config_path: str | Path | None = None) -> AIServiceConfig:
    target = Path(config_path) if config_path else DEFAULT_CONFIG_PATH
    file_data = _load_config_file(target)
    return AIServiceConfig.from_source(file_data, dict(os.environ))


def get_config(config_path: str | Path | None = None, force_reload: bool = False) -> AIServiceConfig:
    global _cached_config
    with _config_lock:
        if _cached_config is None or force_reload:
            _cached_config = load_config(config_path=config_path)
        return _cached_config
