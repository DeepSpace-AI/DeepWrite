import logging
import re
from datetime import datetime
from typing import Any

import httpx

from config.config import get_config
from worker.celery_app import celery_app

logger = logging.getLogger(__name__)

DOI_PATTERN = re.compile(r"10\.\d{4,}/[^\s]+", re.IGNORECASE)
ISBN_PATTERN = re.compile(r"97[89][\d-]{10,}", re.IGNORECASE)
YEAR_PATTERN = re.compile(r"\b(19|20)\d{2}\b")
TITLE_PATTERNS = [
    re.compile(r"^[A-Z][A-Za-z\s\-:,.?!]{10,200}$", re.MULTILINE),
]


def _get_gateway_config() -> tuple[str, str]:
    """获取 Gateway URL 和 Token 配置"""
    cfg = get_config()
    return cfg.gateway_url, cfg.gateway_token


@celery_app.task(
    name="worker.pdf.extract_metadata",
    bind=True,
    max_retries=3,
    default_retry_delay=60,
    autoretry_for=(Exception,),
    retry_backoff=True,
)
def extract_pdf_metadata(self, file_id: str, content: bytes | str) -> dict[str, Any]:
    """
    从 PDF 文件中提取元数据。

    Args:
        file_id: 文件唯一标识
        content: PDF 文件内容 (bytes 或 Base64 编码字符串)

    Returns:
        提取的元数据字典
    """
    logger.info(f"[extract_pdf_metadata] file_id={file_id}")

    try:
        if isinstance(content, str):
            import base64
            content = base64.b64decode(content)

        text = _extract_text_from_pdf(content)
        metadata = _parse_metadata(text)

        result = {
            "file_id": file_id,
            "status": "extracted",
            "extracted_at": datetime.now().isoformat(),
            "text_length": len(text),
            "metadata": metadata,
            "raw_text_preview": text[:1000] if text else None,
        }

        logger.info(f"[extract_pdf_metadata] success file_id={file_id}, found DOI: {metadata.get('doi')}")
        return result

    except Exception as exc:
        logger.error(f"[extract_pdf_metadata] failed file_id={file_id}: {exc}")
        raise


@celery_app.task(
    name="worker.pdf.extract_and_lookup",
    bind=True,
    max_retries=3,
    default_retry_delay=60,
)
def extract_and_lookup_doi(self, file_id: str, content: bytes | str, reference_id: str | None = None) -> dict[str, Any]:
    """
    从 PDF 提取 DOI 并查询 Crossref 获取完整元数据。

    Args:
        file_id: 文件唯一标识
        content: PDF 文件内容
        reference_id: 关联的文献ID（可选，用于回调更新）

    Returns:
        完整的文献元数据
    """
    logger.info(f"[extract_and_lookup_doi] file_id={file_id}, reference_id={reference_id}")

    try:
        if isinstance(content, str):
            import base64
            content = base64.b64decode(content)

        text = _extract_text_from_pdf(content)
        doi = _extract_doi(text)

        result = {
            "file_id": file_id,
            "status": "partial",
            "extracted_at": datetime.now().isoformat(),
            "doi": doi,
            "text_length": len(text),
        }

        if doi:
            logger.info(f"[extract_and_lookup_doi] found DOI: {doi}, querying Crossref")
            crossref_data = _query_crossref(doi)
            if crossref_data:
                result["status"] = "complete"
                result["crossref"] = crossref_data
                result["reference"] = _build_reference(doi, crossref_data)
            else:
                result["status"] = "doi_found_no_match"
        else:
            metadata = _parse_metadata(text)
            result["metadata"] = metadata
            result["status"] = "no_doi"

        logger.info(f"[extract_and_lookup_doi] success file_id={file_id}, status={result['status']}")

        if reference_id:
            _callback_update_reference(reference_id, result)

        return result

    except Exception as exc:
        logger.error(f"[extract_and_lookup_doi] failed file_id={file_id}: {exc}")
        if reference_id:
            _callback_update_reference(reference_id, {
                "status": "failed",
                "error": str(exc),
            })
        raise


def _callback_update_reference(reference_id: str, result: dict[str, Any]) -> None:
    """回调Gateway更新文献信息"""
    gateway_url, gateway_token = _get_gateway_config()
    url = f"{gateway_url}/api/v1/internal/references/extract-callback"
    
    headers = {"Content-Type": "application/json"}
    if gateway_token:
        headers["X-Internal-Token"] = gateway_token

    payload = {
        "reference_id": reference_id,
        "status": result.get("status", "complete"),
        "reference": result.get("reference"),
        "metadata": result.get("metadata"),
        "error": result.get("error"),
    }

    try:
        with httpx.Client(timeout=10.0) as client:
            response = client.post(url, json=payload, headers=headers)
            if response.status_code == 200:
                logger.info(f"[callback] successfully updated reference {reference_id}")
            else:
                logger.warning(f"[callback] failed to update reference {reference_id}: {response.status_code}")
    except Exception as e:
        logger.error(f"[callback] error updating reference {reference_id}: {e}")


def _extract_text_from_pdf(content: bytes) -> str:
    """使用 pdfminer 从 PDF 提取文本"""
    try:
        from io import BytesIO

        from pdfminer.high_level import extract_text

        pdf_file = BytesIO(content)
        text = extract_text(pdf_file)
        return text or ""
    except ImportError:
        logger.warning("pdfminer not installed, returning empty text")
        return ""
    except Exception as e:
        logger.error(f"Failed to extract PDF text: {e}")
        return ""


def _extract_doi(text: str) -> str | None:
    """从文本中提取 DOI"""
    if not text:
        return None

    text = text[:5000]

    match = DOI_PATTERN.search(text)
    if match:
        doi = match.group(0)
        doi = re.sub(r"[.,;:)\]\s]+$", "", doi)
        return doi

    return None


def _parse_metadata(text: str) -> dict[str, Any]:
    """从文本中解析可能的元数据"""
    if not text:
        return {}

    metadata: dict[str, Any] = {}

    doi = _extract_doi(text)
    if doi:
        metadata["doi"] = doi

    isbn_matches = ISBN_PATTERN.findall(text)
    if isbn_matches:
        metadata["isbn"] = isbn_matches[0].replace("-", "")

    years = YEAR_PATTERN.findall(text)
    if years:
        years_int = [int(y) for y in years]
        recent_years = [y for y in years_int if 1990 <= y <= datetime.now().year]
        if recent_years:
            metadata["possible_years"] = sorted(set(recent_years), reverse=True)[:3]

    lines = [l.strip() for l in text.split("\n") if l.strip()]
    if lines:
        for line in lines[:10]:
            if len(line) > 20 and len(line) < 300:
                words = line.split()
                if len(words) >= 3:
                    metadata["possible_title"] = line
                    break

    return metadata


def _query_crossref(doi: str) -> dict[str, Any] | None:
    """查询 Crossref API 获取 DOI 元数据"""
    import httpx

    url = f"https://api.crossref.org/works/{doi}"

    headers = {
        "Accept": "application/json",
        "User-Agent": "DeepWrite/1.0 (mailto:support@deepwrite.work)",
    }

    try:
        with httpx.Client(timeout=30.0) as client:
            response = client.get(url, headers=headers)

            if response.status_code == 200:
                data = response.json()
                return data.get("message", {})
            elif response.status_code == 404:
                logger.warning(f"DOI not found in Crossref: {doi}")
                return None
            else:
                logger.warning(f"Crossref API error: {response.status_code}")
                return None

    except httpx.TimeoutException:
        logger.warning(f"Crossref API timeout for DOI: {doi}")
        return None
    except Exception as e:
        logger.error(f"Crossref API error: {e}")
        return None


def _build_reference(doi: str, crossref_data: dict[str, Any]) -> dict[str, Any]:
    """从 Crossref 数据构建 Reference 对象"""
    ref: dict[str, Any] = {
        "doi": doi,
        "title": "",
        "authors": [],
        "type": "unknown",
    }

    titles = crossref_data.get("title", [])
    if titles:
        ref["title"] = titles[0]

    authors_data = crossref_data.get("author", [])
    authors = []
    for a in authors_data:
        author: dict[str, str] = {}
        if a.get("family"):
            author["family"] = a["family"]
        if a.get("given"):
            author["given"] = a["given"]
        if a.get("name"):
            author["literal"] = a["name"]
        if a.get("ORCID"):
            author["orcid"] = a["ORCID"].replace("https://orcid.org/", "")
        authors.append(author)
    ref["authors"] = authors

    published = crossref_data.get("published-print") or crossref_data.get("published-online")
    if published and published.get("date-parts"):
        date_parts = published["date-parts"][0]
        if date_parts:
            ref["year"] = date_parts[0]

    containers = crossref_data.get("container-title", [])
    if containers:
        ref["source"] = containers[0]

    cr_type = crossref_data.get("type", "")
    ref["type"] = _map_crossref_type(cr_type)

    ref["abstract"] = crossref_data.get("abstract", "")
    ref["volume"] = crossref_data.get("volume", "")
    ref["issue"] = crossref_data.get("issue", "")
    ref["pages"] = crossref_data.get("page", "")
    ref["publisher"] = crossref_data.get("publisher", "")
    ref["url"] = crossref_data.get("URL", "")

    return ref


def _map_crossref_type(cr_type: str) -> str:
    """映射 Crossref 类型到 Reference 类型"""
    type_map = {
        "journal-article": "article",
        "article": "article",
        "book": "book",
        "book-chapter": "book-chapter",
        "proceedings-article": "conference",
        "conference-paper": "conference",
        "dissertation": "thesis",
        "report": "report",
        "posted-content": "preprint",
        "preprint": "preprint",
        "webpage": "web",
    }
    return type_map.get(cr_type, "unknown")