import logging
from dataclasses import dataclass
from datetime import datetime

from sqlalchemy import select, update
from sqlalchemy.ext.asyncio import AsyncSession

from models import AIWorkspaceQuota

logger = logging.getLogger(__name__)


@dataclass
class QuotaResult:
    allowed: bool
    reason: str | None = None
    remaining_tokens: int = 0
    remaining_requests: int = 0
    monthly_token_limit: int = 0
    monthly_request_limit: int = 0


class QuotaChecker:
    def __init__(self, db: AsyncSession):
        self.db = db

    async def get_workspace_quota(self, workspace_id: str) -> AIWorkspaceQuota | None:
        result = await self.db.execute(
            select(AIWorkspaceQuota).where(AIWorkspaceQuota.workspace_id == workspace_id)
        )
        quota = result.scalar_one_or_none()
        if quota and quota.reset_at and quota.reset_at < datetime.utcnow():
            await self.reset_quota(quota)
            await self.db.refresh(quota)
        return quota

    async def check_quota(
        self,
        workspace_id: str,
        model: str,
        estimated_tokens: int = 0,
    ) -> QuotaResult:
        quota = await self.get_workspace_quota(workspace_id)
        if not quota:
            return QuotaResult(allowed=True, remaining_tokens=-1, remaining_requests=-1)

        remaining_tokens = quota.monthly_token_limit - quota.monthly_token_used
        remaining_requests = quota.monthly_request_limit - quota.monthly_request_used

        if quota.monthly_token_limit > 0 and remaining_tokens <= 0:
            return QuotaResult(
                allowed=False,
                reason="token_limit_exceeded",
                remaining_tokens=0,
                remaining_requests=remaining_requests,
                monthly_token_limit=quota.monthly_token_limit,
                monthly_request_limit=quota.monthly_request_limit,
            )

        if quota.monthly_request_limit > 0 and remaining_requests <= 0:
            return QuotaResult(
                allowed=False,
                reason="request_limit_exceeded",
                remaining_tokens=remaining_tokens,
                remaining_requests=0,
                monthly_token_limit=quota.monthly_token_limit,
                monthly_request_limit=quota.monthly_request_limit,
            )

        if quota.enabled_models and model not in quota.enabled_models:
            return QuotaResult(
                allowed=False,
                reason="model_not_allowed",
                remaining_tokens=remaining_tokens,
                remaining_requests=remaining_requests,
                monthly_token_limit=quota.monthly_token_limit,
                monthly_request_limit=quota.monthly_request_limit,
            )

        return QuotaResult(
            allowed=True,
            remaining_tokens=remaining_tokens,
            remaining_requests=remaining_requests,
            monthly_token_limit=quota.monthly_token_limit,
            monthly_request_limit=quota.monthly_request_limit,
        )

    async def consume_quota(self, workspace_id: str, tokens: int) -> None:
        quota = await self.get_workspace_quota(workspace_id)
        if not quota:
            return

        await self.db.execute(
            update(AIWorkspaceQuota)
            .where(AIWorkspaceQuota.workspace_id == workspace_id)
            .values(
                monthly_token_used=AIWorkspaceQuota.monthly_token_used + tokens,
                monthly_request_used=AIWorkspaceQuota.monthly_request_used + 1,
            )
        )
        await self.db.commit()

    async def reset_quota(self, quota: AIWorkspaceQuota) -> None:
        quota.monthly_token_used = 0
        quota.monthly_request_used = 0
        await self.db.commit()
        logger.info(f"Quota reset for workspace {quota.workspace_id}")