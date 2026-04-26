CREATE TABLE users (
                       id UUID PRIMARY KEY,
                       tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
                       first_name TEXT NOT NULL,
                       last_name TEXT NOT NULL,
                       email TEXT NOT NULL,
                       mobile TEXT,
                       password_hash TEXT NOT NULL,

    -- Status & RBAC
                       is_super_admin BOOLEAN NOT NULL DEFAULT FALSE,
                       authz_version INTEGER NOT NULL DEFAULT 1,
                       is_active BOOLEAN NOT NULL DEFAULT TRUE,

    -- Verification Info
                       email_verified_at TIMESTAMPTZ,
                       mobile_verified_at TIMESTAMPTZ,

                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX users_email_idx ON users (tenant_id, email);
CREATE UNIQUE INDEX unique_super_admin ON users (is_super_admin) WHERE is_super_admin = TRUE;