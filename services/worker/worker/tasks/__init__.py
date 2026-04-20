from importlib import import_module
from pkgutil import iter_modules

for module in iter_modules(__path__, prefix=f"{__name__}."):
    import_module(module.name)

from worker.tasks.math_tasks import add
from worker.tasks.ping_tasks import ping
from worker.tasks.document_tasks import parse_document, index_document, export_document
from worker.tasks.ai_tasks import chat_completion, batch_completion, generate_embedding
from worker.tasks.notify_tasks import send_email, send_welcome_email, send_document_shared_notification, send_task_reminder
from worker.tasks.pdf_tasks import extract_pdf_metadata, extract_and_lookup_doi

__all__ = [
    "add",
    "ping",
    "parse_document",
    "index_document",
    "export_document",
    "chat_completion",
    "batch_completion",
    "generate_embedding",
    "send_email",
    "send_welcome_email",
    "send_document_shared_notification",
    "send_task_reminder",
    "extract_pdf_metadata",
    "extract_and_lookup_doi",
]
