BEGIN;
ALTER TABLE users
    DROP COLUMN spotify_access_token,
    DROP COLUMN spotify_refresh_token,
    DROP COLUMN spotify_token_type,
    DROP COLUMN spotify_token_expiry;
COMMIT;
