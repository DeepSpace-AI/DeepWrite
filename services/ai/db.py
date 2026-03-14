from __future__ import annotations

from sqlalchemy.ext.asyncio import AsyncEngine, create_async_engine

from config import AIServiceConfig


def build_engine(config: AIServiceConfig) -> AsyncEngine | None:
    if not config.database_dsn.strip():
        return None
    return create_async_engine(
        config.database_dsn,
        pool_pre_ping=True,
    )
