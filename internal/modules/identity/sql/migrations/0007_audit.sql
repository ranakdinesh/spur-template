CREATE TABLE audit_events (
                              id UUID PRIMARY KEY,
                              actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
                              actor_tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
                              target_tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
                              action TEXT NOT NULL,
                              entity TEXT NOT NULL,
                              payload_json JSONB,
                              request_id TEXT,
                              created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);