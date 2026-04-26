-- name: CreateLead :one
INSERT INTO leads (tenant_id, source, status, created_at, updated_at) 
VALUES ($1, $2, $3, NOW(), NOW()) 
RETURNING id, tenant_id, source, status, created_at, updated_at;

-- name: GetLead :one
SELECT * FROM leads WHERE id = $1 AND tenant_id = $2;
