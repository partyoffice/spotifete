BEGIN;
ALTER TABLE listening_sessions
    RENAME COLUMN spotify_playlist TO queue_playlist;
COMMIT;
