-- Session 表新增字段
ALTER TABLE dw_sessions ADD COLUMN IF NOT EXISTS pinned BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE dw_sessions ADD COLUMN IF NOT EXISTS archived BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE dw_sessions ADD COLUMN IF NOT EXISTS group_id UUID;
ALTER TABLE dw_sessions ADD COLUMN IF NOT EXISTS tags JSONB NOT NULL DEFAULT '[]'::jsonb;

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_sessions_pinned ON dw_sessions(pinned);
CREATE INDEX IF NOT EXISTS idx_sessions_archived ON dw_sessions(archived);
CREATE INDEX IF NOT EXISTS idx_sessions_group_id ON dw_sessions(group_id);

-- SessionGroup 表
CREATE TABLE IF NOT EXISTS dw_session_groups (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  color VARCHAR(20) DEFAULT '#6366f1',
  icon VARCHAR(50),
  sort_order INT DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_session_groups_user_id ON dw_session_groups(user_id);

-- SessionShare 表
CREATE TABLE IF NOT EXISTS dw_session_shares (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id UUID NOT NULL,
  user_id UUID NOT NULL,
  share_token VARCHAR(32) UNIQUE NOT NULL,
  title VARCHAR(255),
  expires_at TIMESTAMP,
  view_count INT DEFAULT 0,
  allow_copy BOOLEAN DEFAULT TRUE,
  is_public BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_session_shares_session_id ON dw_session_shares(session_id);
CREATE INDEX IF NOT EXISTS idx_session_shares_user_id ON dw_session_shares(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_session_shares_token ON dw_session_shares(share_token);