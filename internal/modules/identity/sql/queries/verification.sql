-- name: CreateVerificationChallenge :one
INSERT INTO verification_challenges (
    id,
    tenant_id,
    user_id,
    kind,
    token_hash,
    expires_at
) VALUES (
             $1, $2, $3, $4, $5, $6
         )
RETURNING *;

-- name: GetVerificationChallenge :one
SELECT * FROM verification_challenges
WHERE user_id = $1 AND kind = $2 AND consumed_at IS NULL AND expires_at > NOW();

-- name: GetVerificationChallengeByToken :one
SELECT * FROM verification_challenges
WHERE token_hash = $1 AND kind = $2 AND consumed_at IS NULL AND expires_at > NOW();

-- name: MarkChallengeConsumed :exec
UPDATE verification_challenges
SET consumed_at = NOW()
WHERE id = $1;

-- name: DeleteExpiredChallenges :exec
DELETE FROM verification_challenges
WHERE expires_at < NOW();
