from __future__ import annotations

import base64
import hashlib
from cryptography.hazmat.primitives.ciphers.aead import AESGCM


class CryptoService:
    def __init__(self, secret_key: str):
        key_hash = hashlib.sha256(secret_key.encode()).digest()
        self._key = key_hash
        self._aesgcm = AESGCM(self._key)

    def decrypt(self, ciphertext: str) -> str:
        if not ciphertext:
            return ""
        try:
            data = base64.b64decode(ciphertext)
            nonce = data[:12]
            encrypted = data[12:]
            plaintext = self._aesgcm.decrypt(nonce, encrypted, None)
            return plaintext.decode()
        except Exception:
            return ciphertext

    def encrypt(self, plaintext: str) -> str:
        if not plaintext:
            return ""
        import os
        nonce = os.urandom(12)
        encrypted = self._aesgcm.encrypt(nonce, plaintext.encode(), None)
        return base64.b64encode(nonce + encrypted).decode()


_crypto_service: CryptoService | None = None


def init_crypto_service(secret_key: str) -> None:
    global _crypto_service
    _crypto_service = CryptoService(secret_key)


def get_crypto_service() -> CryptoService | None:
    return _crypto_service


def decrypt_api_key(ciphertext: str) -> str:
    if _crypto_service is None:
        return ciphertext
    return _crypto_service.decrypt(ciphertext)


def encrypt_api_key(plaintext: str) -> str:
    if _crypto_service is None:
        return plaintext
    return _crypto_service.encrypt(plaintext)