BEGIN;

ALTER TABLE song_requests
    DROP COLUMN weight;

COMMIT;
