-- ============================================================
-- DeepWrite - user-service Database Initialization
-- Database: user_db
-- Description: Users, teams, subscriptions, and usage tracking
-- ============================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- Users Table
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    institution VARCHAR(255),
    research_fields TEXT[],
    subscription_tier VARCHAR(50) DEFAULT 'free',
    subscription_expires_at TIMESTAMP WITH TIME ZONE,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE users IS 'User accounts';
COMMENT ON COLUMN users.subscription_tier IS 'free, pro, enterprise';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_subscription_tier ON users(subscription_tier);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- ============================================================
-- Teams Table
-- ============================================================
CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    max_members INTEGER DEFAULT 5,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE teams IS 'Research teams';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_teams_owner_id ON teams(owner_id);
CREATE INDEX IF NOT EXISTS idx_teams_created_at ON teams(created_at);

-- ============================================================
-- Team Members Table
-- ============================================================
CREATE TABLE IF NOT EXISTS team_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, user_id)
);

COMMENT ON TABLE team_members IS 'Team membership';
COMMENT ON COLUMN team_members.role IS 'owner, admin, member, viewer';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id);
CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id);
CREATE INDEX IF NOT EXISTS idx_team_members_role ON team_members(role);

-- ============================================================
-- Subscription Plans Table
-- ============================================================
CREATE TABLE IF NOT EXISTS subscription_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    tier VARCHAR(50) NOT NULL UNIQUE,
    price_monthly DECIMAL(10,2),
    price_yearly DECIMAL(10,2),
    features JSONB NOT NULL DEFAULT '{}',
    limits JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE subscription_plans IS 'Subscription plan definitions';
COMMENT ON COLUMN subscription_plans.tier IS 'free, pro, enterprise';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_subscription_plans_tier ON subscription_plans(tier);
CREATE INDEX IF NOT EXISTS idx_subscription_plans_is_active ON subscription_plans(is_active);

-- Insert default plans
INSERT INTO subscription_plans (name, tier, price_monthly, price_yearly, features, limits) VALUES
('Free', 'free', 0, 0, 
 '{"projects": true, "ai_assist": true, "collaboration": true, "code_editor": true, "image_editor": true}'::jsonb,
 '{"max_projects": 3, "max_storage_gb": 1, "max_ai_calls_per_month": 100, "max_team_members": 1}'::jsonb
),
('Professional', 'pro', 99.00, 990.00,
 '{"projects": true, "ai_assist": true, "collaboration": true, "code_editor": true, "image_editor": true, "advanced_ai": true, "priority_support": true}'::jsonb,
 '{"max_projects": -1, "max_storage_gb": 50, "max_ai_calls_per_month": 1000, "max_team_members": 5}'::jsonb
),
('Enterprise', 'enterprise', 299.00, 2990.00,
 '{"projects": true, "ai_assist": true, "collaboration": true, "code_editor": true, "image_editor": true, "advanced_ai": true, "priority_support": true, "dedicated_support": true, "custom_integration": true}'::jsonb,
 '{"max_projects": -1, "max_storage_gb": 500, "max_ai_calls_per_month": -1, "max_team_members": -1}'::jsonb
)
ON CONFLICT (tier) DO NOTHING;

-- ============================================================
-- Subscriptions Table
-- ============================================================
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES subscription_plans(id),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    current_period_start TIMESTAMP WITH TIME ZONE,
    current_period_end TIMESTAMP WITH TIME ZONE,
    cancel_at_period_end BOOLEAN DEFAULT false,
    payment_provider VARCHAR(50),
    payment_provider_subscription_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE subscriptions IS 'User subscriptions';
COMMENT ON COLUMN subscriptions.status IS 'active, canceled, past_due, unpaid, trialing';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_plan_id ON subscriptions(plan_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_current_period_end ON subscriptions(current_period_end);

-- ============================================================
-- Usage Records Table
-- ============================================================
CREATE TABLE IF NOT EXISTS usage_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    resource_type VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    metadata JSONB DEFAULT '{}',
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE usage_records IS 'Resource usage tracking';
COMMENT ON COLUMN usage_records.resource_type IS 'ai_call, storage_gb, team_member, project';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_usage_records_user_id ON usage_records(user_id);
CREATE INDEX IF NOT EXISTS idx_usage_records_resource_type ON usage_records(resource_type);
CREATE INDEX IF NOT EXISTS idx_usage_records_recorded_at ON usage_records(recorded_at);
CREATE INDEX IF NOT EXISTS idx_usage_records_user_resource_date ON usage_records(user_id, resource_type, recorded_at);

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
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_teams_updated_at BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_subscription_plans_updated_at BEFORE UPDATE ON subscription_plans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- Row Level Security (RLS) Policies
-- ============================================================
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE teams ENABLE ROW LEVEL SECURITY;
ALTER TABLE team_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE usage_records ENABLE ROW LEVEL SECURITY;

-- Users can only see their own data
CREATE POLICY users_self_access ON users
    FOR ALL USING (id = current_setting('app.current_user_id')::UUID);

-- Teams accessible by members
CREATE POLICY teams_member_access ON teams
    FOR ALL USING (
        owner_id = current_setting('app.current_user_id')::UUID
        OR EXISTS (
            SELECT 1 FROM team_members 
            WHERE team_id = teams.id 
            AND user_id = current_setting('app.current_user_id')::UUID
        )
    );

-- Team members accessible by team members
CREATE POLICY team_members_team_access ON team_members
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM team_members tm
            WHERE tm.team_id = team_members.team_id
            AND tm.user_id = current_setting('app.current_user_id')::UUID
        )
    );
