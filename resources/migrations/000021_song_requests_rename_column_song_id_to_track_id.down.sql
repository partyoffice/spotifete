BEGIN;
ALTER TABLE song_requests
    RENAME COLUMN track_id TO song_id;
COMMIT;
