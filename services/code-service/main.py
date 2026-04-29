import os
import uuid
from contextlib import asynccontextmanager
from datetime import datetime
from typing import Optional

import docker
from fastapi import FastAPI, HTTPException, Query
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
from pymongo import MongoClient
from bson import ObjectId

MONGO_URL = os.getenv("MONGO_URL", "mongodb://admin:password@code-db:27017")
MONGO_DB = os.getenv("MONGO_DB", "code_db")

client = None
db = None
docker_client = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global client, db, docker_client
    client = MongoClient(MONGO_URL)
    db = client[MONGO_DB]
    try:
        docker_client = docker.from_env()
    except Exception:
        docker_client = None
    yield
    client.close()


app = FastAPI(title="DeepWrite Code Service", lifespan=lifespan)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


class CodeFileCreate(BaseModel):
    project_id: str
    name: str = Field(..., min_length=1, max_length=255)
    language: str = Field(..., min_length=1)
    content: str = ""


class CodeFileUpdate(BaseModel):
    name: Optional[str] = None
    content: Optional[str] = None


class CodeRunRequest(BaseModel):
    file_id: str
    timeout: int = Field(default=30, ge=1, le=300)


SUPPORTED_LANGUAGES = {
    "python": {"image": "python:3.12-alpine", "extension": ".py", "cmd": "python"},
    "javascript": {"image": "node:20-alpine", "extension": ".js", "cmd": "node"},
    "r": {"image": "r-base:latest", "extension": ".R", "cmd": "Rscript"},
    "julia": {"image": "julia:1.10-alpine", "extension": ".jl", "cmd": "julia"},
    "octave": {"image": "gnuoctave/octave:latest", "extension": ".m", "cmd": "octave --no-gui"},
}


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
        "docker_available": docker_client is not None,
    }


@app.post("/api/code/files")
def create_file(file: CodeFileCreate):
    if file.language not in SUPPORTED_LANGUAGES:
        raise HTTPException(
            status_code=400,
            detail=f"Unsupported language: {file.language}. Supported: {list(SUPPORTED_LANGUAGES.keys())}",
        )

    doc = {
        "project_id": file.project_id,
        "name": file.name,
        "language": file.language,
        "content": file.content,
        "version": 1,
        "created_at": datetime.utcnow(),
        "updated_at": datetime.utcnow(),
    }
    result = db.code_files.insert_one(doc)
    created = db.code_files.find_one({"_id": result.inserted_id})
    return {"success": True, "data": serialize_doc(created)}


@app.get("/api/code/files")
def list_files(
    project_id: str = Query(...),
    page: int = Query(1, ge=1),
    limit: int = Query(20, ge=1, le=100),
):
    skip = (page - 1) * limit
    total = db.code_files.count_documents({"project_id": project_id})
    files = list(
        db.code_files.find({"project_id": project_id})
        .skip(skip)
        .limit(limit)
        .sort("updated_at", -1)
    )
    return {
        "success": True,
        "data": [serialize_doc(f) for f in files],
        "meta": {"page": page, "limit": limit, "total": total},
    }


@app.get("/api/code/files/{file_id}")
def get_file(file_id: str):
    try:
        file = db.code_files.find_one({"_id": ObjectId(file_id)})
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid file ID")
    if not file:
        raise HTTPException(status_code=404, detail="File not found")
    return {"success": True, "data": serialize_doc(file)}


@app.put("/api/code/files/{file_id}")
def update_file(file_id: str, update: CodeFileUpdate):
    update_data = {k: v for k, v in update.model_dump().items() if v is not None}
    if update_data:
        update_data["updated_at"] = datetime.utcnow()
        if "content" in update_data:
            update_data["$inc"] = {"version": 1}
        try:
            result = db.code_files.update_one(
                {"_id": ObjectId(file_id)}, {"$set": update_data}
            )
        except Exception:
            raise HTTPException(status_code=400, detail="Invalid file ID")
        if result.matched_count == 0:
            raise HTTPException(status_code=404, detail="File not found")
    file = db.code_files.find_one({"_id": ObjectId(file_id)})
    return {"success": True, "data": serialize_doc(file)}


@app.delete("/api/code/files/{file_id}")
def delete_file(file_id: str):
    try:
        result = db.code_files.delete_one({"_id": ObjectId(file_id)})
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid file ID")
    if result.deleted_count == 0:
        raise HTTPException(status_code=404, detail="File not found")
    return {"success": True, "message": "File deleted"}


@app.post("/api/code/run")
def run_code(request: CodeRunRequest):
    if not docker_client:
        raise HTTPException(status_code=500, detail="Docker not available")

    try:
        file = db.code_files.find_one({"_id": ObjectId(request.file_id)})
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid file ID")
    if not file:
        raise HTTPException(status_code=404, detail="File not found")

    lang_config = SUPPORTED_LANGUAGES.get(file["language"])
    if not lang_config:
        raise HTTPException(status_code=400, detail="Unsupported language")

    run_id = str(uuid.uuid4())
    run_doc = {
        "run_id": run_id,
        "file_id": request.file_id,
        "project_id": file["project_id"],
        "status": "running",
        "started_at": datetime.utcnow(),
    }
    db.code_runs.insert_one(run_doc)

    try:
        container = docker_client.containers.run(
            image=lang_config["image"],
            command=f'{lang_config["cmd"]} -c """{file["content"]}"""',
            detach=True,
            mem_limit="256m",
            cpu_period=100000,
            cpu_quota=50000,
            network_disabled=True,
        )

        result = container.wait(timeout=request.timeout)
        logs = container.logs().decode("utf-8")
        container.remove()

        status = "completed" if result["StatusCode"] == 0 else "failed"
        db.code_runs.update_one(
            {"run_id": run_id},
            {
                "$set": {
                    "status": status,
                    "exit_code": result["StatusCode"],
                    "output": logs,
                    "completed_at": datetime.utcnow(),
                }
            },
        )

        return {
            "success": True,
            "data": {
                "run_id": run_id,
                "status": status,
                "exit_code": result["StatusCode"],
                "output": logs,
            },
        }

    except docker.errors.ContainerError as e:
        db.code_runs.update_one(
            {"run_id": run_id},
            {"$set": {"status": "failed", "error": str(e), "completed_at": datetime.utcnow()}},
        )
        raise HTTPException(status_code=400, detail=f"Container error: {str(e)}")
    except docker.errors.ImageNotFound:
        raise HTTPException(status_code=400, detail=f"Image not found: {lang_config['image']}")
    except Exception as e:
        db.code_runs.update_one(
            {"run_id": run_id},
            {"$set": {"status": "failed", "error": str(e), "completed_at": datetime.utcnow()}},
        )
        raise HTTPException(status_code=500, detail=f"Execution error: {str(e)}")


@app.get("/api/code/runs/{run_id}")
def get_run_status(run_id: str):
    run = db.code_runs.find_one({"run_id": run_id})
    if not run:
        raise HTTPException(status_code=404, detail="Run not found")
    return {"success": True, "data": serialize_doc(run)}


@app.get("/api/code/files/{file_id}/versions")
def list_versions(file_id: str):
    try:
        file = db.code_files.find_one({"_id": ObjectId(file_id)})
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid file ID")
    if not file:
        raise HTTPException(status_code=404, detail="File not found")

    versions = list(
        db.code_versions.find({"file_id": file_id})
        .sort("version", -1)
        .limit(50)
    )
    return {"success": True, "data": [serialize_doc(v) for v in versions]}


@app.post("/api/code/files/{file_id}/versions")
def create_version(file_id: str, message: str = Query("")):
    try:
        file = db.code_files.find_one({"_id": ObjectId(file_id)})
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid file ID")
    if not file:
        raise HTTPException(status_code=404, detail="File not found")

    version_doc = {
        "file_id": file_id,
        "project_id": file["project_id"],
        "version": file.get("version", 1),
        "content": file["content"],
        "message": message,
        "created_at": datetime.utcnow(),
    }
    db.code_versions.insert_one(version_doc)

    return {"success": True, "message": "Version created"}


@app.get("/api/code/files/{file_id}/versions/{version}")
def get_version(file_id: str, version: int):
    try:
        file = db.code_files.find_one({"_id": ObjectId(file_id)})
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid file ID")
    if not file:
        raise HTTPException(status_code=404, detail="File not found")

    version_doc = db.code_versions.find_one({
        "file_id": file_id,
        "version": version,
    })
    if not version_doc:
        raise HTTPException(status_code=404, detail="Version not found")

    return {"success": True, "data": serialize_doc(version_doc)}


@app.post("/api/code/files/{file_id}/rollback/{version}")
def rollback_version(file_id: str, version: int):
    try:
        file = db.code_files.find_one({"_id": ObjectId(file_id)})
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid file ID")
    if not file:
        raise HTTPException(status_code=404, detail="File not found")

    version_doc = db.code_versions.find_one({
        "file_id": file_id,
        "version": version,
    })
    if not version_doc:
        raise HTTPException(status_code=404, detail="Version not found")

    db.code_files.update_one(
        {"_id": ObjectId(file_id)},
        {
            "$set": {
                "content": version_doc["content"],
                "updated_at": datetime.utcnow(),
            },
            "$inc": {"version": 1},
        },
    )

    updated_file = db.code_files.find_one({"_id": ObjectId(file_id)})
    return {"success": True, "data": serialize_doc(updated_file)}
