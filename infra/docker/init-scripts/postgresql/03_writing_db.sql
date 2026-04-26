-- ============================================================
-- DeepWrite - writing-service Database Initialization
-- Database: writing_db
-- Description: Documents, versions, and writing metadata
-- ============================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- Documents Table (Metadata in PostgreSQL, content in MongoDB)
-- ============================================================
CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL,
    title VARCHAR(500) NOT NULL,
    abstract TEXT,
    author_id UUID NOT NULL,
    status VARCHAR(50) DEFAULT 'draft',
    word_count INTEGER DEFAULT 0,
    citation_count INTEGER DEFAULT 0,
    last_edited_by UUID,
    last_edited_at TIMESTAMP WITH TIME ZONE,
    mongo_content_id VARCHAR(24),
    format VARCHAR(50) DEFAULT 'markdown',
    is_deleted BOOLEAN DEFAULT false,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE documents IS 'Document metadata (content stored in MongoDB)';
COMMENT ON COLUMN documents.status IS 'draft, reviewing, published, archived';
COMMENT ON COLUMN documents.mongo_content_id IS 'Reference to MongoDB document content';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_documents_project_id ON documents(project_id);
CREATE INDEX IF NOT EXISTS idx_documents_author_id ON documents(author_id);
CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);
CREATE INDEX IF NOT EXISTS idx_documents_is_deleted ON documents(is_deleted);
CREATE INDEX IF NOT EXISTS idx_documents_created_at ON documents(created_at);
CREATE INDEX IF NOT EXISTS idx_documents_project_status ON documents(project_id, status) WHERE is_deleted = false;

-- ============================================================
-- Document Versions Table
-- ============================================================
CREATE TABLE IF NOT EXISTS document_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    title VARCHAR(500),
    word_count INTEGER DEFAULT 0,
    change_summary TEXT,
    mongo_snapshot_id VARCHAR(24),
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE document_versions IS 'Document version history snapshots';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_document_versions_document_id ON document_versions(document_id);
CREATE INDEX IF NOT EXISTS idx_document_versions_version_number ON document_versions(document_id, version_number);
CREATE INDEX IF NOT EXISTS idx_document_versions_created_at ON document_versions(created_at);

-- ============================================================
-- Document Citations Table
-- ============================================================
CREATE TABLE IF NOT EXISTS document_citations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    reference_id UUID NOT NULL,
    citation_key VARCHAR(100),
    section VARCHAR(100),
    paragraph INTEGER,
    offset INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE document_citations IS 'Citation positions within documents';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_document_citations_document_id ON document_citations(document_id);
CREATE INDEX IF NOT EXISTS idx_document_citations_reference_id ON document_citations(reference_id);

-- ============================================================
-- Document Collaborators Table
-- ============================================================
CREATE TABLE IF NOT EXISTS document_collaborators (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role VARCHAR(50) DEFAULT 'editor',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(document_id, user_id)
);

COMMENT ON TABLE document_collaborators IS 'Document-level collaboration permissions';
COMMENT ON COLUMN document_collaborators.role IS 'owner, editor, viewer';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_document_collaborators_document_id ON document_collaborators(document_id);
CREATE INDEX IF NOT EXISTS idx_document_collaborators_user_id ON document_collaborators(user_id);

-- ============================================================
-- Updated At Trigger Function
-- ============================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers
CREATE TRIGGER update_documents_updated_at BEFORE UPDATE ON documents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- Row Level Security (RLS) Policies
-- ============================================================
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE document_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE document_citations ENABLE ROW LEVEL SECURITY;
ALTER TABLE document_collaborators ENABLE ROW LEVEL SECURITY;

-- Documents accessible by collaborators
CREATE POLICY documents_collaborator_access ON documents
    FOR ALL USING (
        author_id = current_setting('app.current_user_id')::UUID
        OR EXISTS (
            SELECT 1 FROM document_collaborators
            WHERE document_id = documents.id
            AND user_id = current_setting('app.current_user_id')::UUID
        )
    );

-- Versions accessible by document collaborators
CREATE POLICY document_versions_access ON document_versions
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM document_collaborators
            WHERE document_id = document_versions.document_id
            AND user_id = current_setting('app.current_user_id')::UUID
        )
        OR EXISTS (
            SELECT 1 FROM documents
            WHERE id = document_versions.document_id
            AND author_id = current_setting('app.current_user_id')::UUID
        )
    );
