-- name: CreateRole :one
INSERT INTO roles (
    id,
    tenant_id,
    name,
    code,
    description,
    is_system
) VALUES (
             $1, $2, $3, $4, $5, $6
         )
RETURNING *;

-- name: GetRole :one
SELECT * FROM roles
WHERE id = $1 AND tenant_id = $2;

-- name: GetRoleByID :one
SELECT * FROM roles
WHERE id = $1;

-- name: GetRoleByCode :one
SELECT * FROM roles
WHERE code = $1 AND tenant_id = $2;

-- name: GetSystemRoleByCode :one
SELECT * FROM roles
WHERE code = $1 
  AND is_system = TRUE 
  AND tenant_id IN (SELECT id FROM tenants WHERE kind = 'ops')
LIMIT 1;

-- name: ListRoles :many
SELECT * FROM roles
WHERE tenant_id = $1
   OR (tenant_id IN (SELECT id FROM tenants WHERE kind = 'ops') AND is_system = TRUE)
ORDER BY created_at DESC;

-- name: ListAllRoles :many
SELECT * FROM roles
ORDER BY created_at DESC;

-- name: UpdateRole :exec
UPDATE roles
SET name = $3, description = $4, code = $5
WHERE id = $1 AND tenant_id = $2;

-- name: AssignRoleToUser :exec
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveRoleFromUser :exec
DELETE FROM user_roles
WHERE user_id = $1 AND role_id = $2;

-- name: ListUserRoles :many
SELECT r.* FROM roles r
                    JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = $1 AND tenant_id = $2;

-- Modules

-- name: UpsertModule :one
INSERT INTO modules (id, code, name, description)
VALUES ($1, $2, $3, $4)
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description
RETURNING *;

-- name: GetModuleByCode :one
SELECT * FROM modules WHERE code = $1;

-- name: UpsertPermission :one
INSERT INTO permissions (id, key, description, module, module_id)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (module, key) DO UPDATE SET
    description = EXCLUDED.description
RETURNING *;

-- name: CreateModule :one
INSERT INTO modules (id, code, name, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetModule :one
SELECT * FROM modules WHERE id = $1;

-- name: ListModules :many
SELECT * FROM modules ORDER BY code ASC;

-- name: UpdateModule :exec
UPDATE modules SET name = $2, description = $3 WHERE id = $1;

-- name: DeleteModule :exec
DELETE FROM modules WHERE id = $1;

-- Permissions

-- name: CreatePermission :one
INSERT INTO permissions (id, key, description, module, module_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPermission :one
SELECT * FROM permissions WHERE id = $1;

-- name: ListPermissions :many
SELECT * FROM permissions ORDER BY key ASC;

-- name: ListPermissionsByModule :many
SELECT * FROM permissions WHERE module_id = $1 ORDER BY key ASC;

-- name: UpdatePermission :exec
UPDATE permissions
SET key = $2, description = $3, module = $4, module_id = $5
WHERE id = $1;

-- name: DeletePermission :exec
DELETE FROM permissions WHERE id = $1;

-- Role Permissions

-- name: AssignPermissionToRole :exec
INSERT INTO role_permissions (role_id, permission_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemovePermissionFromRole :exec
DELETE FROM role_permissions
WHERE role_id = $1 AND permission_id = $2;

-- name: ListRolePermissions :many
SELECT p.* FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = $1;

-- Tenant Modules

-- name: EnableModuleForTenant :exec
INSERT INTO tenant_modules (tenant_id, module_id, status)
VALUES ($1, $2, $3)
ON CONFLICT (tenant_id, module_id) DO UPDATE SET status = EXCLUDED.status;

-- name: DisableModuleForTenant :exec
DELETE FROM tenant_modules
WHERE tenant_id = $1 AND module_id = $2;

-- name: ListTenantModules :many
SELECT m.* FROM modules m
JOIN tenant_modules tm ON m.id = tm.module_id
WHERE tm.tenant_id = $1;