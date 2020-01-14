BEGIN;
ALTER TABLE listening_sessions
    ADD COLUMN spotify_playlist VARCHAR;
COMMIT;
