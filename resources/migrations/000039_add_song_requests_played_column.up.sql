BEGIN;

ALTER TABLE song_requests
    ADD COLUMN played BOOLEAN NOT NULL DEFAULT FALSE;

COMMIT;