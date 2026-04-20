import logging
from fastapi import APIRouter, Request
from pydantic import BaseModel

from core import model_config_cache

logger = logging.getLogger(__name__)


class CacheInvalidateRequest(BaseModel):
    model: str | None = None
    all: bool = False


async def get_db(request: Request):
    return request.state.db


async def invalidate_cache(
    request: CacheInvalidateRequest,
    db = None,
) -> dict:
    if request.all:
        await model_config_cache.invalidate_all()
        return {"success": True, "message": "All model configs cache invalidated"}

    if request.model:
        await model_config_cache.invalidate(request.model)
        return {"success": True, "message": f"Model {request.model} cache invalidated"}

    return {"success": False, "message": "No action specified"}


async def refresh_cache(db = None) -> dict:
    await model_config_cache.refresh_all(db)
    return {
        "success": True,
        "cached_models": model_config_cache.get_cached_models(),
    }


router = APIRouter()
router.add_api_route("/invalidate", invalidate_cache, methods=["POST"])
router.add_api_route("/refresh", refresh_cache, methods=["POST"])