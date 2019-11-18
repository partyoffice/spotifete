BEGIN;
ALTER TABLE users
    DROP COLUMN IF EXISTS spotify_access_token,
    DROP COLUMN IF EXISTS spotify_refresh_token,
    DROP COLUMN IF EXISTS spotify_token_type,
    DROP COLUMN IF EXISTS spotify_token_expiry;
COMMIT;