import json
import logging
from datetime import datetime
from typing import Any

import httpx

from config.config import get_config
from worker.celery_app import celery_app

logger = logging.getLogger(__name__)

MEMORY_EXTRACTION_PROMPT = """You are an AI memory extraction assistant. Analyze the following conversation and extract important information that should be remembered for future interactions.

Guidelines:
1. Extract facts about the user (preferences, interests, background)
2. Extract important context mentioned (research topics, projects, goals)
3. Extract any explicit requests for future reference
4. Skip generic greetings and small talk
5. Skip information that is already common knowledge

Conversation:
{conversation}

Respond in JSON format with a list of memories:
{{
  "memories": [
    {{
      "type": "preference|fact|task|insight",
      "content": "The actual content to remember",
      "summary": "A brief summary (max 100 chars)",
      "importance": 0.0-1.0
    }}
  ]
}}

If no important information is found, return: {{"memories": []}}
"""

MEMORY_TYPES = {
    "preference": "User preferences and tastes",
    "fact": "Factual information about user or context",
    "task": "Tasks or reminders to follow up",
    "insight": "Key insights or decisions made",
}


def _get_gateway_config() -> tuple[str, str]:
    """获取 Gateway URL 和 Token 配置"""
    cfg = get_config()
    return cfg.gateway_url, cfg.gateway_token


def _get_ai_service_config() -> tuple[str, str]:
    """获取 AI Service URL 和 Token 配置"""
    cfg = get_config()
    return cfg.ai_service_url, cfg.ai_service_token


def _fetch_session_messages(session_id: str, gateway_url: str, gateway_token: str) -> list[dict[str, Any]]:
    """从 Gateway 获取会话消息"""
    url = f"{gateway_url}/api/v1/sessions/{session_id}"
    
    headers = {"Content-Type": "application/json"}
    if gateway_token:
        headers["X-Internal-Token"] = gateway_token
    
    with httpx.Client(timeout=30.0) as client:
        response = client.get(url, headers=headers)
        if response.status_code == 200:
            data = response.json()
            return data.get("messages", [])
        else:
            logger.warning(f"Failed to fetch session {session_id}: {response.status_code}")
            return []


def _call_ai_for_extraction(conversation: str, ai_url: str, ai_token: str) -> dict[str, Any]:
    """调用 AI Service 进行记忆提取分析"""
    url = f"{ai_url}/v1/chat/completions"
    
    headers = {
        "Content-Type": "application/json",
    }
    if ai_token:
        headers["X-Internal-Token"] = ai_token
    
    prompt = MEMORY_EXTRACTION_PROMPT.format(conversation=conversation)
    
    payload = {
        "model": "gpt-4o-mini",
        "messages": [
            {"role": "system", "content": "You are a memory extraction assistant. Always respond with valid JSON."},
            {"role": "user", "content": prompt}
        ],
        "temperature": 0.3,
        "max_tokens": 1024,
    }
    
    with httpx.Client(timeout=60.0) as client:
        response = client.post(url, json=payload, headers=headers)
        if response.status_code == 200:
            data = response.json()
            content = data.get("choices", [{}])[0].get("message", {}).get("content", "")
            try:
                return json.loads(content)
            except json.JSONDecodeError:
                logger.warning(f"Failed to parse AI response as JSON: {content[:200]}")
                return {"memories": []}
        else:
            logger.warning(f"AI Service error: {response.status_code}")
            return {"memories": []}


def _generate_embedding(text: str, ai_url: str, ai_token: str) -> list[float] | None:
    """调用 AI Service 生成 embedding"""
    url = f"{ai_url}/v1/embeddings"
    
    headers = {
        "Content-Type": "application/json",
    }
    if ai_token:
        headers["X-Internal-Token"] = ai_token
    
    payload = {
        "model": "text-embedding-ada-002",
        "input": text,
    }
    
    with httpx.Client(timeout=30.0) as client:
        response = client.post(url, json=payload, headers=headers)
        if response.status_code == 200:
            data = response.json()
            embeddings = data.get("data", [])
            if embeddings:
                return embeddings[0].get("embedding", [])
        return None


def _store_memory(
    memory_data: dict[str, Any],
    agent_id: str,
    user_id: str,
    workspace_id: str,
    session_id: str,
    embedding: list[float] | None,
    gateway_url: str,
    gateway_token: str,
) -> bool:
    """存储记忆到 Gateway"""
    url = f"{gateway_url}/api/v1/memories"
    
    headers = {"Content-Type": "application/json"}
    if gateway_token:
        headers["X-Internal-Token"] = gateway_token
    
    payload = {
        "agent_id": agent_id,
        "workspace_id": workspace_id,
        "type": memory_data.get("type", "fact"),
        "content": memory_data.get("content", ""),
        "summary": memory_data.get("summary", ""),
        "importance": memory_data.get("importance", 0.5),
        "source_session_id": session_id,
    }
    
    if embedding:
        payload["embedding_vector"] = json.dumps(embedding)
    
    with httpx.Client(timeout=10.0) as client:
        response = client.post(url, json=payload, headers=headers)
        if response.status_code in (200, 201):
            return True
        else:
            logger.warning(f"Failed to store memory: {response.status_code}")
            return False


def _format_conversation(messages: list[dict[str, Any]]) -> str:
    """格式化对话消息为文本"""
    lines = []
    for msg in messages[-20:]:
        role = msg.get("role", "unknown")
        content = msg.get("content", "")
        if role == "user":
            lines.append(f"User: {content}")
        elif role == "assistant":
            lines.append(f"Assistant: {content}")
    return "\n\n".join(lines)


@celery_app.task(
    name="worker.memory.extract",
    bind=True,
    max_retries=3,
    default_retry_delay=60,
    autoretry_for=(Exception,),
    retry_backoff=True,
)
def extract_memories(
    self,
    session_id: str,
    agent_id: str,
    user_id: str,
    workspace_id: str,
) -> dict[str, Any]:
    """
    从对话会话中提取记忆。

    Args:
        session_id: 会话 ID
        agent_id: Agent ID
        user_id: 用户 ID
        workspace_id: 工作区 ID

    Returns:
        提取结果
    """
    logger.info(f"[extract_memories] session={session_id}, agent={agent_id}")

    gateway_url, gateway_token = _get_gateway_config()
    ai_url, ai_token = _get_ai_service_config()

    try:
        messages = _fetch_session_messages(session_id, gateway_url, gateway_token)
        if not messages:
            logger.info(f"[extract_memories] no messages in session {session_id}")
            return {
                "status": "skipped",
                "reason": "no_messages",
                "session_id": session_id,
            }

        conversation = _format_conversation(messages)
        if len(conversation) < 100:
            logger.info(f"[extract_memories] conversation too short: {len(conversation)} chars")
            return {
                "status": "skipped",
                "reason": "conversation_too_short",
                "session_id": session_id,
            }

        extraction_result = _call_ai_for_extraction(conversation, ai_url, ai_token)
        memories = extraction_result.get("memories", [])

        if not memories:
            logger.info(f"[extract_memories] no memories extracted from session {session_id}")
            return {
                "status": "completed",
                "memories_extracted": 0,
                "session_id": session_id,
            }

        stored_count = 0
        for memory_data in memories:
            content = memory_data.get("content", "")
            if not content or len(content) < 10:
                continue

            embedding = _generate_embedding(content, ai_url, ai_token)

            success = _store_memory(
                memory_data=memory_data,
                agent_id=agent_id,
                user_id=user_id,
                workspace_id=workspace_id,
                session_id=session_id,
                embedding=embedding,
                gateway_url=gateway_url,
                gateway_token=gateway_token,
            )
            if success:
                stored_count += 1

        logger.info(f"[extract_memories] extracted {stored_count} memories from session {session_id}")

        return {
            "status": "completed",
            "session_id": session_id,
            "memories_extracted": stored_count,
            "completed_at": datetime.now().isoformat(),
        }

    except Exception as exc:
        logger.error(f"[extract_memories] failed: {exc}")
        raise


@celery_app.task(
    name="worker.memory.batch_extract",
    bind=True,
    max_retries=2,
    default_retry_delay=120,
)
def batch_extract_memories(
    self,
    sessions: list[dict[str, str]],
) -> dict[str, Any]:
    """
    批量提取多个会话的记忆。

    Args:
        sessions: 会话列表，每项包含 session_id, agent_id, user_id, workspace_id

    Returns:
        批量处理结果
    """
    logger.info(f"[batch_extract_memories] processing {len(sessions)} sessions")

    results = []
    for session_info in sessions:
        try:
            result = extract_memories.delay(
                session_id=session_info["session_id"],
                agent_id=session_info["agent_id"],
                user_id=session_info["user_id"],
                workspace_id=session_info["workspace_id"],
            )
            results.append({
                "session_id": session_info["session_id"],
                "task_id": result.id,
                "status": "queued",
            })
        except Exception as e:
            results.append({
                "session_id": session_info.get("session_id", "unknown"),
                "status": "failed",
                "error": str(e),
            })

    return {
        "status": "completed",
        "total": len(sessions),
        "results": results,
        "completed_at": datetime.now().isoformat(),
    }