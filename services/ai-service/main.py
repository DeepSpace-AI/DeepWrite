import os
from typing import Optional

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

app = FastAPI(title="DeepWrite AI Service")

OPENAI_API_KEY = os.getenv("OPENAI_API_KEY", "")
ANTHROPIC_API_KEY = os.getenv("ANTHROPIC_API_KEY", "")

class ChatRequest(BaseModel):
    message: str = Field(..., min_length=1)
    context: Optional[str] = None
    model: str = "gpt-4"

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

@app.get("/health")
def health_check():
    return {"status": "healthy", "openai_configured": bool(OPENAI_API_KEY), "anthropic_configured": bool(ANTHROPIC_API_KEY)}

@app.post("/api/ai/chat")
def chat(request: ChatRequest):
    return {
        "success": True,
        "data": {
            "response": f"这是 AI 对 '{request.message}' 的回复。（AI 服务已连接，但尚未配置真实的 AI 模型）",
            "model": request.model,
            "tokens_used": 0,
        }
    }

@app.post("/api/ai/academic/write")
def academic_write(request: AcademicWriteRequest):
    return {
        "success": True,
        "data": {
            "content": f"# {request.topic}\n\n这是关于 '{request.topic}' 的学术写作示例。\n\n## 引言\n\n在本研究中，我们探讨了...\n\n## 方法\n\n我们采用了定量研究方法...\n\n## 结果\n\n研究发现表明...\n\n## 讨论\n\n这些结果的意义在于...\n\n（此为演示内容，实际使用需要配置 OpenAI/Anthropic API Key）",
            "word_count": request.word_count or 500,
            "section": request.section,
        }
    }

@app.post("/api/ai/academic/polish")
def academic_polish(request: PolishRequest):
    return {
        "success": True,
        "data": {
            "original": request.text,
            "polished": request.text + "\n\n[已润色版本：此处将显示经过学术润色后的文本，包括语法修正、风格优化和表达改进。]",
            "changes": [
                {"type": "grammar", "description": "修正了语法错误"},
                {"type": "style", "description": "优化了学术表达"},
            ]
        }
    }

@app.post("/api/ai/references/analyze")
def analyze_reference(request: ReferenceAnalyzeRequest):
    text = request.text or request.abstract or request.title or ""
    return {
        "success": True,
        "data": {
            "summary": f"这篇文献的主要内容是关于...（基于 '{text[:100]}...' 的分析）",
            "keywords": ["关键词1", "关键词2", "关键词3"],
            "main_findings": ["发现1", "发现2"],
            "methodology": "研究方法概述",
        }
    }

@app.post("/api/ai/code/generate")
def generate_code(request: dict):
    language = request.get("language", "python")
    description = request.get("description", "")
    return {
        "success": True,
        "data": {
            "code": f"# {language} 代码生成示例\n# 描述: {description}\n\ndef example():\n    pass\n",
            "language": language,
            "explanation": "这是根据您的描述生成的代码示例。",
        }
    }
