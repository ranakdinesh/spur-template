-- name: CreateClient :one
INSERT INTO fosite_clients (
    id,
    tenant_id,
    client_secret,
    redirect_uris,
    grant_types,
    response_types,
    scopes,
    audience,
    public,
    active
) VALUES (
             $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
         )
RETURNING *;

-- name: GetClient :one
SELECT * FROM fosite_clients
WHERE id = $1;

-- name: GetActiveClient :one
SELECT * FROM fosite_clients
WHERE id = $1 AND active = true;

-- name: ListClients :many
SELECT * FROM fosite_clients
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: ListPublicClients :many
SELECT * FROM fosite_clients
WHERE public = true
ORDER BY created_at DESC;

-- name: UpdateClientSecret :exec
UPDATE fosite_clients
SET client_secret = $2, updated_at = NOW()
WHERE id = $1;

-- name: ToggleClientStatus :exec
UPDATE fosite_clients
SET active = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateClientConfig :exec
UPDATE fosite_clients
SET
    redirect_uris = $2,
    scopes = $3,
    grant_types = $4,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteClient :exec
DELETE FROM fosite_clients
WHERE id = $1;