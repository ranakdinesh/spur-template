-- name: CreateUser :one
INSERT INTO users (
    id,
    tenant_id,
    first_name,
    last_name,
    email,
    mobile,
    password_hash,
    is_super_admin,
    is_active,
    is_locked,
    verification_grace_period_end
) VALUES (
             $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
         )
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 AND tenant_id = $2;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByMobile :one
SELECT * FROM users
WHERE mobile = $1;

-- name: ListUsers :many
SELECT * FROM users
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: ListAllUsers :many
SELECT * FROM users
ORDER BY created_at DESC;

-- name: UpdateUser :exec
UPDATE users
SET
    first_name = $2,
    last_name = $3,
    mobile = $4,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $5;

-- name: UpdateUserLockStatus :exec
UPDATE users
SET is_locked = $2, updated_at = NOW()
WHERE id = $1 AND tenant_id = $3;

-- name: UpdatePassword :exec
UPDATE users
SET password_hash = $3, updated_at = NOW()
WHERE id = $1 AND tenant_id = $2;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1 AND tenant_id = $2;

-- name: DeleteUserByEmail :exec
DELETE FROM users
WHERE email = $1;

-- name: DeleteUserByMobile :exec
DELETE FROM users
WHERE mobile = $1;
