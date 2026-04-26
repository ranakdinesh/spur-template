-- 1. Enhance API Keys for Public Usage
ALTER TABLE api_keys 
    ADD COLUMN type TEXT NOT NULL DEFAULT 'secret', -- 'secret' or 'publishable'
    ADD COLUMN allowed_origins TEXT[] DEFAULT '{}'; -- e.g. ['https://example.com']

-- 2. Index for faster lookups during ingestion
CREATE INDEX idx_api_keys_type ON api_keys(type);
