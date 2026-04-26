-- name: CreateTenant :one
INSERT INTO tenants (
    id,
    name,
    kind,
    trial_ends_at,
    subscription_plan
) VALUES (
             $1, $2, $3, $4, $5
         )
RETURNING *;

-- name: GetTenant :one
SELECT * FROM tenants
WHERE id = $1;

-- name: ListTenants :many
SELECT * FROM tenants
ORDER BY created_at DESC;

-- name: UpdateTenant :exec
UPDATE tenants
SET name = $2, subscription_plan = $3, updated_at = NOW()
WHERE id = $1;

-- name: DeleteTenant :exec
DELETE FROM tenants
WHERE id = $1;