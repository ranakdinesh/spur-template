CREATE TABLE api_keys (
                          id UUID PRIMARY KEY,
                          tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
                          name TEXT NOT NULL,
                          prefix TEXT NOT NULL,
                          key_hash TEXT NOT NULL,
                          scopes TEXT[] DEFAULT '{}',
                          ip_allowlist INET[] DEFAULT NULL,
                          expires_at TIMESTAMPTZ NOT NULL,
                          created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);