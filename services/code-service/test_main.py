import pytest
import os
from unittest.mock import MagicMock, patch
from fastapi.testclient import TestClient

os.environ["MONGO_URL"] = "mongodb://localhost:27017"
os.environ["MONGO_DB"] = "test_code_db"

from main import app, SUPPORTED_LANGUAGES

client = TestClient(app)


class TestHealthCheck:
    def test_health_check(self):
        response = client.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"


class TestSupportedLanguages:
    def test_python_supported(self):
        assert "python" in SUPPORTED_LANGUAGES
        assert SUPPORTED_LANGUAGES["python"]["image"] == "python:3.12-alpine"
        assert SUPPORTED_LANGUAGES["python"]["extension"] == ".py"

    def test_javascript_supported(self):
        assert "javascript" in SUPPORTED_LANGUAGES
        assert SUPPORTED_LANGUAGES["javascript"]["image"] == "node:20-alpine"
        assert SUPPORTED_LANGUAGES["javascript"]["extension"] == ".js"

    def test_r_supported(self):
        assert "r" in SUPPORTED_LANGUAGES
        assert SUPPORTED_LANGUAGES["r"]["image"] == "r-base:latest"
        assert SUPPORTED_LANGUAGES["r"]["extension"] == ".R"

    def test_julia_supported(self):
        assert "julia" in SUPPORTED_LANGUAGES
        assert SUPPORTED_LANGUAGES["julia"]["image"] == "julia:1.10-alpine"
        assert SUPPORTED_LANGUAGES["julia"]["extension"] == ".jl"

    def test_octave_supported(self):
        assert "octave" in SUPPORTED_LANGUAGES
        assert SUPPORTED_LANGUAGES["octave"]["extension"] == ".m"


class TestCodeFileCRUD:
    @patch("main.db")
    def test_create_file(self, mock_db):
        mock_db.code_files.insert_one.return_value = MagicMock(inserted_id="test_id")
        mock_db.code_files.find_one.return_value = {
            "_id": "test_id",
            "project_id": "proj1",
            "name": "test.py",
            "language": "python",
            "content": "print('hello')",
            "version": 1,
        }

        response = client.post(
            "/api/code/files",
            json={
                "project_id": "proj1",
                "name": "test.py",
                "language": "python",
                "content": "print('hello')",
            },
        )
        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True

    def test_create_file_unsupported_language(self):
        response = client.post(
            "/api/code/files",
            json={
                "project_id": "proj1",
                "name": "test.xyz",
                "language": "unsupported",
                "content": "test",
            },
        )
        assert response.status_code == 400


class TestCodeExecution:
    def test_run_code_invalid_file(self):
        response = client.post(
            "/api/code/run",
            json={"file_id": "invalid_id"},
        )
        assert response.status_code in [400, 404]


class TestVersionManagement:
    def test_list_versions_invalid_file(self):
        response = client.get("/api/code/files/invalid_id/versions")
        assert response.status_code in [400, 404]

    def test_create_version_invalid_file(self):
        response = client.post("/api/code/files/invalid_id/versions")
        assert response.status_code in [400, 404]
