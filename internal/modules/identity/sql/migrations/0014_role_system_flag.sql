-- 1. Add is_system flag to roles
ALTER TABLE roles ADD COLUMN is_system BOOLEAN NOT NULL DEFAULT FALSE;

-- 2. Mark the 'Super Admin' role as a system role
-- We assume it belongs to the Ops tenant
UPDATE roles SET is_system = TRUE WHERE name = 'Super Admin' OR name = 'Tenant Admin';
