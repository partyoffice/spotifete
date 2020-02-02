BEGIN;
ALTER TABLE listening_sessions
    ADD COLUMN fallback_playlist VARCHAR REFERENCES playlist_metadata(spotify_playlist_id);
COMMIT;
