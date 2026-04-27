import os
import uuid
from contextlib import asynccontextmanager
from datetime import datetime
from io import BytesIO
from typing import Optional

from fastapi import FastAPI, HTTPException, Query, UploadFile, File
from fastapi.middleware.cors import CORSMiddleware
from minio import Minio
from PIL import Image
from pydantic import BaseModel, Field
from pymongo import MongoClient
from bson import ObjectId

MONGO_URL = os.getenv("MONGO_URL", "mongodb://admin:password@image-db:27017")
MONGO_DB = os.getenv("MONGO_DB", "image_db")
MINIO_ENDPOINT = os.getenv("MINIO_ENDPOINT", "minio:9000")
MINIO_ACCESS_KEY = os.getenv("MINIO_ACCESS_KEY", "minioadmin")
MINIO_SECRET_KEY = os.getenv("MINIO_SECRET_KEY", "minioadmin")
MINIO_BUCKET = os.getenv("MINIO_BUCKET", "deepwrite-images")

client = None
db = None
minio_client = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global client, db, minio_client
    client = MongoClient(MONGO_URL)
    db = client[MONGO_DB]
    minio_client = Minio(
        MINIO_ENDPOINT,
        access_key=MINIO_ACCESS_KEY,
        secret_key=MINIO_SECRET_KEY,
        secure=False,
    )
    if not minio_client.bucket_exists(MINIO_BUCKET):
        minio_client.make_bucket(MINIO_BUCKET)
    yield
    client.close()


app = FastAPI(title="DeepWrite Image Service", lifespan=lifespan)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


class ImageCreate(BaseModel):
    project_id: str
    name: str = Field(..., min_length=1, max_length=255)
    description: Optional[str] = None
    tags: list[str] = []


class ImageUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    tags: Optional[list[str]] = None


def serialize_doc(doc):
    if doc is None:
        return None
    doc["id"] = str(doc.pop("_id"))
    for key, val in doc.items():
        if isinstance(val, datetime):
            doc[key] = val.isoformat()
    return doc


@app.get("/health")
def health_check():
    return {
        "status": "healthy",
        "minio_connected": minio_client is not None,
    }


@app.post("/api/images/upload")
async def upload_image(
    project_id: str = Query(...),
    file: UploadFile = File(...),
    name: Optional[str] = Query(None),
    description: Optional[str] = Query(None),
):
    content = await file.read()

    try:
        img = Image.open(BytesIO(content))
        width, height = img.size
        format = img.format.lower()
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid image file")

    file_ext = format if format in ["png", "jpg", "jpeg", "gif", "webp"] else "png"
    object_name = f"{project_id}/{uuid.uuid4()}.{file_ext}"

    minio_client.put_object(
        MINIO_BUCKET,
        object_name,
        BytesIO(content),
        len(content),
        content_type=f"image/{file_ext}",
    )

    doc = {
        "project_id": project_id,
        "name": name or file.filename,
        "description": description or "",
        "object_name": object_name,
        "url": f"/api/images/{object_name}",
        "width": width,
        "height": height,
        "size": len(content),
        "format": file_ext,
        "tags": [],
        "created_at": datetime.utcnow(),
        "updated_at": datetime.utcnow(),
    }
    result = db.images.insert_one(doc)
    created = db.images.find_one({"_id": result.inserted_id})
    return {"success": True, "data": serialize_doc(created)}


@app.get("/api/images")
def list_images(
    project_id: str = Query(...),
    page: int = Query(1, ge=1),
    limit: int = Query(20, ge=1, le=100),
):
    skip = (page - 1) * limit
    total = db.images.count_documents({"project_id": project_id})
    images = list(
        db.images.find({"project_id": project_id})
        .skip(skip)
        .limit(limit)
        .sort("created_at", -1)
    )
    return {
        "success": True,
        "data": [serialize_doc(img) for img in images],
        "meta": {"page": page, "limit": limit, "total": total},
    }


@app.get("/api/images/{image_id}")
def get_image(image_id: str):
    try:
        image = db.images.find_one({"_id": ObjectId(image_id)})
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid image ID")
    if not image:
        raise HTTPException(status_code=404, detail="Image not found")
    return {"success": True, "data": serialize_doc(image)}


@app.put("/api/images/{image_id}")
def update_image(image_id: str, update: ImageUpdate):
    update_data = {k: v for k, v in update.model_dump().items() if v is not None}
    if update_data:
        update_data["updated_at"] = datetime.utcnow()
        try:
            result = db.images.update_one(
                {"_id": ObjectId(image_id)}, {"$set": update_data}
            )
        except Exception:
            raise HTTPException(status_code=400, detail="Invalid image ID")
        if result.matched_count == 0:
            raise HTTPException(status_code=404, detail="Image not found")
    image = db.images.find_one({"_id": ObjectId(image_id)})
    return {"success": True, "data": serialize_doc(image)}


@app.delete("/api/images/{image_id}")
def delete_image(image_id: str):
    try:
        image = db.images.find_one({"_id": ObjectId(image_id)})
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid image ID")
    if not image:
        raise HTTPException(status_code=404, detail="Image not found")

    try:
        minio_client.remove_object(MINIO_BUCKET, image["object_name"])
    except Exception:
        pass

    db.images.delete_one({"_id": ObjectId(image_id)})
    return {"success": True, "message": "Image deleted"}


@app.get("/api/images/file/{object_name:path}")
def get_image_file(object_name: str):
    try:
        response = minio_client.get_object(MINIO_BUCKET, object_name)
        content = response.read()
        response.close()
        response.release_conn()
    except Exception:
        raise HTTPException(status_code=404, detail="Image file not found")

    from fastapi.responses import Response

    return Response(content=content, media_type="image/png")
