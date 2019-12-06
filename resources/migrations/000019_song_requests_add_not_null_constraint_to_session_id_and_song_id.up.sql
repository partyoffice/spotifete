BEGIN;
DELETE FROM song_requests;
ALTER TABLE song_requests
    ALTER COLUMN session_id SET NOT NULL,
    ALTER COLUMN song_id SET NOT NULL;
COMMIT;
