BEGIN;

ALTER TABLE song_requests
    DROP COLUMN user_id;

COMMIT;
