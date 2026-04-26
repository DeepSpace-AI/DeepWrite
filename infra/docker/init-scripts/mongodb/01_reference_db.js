// ============================================================
// DeepWrite - MongoDB Initialization Script
// Database: reference_db
// Description: References and literature data
// ============================================================

// Create collections
db.createCollection('references');
db.createCollection('reference_collections');
db.createCollection('import_jobs');

// Create indexes for references collection
db.references.createIndex({ "project_id": 1 });
db.references.createIndex({ "project_id": 1, "created_at": -1 });
db.references.createIndex({ "doi": 1 }, { sparse: true });
db.references.createIndex({ "title": "text", "abstract": "text", "authors": "text" });
db.references.createIndex({ "tags": 1 });
db.references.createIndex({ "year": -1 });
db.references.createIndex({ "journal": 1 });
db.references.createIndex({ "project_id": 1, "citation_key": 1 });

// Create indexes for reference collections
db.reference_collections.createIndex({ "project_id": 1 });
db.reference_collections.createIndex({ "project_id": 1, "name": 1 }, { unique: true });

// Create indexes for import jobs
db.import_jobs.createIndex({ "project_id": 1 });
db.import_jobs.createIndex({ "status": 1 });
db.import_jobs.createIndex({ "created_at": -1 });

// Insert sample reference schema validation (optional, for documentation)
db.runCommand({
    collMod: 'references',
    validator: {
        $jsonSchema: {
            bsonType: 'object',
            required: ['project_id', 'title'],
            properties: {
                project_id: { bsonType: 'string' },
                title: { bsonType: 'string' },
                authors: { bsonType: 'array' },
                year: { bsonType: 'int' },
                journal: { bsonType: 'string' },
                doi: { bsonType: 'string' },
                url: { bsonType: 'string' },
                abstract: { bsonType: 'string' },
                keywords: { bsonType: 'array' },
                citation_key: { bsonType: 'string' },
                bibtex: { bsonType: 'string' },
                pdf_url: { bsonType: 'string' },
                notes: { bsonType: 'string' },
                tags: { bsonType: 'array' },
                created_at: { bsonType: 'date' },
                updated_at: { bsonType: 'date' }
            }
        }
    },
    validationLevel: 'moderate'
});

print('reference_db initialization complete');
