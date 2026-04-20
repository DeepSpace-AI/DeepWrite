import asyncio
import logging
from typing import Any

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from models import ProviderModel

logger = logging.getLogger(__name__)


class ModelConfigCache:
    def __init__(self):
        self._cache: dict[str, ProviderModel] = {}
        self._lock = asyncio.Lock()

    async def get_model(self, db: AsyncSession, model: str) -> ProviderModel | None:
        model = model.strip()
        if not model:
            return None

        if model in self._cache:
            return self._cache[model]

        async with self._lock:
            if model in self._cache:
                return self._cache[model]

            result = await db.execute(
                select(ProviderModel).where(
                    ProviderModel.model == model,
                    ProviderModel.enabled == True,
                )
            )
            record = result.scalar_one_or_none()
            if record:
                self._cache[model] = record
            return record

    async def get_model_by_id(self, db: AsyncSession, model_id: str) -> ProviderModel | None:
        for cached_model in self._cache.values():
            if str(cached_model.id) == model_id:
                return cached_model

        result = await db.execute(
            select(ProviderModel).where(ProviderModel.id == model_id, ProviderModel.enabled == True)
        )
        return result.scalar_one_or_none()

    async def refresh_all(self, db: AsyncSession) -> None:
        async with self._lock:
            result = await db.execute(select(ProviderModel).where(ProviderModel.enabled == True))
            models = result.scalars().all()
            self._cache = {m.model: m for m in models}
            logger.info(f"ModelConfigCache refreshed, {len(self._cache)} models cached")

    async def invalidate(self, model: str) -> None:
        async with self._lock:
            self._cache.pop(model.strip(), None)
            logger.debug(f"ModelConfigCache invalidated model: {model}")

    async def invalidate_all(self) -> None:
        async with self._lock:
            self._cache.clear()
            logger.info("ModelConfigCache cleared")

    def get_cached_models(self) -> list[str]:
        return list(self._cache.keys())


model_config_cache = ModelConfigCache()