-- 1. Create modules table
CREATE TABLE modules (
    id UUID PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Add module_id to permissions and link
ALTER TABLE permissions ADD COLUMN module_id UUID REFERENCES modules(id) ON DELETE CASCADE;

-- 3. Update users table
ALTER TABLE users ADD COLUMN verification_grace_period_end TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN verified_at TIMESTAMPTZ;

-- 4. Update api_keys table
ALTER TABLE api_keys ADD COLUMN last_used_at TIMESTAMPTZ;

-- 5. RLS Policies

-- Enable RLS
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE modules ENABLE ROW LEVEL SECURITY; -- Enable for consistency, but maybe permissive policy.

-- Users Policy
CREATE POLICY users_isolation ON users
    USING (
        tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid
        OR
        current_setting('app.is_super_admin', true) = 'true'
    );

-- Roles Policy
CREATE POLICY roles_isolation ON roles
    USING (
        tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid
        OR
        current_setting('app.is_super_admin', true) = 'true'
    );

-- API Keys Policy
CREATE POLICY api_keys_isolation ON api_keys
    USING (
        tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid
        OR
        current_setting('app.is_super_admin', true) = 'true'
    );

-- Modules Policy (Readable by all authenticated users? Or just global?)
-- Modules are system-wide. Usually we don't need to filter by tenant.
-- But if we enable RLS, we MUST provide a policy, otherwise no one can see anything (default deny).
CREATE POLICY modules_read_all ON modules
    FOR SELECT
    USING (true);

-- Permissions Policy (Same as modules)
ALTER TABLE permissions ENABLE ROW LEVEL SECURITY;
CREATE POLICY permissions_read_all ON permissions
    FOR SELECT
    USING (true);