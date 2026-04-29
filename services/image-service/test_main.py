import pytest
import os
from unittest.mock import MagicMock, patch
from fastapi.testclient import TestClient

os.environ["MONGO_URL"] = "mongodb://localhost:27017"
os.environ["MONGO_DB"] = "test_image_db"
os.environ["MINIO_ENDPOINT"] = "localhost:9000"
os.environ["MINIO_ACCESS_KEY"] = "test"
os.environ["MINIO_SECRET_KEY"] = "test"

from main import app

client = TestClient(app)


class TestHealthCheck:
    def test_health_check(self):
        response = client.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"


class TestImageAPI:
    @patch("main.db")
    def test_list_images(self, mock_db):
        mock_db.images.count_documents.return_value = 0
        mock_db.images.find.return_value.skip.return_value.limit.return_value.sort.return_value = []

        response = client.get("/api/images?project_id=proj1")
        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True
        assert data["data"] == []

    @patch("main.db")
    def test_get_image_not_found(self, mock_db):
        mock_db.images.find_one.return_value = None

        response = client.get("/api/images/nonexistent")
        assert response.status_code in [400, 404]

    @patch("main.db")
    def test_update_image_not_found(self, mock_db):
        mock_db.images.update_one.return_value = MagicMock(matched_count=0)

        response = client.put(
            "/api/images/nonexistent",
            json={"name": "new name"},
        )
        assert response.status_code == 404

    @patch("main.db")
    def test_delete_image_not_found(self, mock_db):
        mock_db.images.find_one.return_value = None

        response = client.delete("/api/images/nonexistent")
        assert response.status_code in [400, 404]


class TestImageUpload:
    def test_upload_no_file(self):
        response = client.post("/api/images/upload?project_id=proj1")
        assert response.status_code == 422
