import base64
import logging
import os

from cryptography.hazmat.backends import default_backend
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes

logger = logging.getLogger(__name__)


def get_encryption_key() -> bytes:
    from config import get_config

    cfg = get_config()
    key = cfg.security.api_key_encryption_key
    if not key:
        raise ValueError("API_KEY_ENCRYPTION_KEY is not configured")
    key_bytes = key.encode("utf-8")
    if len(key_bytes) < 32:
        key_bytes = key_bytes.ljust(32, b"0")
    return key_bytes[:32]


def encrypt_api_key(plaintext: str) -> str:
    if not plaintext:
        return ""

    key = get_encryption_key()
    iv = os.urandom(12)
    plaintext_bytes = plaintext.encode("utf-8")

    cipher = Cipher(algorithms.AES(key), modes.GCM(iv), backend=default_backend())
    encryptor = cipher.encryptor()
    ciphertext = encryptor.update(plaintext_bytes) + encryptor.finalize()

    combined = iv + encryptor.tag + ciphertext
    return base64.b64encode(combined).decode("utf-8")


def decrypt_api_key(encrypted: str) -> str:
    if not encrypted:
        return ""

    try:
        key = get_encryption_key()
        combined = base64.b64decode(encrypted.encode("utf-8"))

        iv = combined[:12]
        tag = combined[12:28]
        ciphertext = combined[28:]

        cipher = Cipher(algorithms.AES(key), modes.GCM(iv, tag), backend=default_backend())
        decryptor = cipher.decryptor()
        plaintext_bytes = decryptor.update(ciphertext) + decryptor.finalize()

        return plaintext_bytes.decode("utf-8")
    except Exception as e:
        logger.error(f"Failed to decrypt API key: {e}")
        raise ValueError("Failed to decrypt API key") from e