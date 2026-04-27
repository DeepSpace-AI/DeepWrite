import os
import re
from contextlib import asynccontextmanager
from datetime import datetime
from typing import Optional

import bibtexparser
import httpx
from fastapi import FastAPI, HTTPException, Query, UploadFile, File
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field
from pymongo import MongoClient, TEXT
from bson import ObjectId

MONGO_URL = os.getenv("MONGO_URL", "mongodb://admin:password@reference-db:27017")
MONGO_DB = os.getenv("MONGO_DB", "reference_db")

client = None
db = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global client, db
    client = MongoClient(MONGO_URL)
    db = client[MONGO_DB]
    db.references.create_index([("project_id", 1), ("created_at", -1)])
    try:
        db.references.create_index([("title", TEXT), ("abstract", TEXT), ("keywords", TEXT)])
    except Exception:
        existing_indexes = db.references.list_indexes()
        for idx in existing_indexes:
            if idx.get("key", {}).get("_fts") == "text":
                db.references.drop_index(idx["name"])
                break
        db.references.create_index([("title", TEXT), ("abstract", TEXT), ("keywords", TEXT)])
    yield
    client.close()


app = FastAPI(title="DeepWrite Reference Service", lifespan=lifespan)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


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


class DOIImportRequest(BaseModel):
    project_id: str
    doi: str = Field(..., min_length=1)


class CitationFormatRequest(BaseModel):
    style: str = Field(..., pattern="^(apa|mla|gb7714|chicago|harvard)$")


def serialize_doc(doc):
    if doc is None:
        return None
    doc["id"] = str(doc.pop("_id"))
    for key, val in doc.items():
        if isinstance(val, datetime):
            doc[key] = val.isoformat()
    return doc


def clean_none_fields(doc: dict) -> dict:
    for key in ["journal", "doi", "url", "abstract", "citation_key", "bibtex", "pdf_url", "notes"]:
        if doc.get(key) is None:
            doc[key] = ""
    return doc


async def fetch_doi_metadata(doi: str) -> dict:
    async with httpx.AsyncClient() as client:
        headers = {"Accept": "application/vnd.citationstyles.csl+json"}
        response = await client.get(
            f"https://doi.org/{doi}",
            headers=headers,
            follow_redirects=True,
            timeout=10.0,
        )
        if response.status_code != 200:
            raise HTTPException(status_code=400, detail=f"Failed to fetch DOI: {response.status_code}")

        data = response.json()

        authors = []
        for author in data.get("author", []):
            name = f"{author.get('given', '')} {author.get('family', '')}".strip()
            if name:
                authors.append(name)

        year = None
        if data.get("published"):
            date_parts = data["published"].get("date-parts", [[]])
            if date_parts and date_parts[0]:
                year = date_parts[0][0]

        return {
            "title": data.get("title", ""),
            "authors": authors,
            "year": year,
            "journal": data.get("container-title", ""),
            "doi": doi,
            "url": data.get("URL", ""),
            "abstract": data.get("abstract", ""),
            "keywords": data.get("subject", []),
        }


def parse_bibtex(content: str) -> list[dict]:
    try:
        bib_database = bibtexparser.loads(content)
        references = []
        for entry in bib_database.entries:
            authors_str = entry.get("author", "")
            authors = [a.strip() for a in authors_str.split(" and ") if a.strip()]

            year = None
            if entry.get("year"):
                try:
                    year = int(entry["year"])
                except ValueError:
                    pass

            ref = {
                "title": entry.get("title", ""),
                "authors": authors,
                "year": year,
                "journal": entry.get("journal", entry.get("booktitle", "")),
                "doi": entry.get("doi", ""),
                "url": entry.get("url", ""),
                "abstract": entry.get("abstract", ""),
                "keywords": [k.strip() for k in entry.get("keywords", "").split(",") if k.strip()],
                "citation_key": entry.get("ID", ""),
                "bibtex": bibtexparser.dumps(bib_database),
            }
            references.append(ref)
        return references
    except Exception as e:
        raise HTTPException(status_code=400, detail=f"Failed to parse BibTeX: {str(e)}")


def format_authors_apa(authors: list[str]) -> str:
    if not authors:
        return ""
    if len(authors) == 1:
        parts = authors[0].split()
        return f"{parts[-1]}, {' '.join(parts[:-1])}" if len(parts) > 1 else authors[0]
    if len(authors) == 2:
        parts1 = authors[0].split()
        name1 = f"{parts1[-1]}, {' '.join(parts1[:-1])}" if len(parts1) > 1 else authors[0]
        return f"{name1}, & {authors[1]}"
    parts1 = authors[0].split()
    name1 = f"{parts1[-1]}, {' '.join(parts1[:-1])}" if len(parts1) > 1 else authors[0]
    return f"{name1}, et al."


def format_authors_mla(authors: list[str]) -> str:
    if not authors:
        return ""
    if len(authors) == 1:
        return authors[0]
    if len(authors) == 2:
        return f"{authors[0]}, and {authors[1]}"
    return f"{authors[0]}, et al."


def format_authors_gb7714(authors: list[str]) -> str:
    if not authors:
        return ""
    if len(authors) <= 3:
        return ", ".join(authors)
    return f"{authors[0]}, {authors[1]}, {authors[2]}, et al"


def format_citation(ref: dict, style: str) -> str:
    authors = ref.get("authors", [])
    title = ref.get("title", "")
    journal = ref.get("journal", "")
    year = ref.get("year", "")
    doi = ref.get("doi", "")
    volume = ref.get("volume", "")
    pages = ref.get("pages", "")

    if style == "apa":
        author_str = format_authors_apa(authors)
        year_str = f"({year})" if year else "(n.d.)"
        doi_str = f" https://doi.org/{doi}" if doi else ""
        return f"{author_str} {year_str}. {title}. *{journal}*{doi_str}"

    elif style == "mla":
        author_str = format_authors_mla(authors)
        return f'{author_str}. "{title}." *{journal}*, {year}.'

    elif style == "gb7714":
        author_str = format_authors_gb7714(authors)
        return f"{author_str}. {title}[J]. {journal}, {year}."

    elif style == "chicago":
        author_str = format_authors_apa(authors)
        return f'{author_str}. "{title}." *{journal}* ({year}).'

    elif style == "harvard":
        author_str = format_authors_apa(authors)
        return f"{author_str} ({year}) '{title}', *{journal}*."

    return f"{', '.join(authors)}. {title}. {journal}, {year}."


@app.get("/health")
def health_check():
    return {"status": "healthy"}


@app.post("/api/references")
def create_reference(ref: ReferenceCreate):
    doc = ref.model_dump()
    now = datetime.utcnow()
    doc["created_at"] = doc["updated_at"] = now
    doc = clean_none_fields(doc)
    result = db.references.insert_one(doc)
    created = db.references.find_one({"_id": result.inserted_id})
    return {"success": True, "data": serialize_doc(created)}


@app.get("/api/references")
def list_references(
    project_id: str = Query(...),
    page: int = Query(1, ge=1),
    limit: int = Query(20, ge=1, le=100),
    search: Optional[str] = Query(None),
    tags: Optional[str] = Query(None),
    year_from: Optional[int] = Query(None),
    year_to: Optional[int] = Query(None),
):
    skip = (page - 1) * limit
    query = {"project_id": project_id}

    if search:
        query["$or"] = [
            {"title": {"$regex": search, "$options": "i"}},
            {"abstract": {"$regex": search, "$options": "i"}},
            {"authors": {"$regex": search, "$options": "i"}},
        ]

    if tags:
        tag_list = [t.strip() for t in tags.split(",") if t.strip()]
        if tag_list:
            query["tags"] = {"$in": tag_list}

    if year_from or year_to:
        year_query = {}
        if year_from:
            year_query["$gte"] = year_from
        if year_to:
            year_query["$lte"] = year_to
        if year_query:
            query["year"] = year_query

    total = db.references.count_documents(query)
    refs = list(
        db.references.find(query)
        .skip(skip)
        .limit(limit)
        .sort("created_at", -1)
    )
    return {
        "success": True,
        "data": [serialize_doc(ref) for ref in refs],
        "meta": {"page": page, "limit": limit, "total": total},
    }


@app.get("/api/references/{ref_id}")
def get_reference(ref_id: str):
    try:
        ref = db.references.find_one({"_id": ObjectId(ref_id)})
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid reference ID")
    if not ref:
        raise HTTPException(status_code=404, detail="Reference not found")
    return {"success": True, "data": serialize_doc(ref)}


@app.put("/api/references/{ref_id}")
def update_reference(ref_id: str, update: ReferenceUpdate):
    update_data = {k: v for k, v in update.model_dump().items() if v is not None}
    if update_data:
        update_data["updated_at"] = datetime.utcnow()
        try:
            result = db.references.update_one(
                {"_id": ObjectId(ref_id)}, {"$set": update_data}
            )
        except Exception:
            raise HTTPException(status_code=400, detail="Invalid reference ID")
        if result.matched_count == 0:
            raise HTTPException(status_code=404, detail="Reference not found")
    ref = db.references.find_one({"_id": ObjectId(ref_id)})
    return {"success": True, "data": serialize_doc(ref)}


@app.delete("/api/references/{ref_id}")
def delete_reference(ref_id: str):
    try:
        result = db.references.delete_one({"_id": ObjectId(ref_id)})
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid reference ID")
    if result.deleted_count == 0:
        raise HTTPException(status_code=404, detail="Reference not found")
    return {"success": True, "message": "Reference deleted"}


@app.post("/api/references/import/doi")
async def import_from_doi(request: DOIImportRequest):
    metadata = await fetch_doi_metadata(request.doi)

    existing = db.references.find_one({"doi": request.doi, "project_id": request.project_id})
    if existing:
        raise HTTPException(status_code=409, detail="Reference with this DOI already exists")

    doc = {
        "project_id": request.project_id,
        "title": metadata["title"],
        "authors": metadata["authors"],
        "year": metadata["year"],
        "journal": metadata["journal"],
        "doi": metadata["doi"],
        "url": metadata["url"],
        "abstract": metadata["abstract"],
        "keywords": metadata["keywords"],
        "citation_key": "",
        "bibtex": "",
        "pdf_url": "",
        "notes": "",
        "tags": [],
        "created_at": datetime.utcnow(),
        "updated_at": datetime.utcnow(),
    }
    doc = clean_none_fields(doc)
    result = db.references.insert_one(doc)
    created = db.references.find_one({"_id": result.inserted_id})
    return {"success": True, "data": serialize_doc(created)}


@app.post("/api/references/import/bibtex")
async def import_from_bibtex(
    project_id: str = Query(...),
    file: UploadFile = File(...),
):
    content = await file.read()
    text = content.decode("utf-8")

    parsed_refs = parse_bibtex(text)
    imported = []

    for ref_data in parsed_refs:
        ref_data["project_id"] = project_id
        ref_data["pdf_url"] = ""
        ref_data["notes"] = ""
        ref_data["tags"] = []
        now = datetime.utcnow()
        ref_data["created_at"] = now
        ref_data["updated_at"] = now
        ref_data = clean_none_fields(ref_data)
        result = db.references.insert_one(ref_data)
        created = db.references.find_one({"_id": result.inserted_id})
        imported.append(serialize_doc(created))

    return {"success": True, "data": imported, "count": len(imported)}


@app.post("/api/references/{ref_id}/citation")
def get_citation(ref_id: str, request: CitationFormatRequest):
    try:
        ref = db.references.find_one({"_id": ObjectId(ref_id)})
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid reference ID")
    if not ref:
        raise HTTPException(status_code=404, detail="Reference not found")

    citation = format_citation(ref, request.style)
    return {"success": True, "data": {"citation": citation, "style": request.style}}


@app.get("/api/references/search/external")
async def search_external(
    query: str = Query(..., min_length=1),
    source: str = Query("crossref", pattern="^(crossref|semantic_scholar)$"),
    limit: int = Query(10, ge=1, le=50),
):
    results = []

    async with httpx.AsyncClient() as client:
        if source == "crossref":
            response = await client.get(
                "https://api.crossref.org/works",
                params={"query": query, "rows": limit},
                timeout=15.0,
            )
            if response.status_code == 200:
                data = response.json()
                for item in data.get("message", {}).get("items", []):
                    authors = []
                    for author in item.get("author", []):
                        name = f"{author.get('given', '')} {author.get('family', '')}".strip()
                        if name:
                            authors.append(name)

                    year = None
                    if item.get("published"):
                        date_parts = item["published"].get("date-parts", [[]])
                        if date_parts and date_parts[0]:
                            year = date_parts[0][0]

                    results.append({
                        "title": item.get("title", [""])[0] if item.get("title") else "",
                        "authors": authors,
                        "year": year,
                        "journal": item.get("container-title", [""])[0] if item.get("container-title") else "",
                        "doi": item.get("DOI", ""),
                        "abstract": item.get("abstract", ""),
                    })

        elif source == "semantic_scholar":
            response = await client.get(
                "https://api.semanticscholar.org/graph/v1/paper/search",
                params={"query": query, "limit": limit, "fields": "title,authors,year,venue,abstract,externalIds"},
                timeout=15.0,
            )
            if response.status_code == 200:
                data = response.json()
                for item in data.get("data", []):
                    authors = [a.get("name", "") for a in item.get("authors", [])]
                    external_ids = item.get("externalIds", {})

                    results.append({
                        "title": item.get("title", ""),
                        "authors": authors,
                        "year": item.get("year"),
                        "journal": item.get("venue", ""),
                        "doi": external_ids.get("DOI", ""),
                        "abstract": item.get("abstract", ""),
                    })

    return {"success": True, "data": results, "source": source}
