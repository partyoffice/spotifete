BEGIN;
ALTER TABLE listening_sessions
    DROP CONSTRAINT sessions_owner_id_fkey;

ALTER TABLE listening_sessions
    ADD CONSTRAINT sessions_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES users (id);
COMMIT;
