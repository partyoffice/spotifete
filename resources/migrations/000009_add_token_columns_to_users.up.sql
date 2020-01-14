BEGIN;
ALTER TABLE users
    ADD COLUMN spotify_access_token VARCHAR,
    ADD COLUMN spotify_refresh_token VARCHAR,
    ADD COLUMN spotify_token_type VARCHAR,
    ADD COLUMN spotify_token_expiry TIMESTAMP;
COMMIT;
