-- name: CreateLeadIdentity :exec
INSERT INTO lead_identities (id, tenant_id, lead_id, type, value)
VALUES ($1, $2, $3, $4, $5);

-- name: FindLeadByIdentity :one
SELECT l.* 
FROM leads l
JOIN lead_identities li ON l.id = li.lead_id
WHERE li.tenant_id = $1 AND li.type = $2 AND li.value = $3
LIMIT 1;
