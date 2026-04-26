-- name: GetFunnelStats :many
SELECT 
    stage,
    COUNT(*) as count
FROM leads
WHERE created_at BETWEEN $1 AND $2
GROUP BY stage
ORDER BY count DESC;

-- name: GetSourceStats :many
SELECT 
    source,
    COUNT(*) as total_leads,
    AVG(score)::float as avg_score
FROM leads
GROUP BY source
ORDER BY total_leads DESC;
