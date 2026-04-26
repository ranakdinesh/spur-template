-- name: GetWebFormBySlug :one
SELECT id, tenant_id, name, slug, schema, settings, is_active, created_at, updated_at
FROM lead_forms
WHERE slug = $1 LIMIT 1;

-- name: CreateLeadSubmission :one
INSERT INTO lead_submissions (id, tenant_id, form_id, lead_id, submitted_at, payload, ip, user_agent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;
