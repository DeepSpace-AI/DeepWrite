from __future__ import annotations

import os
from dataclasses import dataclass, field
from pathlib import Path
from threading import RLock
from typing import Any

import yaml

DEFAULT_CONFIG_PATH = Path(__file__).resolve().parents[1] / "config.yaml"


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


@dataclass(frozen=True, slots=True)
class WorkerConfig:
    """Worker 运行时配置对象。"""

    app_name: str = "worker-service"
    app_env: str = "development"
    host: str = "0.0.0.0"
    port: int = 8000
    reload: bool = True
    log_level: str = "info"
    internal_token: str = ""
    database_dsn: str = ""
    redis_url: str = ""
    gateway_url: str = "http://localhost:8080"
    gateway_token: str = ""
    gateway_callback_url: str = ""
    gateway_callback_token: str = ""
    max_workspace_concurrency: int = 8
    max_user_concurrency: int = 3
    celery_broker_url: str = "redis://localhost:6379/1"
    celery_result_backend: str = "redis://localhost:6379/2"
    celery_default_queue: str = "worker.default"
    raw: dict[str, Any] = field(default_factory=dict, repr=False)

    @classmethod
    def from_source(cls, data: dict[str, Any], env: dict[str, str]) -> "WorkerConfig":
        defaults = cls()

        # 文件配置取值
        app_name = _to_str(data.get("APP_NAME"), defaults.app_name)
        app_env = _to_str(data.get("APP_ENV"), defaults.app_env)
        host = _to_str(_deep_get(data, "SERVER", "HOST", default=data.get("HOST")), defaults.host)
        port = _to_int(_deep_get(data, "SERVER", "PORT", default=data.get("PORT")), defaults.port)
        reload = _to_bool(
            _deep_get(data, "SERVER", "RELOAD", default=data.get("RELOAD")),
            defaults.reload,
        )
        log_level = _to_str(
            _deep_get(data, "LOG", "LEVEL", default=data.get("LOG_LEVEL")),
            defaults.log_level,
        )
        internal_token = _to_str(data.get("INTERNAL_TOKEN"), defaults.internal_token)
        database_dsn = _to_str(data.get("DATABASE_DSN"), defaults.database_dsn)
        redis_url = _to_str(data.get("REDIS_URL"), defaults.redis_url)
        gateway_callback_url = _to_str(
            _deep_get(data, "GATEWAY", "CALLBACK_URL", default=data.get("GATEWAY_CALLBACK_URL")),
            defaults.gateway_callback_url,
        )
        gateway_callback_token = _to_str(
            _deep_get(data, "GATEWAY", "CALLBACK_TOKEN", default=data.get("GATEWAY_CALLBACK_TOKEN")),
            defaults.gateway_callback_token,
        )
        gateway_url = _to_str(
            _deep_get(data, "GATEWAY", "URL", default=data.get("GATEWAY_URL")),
            defaults.gateway_url,
        )
        gateway_token = _to_str(
            _deep_get(data, "GATEWAY", "TOKEN", default=data.get("GATEWAY_TOKEN")),
            defaults.gateway_token,
        )
        max_workspace_concurrency = _to_int(
            _deep_get(
                data,
                "QUOTA",
                "MAX_WORKSPACE_CONCURRENCY",
                default=data.get("MAX_WORKSPACE_CONCURRENCY"),
            ),
            defaults.max_workspace_concurrency,
        )
        max_user_concurrency = _to_int(
            _deep_get(data, "QUOTA", "MAX_USER_CONCURRENCY", default=data.get("MAX_USER_CONCURRENCY")),
            defaults.max_user_concurrency,
        )
        celery_broker_url = _to_str(
            _deep_get(data, "CELERY", "BROKER_URL", default=data.get("CELERY_BROKER_URL")),
            defaults.celery_broker_url,
        )
        celery_result_backend = _to_str(
            _deep_get(data, "CELERY", "RESULT_BACKEND", default=data.get("CELERY_RESULT_BACKEND")),
            defaults.celery_result_backend,
        )
        celery_default_queue = _to_str(
            _deep_get(data, "CELERY", "DEFAULT_QUEUE", default=data.get("CELERY_DEFAULT_QUEUE")),
            defaults.celery_default_queue,
        )

        # 环境变量覆盖（优先级最高）
        app_name = _to_str(env.get("WORKER_APP_NAME", app_name), app_name)
        app_env = _to_str(env.get("WORKER_ENV", app_env), app_env)
        host = _to_str(env.get("WORKER_HOST", host), host)
        port = _to_int(env.get("WORKER_PORT", port), port)
        reload = _to_bool(env.get("WORKER_RELOAD", reload), reload)
        log_level = _to_str(env.get("WORKER_LOG_LEVEL", log_level), log_level)
        internal_token = _to_str(env.get("WORKER_INTERNAL_TOKEN", internal_token), internal_token)
        database_dsn = _to_str(env.get("DATABASE_DSN", database_dsn), database_dsn)
        redis_url = _to_str(env.get("REDIS_URL", redis_url), redis_url)
        gateway_callback_url = _to_str(
            env.get("GATEWAY_CALLBACK_URL", gateway_callback_url),
            gateway_callback_url,
        )
        gateway_callback_token = _to_str(
            env.get("GATEWAY_CALLBACK_TOKEN", gateway_callback_token),
            gateway_callback_token,
        )
        gateway_url = _to_str(env.get("GATEWAY_URL", gateway_url), gateway_url)
        gateway_token = _to_str(env.get("GATEWAY_TOKEN", gateway_token), gateway_token)
        max_workspace_concurrency = _to_int(
            env.get("MAX_WORKSPACE_CONCURRENCY", max_workspace_concurrency),
            max_workspace_concurrency,
        )
        max_user_concurrency = _to_int(
            env.get("MAX_USER_CONCURRENCY", max_user_concurrency),
            max_user_concurrency,
        )
        celery_broker_url = _to_str(env.get("CELERY_BROKER_URL", celery_broker_url), celery_broker_url)
        celery_result_backend = _to_str(
            env.get("CELERY_RESULT_BACKEND", celery_result_backend),
            celery_result_backend,
        )
        celery_default_queue = _to_str(
            env.get("CELERY_DEFAULT_QUEUE", celery_default_queue),
            celery_default_queue,
        )

        return cls(
            app_name=app_name,
            app_env=app_env,
            host=host,
            port=port,
            reload=reload,
            log_level=log_level,
            internal_token=internal_token,
            database_dsn=database_dsn,
            redis_url=redis_url,
            gateway_url=gateway_url,
            gateway_token=gateway_token,
            gateway_callback_url=gateway_callback_url,
            gateway_callback_token=gateway_callback_token,
            max_workspace_concurrency=max_workspace_concurrency,
            max_user_concurrency=max_user_concurrency,
            celery_broker_url=celery_broker_url,
            celery_result_backend=celery_result_backend,
            celery_default_queue=celery_default_queue,
            raw=data,
        )


_config_lock = RLock()
_cached_config: WorkerConfig | None = None


def load_config(config_path: str | Path | None = None) -> WorkerConfig:
    """读取配置文件并叠加环境变量，返回最终配置。"""
    target = Path(config_path) if config_path else DEFAULT_CONFIG_PATH
    file_data = _load_config_file(target)
    return WorkerConfig.from_source(file_data, dict(os.environ))


def get_config(config_path: str | Path | None = None, force_reload: bool = False) -> WorkerConfig:
    """获取全局缓存配置，默认只加载一次。"""
    global _cached_config

    with _config_lock:
        if _cached_config is None or force_reload:
            _cached_config = load_config(config_path=config_path)
        return _cached_config


def reload_config(config_path: str | Path | None = None) -> WorkerConfig:
    """强制重新加载配置。"""
    return get_config(config_path=config_path, force_reload=True)
