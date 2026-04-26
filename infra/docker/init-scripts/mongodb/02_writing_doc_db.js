// ============================================================
// DeepWrite - MongoDB Initialization Script
// Database: writing_doc_db
// Description: Document content and version history
// ============================================================

// Create collections
db.createCollection('document_contents');
db.createCollection('document_snapshots');
db.createCollection('collaboration_sessions');

// Create indexes for document contents
db.document_contents.createIndex({ "project_id": 1 });
db.document_contents.createIndex({ "document_id": 1 }, { unique: true });
db.document_contents.createIndex({ "project_id": 1, "created_at": -1 });
db.document_contents.createIndex({ "last_edited_by": 1 });

// Create indexes for document snapshots (version history)
db.document_snapshots.createIndex({ "document_id": 1 });
db.document_snapshots.createIndex({ "document_id": 1, "version_id": 1 });
db.document_snapshots.createIndex({ "document_id": 1, "created_at": -1 });
db.document_snapshots.createIndex({ "created_by": 1 });

// Create indexes for collaboration sessions
db.collaboration_sessions.createIndex({ "document_id": 1 });
db.collaboration_sessions.createIndex({ "document_id": 1, "user_id": 1 });
db.collaboration_sessions.createIndex({ "last_activity": 1 }, { expireAfterSeconds: 3600 });

// Insert sample document content structure
db.document_contents.insertOne({
    document_id: "sample-doc-id",
    project_id: "sample-project-id",
    title: "Sample Document",
    content: {
        abstract: "",
        introduction: "",
        methods: "",
        results: "",
        discussion: "",
        conclusion: ""
    },
    metadata: {
        word_count: 0,
        citation_count: 0,
        last_edited_by: "",
        format: "markdown"
    },
    citations: [],
    last_edited_at: new Date(),
    created_at: new Date(),
    updated_at: new Date()
});

// Remove sample document
db.document_contents.deleteOne({ document_id: "sample-doc-id" });

print('writing_doc_db initialization complete');
