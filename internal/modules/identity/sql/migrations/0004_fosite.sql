-- 1. Clients Table
CREATE TABLE fosite_clients (
                                id TEXT PRIMARY KEY,
                                client_secret TEXT,
                                tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
                                redirect_uris TEXT[],
                                grant_types TEXT[],
                                response_types TEXT[],
                                scopes TEXT[],
                                audience TEXT[],
                                public BOOLEAN NOT NULL DEFAULT FALSE,
                                active BOOLEAN NOT NULL DEFAULT TRUE,
                                created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Sessions Table (Fixed: Added 'type')
CREATE TABLE fosite_sessions (
                                 signature TEXT NOT NULL,
                                 type TEXT NOT NULL,              -- <--- This was missing!
                                 request_id TEXT NOT NULL,
                                 client_id TEXT NOT NULL REFERENCES fosite_clients(id),
                                 tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
                                 subject TEXT NOT NULL,
                                 active BOOLEAN NOT NULL DEFAULT TRUE,
                                 requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                 expires_at TIMESTAMPTZ NOT NULL,
                                 form_data JSONB NOT NULL,
                                 session_data JSONB NOT NULL,

    -- Composite Primary Key to match our queries
                                 PRIMARY KEY (signature, type)
);

-- Index for revocation by Request ID (speed up Logout/Revoke)
CREATE INDEX idx_fosite_sessions_req_id ON fosite_sessions (request_id);