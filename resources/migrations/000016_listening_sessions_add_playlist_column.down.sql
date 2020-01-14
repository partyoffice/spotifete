BEGIN;
ALTER TABLE listening_sessions
    DROP COLUMN spotify_playlist;
COMMIT;
