import logging

from fastapi import APIRouter

from api.internal.test_connection import test_connection
from api.internal.stats import get_workspace_stats, get_user_stats
from api.internal import router as cache_router

logger = logging.getLogger(__name__)

router = APIRouter()

router.add_api_route("/test-connection", test_connection, methods=["POST"])
router.add_api_route("/stats/workspace", get_workspace_stats, methods=["GET"])
router.add_api_route("/stats/user", get_user_stats, methods=["GET"])
router.include_router(cache_router, prefix="/cache")