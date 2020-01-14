BEGIN;
ALTER TABLE authentication_states
    ADD CONSTRAINT state_unique UNIQUE (state);
COMMIT;
