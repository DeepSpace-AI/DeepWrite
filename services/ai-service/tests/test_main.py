import pytest
from fastapi.testclient import TestClient

from main import app


@pytest.fixture
def client():
    return TestClient(app)


class TestHealthCheck:
    def test_health_returns_200(self, client):
        response = client.get("/health")
        assert response.status_code == 200

    def test_health_returns_status(self, client):
        response = client.get("/health")
        data = response.json()
        assert data["status"] == "healthy"

    def test_health_returns_config_status(self, client):
        response = client.get("/health")
        data = response.json()
        assert "openai_configured" in data
        assert "anthropic_configured" in data


class TestChatEndpoint:
    def test_chat_returns_200(self, client):
        response = client.post("/api/ai/chat", json={"message": "Hello"})
        assert response.status_code == 200

    def test_chat_returns_success(self, client):
        response = client.post("/api/ai/chat", json={"message": "Hello"})
        data = response.json()
        assert data["success"] is True

    def test_chat_returns_response(self, client):
        response = client.post("/api/ai/chat", json={"message": "Hello"})
        data = response.json()
        assert "response" in data["data"]
        assert "model" in data["data"]
        assert "tokens_used" in data["data"]

    def test_chat_with_model(self, client):
        response = client.post(
            "/api/ai/chat", json={"message": "Hello", "model": "gpt-4"}
        )
        data = response.json()
        assert data["data"]["model"] == "gpt-4"

    def test_chat_empty_message_fails(self, client):
        response = client.post("/api/ai/chat", json={"message": ""})
        assert response.status_code == 422

    def test_chat_missing_message_fails(self, client):
        response = client.post("/api/ai/chat", json={})
        assert response.status_code == 422


class TestAcademicWriteEndpoint:
    def test_academic_write_returns_200(self, client):
        response = client.post(
            "/api/ai/academic/write", json={"topic": "Machine Learning"}
        )
        assert response.status_code == 200

    def test_academic_write_returns_content(self, client):
        response = client.post(
            "/api/ai/academic/write", json={"topic": "Machine Learning"}
        )
        data = response.json()
        assert data["success"] is True
        assert "content" in data["data"]
        assert "Machine Learning" in data["data"]["content"]

    def test_academic_write_with_section(self, client):
        response = client.post(
            "/api/ai/academic/write",
            json={"topic": "AI", "section": "introduction"},
        )
        data = response.json()
        assert data["data"]["section"] == "introduction"

    def test_academic_write_with_word_count(self, client):
        response = client.post(
            "/api/ai/academic/write",
            json={"topic": "AI", "word_count": 1000},
        )
        data = response.json()
        assert data["data"]["word_count"] == 1000

    def test_academic_write_empty_topic_fails(self, client):
        response = client.post(
            "/api/ai/academic/write", json={"topic": ""}
        )
        assert response.status_code == 422


class TestAcademicPolishEndpoint:
    def test_polish_returns_200(self, client):
        response = client.post(
            "/api/ai/academic/polish", json={"text": "This is a test."}
        )
        assert response.status_code == 200

    def test_polish_returns_original_and_polished(self, client):
        response = client.post(
            "/api/ai/academic/polish", json={"text": "This is a test."}
        )
        data = response.json()
        assert data["success"] is True
        assert data["data"]["original"] == "This is a test."
        assert "polished" in data["data"]

    def test_polish_returns_changes(self, client):
        response = client.post(
            "/api/ai/academic/polish", json={"text": "This is a test."}
        )
        data = response.json()
        assert "changes" in data["data"]
        assert len(data["data"]["changes"]) > 0

    def test_polish_with_style(self, client):
        response = client.post(
            "/api/ai/academic/polish",
            json={"text": "Test text.", "style": "formal"},
        )
        assert response.status_code == 200

    def test_polish_empty_text_fails(self, client):
        response = client.post(
            "/api/ai/academic/polish", json={"text": ""}
        )
        assert response.status_code == 422


class TestReferenceAnalyzeEndpoint:
    def test_analyze_with_title(self, client):
        response = client.post(
            "/api/ai/references/analyze",
            json={"title": "Machine Learning in Healthcare"},
        )
        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True

    def test_analyze_with_abstract(self, client):
        response = client.post(
            "/api/ai/references/analyze",
            json={"abstract": "This paper explores..."},
        )
        assert response.status_code == 200

    def test_analyze_with_text(self, client):
        response = client.post(
            "/api/ai/references/analyze",
            json={"text": "Full paper text here..."},
        )
        assert response.status_code == 200

    def test_analyze_returns_keywords(self, client):
        response = client.post(
            "/api/ai/references/analyze",
            json={"title": "Test Paper"},
        )
        data = response.json()
        assert "keywords" in data["data"]
        assert isinstance(data["data"]["keywords"], list)

    def test_analyze_returns_findings(self, client):
        response = client.post(
            "/api/ai/references/analyze",
            json={"title": "Test Paper"},
        )
        data = response.json()
        assert "main_findings" in data["data"]


class TestCodeGenerateEndpoint:
    def test_generate_code_returns_200(self, client):
        response = client.post(
            "/api/ai/code/generate",
            json={"language": "python", "description": "Hello world"},
        )
        assert response.status_code == 200

    def test_generate_code_returns_code(self, client):
        response = client.post(
            "/api/ai/code/generate",
            json={"language": "python", "description": "Hello world"},
        )
        data = response.json()
        assert data["success"] is True
        assert "code" in data["data"]
        assert "python" in data["data"]["code"]

    def test_generate_code_default_language(self, client):
        response = client.post(
            "/api/ai/code/generate",
            json={"description": "Hello world"},
        )
        data = response.json()
        assert data["data"]["language"] == "python"

    def test_generate_code_javascript(self, client):
        response = client.post(
            "/api/ai/code/generate",
            json={"language": "javascript", "description": "Hello world"},
        )
        data = response.json()
        assert data["data"]["language"] == "javascript"
