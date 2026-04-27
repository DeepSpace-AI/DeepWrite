import os
import uuid
from contextlib import asynccontextmanager
from datetime import datetime
from typing import Optional

from fastapi import FastAPI, HTTPException, Query, UploadFile, File
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import Response
from minio import Minio
from pydantic import BaseModel, Field

MINIO_ENDPOINT = os.getenv("MINIO_ENDPOINT", "minio:9000")
MINIO_ACCESS_KEY = os.getenv("MINIO_ACCESS_KEY", "minioadmin")
MINIO_SECRET_KEY = os.getenv("MINIO_SECRET_KEY", "minioadmin")
MINIO_BUCKET_DOCUMENTS = os.getenv("MINIO_BUCKET_DOCUMENTS", "deepwrite-documents")
MINIO_BUCKET_CODE = os.getenv("MINIO_BUCKET_CODE", "deepwrite-code")
MINIO_BUCKET_EXPORTS = os.getenv("MINIO_BUCKET_EXPORTS", "deepwrite-exports")

minio_client = None
buckets = {}


@asynccontextmanager
async def lifespan(app: FastAPI):
    global minio_client, buckets
    minio_client = Minio(
        MINIO_ENDPOINT,
        access_key=MINIO_ACCESS_KEY,
        secret_key=MINIO_SECRET_KEY,
        secure=False,
    )

    buckets = {
        "documents": MINIO_BUCKET_DOCUMENTS,
        "code": MINIO_BUCKET_CODE,
        "exports": MINIO_BUCKET_EXPORTS,
    }

    for bucket_name in buckets.values():
        if not minio_client.bucket_exists(bucket_name):
            minio_client.make_bucket(bucket_name)

    yield


app = FastAPI(title="DeepWrite Storage Service", lifespan=lifespan)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


class FileMetadata(BaseModel):
    filename: str
    content_type: str
    size: int
    bucket: str
    object_name: str
    url: str
    created_at: str


def get_bucket(bucket_type: str) -> str:
    if bucket_type not in buckets:
        raise HTTPException(
            status_code=400,
            detail=f"Invalid bucket type: {bucket_type}. Valid types: {list(buckets.keys())}",
        )
    return buckets[bucket_type]


@app.get("/health")
def health_check():
    return {
        "status": "healthy",
        "minio_connected": minio_client is not None,
        "buckets": list(buckets.keys()),
    }


@app.post("/api/storage/upload")
async def upload_file(
    bucket: str = Query(..., pattern="^(documents|code|exports)$"),
    project_id: str = Query(...),
    file: UploadFile = File(...),
):
    bucket_name = get_bucket(bucket)
    content = await file.read()

    file_ext = os.path.splitext(file.filename)[1] if file.filename else ""
    object_name = f"{project_id}/{uuid.uuid4()}{file_ext}"

    minio_client.put_object(
        bucket_name,
        object_name,
        file.file,
        len(content),
        content_type=file.content_type or "application/octet-stream",
    )

    metadata = {
        "filename": file.filename,
        "content_type": file.content_type,
        "size": len(content),
        "bucket": bucket,
        "object_name": object_name,
        "url": f"/api/storage/{bucket}/{object_name}",
        "created_at": datetime.utcnow().isoformat(),
    }

    return {"success": True, "data": metadata}


@app.get("/api/storage/{bucket}/{object_name:path}")
def download_file(bucket: str, object_name: str):
    bucket_name = get_bucket(bucket)

    try:
        response = minio_client.get_object(bucket_name, object_name)
        content = response.read()
        response.close()
        response.release_conn()
    except Exception:
        raise HTTPException(status_code=404, detail="File not found")

    return Response(content=content, media_type="application/octet-stream")


@app.delete("/api/storage/{bucket}/{object_name:path}")
def delete_file(bucket: str, object_name: str):
    bucket_name = get_bucket(bucket)

    try:
        minio_client.remove_object(bucket_name, object_name)
    except Exception:
        raise HTTPException(status_code=404, detail="File not found")

    return {"success": True, "message": "File deleted"}


@app.get("/api/storage/list")
def list_files(
    bucket: str = Query(..., pattern="^(documents|code|exports)$"),
    project_id: str = Query(...),
    prefix: str = Query(""),
):
    bucket_name = get_bucket(bucket)
    full_prefix = f"{project_id}/{prefix}"

    try:
        objects = minio_client.list_objects(bucket_name, prefix=full_prefix, recursive=True)
        files = []
        for obj in objects:
            files.append({
                "object_name": obj.object_name,
                "size": obj.size,
                "last_modified": obj.last_modified.isoformat() if obj.last_modified else None,
                "url": f"/api/storage/{bucket}/{obj.object_name}",
            })
        return {"success": True, "data": files}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to list files: {str(e)}")
