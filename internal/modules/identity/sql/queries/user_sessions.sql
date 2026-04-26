-- name: CreateUserSession :one
INSERT INTO user_sessions (
  id,
  user_id,
  token,
  ip_address,
  user_agent,
  expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetUserSessionByToken :one
SELECT s.*, u.tenant_id
FROM user_sessions s
JOIN users u ON s.user_id = u.id
WHERE s.token = $1 LIMIT 1;

-- name: DeleteUserSession :exec
DELETE FROM user_sessions
WHERE token = $1;

-- name: DeleteUserSessionsByUserID :exec
DELETE FROM user_sessions
WHERE user_id = $1;
