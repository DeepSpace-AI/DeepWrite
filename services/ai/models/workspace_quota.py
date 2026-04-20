from datetime import datetime
from uuid import UUID, uuid4

from sqlalchemy import DateTime, Integer, JSON, String, Boolean
from sqlalchemy.dialects.postgresql import UUID as PGUUID
from sqlalchemy.orm import Mapped, mapped_column

from models.base import Base


class AIWorkspaceQuota(Base):
    __tablename__ = "dw_ai_workspace_quotas"

    id: Mapped[UUID] = mapped_column(PGUUID(as_uuid=True), primary_key=True, default=uuid4)
    workspace_id: Mapped[UUID] = mapped_column(PGUUID(as_uuid=True), unique=True, nullable=False)
    monthly_token_limit: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    monthly_token_used: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    monthly_request_limit: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    monthly_request_used: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    enabled_models: Mapped[list | None] = mapped_column(JSON, nullable=True)
    reset_at: Mapped[datetime | None] = mapped_column(DateTime, nullable=True)
    created_at: Mapped[datetime] = mapped_column(DateTime, nullable=False, default=datetime.utcnow)
    updated_at: Mapped[datetime] = mapped_column(DateTime, nullable=False, default=datetime.utcnow, onupdate=datetime.utcnow)