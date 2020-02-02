BEGIN;
ALTER TABLE listening_sessions
    RENAME COLUMN queue_playlist TO spotify_playlist;
COMMIT;
