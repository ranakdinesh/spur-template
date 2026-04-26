CREATE UNIQUE INDEX users_mobile_idx ON users (tenant_id, mobile) WHERE mobile IS NOT NULL AND mobile != '';

-- Add a new column to verification_challenges if needed, or just use existing generic structure.
-- The existing structure:
-- kind TEXT NOT NULL
-- token_hash TEXT NOT NULL
-- But we might need 'channel' (email/mobile) if not implied by 'kind'.
-- Let's stick to 'kind' = 'email_otp' or 'mobile_otp'.
