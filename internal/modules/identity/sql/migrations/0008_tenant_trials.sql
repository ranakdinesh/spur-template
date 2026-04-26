-- 8. Tenant Trials & Subscriptions
ALTER TABLE tenants
    ADD COLUMN trial_ends_at TIMESTAMPTZ,
    ADD COLUMN subscription_plan TEXT;
