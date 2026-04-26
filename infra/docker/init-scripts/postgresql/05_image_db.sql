-- ============================================================
-- DeepWrite - image-service Database Initialization
-- Database: image_db
-- Description: Images, templates, and metadata
-- ============================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- Images Table
-- ============================================================
CREATE TABLE IF NOT EXISTS images (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    image_type VARCHAR(50) NOT NULL DEFAULT 'upload',
    s3_key VARCHAR(500) NOT NULL,
    s3_bucket VARCHAR(100) DEFAULT 'deepwrite-images',
    mime_type VARCHAR(100),
    size_bytes INTEGER,
    width INTEGER,
    height INTEGER,
    thumbnail_s3_key VARCHAR(500),
    metadata JSONB DEFAULT '{}',
    is_deleted BOOLEAN DEFAULT false,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE images IS 'Scientific images and figures';
COMMENT ON COLUMN images.image_type IS 'upload, generated, template';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_images_project_id ON images(project_id);
CREATE INDEX IF NOT EXISTS idx_images_image_type ON images(image_type);
CREATE INDEX IF NOT EXISTS idx_images_is_deleted ON images(is_deleted);
CREATE INDEX IF NOT EXISTS idx_images_created_at ON images(created_at);
CREATE INDEX IF NOT EXISTS idx_images_project_type ON images(project_id, image_type) WHERE is_deleted = false;

-- ============================================================
-- Image Templates Table
-- ============================================================
CREATE TABLE IF NOT EXISTS image_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    tags TEXT[],
    thumbnail_s3_key VARCHAR(500),
    template_data JSONB NOT NULL DEFAULT '{}',
    is_public BOOLEAN DEFAULT true,
    is_active BOOLEAN DEFAULT true,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE image_templates IS 'Reusable image templates';
COMMENT ON COLUMN image_templates.category IS 'flowchart, architecture, data_viz, diagram, etc.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_image_templates_category ON image_templates(category);
CREATE INDEX IF NOT EXISTS idx_image_templates_is_public ON image_templates(is_public);
CREATE INDEX IF NOT EXISTS idx_image_templates_is_active ON image_templates(is_active);
CREATE INDEX IF NOT EXISTS idx_image_templates_tags ON image_templates USING gin(tags);

-- Insert default templates
INSERT INTO image_templates (name, description, category, tags, template_data) VALUES
('Research Flowchart', 'Standard research methodology flowchart', 'flowchart', 
 ARRAY['research', 'methodology', 'flowchart'],
 '{"nodes": [], "edges": [], "defaultWidth": 800, "defaultHeight": 600}'::jsonb
),
('System Architecture', 'Microservices architecture diagram template', 'architecture',
 ARRAY['system', 'architecture', 'microservices'],
 '{"nodes": [], "edges": [], "defaultWidth": 1000, "defaultHeight": 700}'::jsonb
),
('Data Visualization', 'Standard data visualization layout', 'data_viz',
 ARRAY['data', 'chart', 'visualization'],
 '{"nodes": [], "edges": [], "defaultWidth": 800, "defaultHeight": 500}'::jsonb
)
ON CONFLICT DO NOTHING;

-- ============================================================
-- Image Versions Table
-- ============================================================
CREATE TABLE IF NOT EXISTS image_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    image_id UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    s3_key VARCHAR(500) NOT NULL,
    change_description TEXT,
    width INTEGER,
    height INTEGER,
    size_bytes INTEGER,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE image_versions IS 'Image version history';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_image_versions_image_id ON image_versions(image_id);
CREATE INDEX IF NOT EXISTS idx_image_versions_version_number ON image_versions(image_id, version_number);

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
CREATE TRIGGER update_images_updated_at BEFORE UPDATE ON images
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_image_templates_updated_at BEFORE UPDATE ON image_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- Row Level Security (RLS) Policies
-- ============================================================
ALTER TABLE images ENABLE ROW LEVEL SECURITY;
ALTER TABLE image_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE image_versions ENABLE ROW LEVEL SECURITY;

-- Images accessible by project members
CREATE POLICY images_project_access ON images
    FOR ALL USING (
        created_by = current_setting('app.current_user_id')::UUID
        OR project_id IN (
            SELECT project_id FROM project_members
            WHERE user_id = current_setting('app.current_user_id')::UUID
        )
    );

-- Public templates accessible by all
CREATE POLICY image_templates_public ON image_templates
    FOR SELECT USING (is_public = true AND is_active = true);

-- Private templates only by creator
CREATE POLICY image_templates_private ON image_templates
    FOR ALL USING (
        created_by = current_setting('app.current_user_id')::UUID
        OR is_public = true
    );
