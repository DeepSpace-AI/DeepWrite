import logging
from fastapi import Request
from fastapi.responses import JSONResponse
from starlette.middleware.base import BaseHTTPMiddleware

from config import get_config

logger = logging.getLogger(__name__)


class InternalAuthMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        if request.url.path.startswith("/internal/"):
            cfg = get_config()
            internal_token = request.headers.get("X-Internal-Token", "")
            expected_token = cfg.security.internal_token

            if not expected_token:
                logger.warning("Internal token not configured, skipping internal auth")
                return await call_next(request)

            if not internal_token or internal_token != expected_token:
                return JSONResponse(
                    status_code=401,
                    content={"error": "Unauthorized", "message": "Invalid or missing internal token"},
                )

        return await call_next(request)