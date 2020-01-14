BEGIN;
ALTER TABLE listening_sessions
    DROP COLUMN title;
COMMIT;
