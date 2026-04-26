import os
from contextlib import asynccontextmanager
from datetime import datetime
from typing import Optional

from fastapi import FastAPI, HTTPException, Query
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field
from pymongo import MongoClient
from bson import ObjectId, json_util
import json

MONGO_URL = os.getenv("MONGO_URL", "mongodb://admin:password@reference-db:27017")
MONGO_DB = os.getenv("MONGO_DB", "reference_db")

client = None
db = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    global client, db
    client = MongoClient(MONGO_URL)
    db = client[MONGO_DB]
    yield
    client.close()

app = FastAPI(title="DeepWrite Reference Service", lifespan=lifespan)

class ReferenceCreate(BaseModel):
    project_id: str
    title: str = Field(..., min_length=1, max_length=1000)
    authors: list[str] = []
    year: Optional[int] = None
    journal: Optional[str] = None
    doi: Optional[str] = None
    url: Optional[str] = None
    abstract: Optional[str] = None
    keywords: list[str] = []
    citation_key: Optional[str] = None
    bibtex: Optional[str] = None
    pdf_url: Optional[str] = None
    notes: Optional[str] = None
    tags: list[str] = []

class ReferenceUpdate(BaseModel):
    title: Optional[str] = None
    authors: Optional[list[str]] = None
    year: Optional[int] = None
    journal: Optional[str] = None
    doi: Optional[str] = None
    url: Optional[str] = None
    abstract: Optional[str] = None
    keywords: Optional[list[str]] = None
    citation_key: Optional[str] = None
    bibtex: Optional[str] = None
    pdf_url: Optional[str] = None
    notes: Optional[str] = None
    tags: Optional[list[str]] = None

@app.get("/health")
def health_check():
    return {"status": "healthy"}

@app.post("/api/references")
def create_reference(ref: ReferenceCreate):
    doc = ref.model_dump()
    now = datetime.utcnow()
    doc["created_at"] = doc["updated_at"] = now
    for key in ["journal", "doi", "url", "abstract", "citation_key", "bibtex", "pdf_url", "notes"]:
        if doc.get(key) is None:
            doc[key] = ""
    result = db.references.insert_one(doc)
    created = db.references.find_one({"_id": result.inserted_id})
    return {"success": True, "data": serialize_doc(created)}

def serialize_doc(doc):
    doc["id"] = str(doc.pop("_id"))
    for key, val in doc.items():
        if isinstance(val, datetime):
            doc[key] = val.isoformat()
    return doc

@app.get("/api/references")
def list_references(
    project_id: str = Query(...),
    page: int = Query(1, ge=1),
    limit: int = Query(20, ge=1, le=100),
):
    skip = (page - 1) * limit
    total = db.references.count_documents({"project_id": project_id})
    refs = list(db.references.find({"project_id": project_id}).skip(skip).limit(limit).sort("created_at", -1))
    return {"success": True, "data": [serialize_doc(ref) for ref in refs], "meta": {"page": page, "limit": limit, "total": total}}

@app.get("/api/references/{ref_id}")
def get_reference(ref_id: str):
    ref = db.references.find_one({"_id": ObjectId(ref_id)})
    if not ref:
        raise HTTPException(status_code=404, detail="Reference not found")
    return {"success": True, "data": serialize_doc(ref)}

@app.put("/api/references/{ref_id}")
def update_reference(ref_id: str, update: ReferenceUpdate):
    update_data = {k: v for k, v in update.model_dump().items() if v is not None}
    if update_data:
        update_data["updated_at"] = datetime.utcnow()
        result = db.references.update_one({"_id": ObjectId(ref_id)}, {"$set": update_data})
        if result.matched_count == 0:
            raise HTTPException(status_code=404, detail="Reference not found")
    ref = db.references.find_one({"_id": ObjectId(ref_id)})
    return {"success": True, "data": serialize_doc(ref)}

@app.delete("/api/references/{ref_id}")
def delete_reference(ref_id: str):
    result = db.references.delete_one({"_id": ObjectId(ref_id)})
    if result.deleted_count == 0:
        raise HTTPException(status_code=404, detail="Reference not found")
    return {"success": True, "message": "Reference deleted"}
