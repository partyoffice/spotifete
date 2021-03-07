BEGIN;

DELETE
FROM sessions;

ALTER TABLE sessions
    ADD COLUMN owner_id INTEGER NOT NULL DEFAULT -1;;

ALTER TABLE sessions
    ADD CONSTRAINT sessions_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES sessions (id);

COMMIT;