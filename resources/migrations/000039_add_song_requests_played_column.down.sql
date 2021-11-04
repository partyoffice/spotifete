BEGIN;

ALTER TABLE song_requests
    DROP COLUMN played;

COMMIT;
