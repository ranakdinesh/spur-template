-- name: CreateContact :one
INSERT INTO contacts (id, tenant_id, lead_id, first_name, last_name, email, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
RETURNING id;

-- name: CreateAccount :one
INSERT INTO accounts (id, tenant_id, name, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id;

-- name: GetLeadForUpdate :one
SELECT * FROM leads WHERE id = $1 FOR UPDATE;
