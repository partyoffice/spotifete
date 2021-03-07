BEGIN;

ALTER TABLE listening_sessions
    ADD COLUMN fallback_playlist_shuffle BOOLEAN DEFAULT FALSE;

COMMIT;
