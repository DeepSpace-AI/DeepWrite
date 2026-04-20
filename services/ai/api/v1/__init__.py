from fastapi import APIRouter

from api.v1 import chat, embeddings, rerank, audio

router = APIRouter()

router.add_api_route("/chat/completions", chat.chat_completions, methods=["POST"], response_model=None)
router.add_api_route("/embeddings", embeddings.embeddings, methods=["POST"])
router.add_api_route("/rerank", rerank.rerank, methods=["POST"])
router.add_api_route("/audio/speech", audio.speech, methods=["POST"], response_class=None)
router.add_api_route("/audio/transcriptions", audio.speech, methods=["POST"])