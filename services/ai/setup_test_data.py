"""Create test data for AI service testing"""

import asyncio
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from sqlalchemy import text
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession, async_sessionmaker
from config import get_config
from core import encrypt_api_key
from models import Base, Provider, ProviderModel, AIWorkspaceQuota
from uuid import uuid4


async def create_tables(engine):
    """Create all tables"""
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    print("Tables created successfully")


async def insert_test_data(session: AsyncSession):
    """Insert test data"""
    cfg = get_config()
    
    # Create test provider
    provider_id = uuid4()
    encrypted_key = encrypt_api_key("sk-test-api-key-12345")
    
    provider = Provider(
        id=provider_id,
        name="test-openai",
        base_url="https://api.openai.com",
        api_key=encrypted_key,
        organization=None,
        chat_completions_path="/v1/chat/completions",
        embeddings_path="/v1/embeddings",
        rerank_path="/v1/rerank",
        audio_speech_path="/v1/audio/speech",
        audio_transcriptions_path="/v1/audio/transcriptions",
        models_path="/v1/models",
        extra_headers={},
        enabled=True,
    )
    session.add(provider)
    
    # Create test model
    model_id = uuid4()
    model = ProviderModel(
        id=model_id,
        provider_id=provider_id,
        model="gpt-4o-mini",
        request_model="gpt-4o-mini",
        supports_chat_completions=True,
        supports_embeddings=False,
        supports_rerank=False,
        supports_audio_speech=True,
        supports_audio_transcriptions=False,
        provider="test-openai",
        base_url="https://api.openai.com",
        api_key=encrypted_key,
        chat_completions_path="/v1/chat/completions",
        embeddings_path="/v1/embeddings",
        audio_speech_path="/v1/audio/speech",
        models_path="/v1/models",
        extra_headers={},
        enabled=True,
    )
    session.add(model)
    
    # Create test workspace quota
    workspace_id = uuid4()
    quota = AIWorkspaceQuota(
        workspace_id=workspace_id,
        monthly_token_limit=100000,
        monthly_token_used=0,
        monthly_request_limit=1000,
        monthly_request_used=0,
        enabled_models=["gpt-4o-mini"],
    )
    session.add(quota)
    
    await session.commit()
    
    print(f"Test provider created: {provider_id}")
    print(f"Test model created: gpt-4o-mini")
    print(f"Test workspace created: {workspace_id}")
    
    return provider_id, model_id, workspace_id


async def main():
    cfg = get_config()
    print(f"Connecting to database: {cfg.database.host}:{cfg.database.port}/{cfg.database.name}")
    
    engine = create_async_engine(cfg.database_url, echo=True)
    async_session = async_sessionmaker(engine, class_=AsyncSession, expire_on_commit=False)
    
    # Create tables
    await create_tables(engine)
    
    # Insert test data
    async with async_session() as session:
        await insert_test_data(session)
    
    await engine.dispose()
    print("\nTest data setup complete!")


if __name__ == "__main__":
    asyncio.run(main())