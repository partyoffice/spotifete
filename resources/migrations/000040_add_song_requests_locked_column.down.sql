BEGIN;

ALTER TABLE song_requests
    DROP COLUMN locked;

COMMIT;
