BEGIN;
ALTER TABLE sessions
    DROP CONSTRAINT join_id_unique;
COMMIT;
