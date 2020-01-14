BEGIN;
ALTER TABLE song_requests
    RENAME COLUMN requested_by TO user_id;
COMMIT;
