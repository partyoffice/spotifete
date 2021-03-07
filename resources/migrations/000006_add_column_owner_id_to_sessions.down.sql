BEGIN;

ALTER TABLE sessions
    DROP CONSTRAINT sessions_owner_id_fkey;

ALTER TABLE sessions
    DROP COLUMN owner_id;

COMMIT;
