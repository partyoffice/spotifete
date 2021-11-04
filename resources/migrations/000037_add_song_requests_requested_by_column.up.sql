BEGIN;

ALTER TABLE song_requests
    ADD COLUMN requested_by VARCHAR(63);

COMMIT;
