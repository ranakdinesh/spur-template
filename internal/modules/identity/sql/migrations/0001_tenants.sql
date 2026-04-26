CREATE TYPE tenant_kind AS ENUM ('ops', 'customer');

CREATE TABLE tenants (
                         id UUID PRIMARY KEY,
                         name TEXT NOT NULL,
                         kind tenant_kind NOT NULL DEFAULT 'customer',
                         created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                         updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Invariant: Only one 'ops' tenant allowed
CREATE UNIQUE INDEX unique_ops_tenant ON tenants (kind) WHERE kind = 'ops';