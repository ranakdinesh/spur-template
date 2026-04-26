-- 1. Remove the old unique constraint on 'key'
-- First, find the constraint name if possible, or just drop index if created as such.
-- In migration 0003 it was defined as: key TEXT NOT NULL UNIQUE
-- Usually PostgreSQL creates an implicit index named 'permissions_key_key'
ALTER TABLE permissions DROP CONSTRAINT IF EXISTS permissions_key_key;

-- 2. Add composite unique constraint
ALTER TABLE permissions ADD CONSTRAINT unique_module_key UNIQUE (module, key);
