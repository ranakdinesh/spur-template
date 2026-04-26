-- name: CreateSession :exec
INSERT INTO fosite_sessions (
    signature,
    request_id,
    client_id,
    tenant_id,
    subject,
    type,
    active,
    requested_at,
    expires_at,
    form_data,
    session_data
) VALUES (
             $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
         );

-- name: GetSession :one
SELECT * FROM fosite_sessions
WHERE signature = $1 AND type = $2;

-- name: DeleteSessionByType :execrows
DELETE FROM fosite_sessions
WHERE signature = $1 AND type = $2;

-- name: RevokeSessionByRequestIdAndType :exec
UPDATE fosite_sessions
SET active = false
WHERE request_id = $1 AND type = $2;

-- name: RevokeSessionByRequestId :exec
UPDATE fosite_sessions
SET active = false
WHERE request_id = $1;