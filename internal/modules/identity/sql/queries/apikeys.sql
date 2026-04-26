-- name: CreateAPIKey :one
INSERT INTO api_keys (
    id,
    tenant_id,
    name,
    type,
    prefix,
    key_hash,
    scopes,
    allowed_origins,
    expires_at
) VALUES (
             $1, $2, $3, $4, $5, $6, $7, $8, $9
         )
RETURNING *;

-- name: GetAPIKey :one
SELECT * FROM api_keys
WHERE id = $1 AND tenant_id = $2;

-- name: GetAPIKeyByPrefix :one
SELECT * FROM api_keys
WHERE prefix = $1;

-- name: ListAPIKeys :many
SELECT * FROM api_keys
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys
SET last_used_at = NOW()
WHERE id = $1;

-- name: DeleteAPIKey :exec
DELETE FROM api_keys
WHERE id = $1 AND tenant_id = $2;
