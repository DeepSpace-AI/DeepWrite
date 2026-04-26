-- ============================================================
-- DeepWrite - code-service Database Initialization
-- Database: code_db
-- Description: Code files and execution runs
-- ============================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- Code Files Table
-- ============================================================
CREATE TABLE IF NOT EXISTS code_files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    language VARCHAR(50) NOT NULL,
    content TEXT,
    size_bytes INTEGER DEFAULT 0,
    version_count INTEGER DEFAULT 1,
    is_deleted BOOLEAN DEFAULT false,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE code_files IS 'Code files within projects';
COMMENT ON COLUMN code_files.language IS 'python, r, matlab, julia, etc.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_code_files_project_id ON code_files(project_id);
CREATE INDEX IF NOT EXISTS idx_code_files_language ON code_files(language);
CREATE INDEX IF NOT EXISTS idx_code_files_is_deleted ON code_files(is_deleted);
CREATE INDEX IF NOT EXISTS idx_code_files_created_at ON code_files(created_at);
CREATE INDEX IF NOT EXISTS idx_code_files_project_path ON code_files(project_id, file_path) WHERE is_deleted = false;

-- ============================================================
-- Code Runs Table
-- ============================================================
CREATE TABLE IF NOT EXISTS code_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code_file_id UUID NOT NULL REFERENCES code_files(id) ON DELETE CASCADE,
    run_number INTEGER NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    input_data TEXT,
    output_data TEXT,
    error_message TEXT,
    execution_time_ms INTEGER,
    memory_usage_mb INTEGER,
    container_id VARCHAR(100),
    triggered_by UUID NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE code_runs IS 'Code execution history';
COMMENT ON COLUMN code_runs.status IS 'pending, running, success, failed, timeout';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_code_runs_code_file_id ON code_runs(code_file_id);
CREATE INDEX IF NOT EXISTS idx_code_runs_status ON code_runs(status);
CREATE INDEX IF NOT EXISTS idx_code_runs_triggered_by ON code_runs(triggered_by);
CREATE INDEX IF NOT EXISTS idx_code_runs_created_at ON code_runs(created_at);
CREATE INDEX IF NOT EXISTS idx_code_runs_file_run_number ON code_runs(code_file_id, run_number);

-- ============================================================
-- Code File Versions Table
-- ============================================================
CREATE TABLE IF NOT EXISTS code_file_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code_file_id UUID NOT NULL REFERENCES code_files(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    content TEXT NOT NULL,
    change_description TEXT,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE code_file_versions IS 'Code file version history';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_code_file_versions_code_file_id ON code_file_versions(code_file_id);
CREATE INDEX IF NOT EXISTS idx_code_file_versions_version_number ON code_file_versions(code_file_id, version_number);

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
CREATE TRIGGER update_code_files_updated_at BEFORE UPDATE ON code_files
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- Row Level Security (RLS) Policies
-- ============================================================
ALTER TABLE code_files ENABLE ROW LEVEL SECURITY;
ALTER TABLE code_runs ENABLE ROW LEVEL SECURITY;
ALTER TABLE code_file_versions ENABLE ROW LEVEL SECURITY;

-- Code files accessible by project members
CREATE POLICY code_files_project_access ON code_files
    FOR ALL USING (
        created_by = current_setting('app.current_user_id')::UUID
        OR project_id IN (
            SELECT project_id FROM project_members
            WHERE user_id = current_setting('app.current_user_id')::UUID
        )
    );

-- Code runs accessible by code file owners
CREATE POLICY code_runs_access ON code_runs
    FOR ALL USING (
        triggered_by = current_setting('app.current_user_id')::UUID
        OR code_file_id IN (
            SELECT id FROM code_files
            WHERE created_by = current_setting('app.current_user_id')::UUID
        )
    );
