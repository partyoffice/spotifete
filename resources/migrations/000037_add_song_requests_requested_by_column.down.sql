BEGIN;

ALTER TABLE song_requests
    DROP COLUMN requested_by;

COMMIT;
