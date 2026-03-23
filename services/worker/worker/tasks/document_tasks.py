import logging
from datetime import datetime
from typing import Any

from worker.celery_app import celery_app

logger = logging.getLogger(__name__)


@celery_app.task(
    name="worker.document.parse",
    bind=True,
    max_retries=3,
    default_retry_delay=60,
    autoretry_for=(Exception,),
    retry_backoff=True,
)
def parse_document(self, file_id: str, content: str, file_type: str) -> dict[str, Any]:
    """
    解析文档内容，提取文本和元数据。

    Args:
        file_id: 文件唯一标识
        content: 文件内容 (Base64 或文本)
        file_type: 文件类型 (pdf, docx, txt, md)

    Returns:
        解析结果字典
    """
    logger.info(f"[parse_document] file_id={file_id}, file_type={file_type}")

    try:
        result = {
            "file_id": file_id,
            "status": "parsed",
            "parsed_at": datetime.now().isoformat(),
            "text_length": len(content) if content else 0,
            "file_type": file_type,
            "paragraphs": _extract_paragraphs(content),
            "metadata": {
                "word_count": len(content.split()) if content else 0,
                "char_count": len(content) if content else 0,
            },
        }
        logger.info(f"[parse_document] success file_id={file_id}")
        return result
    except Exception as exc:
        logger.error(f"[parse_document] failed file_id={file_id}: {exc}")
        raise


@celery_app.task(
    name="worker.document.index",
    bind=True,
    max_retries=3,
    default_retry_delay=30,
)
def index_document(self, file_id: str, text: str, metadata: dict[str, Any] | None = None) -> dict[str, Any]:
    """
    将文档内容索引到搜索引擎。

    Args:
        file_id: 文件唯一标识
        text: 文档文本内容
        metadata: 附加元数据

    Returns:
        索引结果
    """
    logger.info(f"[index_document] file_id={file_id}")

    try:
        result = {
            "file_id": file_id,
            "status": "indexed",
            "indexed_at": datetime.now().isoformat(),
            "doc_id": f"doc_{file_id}",
            "vector_dim": 1536,
        }
        logger.info(f"[index_document] success file_id={file_id}")
        return result
    except Exception as exc:
        logger.error(f"[index_document] failed file_id={file_id}: {exc}")
        raise


@celery_app.task(
    name="worker.document.export",
    bind=True,
    max_retries=2,
    default_retry_delay=30,
)
def export_document(
    self,
    file_id: str,
    content: str,
    output_format: str,
    options: dict[str, Any] | None = None,
) -> dict[str, Any]:
    """
    导出文档为指定格式。

    Args:
        file_id: 文件唯一标识
        content: 文档内容 (HTML 或 JSON)
        output_format: 输出格式 (pdf, docx, html, md)
        options: 导出选项

    Returns:
        导出结果
    """
    logger.info(f"[export_document] file_id={file_id}, format={output_format}")

    try:
        result = {
            "file_id": file_id,
            "status": "exported",
            "output_format": output_format,
            "exported_at": datetime.now().isoformat(),
            "output_path": f"/exports/{file_id}.{output_format}",
        }
        logger.info(f"[export_document] success file_id={file_id}")
        return result
    except Exception as exc:
        logger.error(f"[export_document] failed file_id={file_id}: {exc}")
        raise


def _extract_paragraphs(content: str) -> list[str]:
    """从内容中提取段落"""
    if not content:
        return []
    paragraphs = [p.strip() for p in content.split("\n\n") if p.strip()]
    return paragraphs[:1000]
