BEGIN;
ALTER TABLE sessions
    DROP COLUMN owner_id;
COMMIT;
