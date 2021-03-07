BEGIN;

-- This index also acts as a unique constraint to ensure that only one
-- CURRENTLY_PLAYING and one UP_NEXT request are present per session
CREATE UNIQUE INDEX song_requests_status_index
    ON song_requests (session_id, status)
    WHERE status in ('CURRENTLY_PLAYING', 'UP_NEXT');

CREATE INDEX song_requests_status_in_queue_index
    ON song_requests (session_id, status)
    WHERE status = 'IN_QUEUE';

COMMIT;
