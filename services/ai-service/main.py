import os
from typing import Optional

import httpx
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field

app = FastAPI(title="DeepWrite AI Service")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

OPENAI_API_KEY = os.getenv("OPENAI_API_KEY", "")
ANTHROPIC_API_KEY = os.getenv("ANTHROPIC_API_KEY", "")
OPENAI_BASE_URL = os.getenv("OPENAI_BASE_URL", "https://api.openai.com/v1")
ANTHROPIC_BASE_URL = os.getenv("ANTHROPIC_BASE_URL", "https://api.anthropic.com")


class ChatRequest(BaseModel):
    message: str = Field(..., min_length=1)
    context: Optional[str] = None
    model: str = "gpt-4o"


class AcademicWriteRequest(BaseModel):
    topic: str = Field(..., min_length=1)
    section: Optional[str] = None
    style: str = "academic"
    word_count: Optional[int] = 500
    references: Optional[list[str]] = None


class PolishRequest(BaseModel):
    text: str = Field(..., min_length=1)
    style: str = "academic"
    focus: Optional[str] = None


class ReferenceAnalyzeRequest(BaseModel):
    title: Optional[str] = None
    abstract: Optional[str] = None
    text: Optional[str] = None


class CodeGenerateRequest(BaseModel):
    language: str = "python"
    description: str = Field(..., min_length=1)


async def call_openai(messages: list[dict], model: str = "gpt-4o", max_tokens: int = 2000) -> str:
    if not OPENAI_API_KEY:
        raise HTTPException(status_code=500, detail="OpenAI API key not configured")

    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{OPENAI_BASE_URL}/chat/completions",
            headers={
                "Authorization": f"Bearer {OPENAI_API_KEY}",
                "Content-Type": "application/json",
            },
            json={
                "model": model,
                "messages": messages,
                "max_tokens": max_tokens,
                "temperature": 0.7,
            },
            timeout=60.0,
        )

        if response.status_code != 200:
            error_detail = response.json().get("error", {}).get("message", "Unknown error")
            raise HTTPException(status_code=response.status_code, detail=f"OpenAI error: {error_detail}")

        data = response.json()
        return data["choices"][0]["message"]["content"]


async def call_anthropic(messages: list[dict], model: str = "claude-3-5-sonnet-20241022", max_tokens: int = 2000) -> str:
    if not ANTHROPIC_API_KEY:
        raise HTTPException(status_code=500, detail="Anthropic API key not configured")

    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{ANTHROPIC_BASE_URL}/v1/messages",
            headers={
                "x-api-key": ANTHROPIC_API_KEY,
                "anthropic-version": "2023-06-01",
                "Content-Type": "application/json",
            },
            json={
                "model": model,
                "max_tokens": max_tokens,
                "messages": messages,
            },
            timeout=60.0,
        )

        if response.status_code != 200:
            error_detail = response.json().get("error", {}).get("message", "Unknown error")
            raise HTTPException(status_code=response.status_code, detail=f"Anthropic error: {error_detail}")

        data = response.json()
        return data["content"][0]["text"]


async def generate_ai_response(messages: list[dict], model: str = "gpt-4o", max_tokens: int = 2000) -> str:
    if model.startswith("claude"):
        return await call_anthropic(messages, model, max_tokens)
    return await call_openai(messages, model, max_tokens)


@app.get("/health")
def health_check():
    return {
        "status": "healthy",
        "openai_configured": bool(OPENAI_API_KEY),
        "anthropic_configured": bool(ANTHROPIC_API_KEY),
    }


@app.post("/api/ai/chat")
async def chat(request: ChatRequest):
    messages = []
    if request.context:
        messages.append({"role": "system", "content": f"Context: {request.context}"})
    messages.append({"role": "user", "content": request.message})

    response_text = await generate_ai_response(messages, request.model)

    return {
        "success": True,
        "data": {
            "response": response_text,
            "model": request.model,
        },
    }


@app.post("/api/ai/academic/write")
async def academic_write(request: AcademicWriteRequest):
    section_prompt = f"请撰写{request.section}部分。" if request.section else "请撰写完整的学术论文。"

    ref_context = ""
    if request.references:
        ref_context = "\n\n参考文献：\n" + "\n".join(f"- {ref}" for ref in request.references)

    prompt = f"""你是一位资深的学术写作助手。请根据以下要求进行学术写作：

主题：{request.topic}
{section_prompt}
写作风格：{request.style}
目标字数：约{request.word_count}字{ref_context}

请确保：
1. 使用严谨的学术语言
2. 逻辑清晰，论证有力
3. 符合学术论文的写作规范
4. 适当引用参考文献（如果有）"""

    messages = [{"role": "user", "content": prompt}]
    response_text = await generate_ai_response(messages, max_tokens=3000)

    return {
        "success": True,
        "data": {
            "content": response_text,
            "word_count": len(response_text),
            "section": request.section,
        },
    }


@app.post("/api/ai/academic/polish")
async def academic_polish(request: PolishRequest):
    focus_prompt = f"特别关注{request.focus}方面的改进。" if request.focus else ""

    prompt = f"""你是一位资深的学术润色专家。请对以下文本进行学术润色：

{request.text}

润色要求：
1. 保持原意不变
2. 使用{request.style}风格
3. 改进语法、用词和表达
4. 增强学术性和专业性{focus_prompt}

请返回润色后的文本，并简要说明主要修改。"""

    messages = [{"role": "user", "content": prompt}]
    response_text = await generate_ai_response(messages)

    return {
        "success": True,
        "data": {
            "original": request.text,
            "polished": response_text,
        },
    }


@app.post("/api/ai/references/analyze")
async def analyze_reference(request: ReferenceAnalyzeRequest):
    text = request.text or request.abstract or request.title or ""
    if not text:
        raise HTTPException(status_code=400, detail="At least one of title, abstract, or text is required")

    prompt = f"""请分析以下学术文献内容，并提供：

{text}

请提供：
1. 简要摘要（100-200字）
2. 关键词（3-5个）
3. 主要发现/论点
4. 研究方法
5. 研究局限性
6. 与领域的相关性"""

    messages = [{"role": "user", "content": prompt}]
    response_text = await generate_ai_response(messages)

    return {
        "success": True,
        "data": {
            "analysis": response_text,
        },
    }


@app.post("/api/ai/code/generate")
async def generate_code(request: CodeGenerateRequest):
    prompt = f"""请生成{request.language}代码来实现以下功能：

{request.description}

要求：
1. 代码清晰、可读
2. 包含必要的注释
3. 遵循{request.language}最佳实践
4. 包含错误处理"""

    messages = [{"role": "user", "content": prompt}]
    response_text = await generate_ai_response(messages)

    return {
        "success": True,
        "data": {
            "code": response_text,
            "language": request.language,
        },
    }


@app.post("/api/ai/code/explain")
async def explain_code(request: dict):
    code = request.get("code", "")
    language = request.get("language", "python")

    if not code:
        raise HTTPException(status_code=400, detail="Code is required")

    prompt = f"""请解释以下{language}代码的功能和逻辑：

```{language}
{code}
```

请提供：
1. 代码功能概述
2. 逐行/逐块解释
3. 关键概念说明
4. 潜在的改进点"""

    messages = [{"role": "user", "content": prompt}]
    response_text = await generate_ai_response(messages)

    return {
        "success": True,
        "data": {
            "explanation": response_text,
        },
    }
