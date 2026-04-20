from models import ProviderModel
from services.base_client import BaseAIClient
from services.openai_client import OpenAICompatibleClient


def create_client(config: ProviderModel, decrypted_api_key: str) -> BaseAIClient:
    provider = (config.provider or "").lower()

    return OpenAICompatibleClient(config, decrypted_api_key)


__all__ = ["create_client", "BaseAIClient", "OpenAICompatibleClient"]