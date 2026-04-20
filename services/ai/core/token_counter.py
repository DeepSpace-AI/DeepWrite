import logging
from typing import Any

import tiktoken

logger = logging.getLogger(__name__)

_encoders: dict[str, tiktoken.Encoding] = {}


def get_encoder(model: str) -> tiktoken.Encoding:
    if model in _encoders:
        return _encoders[model]

    try:
        if model.startswith("gpt-4") or model.startswith("gpt-4o"):
            encoding_name = "cl100k_base"
        elif model.startswith("gpt-3.5"):
            encoding_name = "cl100k_base"
        elif model.startswith("text-embedding"):
            encoding_name = "cl100k_base"
        elif "claude" in model.lower():
            encoding_name = "cl100k_base"
        else:
            encoding_name = "cl100k_base"

        encoder = tiktoken.get_encoding(encoding_name)
        _encoders[model] = encoder
        return encoder
    except Exception as e:
        logger.warning(f"Failed to get encoder for model {model}: {e}, using cl100k_base")
        encoder = tiktoken.get_encoding("cl100k_base")
        _encoders[model] = encoder
        return encoder


def count_tokens(text: str, model: str = "gpt-4") -> int:
    if not text:
        return 0

    try:
        encoder = get_encoder(model)
        return len(encoder.encode(text))
    except Exception as e:
        logger.error(f"Failed to count tokens: {e}")
        return len(text.split())


def count_messages_tokens(messages: list[dict[str, Any]], model: str = "gpt-4") -> int:
    if not messages:
        return 0

    tokens_per_message = 3
    tokens_per_name = 1

    total = 0
    for message in messages:
        total += tokens_per_message
        for key, value in message.items():
            if isinstance(value, str):
                total += count_tokens(value, model)
            elif isinstance(value, list):
                for item in value:
                    if isinstance(item, dict) and "text" in item:
                        total += count_tokens(item["text"], model)
            if key == "name":
                total += tokens_per_name

    total += 3
    return total