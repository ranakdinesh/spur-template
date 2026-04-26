CREATE TABLE verification_challenges (
                                         id UUID PRIMARY KEY,
                                         tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
                                         user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                         kind TEXT NOT NULL, -- 'email_verify', 'password_reset'
                                         token_hash TEXT NOT NULL,
                                         expires_at TIMESTAMPTZ NOT NULL,
                                         consumed_at TIMESTAMPTZ,
                                         created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_active_challenge ON verification_challenges (user_id, kind)
    WHERE consumed_at IS NULL;