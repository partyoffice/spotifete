BEGIN;
ALTER TABLE song_requests
    RENAME COLUMN song_id TO track_id;
COMMIT;
