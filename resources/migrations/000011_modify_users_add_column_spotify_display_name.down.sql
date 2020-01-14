BEGIN;
ALTER TABLE users
    DROP COLUMN spotify_display_name;
COMMIT;
