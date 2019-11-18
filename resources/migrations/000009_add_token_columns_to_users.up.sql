BEGIN;
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS spotify_access_token VARCHAR,
    ADD COLUMN IF NOT EXISTS spotify_refresh_token VARCHAR,
    ADD COLUMN IF NOT EXISTS spotify_token_type VARCHAR,
    ADD COLUMN IF NOT EXISTS spotify_token_expiry TIMESTAMP;
COMMIT;