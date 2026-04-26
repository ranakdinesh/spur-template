-- 1. Create tenant_modules table
CREATE TABLE tenant_modules (
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    module_id UUID NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active',
    enabled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, module_id)
);

-- 2. Enable RLS
ALTER TABLE tenant_modules ENABLE ROW LEVEL SECURITY;

-- 3. RLS Policies
CREATE POLICY tenant_modules_isolation ON tenant_modules
    USING (
        tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid
        OR
        current_setting('app.is_super_admin', true) = 'true'
    );
