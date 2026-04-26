CREATE TABLE roles (
                       id UUID PRIMARY KEY,
                       tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
                       name TEXT NOT NULL,
                       code TEXT, -- Stable identifier like "ADMIN"
                       description TEXT,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX roles_tenant_name_idx ON roles (tenant_id, name);

CREATE TABLE permissions (
                             id UUID PRIMARY KEY,
                             key TEXT NOT NULL UNIQUE, -- e.g. "users:create"
                             description TEXT,
                             module TEXT NOT NULL -- e.g. "identity", "crm"
);

CREATE TABLE role_permissions (
                                  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
                                  permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
                                  PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE user_roles (
                            user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                            role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
                            PRIMARY KEY (user_id, role_id)
);